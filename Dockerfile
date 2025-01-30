# 后端构建阶段
FROM golang:1.23rc1-alpine AS backend-builder

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

# 前端构建阶段
FROM node:18-alpine AS frontend-builder

WORKDIR /build

# 复制前端依赖文件
COPY package.json package-lock.json ./

# 安装依赖
RUN npm ci

# 复制前端源代码
COPY . .

# 构建前端应用
RUN npm run build

# 运行阶段
FROM alpine:latest

# 安装基本的运行时依赖
RUN apk add --no-cache ca-certificates tzdata nginx

WORKDIR /app

# 从后端构建阶段复制编译好的应用
COPY --from=backend-builder /build/wechat-reader .

# 从前端构建阶段复制构建好的静态文件
COPY --from=frontend-builder /build/dist /usr/share/nginx/html

# 复制 Nginx 配置文件
COPY nginx.conf /etc/nginx/http.d/default.conf

# 暴露端口
EXPOSE 80

# 启动Nginx和后端服务
CMD ["sh", "-c", "nginx && ./wechat-reader"]
COPY nginx.conf /etc/nginx/http.d/default.conf

# 设置时区
ENV TZ=Asia/Shanghai

EXPOSE 8080
EXPOSE 80

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s \
  CMD wget --spider http://localhost:8080/health || exit 1

# 启动 Nginx 和后端应用
CMD nginx && ./wechat-reader