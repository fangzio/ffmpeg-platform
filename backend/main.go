package main

import (
	"ffmpeg-platform/api/handler"
	"ffmpeg-platform/api/middleware"
	"ffmpeg-platform/config"
	"ffmpeg-platform/model"
	"ffmpeg-platform/pkg/storage"
	"ffmpeg-platform/service"
	"ffmpeg-platform/worker"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 初始化数据库
	db, err := initDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 初始化存储
	var storageImpl storage.Storage
	if cfg.Storage.Type == "qiniu" && cfg.Qiniu.Enabled {
		qiniuStorage, err := storage.NewQiniuStorage(storage.QiniuConfig{
			AccessKey: cfg.Qiniu.AccessKey,
			SecretKey: cfg.Qiniu.SecretKey,
			Bucket:    cfg.Qiniu.Bucket,
			Domain:    cfg.Qiniu.Domain,
			Region:    cfg.Qiniu.Region,
			UploadDir: cfg.Storage.UploadDir,
			OutputDir: cfg.Storage.OutputDir,
		})
		if err != nil {
			log.Fatalf("Failed to initialize qiniu storage: %v", err)
		}
		storageImpl = qiniuStorage
		log.Println("Using Qiniu Cloud Storage")
	} else {
		localStorage, err := storage.NewLocalStorage(cfg.Storage.UploadDir, cfg.Storage.OutputDir)
		if err != nil {
			log.Fatalf("Failed to initialize local storage: %v", err)
		}
		storageImpl = localStorage
		log.Println("Using Local Storage")
	}

	// 初始化Asynq客户端
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer asynqClient.Close()

	// 初始化服务
	taskService := service.NewTaskService(db, asynqClient, cfg)

	// 初始化Worker
	w := worker.NewWorker(db, taskService, cfg, storageImpl)

	// 启动Asynq Worker
	srv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     cfg.Redis.Addr,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		},
		asynq.Config{
			Concurrency: 10, // 并发处理数
		},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc("task:process", w.ProcessTask)

	// 启动Worker（在后台goroutine）
	go func() {
		if err := srv.Run(mux); err != nil {
			log.Fatalf("Failed to run asynq server: %v", err)
		}
	}()

	// 初始化Gin
	r := gin.Default()

	// 中间件
	r.Use(middleware.CORS())

	// 注册路由
	taskHandler := handler.NewTaskHandler(taskService, w)

	api := r.Group("/api")
	{
		api.POST("/tasks", taskHandler.CreateTask)
		api.GET("/tasks", taskHandler.ListTasks)
		api.GET("/tasks/:id", taskHandler.GetTask)
		api.GET("/tasks/:id/progress", taskHandler.WatchProgress) // WebSocket
	}

	// 上传接口
	api.POST("/upload", func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(400, gin.H{"error": "no file uploaded"})
			return
		}

		// 生成文件名
		filename := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)

		// 打开上传的文件
		src, err := file.Open()
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to open uploaded file"})
			return
		}
		defer src.Close()

		// 保存文件到本地（临时）
		localPath, err := storageImpl.SaveUploadedFile(src, filename)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to save file"})
			return
		}

		// 如果是七牛云存储，上传到云端并删除本地文件
		var fileURL string
		if cfg.Storage.Type == "qiniu" && cfg.Qiniu.Enabled {
			// 生成七牛云存储key（使用uploads目录前缀）
			key := fmt.Sprintf("uploads/%s", filename)
			cloudURL, err := storageImpl.UploadFile(localPath, key)
			if err != nil {
				c.JSON(500, gin.H{"error": fmt.Sprintf("failed to upload to cloud: %v", err)})
				return
			}
			fileURL = cloudURL

			// 删除本地临时文件
			if err := storageImpl.DeleteLocalFile(localPath); err != nil {
				log.Printf("Warning: failed to delete local file %s: %v", localPath, err)
			}
		} else {
			fileURL = fmt.Sprintf("/api/uploads/%s", filename)
		}

		c.JSON(200, gin.H{
			"filename": filename,
			"path":     localPath,
			"url":      fileURL,
		})
	})

	// 静态文件服务（提供上传和输出文件下载，仅用于本地存储）
	if cfg.Storage.Type == "local" {
		api.Static("/uploads", cfg.Storage.UploadDir)
		api.Static("/outputs", cfg.Storage.OutputDir)
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 启动服务
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Server starting on %s", addr)
	go func() {
		if err := r.Run(addr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 接收关闭信号
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	fmt.Println("Shutdown Server ...")
}

func initDatabase(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.Database.DSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// 自动迁移
	if err := db.AutoMigrate(&model.Task{}); err != nil {
		return nil, err
	}

	return db, nil
}
