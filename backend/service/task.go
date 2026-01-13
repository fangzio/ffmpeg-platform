package service

import (
	"encoding/json"
	"ffmpeg-platform/config"
	"ffmpeg-platform/model"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

type TaskService struct {
	db            *gorm.DB
	asynqClient   *asynq.Client
	ffmpegService *FFmpegService
	config        *config.Config
}

func NewTaskService(db *gorm.DB, asynqClient *asynq.Client, cfg *config.Config) *TaskService {
	return &TaskService{
		db:            db,
		asynqClient:   asynqClient,
		ffmpegService: NewFFmpegService(cfg),
		config:        cfg,
	}
}

// CreateTask 创建任务
func (s *TaskService) CreateTask(taskType string, params model.TaskInputParams) (*model.Task, error) {
	// 验证输入
	if err := s.ffmpegService.ValidateInputs(params); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	task := &model.Task{
		ID:          uuid.New().String(),
		Type:        taskType,
		Status:      model.TaskStatusPending,
		InputParams: params,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.db.Create(task).Error; err != nil {
		return nil, fmt.Errorf("create task failed: %w", err)
	}

	// 提交到异步队列
	payload, _ := json.Marshal(map[string]string{"task_id": task.ID})
	taskInfo := asynq.NewTask("task:process", payload)
	fmt.Println("after ad taskInfo")
	if _, err := s.asynqClient.Enqueue(taskInfo); err != nil {
		return nil, fmt.Errorf("enqueue task failed: %w", err)
	}

	return task, nil
}

// GetTask 获取任务详情
func (s *TaskService) GetTask(taskID string) (*model.Task, error) {
	var task model.Task
	if err := s.db.First(&task, "id = ?", taskID).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// ListTasks 获取任务列表
func (s *TaskService) ListTasks(page, pageSize int) ([]model.Task, int64, error) {
	var tasks []model.Task
	var total int64

	query := s.db.Model(&model.Task{})
	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// UpdateTaskProgress 更新任务进度
func (s *TaskService) UpdateTaskProgress(taskID string, progress model.TaskProgress) error {
	updates := map[string]interface{}{
		"status":        progress.Status,
		"progress":      progress.Progress,
		"current_frame": progress.CurrentFrame,
		"total_frames":  progress.TotalFrames,
		"eta":           progress.ETA,
		"updated_at":    time.Now(),
	}

	return s.db.Model(&model.Task{}).Where("id = ?", taskID).Updates(updates).Error
}

// CompleteTask 完成任务（保存完整执行信息）
func (s *TaskService) CompleteTask(taskID string, result TaskResult) error {
	updates := map[string]interface{}{
		"status":         model.TaskStatusCompleted,
		"progress":       100,
		"current_frame":  result.TotalFrames, // 确保完成时帧数正确
		"total_frames":   result.TotalFrames,
		"ffmpeg_command": result.FFmpegCommand,
		"filter_graph":   result.FilterGraph,
		"stderr_log":     result.StderrLog,
		"output_file":    result.OutputFile,
		"output_url":     result.OutputURL,
		"updated_at":     time.Now(),
	}

	return s.db.Model(&model.Task{}).Where("id = ?", taskID).Updates(updates).Error
}

// FailTask 任务失败（保存失败原因）
func (s *TaskService) FailTask(taskID string, ffmpegCommand, filterGraph, stderrLog, errorMessage string) error {
	updates := map[string]interface{}{
		"status":         model.TaskStatusFailed,
		"ffmpeg_command": ffmpegCommand,
		"filter_graph":   filterGraph,
		"stderr_log":     stderrLog,
		"error_message":  errorMessage,
		"updated_at":     time.Now(),
	}

	return s.db.Model(&model.Task{}).Where("id = ?", taskID).Updates(updates).Error
}

type TaskResult struct {
	FFmpegCommand string
	FilterGraph   string
	StderrLog     string
	OutputFile    string
	OutputURL     string
	TotalFrames   int // 总帧数
}
