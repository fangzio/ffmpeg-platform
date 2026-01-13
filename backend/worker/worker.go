package worker

import (
	"context"
	"encoding/json"
	"ffmpeg-platform/config"
	"ffmpeg-platform/model"
	"ffmpeg-platform/pkg/ffmpeg"
	"ffmpeg-platform/pkg/storage"
	"ffmpeg-platform/service"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

// Worker 异步任务处理器
type Worker struct {
	db            *gorm.DB
	taskService   *service.TaskService
	ffmpegService *service.FFmpegService
	config        *config.Config
	storage       storage.Storage
	progressHubs  map[string]*ProgressHub
	mu            sync.RWMutex
}

func NewWorker(db *gorm.DB, taskService *service.TaskService, cfg *config.Config, storageImpl storage.Storage) *Worker {
	return &Worker{
		db:            db,
		taskService:   taskService,
		ffmpegService: service.NewFFmpegService(cfg),
		config:        cfg,
		storage:       storageImpl,
		progressHubs:  make(map[string]*ProgressHub),
	}
}

// ProcessTask 处理任务
func (w *Worker) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload map[string]string
	fmt.Println("in ProcessTask")
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		fmt.Println("unmarshal payload error", err)
		return fmt.Errorf("unmarshal payload failed: %w", err)
	}

	taskID := payload["task_id"]
	log.Printf("Processing task: %s", taskID)

	// 获取任务
	task, err := w.taskService.GetTask(taskID)
	if err != nil {
		return fmt.Errorf("get task failed: %w", err)
	}

	// 提前创建 ProgressHub，确保进度更新能被广播
	w.GetProgressHub(taskID)

	// 更新状态为处理中
	w.taskService.UpdateTaskProgress(taskID, model.TaskProgress{
		TaskID: taskID,
		Status: model.TaskStatusProcessing,
	})

	// 广播初始状态
	w.broadcastProgress(taskID, model.TaskProgress{
		TaskID:   taskID,
		Status:   model.TaskStatusProcessing,
		Progress: 0,
		Message:  "Task started processing",
	})

	// 添加超时上下文 (30分钟超时)
	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	// 监控超时
	done := make(chan error, 1)
	go func() {
		// 根据任务类型执行
		switch task.Type {
		case "image_audio_to_video":
			done <- w.processImageAudioToVideo(ctx, task)
		case "image_slideshow":
			done <- w.processImageSlideshow(ctx, task)
		default:
			done <- fmt.Errorf("unknown task type: %s", task.Type)
		}
	}()

	// 等待任务完成或超时
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		errMsg := "Task timeout: exceeded 30 minutes"
		log.Printf("Task %s timeout", taskID)

		// 更新任务为失败状态
		w.taskService.FailTask(taskID, "", "", "", errMsg)
		w.broadcastProgress(taskID, model.TaskProgress{
			TaskID:  taskID,
			Status:  model.TaskStatusFailed,
			Message: errMsg,
		})

		return fmt.Errorf("%s", errMsg)
	}
}

