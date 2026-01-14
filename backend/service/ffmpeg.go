package service

import (
	"context"
	"fmt"
	"github.com/fangzio/ffmpeg-platform/config"
	"github.com/fangzio/ffmpeg-platform/model"
	"github.com/fangzio/ffmpeg-platform/pkg/downloader"
	"github.com/fangzio/ffmpeg-platform/pkg/ffmpeg"
	"path/filepath"
)

// FFmpegService FFmpeg服务 - 构建语义化参数到ffmpeg命令的映射
type FFmpegService struct {
	executor   *ffmpeg.Executor
	parser     *ffmpeg.Parser
	downloader *downloader.Downloader
	config     *config.Config
}

func NewFFmpegService(cfg *config.Config) *FFmpegService {
	return &FFmpegService{
		executor:   ffmpeg.NewExecutor(cfg.FFmpeg.BinaryPath, cfg.FFmpeg.LogLevel),
		parser:     ffmpeg.NewParser(cfg.FFmpeg.BinaryPath),
		downloader: downloader.NewDownloader(cfg.Storage.TempDir),
		config:     cfg,
	}
}

// BuildImageAudioToVideoCommand 构建图片+音频生成视频的ffmpeg命令
// 返回值：命令参数、总帧数、临时文件列表（需要清理）、错误
func (s *FFmpegService) BuildImageAudioToVideoCommand(params model.TaskInputParams, outputPath string) ([]string, int, []string, error) {
	var tempFiles []string

	// 下载远程图片到本地（解决网络IO瓶颈）
	localImagePath, err := s.downloader.DownloadFile(params.ImagePath)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("download image failed: %w", err)
	}
	// 如果是下载的临时文件，记录下来以便后续清理
	if localImagePath != params.ImagePath {
		tempFiles = append(tempFiles, localImagePath)
	}

	// 下载远程音频到本地
	localAudioPath, err := s.downloader.DownloadFile(params.AudioPath)
	if err != nil {
		// 清理已下载的图片
		for _, f := range tempFiles {
			s.downloader.CleanupFile(f)
		}
		return nil, 0, nil, fmt.Errorf("download audio failed: %w", err)
	}
	// 如果是下载的临时文件，记录下来以便后续清理
	if localAudioPath != params.AudioPath {
		tempFiles = append(tempFiles, localAudioPath)
	}

	// 获取音频时长用于计算总帧数（使用本地文件）
	audioDuration, err := s.parser.GetAudioDuration(localAudioPath)
	if err != nil {
		// 清理已下载的文件
		for _, f := range tempFiles {
			s.downloader.CleanupFile(f)
		}
		return nil, 0, nil, fmt.Errorf("get audio duration failed: %w", err)
	}

	// 计算总帧数
	fps := params.FPS
	if fps == 0 {
		fps = 25 // 默认25fps
	}
	totalFrames := int(audioDuration * float64(fps))

	// 构建ffmpeg参数（语义化 -> 命令行参数）
	args := []string{
		// 移除网络参数，因为现在使用本地文件
		// 日志级别和进度输出
		"-loglevel", "info", // 确保有日志输出
		"-stats", // 输出统计信息到 stderr

		"-loop", "1", // 循环图片
		"-i", localImagePath, // 使用本地图片路径
		"-i", localAudioPath, // 使用本地音频路径
		"-c:v", s.getVideoCodec(params.VideoCodec), // 视频编码器
		"-preset", "ultrafast", // 使用最快的编码preset，提升处理速度
		"-c:a", s.getAudioCodec(params.AudioCodec), // 音频编码器
		"-b:v", s.getVideoBitrate(params.VideoBitrate), // 视频码率
		"-b:a", s.getAudioBitrate(params.AudioBitrate), // 音频码率
		"-r", fmt.Sprintf("%d", fps), // 帧率
		"-pix_fmt", "yuv420p", // 像素格式（兼容性）
	}

	// 视频缩放
	if params.Width > 0 && params.Height > 0 {
		args = append(args, "-vf", fmt.Sprintf("scale=%d:%d", params.Width, params.Height))
	}

	// 音频循环
	if params.AudioLoop {
		args = append(args, "-stream_loop", "-1") // 无限循环音频
	} else {
		args = append(args, "-shortest") // 以最短流为准
	}

	// 输出格式
	args = append(args,
		"-f", s.getOutputFormat(params.OutputFormat),
		"-y", // 覆盖输出文件
		outputPath,
	)

	return args, totalFrames, tempFiles, nil
}

