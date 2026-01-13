package ffmpeg

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ProgressCallback 进度回调函数
type ProgressCallback func(progress Progress)

// Progress 进度信息
type Progress struct {
	Frame     int       // 当前帧
	FPS       float64   // 当前帧率
	Bitrate   string    // 当前码率
	TotalSize int64     // 已生成大小
	Time      string    // 已处理时长
	Speed     float64   // 处理速度倍率
	Progress  float64   // 进度百分比
	ETA       int       // 预计剩余秒数
	Timestamp time.Time // 时间戳
}

// ExecuteResult 执行结果
type ExecuteResult struct {
	Command      string  // 完整命令
	FilterGraph  string  // filter graph
	StderrLog    string  // 完整stderr日志
	Success      bool    // 是否成功
	ErrorMessage string  // 错误信息
	Duration     float64 // 执行耗时（秒）
}

// Executor FFmpeg执行器 - 核心差异点实现
type Executor struct {
	binaryPath string
	logLevel   string
}

func NewExecutor(binaryPath, logLevel string) *Executor {
	return &Executor{
		binaryPath: binaryPath,
		logLevel:   logLevel,
	}
}

// Execute 执行ffmpeg命令并实时解析进度
func (e *Executor) Execute(ctx context.Context, args []string, totalFrames int, callback ProgressCallback) *ExecuteResult {
	startTime := time.Now()
	result := &ExecuteResult{
		Command: fmt.Sprintf("%s %s", e.binaryPath, strings.Join(args, " ")),
	}

	log.Printf("[FFmpeg] Executing command: %s", result.Command)

	// 提取filter graph（如果有）
	result.FilterGraph = e.extractFilterGraph(args)

	// 构建命令
	cmd := exec.CommandContext(ctx, e.binaryPath, args...)

	// 获取stderr pipe用于解析进度
	stderr, err := cmd.StderrPipe()
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("failed to get stderr pipe: %v", err)
		log.Printf("[FFmpeg] Failed to get stderr pipe: %v", err)
		return result
	}

	// 同时获取 stdout，防止 stdout 被阻塞
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("failed to get stdout pipe: %v", err)
		log.Printf("[FFmpeg] Failed to get stdout pipe: %v", err)
		return result
	}

	// 启动命令
	if err := cmd.Start(); err != nil {
		result.ErrorMessage = fmt.Sprintf("failed to start ffmpeg: %v", err)
		log.Printf("[FFmpeg] Failed to start: %v", err)
		return result
	}

	log.Printf("[FFmpeg] Process started, PID: %d", cmd.Process.Pid)

	// 实时解析stderr
	var stderrLog strings.Builder
	var stdoutLog strings.Builder
	var wg sync.WaitGroup
	var progressCount int
	var mu sync.Mutex
	lastOutputTime := time.Now()
	lastProgressFrame := 0      // 上次进度更新的帧数
	lastProgressTime := time.Now() // 上次进度更新的时间

	// 读取 stdout（防止阻塞）
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		lineCount := 0
		for scanner.Scan() {
			line := scanner.Text()
			lineCount++
			stdoutLog.WriteString(line + "\n")
			if lineCount == 1 {
				log.Printf("[FFmpeg] First stdout line: %s", line)
			}
		}
		if lineCount > 0 {
			log.Printf("[FFmpeg] Stdout reading completed, total lines: %d", lineCount)
		}
	}()

	// 监控超时 - 检测进度是否真正停滞
	timeoutCtx, timeoutCancel := context.WithCancel(ctx)
	defer timeoutCancel()

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		checkCount := 0
		for {
			select {
			case <-timeoutCtx.Done():
				return
			case <-ticker.C:
				checkCount++
				mu.Lock()
				elapsedSinceOutput := time.Since(lastOutputTime)
				elapsedSinceProgress := time.Since(lastProgressTime)
				currentProgressCount := progressCount
				currentFrame := lastProgressFrame
				mu.Unlock()

				// 每20秒输出一次状态检查
				if checkCount%2 == 0 {
					log.Printf("[FFmpeg] Status check #%d: %.1fs since output, %.1fs since progress, frame: %d, progress count: %d",
						checkCount, elapsedSinceOutput.Seconds(), elapsedSinceProgress.Seconds(), currentFrame, currentProgressCount)
				}

				// 超时策略：
				// 1. 如果从未有进度更新，且60秒没有输出 -> 认为启动失败
				// 2. 如果有进度更新，但5分钟内进度没有增加 -> 认为卡住了
				if currentProgressCount == 0 {
					// 从未有进度更新
					if elapsedSinceOutput > 60*time.Second {
						log.Printf("[FFmpeg] ERROR: No progress after 60 seconds, killing process")
						if cmd.Process != nil {
							cmd.Process.Kill()
						}
						return
					}
				} else {
					// 已有进度更新，检查进度是否停滞
					if elapsedSinceProgress > 5*time.Minute {
						log.Printf("[FFmpeg] ERROR: Progress stalled for 5 minutes (frame stuck at %d), killing process", currentFrame)
						if cmd.Process != nil {
							cmd.Process.Kill()
						}
						return
					}
				}
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		scanner.Buffer(make([]byte, 64*1024), 1024*1024) // 增大缓冲区

		lineCount := 0
		for scanner.Scan() {
			line := scanner.Text()
			lineCount++
			stderrLog.WriteString(line + "\n")

			// 记录第一行输出
			if lineCount == 1 {
				log.Printf("[FFmpeg] First stderr line: %s", line)
			}

			// 更新最后输出时间
			mu.Lock()
			lastOutputTime = time.Now()
			mu.Unlock()

			// 每 100 行输出一次调试信息
			if lineCount%100 == 0 {
				log.Printf("[FFmpeg] Processed %d lines from stderr", lineCount)
			}

			// 解析进度信息
			if progress := e.parseProgress(line, totalFrames); progress != nil {
				progressCount++

				// 更新进度跟踪
				mu.Lock()
				if progress.Frame > lastProgressFrame {
					lastProgressFrame = progress.Frame
					lastProgressTime = time.Now()
				}
				mu.Unlock()

				// 每 10 次进度更新输出一次日志
				if progressCount%10 == 0 || progressCount == 1 {
					log.Printf("[FFmpeg] Progress update #%d: %.1f%% (Frame %d/%d, Speed: %.2fx, FPS: %.1f)",
						progressCount, progress.Progress, progress.Frame, totalFrames, progress.Speed, progress.FPS)
				}

				if callback != nil {
					callback(*progress)
				}
			}
		}

		if err := scanner.Err(); err != nil {
			log.Printf("[FFmpeg] Scanner error: %v", err)
		}

		log.Printf("[FFmpeg] Stderr reading completed, total lines: %d, progress updates: %d",
			lineCount, progressCount)

		// 停止超时监控
		timeoutCancel()
	}()

	// 等待命令完成
	log.Printf("[FFmpeg] Waiting for process to complete...")
	err = cmd.Wait()

	// 等待 stderr 读取完成
	log.Printf("[FFmpeg] Waiting for stderr/stdout goroutines to finish...")
	wg.Wait()

	result.StderrLog = stderrLog.String()

	// 如果有 stdout 输出，也记录下来
	if stdoutLog.Len() > 0 {
		result.StderrLog += "\n=== STDOUT ===\n" + stdoutLog.String()
	}

	result.Duration = time.Since(startTime).Seconds()

	log.Printf("[FFmpeg] Process completed in %.2f seconds, stderr length: %d bytes, stdout length: %d bytes",
		result.Duration, stderrLog.Len(), stdoutLog.Len())

	if err != nil {
		result.Success = false
		result.ErrorMessage = e.extractError(result.StderrLog)
		log.Printf("[FFmpeg] Execution failed: %v", result.ErrorMessage)
		log.Printf("[FFmpeg] Last 500 chars of stderr: %s", truncate(result.StderrLog, 500))
	} else {
		result.Success = true
		log.Printf("[FFmpeg] Execution succeeded")
	}

	if progressCount == 0 {
		log.Printf("[FFmpeg] WARNING: No progress updates received!")
		log.Printf("[FFmpeg] First 500 chars of stderr: %s", truncate(result.StderrLog, 500))
	}

	return result
}