// processImageAudioToVideo 处理图片+音频生成视频任务
func (w *Worker) processImageAudioToVideo(ctx context.Context, task *model.Task) (err error) {
	// 添加 panic 恢复机制，确保任务状态能正确更新
	defer func() {
		if r := recover(); r != nil {
			errMsg := fmt.Sprintf("Task panic: %v", r)
			log.Printf("Task %s panic: %v", task.ID, r)

			// 更新任务为失败状态
			w.taskService.FailTask(task.ID, "", "", "", errMsg)
			w.broadcastProgress(task.ID, model.TaskProgress{
				TaskID:  task.ID,
				Status:  model.TaskStatusFailed,
				Message: errMsg,
			})

			err = fmt.Errorf("%s", errMsg)
		}
	}()

	log.Printf("Task %s: Starting image+audio to video processing", task.ID)

	// 生成输出路径
	outputPath := w.ffmpegService.GenerateOutputPath(task.ID, task.InputParams.OutputFormat)
	log.Printf("Task %s: Output path: %s", task.ID, outputPath)

	// 构建ffmpeg命令
	args, totalFrames, tempFiles, err := w.ffmpegService.BuildImageAudioToVideoCommand(task.InputParams, outputPath)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to build ffmpeg command: %v", err)
		log.Printf("Task %s: %s", task.ID, errMsg)
		w.taskService.FailTask(task.ID, "", "", "", errMsg)
		w.broadcastProgress(task.ID, model.TaskProgress{
			TaskID:  task.ID,
			Status:  model.TaskStatusFailed,
			Message: errMsg,
		})
		return fmt.Errorf("%s", errMsg)
	}

	// 确保临时文件在函数结束时被清理
	defer func() {
		if len(tempFiles) > 0 {
			log.Printf("Task %s: Cleaning up %d temporary files", task.ID, len(tempFiles))
			w.ffmpegService.CleanupTempFiles(tempFiles)
		}
	}()

	log.Printf("Task %s: Total frames: %d, Command: ffmpeg %v", task.ID, totalFrames, args)

	// 创建进度回调
	progressCallback := func(progress ffmpeg.Progress) {
		// 广播进度到WebSocket
		w.broadcastProgress(task.ID, model.TaskProgress{
			TaskID:       task.ID,
			Status:       model.TaskStatusProcessing,
			Progress:     progress.Progress,
			CurrentFrame: progress.Frame,
			TotalFrames:  totalFrames,
			ETA:          progress.ETA,
			Message:      fmt.Sprintf("Processing: %.1f%% (Frame %d/%d, Speed: %.2fx)", progress.Progress, progress.Frame, totalFrames, progress.Speed),
		})

		// 更新数据库
		w.taskService.UpdateTaskProgress(task.ID, model.TaskProgress{
			TaskID:       task.ID,
			Status:       model.TaskStatusProcessing,
			Progress:     progress.Progress,
			CurrentFrame: progress.Frame,
			TotalFrames:  totalFrames,
			ETA:          progress.ETA,
		})
	}

	// 执行ffmpeg命令
	log.Printf("Task %s: Starting ffmpeg execution", task.ID)
	result := w.ffmpegService.ExecuteWithProgress(ctx, args, totalFrames, progressCallback)
	log.Printf("Task %s: FFmpeg execution finished, success: %v", task.ID, result.Success)

	if result.Success {
		// 任务成功 - 在标记完成前，强制广播100%进度
		log.Printf("Task %s: FFmpeg succeeded, broadcasting final progress", task.ID)
		w.broadcastProgress(task.ID, model.TaskProgress{
			TaskID:       task.ID,
			Status:       model.TaskStatusProcessing,
			Progress:     100,
			CurrentFrame: totalFrames,
			TotalFrames:  totalFrames,
			ETA:          0,
			Message:      "Processing completed, finalizing...",
		})

		// 更新数据库进度为100%
		w.taskService.UpdateTaskProgress(task.ID, model.TaskProgress{
			TaskID:       task.ID,
			Status:       model.TaskStatusProcessing,
			Progress:     100,
			CurrentFrame: totalFrames,
			TotalFrames:  totalFrames,
			ETA:          0,
		})

		// 生成输出文件URL
		outputURL := ""
		outputFilename := w.ffmpegService.GenerateOutputName(task.ID, task.InputParams.OutputFormat)

		// 如果是七牛云存储，上传到云端
		if w.config.Storage.Type == "qiniu" && w.config.Qiniu.Enabled {
			log.Printf("Task %s: Uploading to Qiniu cloud storage", task.ID)
			// 生成七牛云存储key（使用outputs目录前缀）
			key := fmt.Sprintf("outputs/%s", outputFilename)
			cloudURL, err := w.storage.UploadFile(outputPath, key)
			if err != nil {
				log.Printf("Task %s: Warning - failed to upload output to cloud: %v", task.ID, err)
				// 上传失败不影响任务完成，使用本地路径
				outputURL = fmt.Sprintf("/api/outputs/%s", outputFilename)
			} else {
				log.Printf("Task %s: Upload success, URL: %s", task.ID, cloudURL)
				outputURL = cloudURL
				// 删除本地临时文件
				if err := w.storage.DeleteLocalFile(outputPath); err != nil {
					log.Printf("Task %s: Warning - failed to delete local output file %s: %v", task.ID, outputPath, err)
				}
			}
		} else {
			outputURL = fmt.Sprintf("/api/outputs/%s", outputFilename)
		}

		// 任务成功
		log.Printf("Task %s: Marking task as completed", task.ID)
		w.taskService.CompleteTask(task.ID, service.TaskResult{
			FFmpegCommand: result.Command,
			FilterGraph:   result.FilterGraph,
			StderrLog:     result.StderrLog,
			OutputFile:    outputPath,
			OutputURL:     outputURL,
			TotalFrames:   totalFrames, // 传入总帧数
		})

		// 广播完成消息（最终状态）
		w.broadcastProgress(task.ID, model.TaskProgress{
			TaskID:       task.ID,
			Status:       model.TaskStatusCompleted,
			Progress:     100,
			CurrentFrame: totalFrames,
			TotalFrames:  totalFrames,
			Message:      "Task completed successfully",
		})

		// 短暂延迟确保WebSocket消息发送完成
		time.Sleep(100 * time.Millisecond)

		log.Printf("Task %s completed successfully, output: %s", task.ID, outputURL)
		return nil
	} else {
		// 任务失败 - 记录详细的错误信息
		log.Printf("Task %s failed with error: %s", task.ID, result.ErrorMessage)
		log.Printf("Task %s - FFmpeg command: %s", task.ID, result.Command)
		log.Printf("Task %s - Stderr log (last 500 chars): %s", task.ID, truncateString(result.StderrLog, 500))

		// 任务失败
		w.taskService.FailTask(task.ID, result.Command, result.FilterGraph, result.StderrLog, result.ErrorMessage)

		// 广播失败消息
		w.broadcastProgress(task.ID, model.TaskProgress{
			TaskID:  task.ID,
			Status:  model.TaskStatusFailed,
			Message: result.ErrorMessage,
		})

		return fmt.Errorf("ffmpeg execution failed: %s", result.ErrorMessage)
	}
}

