#!/bin/bash
# =============================================================
# Linux 一键部署脚本
#
# 用法（仓库根目录）：
#   bash deploy/deploy.sh
#   bash deploy/deploy.sh --admin-email me@corp.com --admin-user kevin
#
# 流程：
#   1. 从 .env.example.docker-compose 生成 .env（已存在则保留，绝不覆盖已生效的密码）
#   2. 随机生成 POSTGRES_PASSWORD / REDIS_PASSWORD / JWT_SECRET
#      （纯十六进制，不含 # $ 空格 引号，写进 .env 与生成的 config.yaml 都安全）
#   3. 生成初始管理员账号并打开 seed，密码随机生成、部署结果里会打印出来
#   4. 若 HTTP_PORT 已被宿主机占用，自动换一个空闲端口
#   5. docker compose up -d --build
#
# 重复执行是安全的：
#   - 已存在的 .env 与已生成的密码不会被改动
#   - 初始管理员只在「用户表为空」时创建，重复部署不会重建、也不会改掉已设置的密码
# =============================================================
set -eu

# ---------- 命令行参数 ----------
ARG_ADMIN_EMAIL=""
ARG_ADMIN_USER=""
RESET_VOLUMES=0
RESET_ADMIN=0

usage() {
    cat <<'USAGE'
用法：bash deploy/deploy.sh [选项]

  --admin-email <邮箱>    初始管理员登录邮箱（默认 admin@mail.example.com）
  --admin-user  <用户名>  初始管理员显示名（默认 Admin）
  --reset                 启动前清空数据卷（数据库 / Redis / 附件都会被删除）
  --reset-admin           强制重置 / 重建初始管理员账号（不动其他数据）
  -h, --help              显示本帮助

说明：初始管理员的密码由脚本随机生成并打印在部署结果里。
      该账号只在「用户表为空」时创建，重复部署不会重建或改密码。

什么时候需要 --reset：数据库密码只在数据卷首次初始化时写入，
若卷里的密码与 .env 对不上（表现为服务反复报 password authentication
failed），而库里没有需要保留的数据，就用 --reset 清卷重来。

什么时候需要 --reset-admin：忘了管理员密码，但库里已有数据、不想清空重来。
它会把 SEED_ADMIN_EMAIL 指定的账号重置为管理员并换一个随机密码，
该邮箱已不存在（例如被改掉或删掉）时则重新创建，其他数据不受影响。
USAGE
}

while [ $# -gt 0 ]; do
    case "$1" in
        --admin-email)
            [ $# -ge 2 ] || { echo "--admin-email 需要一个值" >&2; exit 1; }
            ARG_ADMIN_EMAIL="$2"
            shift 2
            ;;
        --admin-user)
            [ $# -ge 2 ] || { echo "--admin-user 需要一个值" >&2; exit 1; }
            ARG_ADMIN_USER="$2"
            shift 2
            ;;
        --reset)
            RESET_VOLUMES=1
            shift
            ;;
        --reset-admin)
            RESET_ADMIN=1
            shift
            ;;
        -h | --help)
            usage
            exit 0
            ;;
        *)
            echo "未知参数：$1" >&2
            usage >&2
            exit 1
            ;;
    esac
done

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${ROOT_DIR}"

ENV_FILE="${ROOT_DIR}/.env"
TEMPLATE="${ROOT_DIR}/.env.example.docker-compose"

# ---------- 前置检查 ----------
if [ ! -f docker-compose.yml ]; then
    echo "没找到 docker-compose.yml，请在仓库根目录执行本脚本" >&2
    exit 1
fi
if [ ! -f "${TEMPLATE}" ]; then
    echo "没找到 .env.example.docker-compose 模板" >&2
    exit 1
fi

# 注意：这里用 --version 实际执行一次来验证可用性。
# 只判断 command -v 是不够的：装了 Docker Desktop 但没开启 WSL 集成时，
# docker / docker-compose 会存在一个只会打印提示的 stub，看着有、跑起来必失败。
if docker compose version >/dev/null 2>&1; then
    DC="docker compose"
elif docker-compose version >/dev/null 2>&1; then
    DC="docker-compose"
else
    echo "未检测到可用的 docker compose" >&2
    echo "  - 用 Docker Desktop 时：确认已启动，并在 Settings → Resources → WSL integration 里为本发行版开启集成" >&2
    echo "  - 原生 Docker 时：安装 compose 插件（docker compose）或 docker-compose" >&2
    exit 1
fi

# ---------- 工具函数 ----------
rand_hex() {
    if command -v openssl >/dev/null 2>&1; then
        openssl rand -hex "$1"
    else
        head -c "$1" /dev/urandom | od -An -tx1 | tr -d ' \n'
    fi
}

# 生成 16 位管理员密码：
#   - 含大小写字母与数字，便于人工输入
#   - 排除易混淆的 0 O 1 l I，避免手抄出错
#   - 只用字母数字，不含 # $ 空格 引号，不会破坏 .env 与生成的 config.yaml
rand_password() {
    LC_ALL=C tr -dc 'A-HJ-NP-Za-km-z2-9' < /dev/urandom 2>/dev/null | head -c 16
}

