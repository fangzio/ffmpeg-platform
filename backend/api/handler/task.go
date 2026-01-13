package handler

import (
	"ffmpeg-platform/model"
	"ffmpeg-platform/service"
	"ffmpeg-platform/worker"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type TaskHandler struct {
	taskService *service.TaskService
	worker      *worker.Worker
}

func NewTaskHandler(taskService *service.TaskService, w *worker.Worker) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
		worker:      w,
	}
}

// CreateTaskRequest 创建任务请求
type CreateTaskRequest struct {
	Type        string                `json:"type" binding:"required"`
	InputParams model.TaskInputParams `json:"input_params" binding:"required"`
}

// CreateTask 创建任务
// POST /api/tasks
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := h.taskService.CreateTask(req.Type, req.InputParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, task)
}

// GetTask 获取任务详情
// GET /api/tasks/:id
func (h *TaskHandler) GetTask(c *gin.Context) {
	taskID := c.Param("id")

	task, err := h.taskService.GetTask(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// ListTasks 获取任务列表
// GET /api/tasks
func (h *TaskHandler) ListTasks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	tasks, total, err := h.taskService.ListTasks(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks":     tasks,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// WebSocket升级器
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 生产环境需要验证origin
	},
}

// WatchProgress WebSocket连接 - 实时监听任务进度
// GET /api/tasks/:id/progress
func (h *TaskHandler) WatchProgress(c *gin.Context) {
	taskID := c.Param("id")

	// 验证任务是否存在并获取当前状态
	task, err := h.taskService.GetTask(taskID)
	if err != nil {
		log.Printf("WebSocket: task %s not found", taskID)
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	log.Printf("WebSocket: Attempting to upgrade connection for task %s", taskID)

	// 升级到WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket: Failed to upgrade connection for task %s: %v", taskID, err)
		return
	}
	defer conn.Close()

	log.Printf("WebSocket: Connection established for task %s", taskID)

	// 获取进度Hub
	hub := h.worker.GetProgressHub(taskID)
	client := &worker.ProgressClient{
		Hub:  hub,
		Send: make(chan model.TaskProgress, 256),
	}

	hub.Register <- client
	defer func() {
		hub.UnRegister <- client
		log.Printf("WebSocket: Connection closed for task %s", taskID)
	}()

	// 立即发送当前任务状态（确保新连接能同步最新进度）
	currentProgress := model.TaskProgress{
		TaskID:       task.ID,
		Status:       task.Status,
		Progress:     task.Progress,
		CurrentFrame: task.CurrentFrame,
		TotalFrames:  task.TotalFrames,
		ETA:          task.Eta,
		Message:      "Connected to progress stream",
		Timestamp:    time.Now(),
	}
	if err := conn.WriteJSON(currentProgress); err != nil {
		log.Printf("WebSocket: Failed to send initial progress for task %s: %v", taskID, err)
		return
	}

	log.Printf("WebSocket: Sent initial progress for task %s (status: %s, progress: %.1f%%)",
		taskID, task.Status, task.Progress)

	// 发送进度更新到WebSocket
	for progress := range client.Send {
		progress.Timestamp = time.Now()
		if err := conn.WriteJSON(progress); err != nil {
			log.Printf("WebSocket: Failed to send progress update for task %s: %v", taskID, err)
			break
		}
	}
}