// processImageSlideshow 处理多图片轮播视频任务
func (w *Worker) processImageSlideshow(ctx context.Context, task *model.Task) (err error) {
	// 添加 panic 恢复机制
	defer func() {
		if r := recover(); r != nil {
			errMsg := fmt.Sprintf("Task panic: %v", r)
			log.Printf("Task %s panic: %v", task.ID, r)
			w.taskService.FailTask(task.ID, "", "", "", errMsg)
			w.broadcastProgress(task.ID, model.TaskProgress{
				TaskID:  task.ID,
				Status:  model.TaskStatusFailed,
				Message: errMsg,
			})
			err = fmt.Errorf("%s", errMsg)
		}
	}()

	log.Printf("Task %s: Starting image slideshow processing", task.ID)

	// 生成输出路径
	outputPath := w.ffmpegService.GenerateOutputPath(task.ID, task.InputParams.OutputFormat)
	log.Printf("Task %s: Output path: %s", task.ID, outputPath)

	// 构建ffmpeg命令
	args, totalFrames, tempFiles, err := w.ffmpegService.BuildImageSlideshowCommand(task.InputParams, outputPath)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to build ffmpeg command: %v", err)
		log.Printf("Task %s: %s", task.ID, errMsg)
		w.taskService.FailTask(task.ID, "", "", "", errMsg)
		w.broadcastProgress(task.ID, model.TaskProgress{
			TaskID:  task.ID,
			Status:  model.TaskStatusFailed,
			Message: errMsg,
		})
		return fmt.Errorf("%s", errMsg)
	}

	// 确保临时文件在函数结束时被清理
	defer func() {
		if len(tempFiles) > 0 {
			log.Printf("Task %s: Cleaning up %d temporary files", task.ID, len(tempFiles))
			w.ffmpegService.CleanupTempFiles(tempFiles)
		}
	}()

	log.Printf("Task %s: Total frames: %d, Images: %d, Command: ffmpeg %v", task.ID, totalFrames, len(task.InputParams.ImagePaths), args)

	// 创建进度回调
	progressCallback := func(progress ffmpeg.Progress) {
		w.broadcastProgress(task.ID, model.TaskProgress{
			TaskID:       task.ID,
			Status:       model.TaskStatusProcessing,
			Progress:     progress.Progress,
			CurrentFrame: progress.Frame,
			TotalFrames:  totalFrames,
			ETA:          progress.ETA,
			Message:      fmt.Sprintf("Processing slideshow: %.1f%% (Frame %d/%d, Speed: %.2fx)", progress.Progress, progress.Frame, totalFrames, progress.Speed),
		})

		w.taskService.UpdateTaskProgress(task.ID, model.TaskProgress{
			TaskID:       task.ID,
			Status:       model.TaskStatusProcessing,
			Progress:     progress.Progress,
			CurrentFrame: progress.Frame,
			TotalFrames:  totalFrames,
			ETA:          progress.ETA,
		})
	}

	// 执行ffmpeg命令
	log.Printf("Task %s: Starting ffmpeg execution", task.ID)
	result := w.ffmpegService.ExecuteWithProgress(ctx, args, totalFrames, progressCallback)
	log.Printf("Task %s: FFmpeg execution finished, success: %v", task.ID, result.Success)

	if result.Success {
		// 任务成功 - 强制广播100%进度
		log.Printf("Task %s: FFmpeg succeeded, broadcasting final progress", task.ID)
		w.broadcastProgress(task.ID, model.TaskProgress{
			TaskID:       task.ID,
			Status:       model.TaskStatusProcessing,
			Progress:     100,
			CurrentFrame: totalFrames,
			TotalFrames:  totalFrames,
			ETA:          0,
			Message:      "Processing completed, finalizing...",
		})

		w.taskService.UpdateTaskProgress(task.ID, model.TaskProgress{
			TaskID:       task.ID,
			Status:       model.TaskStatusProcessing,
			Progress:     100,
			CurrentFrame: totalFrames,
			TotalFrames:  totalFrames,
			ETA:          0,
		})

		// 生成输出文件URL
		outputURL := ""
		outputFilename := w.ffmpegService.GenerateOutputName(task.ID, task.InputParams.OutputFormat)

		if w.config.Storage.Type == "qiniu" && w.config.Qiniu.Enabled {
			log.Printf("Task %s: Uploading to Qiniu cloud storage", task.ID)
			key := fmt.Sprintf("outputs/%s", outputFilename)
			cloudURL, err := w.storage.UploadFile(outputPath, key)
			if err != nil {
				log.Printf("Task %s: Warning - failed to upload output to cloud: %v", task.ID, err)
				outputURL = fmt.Sprintf("/api/outputs/%s", outputFilename)
			} else {
				log.Printf("Task %s: Upload success, URL: %s", task.ID, cloudURL)
				outputURL = cloudURL
				if err := w.storage.DeleteLocalFile(outputPath); err != nil {
					log.Printf("Task %s: Warning - failed to delete local output file %s: %v", task.ID, outputPath, err)
				}
			}
		} else {
			outputURL = fmt.Sprintf("/api/outputs/%s", outputFilename)
		}

		// 任务成功
		log.Printf("Task %s: Marking task as completed", task.ID)
		w.taskService.CompleteTask(task.ID, service.TaskResult{
			FFmpegCommand: result.Command,
			FilterGraph:   result.FilterGraph,
			StderrLog:     result.StderrLog,
			OutputFile:    outputPath,
			OutputURL:     outputURL,
			TotalFrames:   totalFrames,
		})

		// 广播完成消息
		w.broadcastProgress(task.ID, model.TaskProgress{
			TaskID:       task.ID,
			Status:       model.TaskStatusCompleted,
			Progress:     100,
			CurrentFrame: totalFrames,
			TotalFrames:  totalFrames,
			Message:      "Slideshow video completed successfully",
		})

		time.Sleep(100 * time.Millisecond)
		log.Printf("Task %s completed successfully, output: %s", task.ID, outputURL)
		return nil
	} else {
		// 任务失败
		log.Printf("Task %s failed with error: %s", task.ID, result.ErrorMessage)
		w.taskService.FailTask(task.ID, result.Command, result.FilterGraph, result.StderrLog, result.ErrorMessage)
		w.broadcastProgress(task.ID, model.TaskProgress{
			TaskID:  task.ID,
			Status:  model.TaskStatusFailed,
			Message: result.ErrorMessage,
		})
		return fmt.Errorf("ffmpeg execution failed: %s", result.ErrorMessage)
	}
}

