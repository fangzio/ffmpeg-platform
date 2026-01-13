package ffmpeg

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// MediaInfo 媒体文件信息
type MediaInfo struct {
	Duration    float64 // 时长（秒）
	Width       int     // 宽度
	Height      int     // 高度
	FPS         float64 // 帧率
	TotalFrames int     // 总帧数
	AudioCodec  string  // 音频编码
	VideoCodec  string  // 视频编码
}

// Parser FFmpeg解析器
type Parser struct {
	binaryPath string
}

func NewParser(binaryPath string) *Parser {
	return &Parser{
		binaryPath: binaryPath,
	}
}

// GetMediaInfo 获取媒体文件信息（用于计算总帧数）
func (p *Parser) GetMediaInfo(filePath string) (*MediaInfo, error) {
	// 使用ffprobe获取媒体信息
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=duration,width,height,r_frame_rate,codec_name",
		"-of", "default=noprint_wrappers=1",
		filePath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ffprobe failed: %w, output: %s", err, string(output))
	}

	info := &MediaInfo{}
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		value := parts[1]

		switch key {
		case "duration":
			info.Duration, _ = strconv.ParseFloat(value, 64)
		case "width":
			info.Width, _ = strconv.Atoi(value)
		case "height":
			info.Height, _ = strconv.Atoi(value)
		case "r_frame_rate":
			info.FPS = p.parseFPS(value)
		case "codec_name":
			info.VideoCodec = value
		}
	}

	// 计算总帧数
	if info.Duration > 0 && info.FPS > 0 {
		info.TotalFrames = int(info.Duration * info.FPS)
	}

	return info, nil
}

// GetAudioDuration 获取音频时长
func (p *Parser) GetAudioDuration(filePath string) (float64, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		filePath,
	)

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe failed: %w", err)
	}

	duration, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	if err != nil {
		return 0, fmt.Errorf("parse duration failed: %w", err)
	}

	return duration, nil
}

// parseFPS 解析帧率（例如：30/1 -> 30.0）
func (p *Parser) parseFPS(fpsStr string) float64 {
	parts := strings.Split(fpsStr, "/")
	if len(parts) != 2 {
		return 0
	}

	numerator, _ := strconv.ParseFloat(parts[0], 64)
	denominator, _ := strconv.ParseFloat(parts[1], 64)

	if denominator == 0 {
		return 0
	}

	return numerator / denominator
}

// ValidateFile 验证文件是否为有效的媒体文件
func (p *Parser) ValidateFile(filePath string) error {
	cmd := exec.Command("ffprobe", "-v", "error", filePath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("invalid media file: %w", err)
	}
	return nil
}

// ExtractFilterComplexity 分析filter_complex复杂度（用于估算性能）
func (p *Parser) ExtractFilterComplexity(filterGraph string) int {
	complexity := 0

	// 统计常见耗性能的滤镜
	expensiveFilters := []string{
		"scale", "overlay", "blur", "transpose",
		"drawtext", "colorkey", "chromakey",
	}

	for _, filter := range expensiveFilters {
		complexity += strings.Count(filterGraph, filter)
	}

	return complexity
}
