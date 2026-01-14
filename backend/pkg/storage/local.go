package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type LocalStorage struct {
	uploadDir string
	outputDir string
}

func NewLocalStorage(uploadDir, outputDir string) (*LocalStorage, error) {
	// 创建目录
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("create upload dir failed: %w", err)
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("create output dir failed: %w", err)
	}

	return &LocalStorage{
		uploadDir: uploadDir,
		outputDir: outputDir,
	}, nil
}

func (s *LocalStorage) SaveUploadedFile(file io.Reader, filename string) (string, error) {
	filepath := filepath.Join(s.uploadDir, filename)
	dst, err := os.Create(filepath)
	if err != nil {
		return "", fmt.Errorf("create file failed: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("save file failed: %w", err)
	}

	return filepath, nil
}

// UploadFile 本地存储不需要上传，直接返回本地路径
func (s *LocalStorage) UploadFile(localPath string, key string) (string, error) {
	return localPath, nil
}

// DeleteLocalFile 删除本地文件
func (s *LocalStorage) DeleteLocalFile(path string) error {
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("delete local file failed: %w", err)
	}
	return nil
}

func (s *LocalStorage) GetUploadPath(filename string) string {
	return filepath.Join(s.uploadDir, filename)
}

func (s *LocalStorage) GetOutputPath(filename string) string {
	return filepath.Join(s.outputDir, filename)
}

// GetPublicURL 本地存储返回相对路径
func (s *LocalStorage) GetPublicURL(key string) string {
	return fmt.Sprintf("/uploads/%s", key)
}
