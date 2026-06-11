# ============================================================
# Offer捕手 - 前端生产镜像（多阶段构建）
# 阶段 1: Node.js 构建
# 阶段 2: Nginx 托管静态产物
# ============================================================

FROM node:22-alpine AS builder

WORKDIR /app

# 安装依赖（优先 npm ci）
COPY web/package.json web/package-lock.json ./
RUN npm ci

# 复制前端源码
COPY web/ ./

# 构建参数：生产环境建议 /api（同域反代）
ARG VITE_API_BASE_URL=/api
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}

# 执行构建
RUN npm run build

# ============================================================
# 阶段 2: Nginx 运行阶段
# ============================================================
FROM nginx:1.27-alpine AS runner

# 复制 Nginx 配置
COPY web/nginx.conf /etc/nginx/conf.d/default.conf

# 复制构建产物
COPY --from=builder /app/dist /usr/share/nginx/html

# 健康检查依赖
RUN apk add --no-cache curl

EXPOSE 80

HEALTHCHECK --interval=15s --timeout=5s --retries=5 --start-period=5s \
    CMD curl -f http://127.0.0.1:80/ || exit 1

CMD ["nginx", "-g", "daemon off;"]
