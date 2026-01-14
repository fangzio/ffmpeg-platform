package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
)

// QiniuStorage 七牛云存储实现
type QiniuStorage struct {
	accessKey  string
	secretKey  string
	bucket     string
	domain     string
	region     *storage.Region
	uploadDir  string // 临时上传目录
	outputDir  string // 临时输出目录
	mac        *qbox.Mac
	bucketManager *storage.BucketManager
}

// QiniuConfig 七牛云配置
type QiniuConfig struct {
	AccessKey string
	SecretKey string
	Bucket    string
	Domain    string
	Region    string // z0=华东, z1=华北, z2=华南, na0=北美, as0=东南亚
	UploadDir string
	OutputDir string
}

// NewQiniuStorage 创建七牛云存储实例
func NewQiniuStorage(config QiniuConfig) (*QiniuStorage, error) {
	// 创建临时目录
	if err := os.MkdirAll(config.UploadDir, 0755); err != nil {
		return nil, fmt.Errorf("create upload dir failed: %w", err)
	}
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("create output dir failed: %w", err)
	}

	// 获取区域配置
	region, err := getRegion(config.Region)
	if err != nil {
		return nil, fmt.Errorf("get region failed: %w", err)
	}

	mac := qbox.NewMac(config.AccessKey, config.SecretKey)
	cfg := storage.Config{
		Region: region,
		UseHTTPS: true,
		UseCdnDomains: false,
	}
	bucketManager := storage.NewBucketManager(mac, &cfg)

	return &QiniuStorage{
		accessKey:     config.AccessKey,
		secretKey:     config.SecretKey,
		bucket:        config.Bucket,
		domain:        config.Domain,
		region:        region,
		uploadDir:     config.UploadDir,
		outputDir:     config.OutputDir,
		mac:           mac,
		bucketManager: bucketManager,
	}, nil
}

// SaveUploadedFile 保存上传的文件到本地临时目录
func (s *QiniuStorage) SaveUploadedFile(file io.Reader, filename string) (string, error) {
	localPath := filepath.Join(s.uploadDir, filename)
	dst, err := os.Create(localPath)
	if err != nil {
		return "", fmt.Errorf("create file failed: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("save file failed: %w", err)
	}

	return localPath, nil
}

// UploadFile 上传文件到七牛云
func (s *QiniuStorage) UploadFile(localPath string, key string) (string, error) {
	putPolicy := storage.PutPolicy{
		Scope: s.bucket,
	}
	upToken := putPolicy.UploadToken(s.mac)

	cfg := storage.Config{
		Region:        s.region,
		UseHTTPS:      true,
		UseCdnDomains: false,
	}

	formUploader := storage.NewFormUploader(&cfg)
	ret := storage.PutRet{}

	err := formUploader.PutFile(context.Background(), &ret, upToken, key, localPath, nil)
	if err != nil {
		return "", fmt.Errorf("upload to qiniu failed: %w", err)
	}

	// 返回CDN访问URL
	url := s.GetPublicURL(key)
	return url, nil
}

// DeleteLocalFile 删除本地文件
func (s *QiniuStorage) DeleteLocalFile(path string) error {
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("delete local file failed: %w", err)
	}
	return nil
}

// GetUploadPath 获取上传文件的本地路径
func (s *QiniuStorage) GetUploadPath(filename string) string {
	return filepath.Join(s.uploadDir, filename)
}

// GetOutputPath 获取输出文件的本地路径
func (s *QiniuStorage) GetOutputPath(filename string) string {
	return filepath.Join(s.outputDir, filename)
}

// GetPublicURL 获取文件的公开访问URL
func (s *QiniuStorage) GetPublicURL(key string) string {
	return fmt.Sprintf("https://%s/%s", s.domain, key)
}

// getRegion 根据区域代码返回七牛云区域配置
func getRegion(regionCode string) (*storage.Region, error) {
	switch regionCode {
	case "z0", "华东", "华东-浙江":
		return &storage.Region{
			SrcUpHosts: []string{"up-z0.qiniup.com"},
			CdnUpHosts: []string{"upload-z0.qiniup.com"},
			RsHost:     "rs-z0.qbox.me",
			RsfHost:    "rsf-z0.qbox.me",
			ApiHost:    "api-z0.qiniu.com",
		}, nil
	case "z1", "华北", "华北-河北":
		return &storage.Region{
			SrcUpHosts: []string{"up-z1.qiniup.com"},
			CdnUpHosts: []string{"upload-z1.qiniup.com"},
			RsHost:     "rs-z1.qbox.me",
			RsfHost:    "rsf-z1.qbox.me",
			ApiHost:    "api-z1.qiniu.com",
		}, nil
	case "z2", "华南", "华南-广东":
		return &storage.Region{
			SrcUpHosts: []string{"up-z2.qiniup.com"},
			CdnUpHosts: []string{"upload-z2.qiniup.com"},
			RsHost:     "rs-z2.qbox.me",
			RsfHost:    "rsf-z2.qbox.me",
			ApiHost:    "api-z2.qiniu.com",
		}, nil
	case "na0", "北美":
		return &storage.Region{
			SrcUpHosts: []string{"up-na0.qiniup.com"},
			CdnUpHosts: []string{"upload-na0.qiniup.com"},
			RsHost:     "rs-na0.qbox.me",
			RsfHost:    "rsf-na0.qbox.me",
			ApiHost:    "api-na0.qiniu.com",
		}, nil
	case "as0", "东南亚":
		return &storage.Region{
			SrcUpHosts: []string{"up-as0.qiniup.com"},
			CdnUpHosts: []string{"upload-as0.qiniup.com"},
			RsHost:     "rs-as0.qbox.me",
			RsfHost:    "rsf-as0.qbox.me",
			ApiHost:    "api-as0.qiniu.com",
		}, nil
	default:
		return nil, fmt.Errorf("unsupported region: %s", regionCode)
	}
}