// ExecuteWithProgress 执行ffmpeg命令并实时报告进度
func (s *FFmpegService) ExecuteWithProgress(
	ctx context.Context,
	args []string,
	totalFrames int,
	callback ffmpeg.ProgressCallback,
) *ffmpeg.ExecuteResult {
	return s.executor.Execute(ctx, args, totalFrames, callback)
}

// 辅助方法：提供默认值和参数验证

func (s *FFmpegService) getVideoCodec(codec string) string {
	if codec == "" {
		return "libx264" // 默认H.264
	}
	return codec
}

func (s *FFmpegService) getAudioCodec(codec string) string {
	if codec == "" {
		return "aac" // 默认AAC
	}
	return codec
}

func (s *FFmpegService) getVideoBitrate(bitrate string) string {
	if bitrate == "" {
		return "1M" // 默认1Mbps
	}
	return bitrate
}

func (s *FFmpegService) getAudioBitrate(bitrate string) string {
	if bitrate == "" {
		return "128k" // 默认128kbps
	}
	return bitrate
}

func (s *FFmpegService) getOutputFormat(format string) string {
	if format == "" {
		return "mp4" // 默认MP4
	}
	return format
}

// ValidateInputs 验证输入文件（智能判断不同任务类型）
func (s *FFmpegService) ValidateInputs(params model.TaskInputParams) error {
	// 验证图片文件
	// 优先检查多图片场景（ImagePaths），如果不存在则检查单图片场景（ImagePath）
	if len(params.ImagePaths) > 0 {
		// 多图片轮播场景
		for i, imagePath := range params.ImagePaths {
			if imagePath == "" {
				continue // 跳过空路径
			}
			if err := s.parser.ValidateFile(imagePath); err != nil {
				return fmt.Errorf("invalid image file at index %d: %w", i, err)
			}
		}
	} else if params.ImagePath != "" {
		// 单图片场景
		if err := s.parser.ValidateFile(params.ImagePath); err != nil {
			return fmt.Errorf("invalid image file: %w", err)
		}
	}

	// 验证音频文件
	// 优先检查单图片+音频场景的 AudioPath，其次检查多图片轮播的 BackgroundAudio
	if params.AudioPath != "" {
		if err := s.parser.ValidateFile(params.AudioPath); err != nil {
			return fmt.Errorf("invalid audio file: %w", err)
		}
	} else if params.BackgroundAudio != "" {
		if err := s.parser.ValidateFile(params.BackgroundAudio); err != nil {
			return fmt.Errorf("invalid background audio file: %w", err)
		}
	}

	return nil
}

// GenerateOutputPath 生成输出文件路径
func (s *FFmpegService) GenerateOutputPath(taskID string, format string) string {
	filename := s.GenerateOutputName(taskID, format)
	return filepath.Join(s.config.Storage.OutputDir, filename)
}

// GenerateOutputName 生成输出文件名称
func (s *FFmpegService) GenerateOutputName(taskID string, format string) string {
	if format == "" {
		format = "mp4"
	}
	return fmt.Sprintf("%s.%s", taskID, format)
}

// CleanupTempFiles 清理临时文件
func (s *FFmpegService) CleanupTempFiles(tempFiles []string) {
	for _, f := range tempFiles {
		if err := s.downloader.CleanupFile(f); err != nil {
			// 记录错误但不影响主流程
			fmt.Printf("Warning: failed to cleanup temp file %s: %v\n", f, err)
		}
	}
}

