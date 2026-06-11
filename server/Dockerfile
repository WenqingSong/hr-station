# ============================================================
# Offer捕手 - 后端生产镜像（多阶段构建）
# 阶段 1: Go 编译
# 阶段 2: Alpine 运行
# ============================================================

FROM golang:1.24-alpine AS builder

WORKDIR /src/server

# 构建依赖
RUN apk add --no-cache git ca-certificates tzdata

# 下载 Go 模块
COPY server/go.mod server/go.sum ./
RUN go mod download

# 复制后端源码
COPY server/ ./

# 编译（静态链接，剥离调试信息）
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" \
    -o /out/offer-catcher \
    ./cmd/server

# ============================================================
# 阶段 2: 运行阶段
# ============================================================
FROM alpine:3.21 AS runner

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata curl

# 复制编译产物
COPY --from=builder /out/offer-catcher /app/offer-catcher

# 复制 LLM prompt / schema 模板（运行时依赖）
COPY --from=builder /src/server/prompts /app/server/prompts

# 健康检查
HEALTHCHECK --interval=15s --timeout=5s --retries=10 --start-period=20s \
    CMD curl -f http://127.0.0.1:8080/healthz || exit 1

EXPOSE 8080

CMD ["/app/offer-catcher"]