env_get() {
    grep "^$1=" "${ENV_FILE}" 2>/dev/null | head -n1 | cut -d= -f2- || true
}

env_set() {
    if grep -q "^$1=" "${ENV_FILE}"; then
        sed -i "s|^$1=.*|$1=$2|" "${ENV_FILE}"
    else
        printf '%s=%s\n' "$1" "$2" >> "${ENV_FILE}"
    fi
}

# 只认模板里的占位值，避免重复执行时把已生效的密码改掉
is_placeholder() {
    case "$(env_get "$1")" in
        "" | change-me* | change-this*) return 0 ;;
        *) return 1 ;;
    esac
}

# 数据卷是否已存在（卷名 = compose 项目名 + 卷名）
volume_exists() {
    docker volume ls --format '{{.Name}}' 2>/dev/null | grep -qx "$1"
}

port_in_use() {
    if command -v ss >/dev/null 2>&1; then
        ss -lntH 2>/dev/null | awk '{print $4}' | sed 's/.*://' | grep -qx "$1"
    elif command -v netstat >/dev/null 2>&1; then
        netstat -lnt 2>/dev/null | awk 'NR>2 {print $4}' | sed 's/.*://' | grep -qx "$1"
    else
        return 1
    fi
}

# ---------- 1. 生成 .env ----------
if [ -f "${ENV_FILE}" ]; then
    echo "[1/4] 已存在 .env：保留现有配置（不会覆盖已生效的密码）"
else
    echo "[1/4] 从 .env.example.docker-compose 生成 .env"
    cp "${TEMPLATE}" "${ENV_FILE}"
fi

# ---------- 可选：清空数据卷（--reset） ----------
if [ "${RESET_VOLUMES}" = "1" ]; then
    echo
    echo "[reset] 停止服务并清空数据卷（数据库 / Redis / 附件都会被删除）"
    # shellcheck disable=SC2086  # DC 需要按空格拆成两个词
    ${DC} down -v
    echo
fi

# ---------- 2. 随机密码 / 密钥 ----------
echo "[2/4] 生成随机密码 / 密钥"

# PostgreSQL 的密码只在「数据卷首次初始化」时写入，之后改 .env 不会同步到库里。
# 若卷已存在而密码仍是模板占位值，继续生成新密码只会让服务一直报
# password authentication failed (SQLSTATE 28P01)，所以这里先拦住。
PROJECT_NAME="$(awk '/^name:/{print $2; exit}' docker-compose.yml)"
PG_VOLUME="${PROJECT_NAME:-mini-kanban}_pgdata"
# 记录"启动前"卷是否存在：全新的卷意味着这次会重新初始化数据库，
# 后端也就会创建初始管理员；卷已存在则说明账号早就建过了。
# 用卷而不是查 users 表来判断，是因为后者存在竞态（后端建表与 seed 只需几毫秒，
# 查晚了就会把首次部署该显示的密码吞掉）。
PG_VOLUME_FRESH=1
if volume_exists "${PG_VOLUME}"; then
    PG_VOLUME_FRESH=0
    if is_placeholder POSTGRES_PASSWORD; then
        cat >&2 <<EOF
检测到数据卷 ${PG_VOLUME} 已存在，但 .env 里的 POSTGRES_PASSWORD 仍是模板占位值。

数据库密码只在数据卷首次初始化时写入，之后改 .env 不会同步到库里。
现在这个卷里存的是更早某次初始化写进去的密码，和 .env 对不上，
继续启动只会一直报 password authentication failed (SQLSTATE 28P01)。

请二选一：

  1) 保留数据 —— 把 .env 的 POSTGRES_PASSWORD 改回当初初始化时用的那个密码。
     如果你根本不知道那个密码是什么（例如 .env 是刚生成的、从没成功配置过），
     就说明这个卷里没有你要保留的东西，直接选第 2 条。

  2) 放弃数据 —— 清掉数据卷后重新运行，本脚本加 --reset 即可一步完成：

       bash deploy/deploy.sh --reset

     等效的手工命令是 ${DC} down -v
EOF
        exit 1
    fi
    echo "      注意：数据卷 ${PG_VOLUME} 已存在，POSTGRES_PASSWORD 沿用 .env 现有值（改动它会导致认证失败）"
fi

for item in "POSTGRES_PASSWORD:24:数据库密码" "REDIS_PASSWORD:24:Redis 密码" "JWT_SECRET:32:JWT 密钥"; do
    key="${item%%:*}"
    rest="${item#*:}"
    len="${rest%%:*}"
    label="${rest#*:}"
    if is_placeholder "${key}"; then
        env_set "${key}" "$(rand_hex "${len}")"
        echo "      ${key}（${label}）已生成"
    else
        echo "      ${key}（${label}）已存在，跳过"
    fi
done

