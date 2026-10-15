#!/bin/sh
# =============================================================
# mini-kanban 容器入口
#
# 为什么需要这个脚本：
#   后端把 config.yaml 当作配置的唯一来源，环境变量只能覆盖
#   PORT / DB_DSN / REDIS_ADDR / REDIS_PASSWORD / JWT_SECRET 五项，
#   seed 段更是完全没有环境变量入口。为了让 compose 里改几个变量
#   就能跑起来，这里按环境变量生成配置文件。
#
# 优先级：
#   若 $CONFIG_FILE（默认 /app/config.yaml）已存在——例如 compose 里
#   挂载了自定义配置——则原样使用，不生成、不覆盖。
# =============================================================
set -e

CONFIG_PATH="${CONFIG_FILE:-/app/config.yaml}"

# 转义 YAML 双引号字符串中的反斜杠与双引号，避免密码 / 密钥里的字符破坏文件结构
yaml_escape() {
    printf '%s' "$1" | sed 's/\\/\\\\/g; s/"/\\"/g'
}

if [ -f "$CONFIG_PATH" ]; then
    echo "[entrypoint] 检测到已有配置 ${CONFIG_PATH}，直接使用（未按环境变量生成）"
else
    echo "[entrypoint] 未发现 ${CONFIG_PATH}，按环境变量生成配置"
    mkdir -p "$(dirname "${CONFIG_PATH}")"

    cat > "${CONFIG_PATH}" <<EOF
# 本文件由 docker-entrypoint.sh 依据环境变量自动生成，请勿手工编辑。
# 需要完全自定义（例如只用本机已有数据库）时，把自定义 config.yaml
# 挂载到 ${CONFIG_PATH} 即可，入口脚本会跳过生成。
server:
  port: "$(yaml_escape "${SERVER_PORT:-8281}")"
  mode: "$(yaml_escape "${SERVER_MODE:-release}")"

database:
  # dsn 非空时优先使用，下面 host/port/... 不再校验；留空则按字段拼装
  dsn: "$(yaml_escape "${DB_DSN:-}")"
  host: "$(yaml_escape "${DB_HOST:-postgres}")"
  port: ${DB_PORT:-5432}
  user: "$(yaml_escape "${DB_USER:-mini-kanban}")"
  password: "$(yaml_escape "${DB_PASSWORD:-}")"
  dbname: "$(yaml_escape "${DB_NAME:-mini-kanban}")"
  sslmode: "$(yaml_escape "${DB_SSLMODE:-disable}")"
  preferSimpleProtocol: ${DB_PREFER_SIMPLE_PROTOCOL:-false}
  logLevel: "$(yaml_escape "${DB_LOG_LEVEL:-warn}")"

redis:
  # host 留空表示关闭缓存（服务照常运行，/healthz 的 redis 字段为 false）
  host: "$(yaml_escape "${REDIS_HOST:-redis}")"
  port: ${REDIS_PORT:-6379}
  password: "$(yaml_escape "${REDIS_PASSWORD:-}")"
  db: ${REDIS_DB:-0}
  cacheTTL: "$(yaml_escape "${REDIS_CACHE_TTL:-30s}")"
  userCacheTTL: "$(yaml_escape "${REDIS_USER_CACHE_TTL:-1m}")"

jwt:
  secret: "$(yaml_escape "${JWT_SECRET:-}")"
  ttlHours: ${JWT_TTL_HOURS:-24}

seed:
  # 仅在「用户表为空」时写入初始管理员；写入成功后本段会被服务自动清空，
  # 所以首次部署完成后建议把 SEED_ENABLED 改回 false
  enabled: ${SEED_ENABLED:-false}
  # 置 true 时忽略「用户表为空」的前提，强制重置 / 重建该管理员（仅本次启动生效）
  forceReset: ${SEED_FORCE_RESET:-false}
  adminEmail: "$(yaml_escape "${SEED_ADMIN_EMAIL:-}")"
  adminUsername: "$(yaml_escape "${SEED_ADMIN_USERNAME:-}")"
  adminPassword: "$(yaml_escape "${SEED_ADMIN_PASSWORD:-}")"

storage:
  # 容器内固定用绝对路径，避免受工作目录影响
  dir: "$(yaml_escape "${STORAGE_DIR:-/data/uploads}")"
  maxSizeMB: ${STORAGE_MAX_SIZE_MB:-10}
EOF
fi

# 附件目录：storage.dir 不存在会自动创建，这里提前建好并兜底
UPLOAD_DIR="${STORAGE_DIR:-/data/uploads}"
mkdir -p "${UPLOAD_DIR}" 2>/dev/null || true

exec "$@"
