# 构建阶段
FROM golang:1.24-alpine AS builder

WORKDIR /app

# 先复制依赖文件，利用缓存
COPY go.mod go.sum ./
RUN go mod download

# 复制源码
COPY . .

# 编译
RUN CGO_ENABLED=0 GOOS=linux go build -o frozeel ./cmd/frozeel

# 运行阶段
FROM alpine:latest

WORKDIR /app

# 复制二进制和配置
COPY --from=builder /app/frozeel .
COPY --from=builder /app/config ./config
# COPY --from=builder /app/.env .

CMD ["./frozeel"]