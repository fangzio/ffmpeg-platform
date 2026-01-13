# 多阶段构建 - 构建阶段
FROM golang:1.23.6-alpine AS builder

WORKDIR /app

# 安装构建依赖
RUN apk add --no-cache git

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建二进制文件
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# 运行阶段
FROM alpine:latest

# 安装FFmpeg和ffprobe
RUN apk add --no-cache \
    ffmpeg \
    ca-certificates \
    tzdata

# 设置时区
ENV TZ=Asia/Shanghai

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/main .

# 创建存储目录
RUN mkdir -p /app/storage/uploads /app/storage/outputs

# 暴露端口
EXPOSE 8008

# 运行服务
CMD ["./main"]
