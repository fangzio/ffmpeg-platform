package model

import (
	"time"

	"gorm.io/gorm"
)

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusProcessing TaskStatus = "processing"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
)

// Task 任务模型 - 核心差异点：完整记录执行信息
type Task struct {
	ID            string          `json:"id" xorm:"not null text 'id'" gorm:"id"`
	Type          string          `json:"type" xorm:"text 'type'"`
	Status        TaskStatus      `json:"status" xorm:"text 'status'"`
	Progress      float64         `json:"progress" xorm:"numeric 'progress'"`
	CurrentFrame  int             `json:"current_frame" xorm:"int8 'current_frame'"`
	TotalFrames   int             `json:"total_frames" xorm:"int8 'total_frames'"`
	Eta           int             `json:"eta" xorm:"int8 'eta'"`
	InputParams   TaskInputParams `json:"input_params" xorm:"jsonb 'input_params'" gorm:"serializer:json"`
	FfmpegCommand string          `json:"ffmpeg_command" xorm:"text 'ffmpeg_command'"` // 完整的ffmpeg命令
	FilterGraph   string          `json:"filter_graph" xorm:"text 'filter_graph'"`     // filter_complex图
	StderrLog     string          `json:"stderr_log" xorm:"text 'stderr_log'"`
	ErrorMessage  string          `json:"error_message" xorm:"text 'error_message'"` // 错误摘要
	OutputFile    string          `json:"output_file" xorm:"text 'output_file'"`
	OutputUrl     string          `json:"output_url" xorm:"text 'output_url'"`
	CreatedAt     time.Time       `json:"created_at" xorm:"timestamptz 'created_at'"`
	UpdatedAt     time.Time       `json:"updated_at" xorm:"timestamptz 'updated_at'"`
	DeletedAt     gorm.DeletedAt  `json:"-" xorm:"timestamptz 'deleted_at'" gorm:"index"`
}

// TaskInputParams 输入参数（语义化设计）
// 使用灵活的结构支持多种任务类型
type TaskInputParams struct {
	// 通用参数
	OutputFormat string `json:"output_format"` // 输出格式：mp4, mov等
	VideoCodec   string `json:"video_codec"`   // 视频编码：libx264, libx265
	AudioCodec   string `json:"audio_codec"`   // 音频编码：aac, mp3
	Width        int    `json:"width"`         // 视频宽度
	Height       int    `json:"height"`        // 视频高度
	FPS          int    `json:"fps"`           // 帧率
	VideoBitrate string `json:"video_bitrate"` // 视频码率：1M, 2M
	AudioBitrate string `json:"audio_bitrate"` // 音频码率：128k, 192k

	// 单图片+音频任务参数
	ImagePath string `json:"image_path"` // 单张图片路径
	AudioPath string `json:"audio_path"` // 音频路径
	AudioLoop bool   `json:"audio_loop"` // 是否循环播放音频

	// 多图片轮播任务参数
	ImagePaths      []string `json:"image_paths"`      // 多张图片路径列表
	ImageDuration   float64  `json:"image_duration"`   // 每张图片显示时长（秒）
	TransitionType  string   `json:"transition_type"`  // 转场效果：fade, slide, none
	TransitionDur   float64  `json:"transition_dur"`   // 转场持续时间（秒）
	BackgroundAudio string   `json:"background_audio"` // 背景音乐路径
}

// TaskProgress WebSocket实时进度推送
type TaskProgress struct {
	TaskID       string     `json:"task_id"`
	Status       TaskStatus `json:"status"`
	Progress     float64    `json:"progress"`
	CurrentFrame int        `json:"current_frame"`
	TotalFrames  int        `json:"total_frames"`
	ETA          int        `json:"eta"`
	Message      string     `json:"message"`
	Timestamp    time.Time  `json:"timestamp"`
}
