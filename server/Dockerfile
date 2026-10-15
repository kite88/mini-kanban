# syntax=docker/dockerfile:1

# =============================================================
# mini-kanban-server（Go + Gin + GORM）
# 构建：docker build -t mini-kanban-server ./server
# 说明：容器内由 docker-entrypoint.sh 按环境变量生成 config.yaml
# =============================================================

# ---------- 构建阶段 ----------
FROM golang:1.22-alpine AS builder

# 国内网络可换成镜像加速；置空则用官方 proxy
ARG GOPROXY=https://goproxy.cn,direct

ENV GOPROXY=${GOPROXY} \
    GOTOOLCHAIN=local \
    CGO_ENABLED=0

WORKDIR /src

# 依赖单独一层：只改业务代码时不会重新下载依赖
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# pgx / gin 均为纯 Go 实现，关闭 CGO 后产物可静态运行在 alpine 上
RUN go build -trimpath -ldflags="-s -w" -o /out/mini-kanban .

# ---------- 运行阶段 ----------
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata wget \
 && addgroup -g 10001 -S mini_kanban \
 && adduser  -u 10001 -S -G mini_kanban -H mini_kanban

ENV TZ=Asia/Shanghai

WORKDIR /app

COPY --from=builder /out/mini-kanban /app/mini-kanban
COPY docker-entrypoint.sh /app/docker-entrypoint.sh

# 顺手把脚本里的 CRLF 归一化（在 Windows 上编辑容易带上 \r，shell 会报错）
RUN sed -i 's/\r$//' /app/docker-entrypoint.sh \
 && chmod +x /app/docker-entrypoint.sh \
 && mkdir -p /data/uploads \
 && chown -R mini_kanban:mini_kanban /app /data

USER mini_kanban

# 只在内网被 nginx 访问，不映射到宿主机
EXPOSE 8281

HEALTHCHECK --interval=15s --timeout=3s --start-period=30s --retries=5 \
  CMD wget -q -O /dev/null "http://127.0.0.1:${SERVER_PORT:-8281}/healthz" || exit 1

ENTRYPOINT ["/app/docker-entrypoint.sh"]
CMD ["/app/mini-kanban"]