# 初始管理员：邮箱 / 用户名可由命令行参数指定，密码始终随机生成
if [ -n "${ARG_ADMIN_EMAIL}" ]; then
    env_set SEED_ADMIN_EMAIL "${ARG_ADMIN_EMAIL}"
fi
if [ -n "${ARG_ADMIN_USER}" ]; then
    env_set SEED_ADMIN_USERNAME "${ARG_ADMIN_USER}"
fi
if [ -z "$(env_get SEED_ADMIN_EMAIL)" ]; then
    env_set SEED_ADMIN_EMAIL "admin@mail.example.com"
fi
if [ -z "$(env_get SEED_ADMIN_USERNAME)" ]; then
    env_set SEED_ADMIN_USERNAME "Admin"
fi

if is_placeholder SEED_ADMIN_PASSWORD; then
    env_set SEED_ADMIN_PASSWORD "$(rand_password)"
    echo "      SEED_ADMIN_PASSWORD（初始管理员密码）已生成"
else
    echo "      SEED_ADMIN_PASSWORD（初始管理员密码）已存在，跳过"
fi

# seed 只在用户表为空时生效，打开它不会影响已有数据；
# 后端创建成功后会把容器内 config.yaml 的 seed 段自动清空
env_set SEED_ENABLED "true"

# --reset-admin：让后端在本次启动时强制重置该管理员（无视用户表是否为空）
if [ "${RESET_ADMIN}" = "1" ]; then
    env_set SEED_FORCE_RESET "true"
    echo "      SEED_FORCE_RESET 已打开：启动时会强制重置 $(env_get SEED_ADMIN_EMAIL)"
fi

# ---------- 3. 宿主机端口 ----------
echo "[3/4] 检查宿主机端口"
PORT="$(env_get HTTP_PORT)"
PORT="${PORT:-8080}"

if port_in_use "${PORT}"; then
    OLD_PORT="${PORT}"
    NEW_PORT="$(bash "${SCRIPT_DIR}/pick-port.sh" --print)"
    env_set HTTP_PORT "${NEW_PORT}"
    PORT="${NEW_PORT}"
    echo "      ${OLD_PORT} 已被占用，改用 ${PORT}"
else
    echo "      ${PORT} 可用"
fi

# ---------- 4. 构建并启动 ----------
echo "[4/4] 构建并启动服务（首次构建需要几分钟）"
# shellcheck disable=SC2086  # DC 需要按空格拆成两个词
${DC} up -d --build

# --reset-admin 是一次性动作：容器已经带着 SEED_FORCE_RESET=true 启动了，
# 这里立刻把 .env 改回 false，避免以后再重建容器时又重置一次密码
if [ "${RESET_ADMIN}" = "1" ]; then
    env_set SEED_FORCE_RESET "false"
fi

# ---------- 结果 ----------
IP="$(hostname -I 2>/dev/null | awk '{print $1}' || true)"

echo
echo "=============================================================================="
echo " 部署完成，访问：http://${IP:-<服务器IP>}:${PORT}"
echo "=============================================================================="
echo
# shellcheck disable=SC2086
${DC} ps || true
echo
echo "常用命令："
echo "  查看后端日志   ${DC} logs -f server"
echo "  停止服务       ${DC} down"
echo "  全部选项       bash deploy/deploy.sh --help"
echo
if [ "${RESET_ADMIN}" = "1" ]; then
    echo "已按 --reset-admin 强制重置管理员账号（随刚才的启动执行）："
    echo "  登录邮箱：$(env_get SEED_ADMIN_EMAIL)"
    echo "  登录密码：$(env_get SEED_ADMIN_PASSWORD)"
    echo "  （用户名与 admin 角色一并确保；其他数据未受影响）"
    echo
elif [ "$(env_get SEED_ENABLED)" = "true" ]; then
    if [ "${PG_VOLUME_FRESH}" = "1" ]; then
        echo "初始管理员账号（本次是全新数据库，启动时会创建该账号）："
        echo "  登录邮箱：$(env_get SEED_ADMIN_EMAIL)"
        echo "  显示名称：$(env_get SEED_ADMIN_USERNAME)"
        echo "  登录密码：$(env_get SEED_ADMIN_PASSWORD)"
        echo "  （同样记录在 .env 的 SEED_ADMIN_* 中；后端首次启动的日志里也会打印创建的邮箱）"
    else
        echo "初始管理员账号：数据库卷已存在，后端只在「用户表为空」时创建，"
        echo "  所以本次不会重建账号、也不会改密码。"
        echo "  忘记密码：bash deploy/deploy.sh --reset-admin   （强制重置，不动其他数据）"
        echo "  密码变量：.env 里的 SEED_ADMIN_PASSWORD"
        echo "  全部选项：bash deploy/deploy.sh --help"
    fi
    echo
fi
echo
echo "注意：POSTGRES_PASSWORD 只在数据库卷首次初始化时生效。"
echo "      之后若要修改它，需清掉数据卷（会删除全部数据）再重新启动："
echo "        bash deploy/deploy.sh --reset     （等效于先执行 ${DC} down -v）"
