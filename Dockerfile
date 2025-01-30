# 构建阶段
FROM golang:1.23rc1-alpine AS builder

# 安装必要的系统依赖
RUN apk add --no-cache gcc musl-dev

WORKDIR /build

# 首先复制依赖文件，利用 Docker 缓存
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=1 go build -o wechat-reader ./cmd/server

# 运行阶段
FROM alpine:latest

# 安装基本的运行时依赖
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# 从构建阶段复制编译好的应用
COPY --from=builder /build/wechat-reader .

# 设置时区
ENV TZ=Asia/Shanghai

EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s \
  CMD wget --spider http://localhost:8080/health || exit 1

# 运行应用
CMD ["./wechat-reader"]