// truncate 截断字符串
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return "..." + s[len(s)-maxLen:]
}

// parseProgress 解析FFmpeg输出的进度信息
// FFmpeg输出示例: frame=  120 fps=30 q=28.0 size=    256kB time=00:00:04.00 bitrate= 524.3kbits/s speed=1.0x
func (e *Executor) parseProgress(line string, totalFrames int) *Progress {
	// 正则匹配进度行
	if !strings.Contains(line, "frame=") {
		return nil
	}

	progress := &Progress{
		Timestamp: time.Now(),
	}

	// 提取frame
	if match := regexp.MustCompile(`frame=\s*(\d+)`).FindStringSubmatch(line); len(match) > 1 {
		progress.Frame, _ = strconv.Atoi(match[1])
	}

	// 提取fps
	if match := regexp.MustCompile(`fps=\s*([\d.]+)`).FindStringSubmatch(line); len(match) > 1 {
		progress.FPS, _ = strconv.ParseFloat(match[1], 64)
	}

	// 提取bitrate
	if match := regexp.MustCompile(`bitrate=\s*([\d.]+\s*\w+/s)`).FindStringSubmatch(line); len(match) > 1 {
		progress.Bitrate = strings.TrimSpace(match[1])
	}

	// 提取time
	if match := regexp.MustCompile(`time=\s*([\d:\.]+)`).FindStringSubmatch(line); len(match) > 1 {
		progress.Time = match[1]
	}

	// 提取speed
	if match := regexp.MustCompile(`speed=\s*([\d.]+)x`).FindStringSubmatch(line); len(match) > 1 {
		progress.Speed, _ = strconv.ParseFloat(match[1], 64)
	}

	// 计算进度百分比和ETA
	if totalFrames > 0 && progress.Frame > 0 {
		progress.Progress = float64(progress.Frame) / float64(totalFrames) * 100
		if progress.Progress > 100 {
			progress.Progress = 100
		}

		// 计算ETA
		if progress.Speed > 0 {
			remainingFrames := totalFrames - progress.Frame
			remainingSeconds := float64(remainingFrames) / (progress.FPS * progress.Speed)
			progress.ETA = int(remainingSeconds)
		}
	}

	return progress
}

// extractFilterGraph 从参数中提取filter graph
func (e *Executor) extractFilterGraph(args []string) string {
	for i, arg := range args {
		if (arg == "-filter_complex" || arg == "-vf" || arg == "-af") && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

// extractError 从stderr日志中提取错误信息
func (e *Executor) extractError(stderrLog string) string {
	lines := strings.Split(stderrLog, "\n")
	var errors []string

	for _, line := range lines {
		if strings.Contains(line, "Error") ||
			strings.Contains(line, "error") ||
			strings.Contains(line, "Invalid") ||
			strings.Contains(line, "failed") {
			errors = append(errors, strings.TrimSpace(line))
		}
	}

	if len(errors) > 0 {
		return strings.Join(errors, "\n")
	}

	// 如果没有明确的错误信息，返回最后几行
	if len(lines) > 5 {
		return strings.Join(lines[len(lines)-5:], "\n")
	}
	return stderrLog
}
