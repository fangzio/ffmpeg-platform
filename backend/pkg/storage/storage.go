package storage

import "io"

// Storage 定义统一的存储接口
type Storage interface {
	// SaveUploadedFile 保存上传的文件，返回文件路径和错误
	SaveUploadedFile(file io.Reader, filename string) (string, error)

	// UploadFile 上传文件到云存储，返回访问URL和错误
	UploadFile(localPath string, key string) (string, error)

	// DeleteLocalFile 删除本地文件
	DeleteLocalFile(path string) error

	// GetUploadPath 获取上传文件的本地路径
	GetUploadPath(filename string) string

	// GetOutputPath 获取输出文件的本地路径
	GetOutputPath(filename string) string

	// GetPublicURL 获取文件的公开访问URL
	GetPublicURL(key string) string
}