// truncateString 截断字符串到指定长度
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return "..." + s[len(s)-maxLen:]
}

// ProgressHub WebSocket连接管理
type ProgressHub struct {
	clients    map[*ProgressClient]bool
	broadcast  chan model.TaskProgress
	Register   chan *ProgressClient
	UnRegister chan *ProgressClient
	mu         sync.RWMutex
}

type ProgressClient struct {
	Hub  *ProgressHub
	Send chan model.TaskProgress
}

func NewProgressHub() *ProgressHub {
	return &ProgressHub{
		clients:    make(map[*ProgressClient]bool),
		broadcast:  make(chan model.TaskProgress, 256),
		Register:   make(chan *ProgressClient),
		UnRegister: make(chan *ProgressClient),
	}
}

func (h *ProgressHub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
		case client := <-h.UnRegister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
			h.mu.Unlock()
		case progress := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.Send <- progress:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// broadcastProgress 广播进度到WebSocket
func (w *Worker) broadcastProgress(taskID string, progress model.TaskProgress) {
	w.mu.RLock()
	hub, exists := w.progressHubs[taskID]
	w.mu.RUnlock()

	if exists {
		hub.broadcast <- progress
	}
}

// GetProgressHub 获取或创建进度Hub
func (w *Worker) GetProgressHub(taskID string) *ProgressHub {
	w.mu.Lock()
	defer w.mu.Unlock()

	if hub, exists := w.progressHubs[taskID]; exists {
		return hub
	}

	hub := NewProgressHub()
	go hub.Run()
	w.progressHubs[taskID] = hub
	return hub
}
