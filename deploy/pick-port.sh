#!/bin/bash
# =============================================================
# 挑选一个空闲的宿主机端口
#
# 用法一（自动写入 .env）：
#   bash deploy/pick-port.sh
# 用法二（只打印端口号，供其他脚本调用）：
#   PORT=$(bash deploy/pick-port.sh --print)
#
# 背景：整份 compose 里只有 nginx 会把端口映射到宿主机（HTTP_PORT，默认 8080）；
#       SERVER_PORT / DB_PORT / REDIS_PORT / 前端 3201 都是容器内端口，
#       与宿主机上跑了什么服务无关，不会冲突。
#
# 仅适用于 Linux（依赖 ss 或 netstat）。
# =============================================================
set -eu

if ! command -v ss >/dev/null 2>&1 && ! command -v netstat >/dev/null 2>&1; then
    echo "系统里既没有 ss 也没有 netstat，无法检测端口占用" >&2
    exit 1
fi

# ---------- 宿主机已监听的端口 ----------
list_busy_ports() {
    if command -v ss >/dev/null 2>&1; then
        ss -lntH 2>/dev/null | awk '{print $4}' || true
    else
        netstat -lnt 2>/dev/null | awk 'NR>2 {print $4}' || true
    fi
}

# 0.0.0.0:8080 / [::]:8080 / 127.0.0.1:8080 → 8080
BUSY=$(list_busy_ports | sed 's/.*://' | sort -u)

is_busy() {
    printf '%s\n' "${BUSY}" | grep -qx "$1"
}

# ---------- 选端口 ----------
# 先用不易撞车的固定候选（避开 80/443/3000/8000/8080/8888/9000 等常见端口），
# 都被占用时落到 IANA 动态端口段 49152-65535 随机试
pick_port() {
    local p i
    for p in 18080 28080 32100 39218 41000 45678; do
        if ! is_busy "$p"; then
            echo "$p"
            return 0
        fi
    done
    for ((i = 0; i < 100; i++)); do
        p=$(( (RANDOM % 16384) + 49152 ))
        if ! is_busy "$p"; then
            echo "$p"
            return 0
        fi
    done
    return 1
}

# ---------- 只打印端口号 ----------
if [ "${1:-}" = "--print" ]; then
    pick_port || { echo "没找到空闲端口" >&2; exit 1; }
    exit 0
fi

# ---------- 写入 .env ----------
ENV_FILE="${1:-.env}"
if [ ! -f "${ENV_FILE}" ]; then
    echo "找不到 ${ENV_FILE}，请先执行：cp .env.example.docker-compose .env" >&2
    exit 1
fi

PORT=$(pick_port) || { echo "没找到空闲端口，请手工指定 HTTP_PORT" >&2; exit 1; }

if grep -q '^HTTP_PORT=' "${ENV_FILE}"; then
    sed -i "s|^HTTP_PORT=.*|HTTP_PORT=${PORT}|" "${ENV_FILE}"
else
    printf '\nHTTP_PORT=%s\n' "${PORT}" >> "${ENV_FILE}"
fi

echo "已将 HTTP_PORT 设为 ${PORT}"
echo "访问地址：http://<服务器IP>:${PORT}"
echo "提示：建议在防火墙里只放行该端口。"