// BuildImageSlideshowCommand 构建多图片轮播视频的ffmpeg命令
// 返回值：命令参数、总帧数、临时文件列表（需要清理）、错误
func (s *FFmpegService) BuildImageSlideshowCommand(params model.TaskInputParams, outputPath string) ([]string, int, []string, error) {
	var tempFiles []string

	if len(params.ImagePaths) == 0 {
		return nil, 0, nil, fmt.Errorf("no images provided")
	}

	// 下载所有图片到本地
	localImagePaths := make([]string, 0, len(params.ImagePaths))
	for _, imagePath := range params.ImagePaths {
		localPath, err := s.downloader.DownloadFile(imagePath)
		if err != nil {
			// 清理已下载的文件
			for _, f := range tempFiles {
				s.downloader.CleanupFile(f)
			}
			return nil, 0, nil, fmt.Errorf("download image %s failed: %w", imagePath, err)
		}
		localImagePaths = append(localImagePaths, localPath)
		if localPath != imagePath {
			tempFiles = append(tempFiles, localPath)
		}
	}

	// 下载背景音乐（如果有）
	var localAudioPath string
	var audioDuration float64
	if params.BackgroundAudio != "" {
		var err error
		localAudioPath, err = s.downloader.DownloadFile(params.BackgroundAudio)
		if err != nil {
			// 清理已下载的文件
			for _, f := range tempFiles {
				s.downloader.CleanupFile(f)
			}
			return nil, 0, nil, fmt.Errorf("download audio failed: %w", err)
		}
		if localAudioPath != params.BackgroundAudio {
			tempFiles = append(tempFiles, localAudioPath)
		}

		// 获取音频时长
		audioDuration, err = s.parser.GetAudioDuration(localAudioPath)
		if err != nil {
			// 清理已下载的文件
			for _, f := range tempFiles {
				s.downloader.CleanupFile(f)
			}
			return nil, 0, nil, fmt.Errorf("get audio duration failed: %w", err)
		}
	}

	// 设置默认值
	fps := params.FPS
	if fps == 0 {
		fps = 25
	}

	imageDuration := params.ImageDuration
	if imageDuration == 0 {
		imageDuration = 3.0 // 默认每张图片3秒
	}

	transitionDur := params.TransitionDur
	if transitionDur == 0 {
		transitionDur = 0.5 // 默认转场0.5秒
	}

	transitionType := params.TransitionType
	if transitionType == "" {
		transitionType = "fade" // 默认淡入淡出
	}

	// 计算视频总时长和总帧数
	numImages := len(localImagePaths)
	var totalDuration float64

	if transitionType == "none" {
		// 无转场，直接拼接
		totalDuration = float64(numImages) * imageDuration
	} else {
		// 有转场，图片之间有重叠
		totalDuration = float64(numImages)*imageDuration - float64(numImages-1)*transitionDur
	}

	// 如果有音频，视频时长以音频为准
	if audioDuration > 0 {
		totalDuration = audioDuration
	}

	totalFrames := int(totalDuration * float64(fps))

	// 构建filter_complex
	filterComplex := s.buildSlideshowFilter(localImagePaths, imageDuration, transitionDur, transitionType, params.Width, params.Height, fps)

	// 构建ffmpeg命令
	args := []string{
		"-loglevel", "info",
		"-stats",
	}

	// 添加所有图片作为输入
	for _, imgPath := range localImagePaths {
		args = append(args, "-loop", "1", "-t", fmt.Sprintf("%.2f", imageDuration), "-i", imgPath)
	}

	// 添加音频输入（如果有）
	if localAudioPath != "" {
		args = append(args, "-i", localAudioPath)
	}

	// 添加filter_complex
	args = append(args,
		"-filter_complex", filterComplex,
		"-map", "[v]", // 映射视频流
	)

	// 映射音频流
	if localAudioPath != "" {
		args = append(args, "-map", fmt.Sprintf("%d:a", len(localImagePaths)))
	}

	// 编码参数
	args = append(args,
		"-c:v", s.getVideoCodec(params.VideoCodec),
		"-preset", "ultrafast",
		"-b:v", s.getVideoBitrate(params.VideoBitrate),
		"-r", fmt.Sprintf("%d", fps),
		"-pix_fmt", "yuv420p",
	)

	// 音频编码参数
	if localAudioPath != "" {
		args = append(args,
			"-c:a", s.getAudioCodec(params.AudioCodec),
			"-b:a", s.getAudioBitrate(params.AudioBitrate),
			"-shortest", // 以最短流为准
		)
	}

	// 输出参数
	args = append(args,
		"-f", s.getOutputFormat(params.OutputFormat),
		"-y",
		outputPath,
	)

	return args, totalFrames, tempFiles, nil
}

