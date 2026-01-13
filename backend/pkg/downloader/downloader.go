package downloader

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Downloader 文件下载器
type Downloader struct {
	tempDir       string
	timeout       time.Duration
	maxRetries    int
	client        *http.Client
}

// NewDownloader 创建下载器
func NewDownloader(tempDir string) *Downloader {
	return &Downloader{
		tempDir:    tempDir,
		timeout:    30 * time.Second,
		maxRetries: 3,
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				IdleConnTimeout:     30 * time.Second,
				DisableCompression:  false,
				DisableKeepAlives:   false,
			},
		},
	}
}

// DownloadFile 下载文件到临时目录，返回本地路径
func (d *Downloader) DownloadFile(urlOrPath string) (string, error) {
	// 如果不是URL，直接返回原路径
	if !d.isURL(urlOrPath) {
		return urlOrPath, nil
	}

	// 生成临时文件路径
	ext := d.getFileExtension(urlOrPath)
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	localPath := filepath.Join(d.tempDir, filename)

	// 确保临时目录存在
	if err := os.MkdirAll(d.tempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// 重试下载
	var lastErr error
	for i := 0; i < d.maxRetries; i++ {
		if i > 0 {
			// 重试前等待
			time.Sleep(time.Duration(i) * time.Second)
		}

		err := d.downloadToFile(urlOrPath, localPath)
		if err == nil {
			return localPath, nil
		}
		lastErr = err
	}

	return "", fmt.Errorf("failed to download after %d retries: %w", d.maxRetries, lastErr)
}

// downloadToFile 执行实际下载
func (d *Downloader) downloadToFile(url, localPath string) error {
	// 发送HTTP请求
	resp, err := d.client.Get(url)
	if err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http status %d: %s", resp.StatusCode, resp.Status)
	}

	// 创建本地文件
	out, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	// 复制数据
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		// 删除不完整的文件
		os.Remove(localPath)
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// isURL 检查是否为URL
func (d *Downloader) isURL(path string) bool {
	return strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://")
}

// getFileExtension 从URL或文件名获取扩展名
func (d *Downloader) getFileExtension(urlOrPath string) string {
	// 去掉查询参数
	if idx := strings.Index(urlOrPath, "?"); idx != -1 {
		urlOrPath = urlOrPath[:idx]
	}

	ext := filepath.Ext(urlOrPath)
	if ext == "" {
		// 尝试从Content-Type推断（暂时返回空）
		return ""
	}
	return ext
}

// CleanupFile 清理下载的临时文件
func (d *Downloader) CleanupFile(path string) error {
	// 只清理临时目录下的文件
	if !strings.HasPrefix(path, d.tempDir) {
		return nil
	}

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove temp file: %w", err)
	}

	return nil
}
