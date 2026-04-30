# Dockerfile for Weini Quantum Proxy

FROM golang:1.21-alpine AS builder

# 安装构建依赖
RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建非GUI版本（适合服务器）
RUN cd cmd/proxy-node-studio && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/weini-quantum-proxy

# 运行时镜像
FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/weini-quantum-proxy .

# 复制必要的配置和脚本
COPY --from=builder /app/scripts ./scripts
COPY --from=builder /app/skills ./skills

# 创建运行时目录
RUN mkdir -p /app/runtime /app/data

# 暴露端口
EXPOSE 8080 1080 7890

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 运行
ENTRYPOINT ["/app/weini-quantum-proxy"]
CMD ["--config", "/app/data/config.json"]