// buildSlideshowFilter 构建幻灯片的filter_complex
func (s *FFmpegService) buildSlideshowFilter(imagePaths []string, duration, transitionDur float64, transitionType string, width, height, fps int) string {
	numImages := len(imagePaths)

	// 设置默认尺寸
	if width == 0 {
		width = 1280
	}
	if height == 0 {
		height = 720
	}

	var filter string

	if transitionType == "none" {
		// 无转场效果 - 简单拼接
		for i := 0; i < numImages; i++ {
			filter += fmt.Sprintf("[%d:v]scale=%d:%d,setsar=1,fps=%d,settb=AVTB[v%d];", i, width, height, fps, i)
		}
		// 拼接所有视频片段
		for i := 0; i < numImages; i++ {
			filter += fmt.Sprintf("[v%d]", i)
		}
		filter += fmt.Sprintf("concat=n=%d:v=1:a=0[v]", numImages)
	} else if transitionType == "fade" {
		// 淡入淡出转场
		// 首先缩放所有图片
		for i := 0; i < numImages; i++ {
			filter += fmt.Sprintf("[%d:v]scale=%d:%d,setsar=1,fps=%d,settb=AVTB[v%d];", i, width, height, fps, i)
		}

		// 构建淡入淡出链 - 正确的xfade链式语法
		// xfade需要两个输入，产生一个输出，然后这个输出再和下一个输入做xfade
		if numImages == 2 {
			// 只有2张图片的简单情况
			offset := duration - transitionDur
			filter += fmt.Sprintf("[v0][v1]xfade=transition=fade:duration=%.2f:offset=%.2f[v]", transitionDur, offset)
		} else {
			// 多张图片的链式xfade
			// 第一个xfade: [v0] + [v1] -> [vt1]
			offset := duration - transitionDur
			filter += fmt.Sprintf("[v0][v1]xfade=transition=fade:duration=%.2f:offset=%.2f[vt1];", transitionDur, offset)

			// 后续的xfade: [vt(i-1)] + [vi] -> [vti] 或 [v]
			for i := 2; i < numImages; i++ {
				// 计算累积的offset
				// 每次xfade的输入流长度 = 前面所有图片的总时长 - 前面所有转场的重叠时长
				// 对于第i个图片(从0开始): 累积长度 = i * duration - (i-1) * transitionDur
				cumulativeLength := float64(i)*duration - float64(i-1)*transitionDur
				offset := cumulativeLength - transitionDur

				if i == numImages-1 {
					// 最后一个xfade，输出为[v]
					filter += fmt.Sprintf("[vt%d][v%d]xfade=transition=fade:duration=%.2f:offset=%.2f[v]", i-1, i, transitionDur, offset)
				} else {
					// 中间的xfade，输出为[vt(i)]
					filter += fmt.Sprintf("[vt%d][v%d]xfade=transition=fade:duration=%.2f:offset=%.2f[vt%d];", i-1, i, transitionDur, offset, i)
				}
			}
		}
	} else {
		// 默认使用fade
		return s.buildSlideshowFilter(imagePaths, duration, transitionDur, "fade", width, height, fps)
	}

	return filter
}
