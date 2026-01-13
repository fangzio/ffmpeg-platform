package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Storage  StorageConfig
	Qiniu    QiniuConfig
	FFmpeg   FFmpegConfig
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	DSN string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type StorageConfig struct {
	Type      string // "local" or "qiniu"
	UploadDir string
	OutputDir string
	TempDir   string // 临时文件目录（用于下载远程文件）
}

type QiniuConfig struct {
	Enabled   bool
	AccessKey string
	SecretKey string
	Bucket    string
	Domain    string
	Region    string
}

type FFmpegConfig struct {
	BinaryPath string
	LogLevel   string
}

func Load() *Config {
	// 加载 .env 文件
	err := godotenv.Load()
	if err != nil {
		// 如果找不到 .env 文件，可以尝试加载 .env.local
		err = godotenv.Load("config.env")
		if err != nil {
			log.Fatal("Error loading .env files")
		}
	}
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8008"),
		},
		Database: DatabaseConfig{
			DSN: getEnv("DATABASE_DSN", "host=localhost user=ffmpeg password=ffmpeg dbname=ffmpeg port=5432 sslmode=disable"),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       0,
		},
		Storage: StorageConfig{
			Type:      getEnv("STORAGE_TYPE", "qiniu"),
			UploadDir: getEnv("UPLOAD_DIR", "./storage/uploads"),
			OutputDir: getEnv("OUTPUT_DIR", "./storage/outputs"),
			TempDir:   getEnv("TEMP_DIR", "./storage/temp"),
		},
		Qiniu: QiniuConfig{
			Enabled:   getEnv("QINIU_ENABLED", "true") == "true",
			AccessKey: getEnv("QINIU_ACCESS_KEY", ""),
			SecretKey: getEnv("QINIU_SECRET_KEY", ""),
			Bucket:    getEnv("QINIU_BUCKET", ""),
			Domain:    getEnv("QINIU_DOMAIN", ""),
			Region:    getEnv("QINIU_REGION", "z2"),
		},
		FFmpeg: FFmpegConfig{
			BinaryPath: getEnv("FFMPEG_PATH", "ffmpeg"),
			LogLevel:   getEnv("FFMPEG_LOG_LEVEL", "info"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
