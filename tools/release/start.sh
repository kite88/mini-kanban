#!/usr/bin/env bash
#
# 启动 mini-kanban 后端。
#
# 脚本会切到自身所在目录再启动，因此在任意位置调用都可以；
# 附件目录（storage.dir 默认 data/uploads）、config.yaml 与前端页面目录
# （web-dist，可用 WEB_DIST_DIR 改）都按该目录解析。
#
# 首次运行目录下没有 config.yaml，程序会自动进入初始化向导，
# 按提示在浏览器打开 http://localhost:8281/setup 填写数据库 / Redis 即可；
# 想重新配置，用 SETUP_FORCE=1 ./start.sh 再跑一次。
#
# 服务异常退出时会打印退出码与常见原因。

set -euo pipefail

cd "$(dirname "$0")"

if [ ! -x ./mini-kanban ]; then
    echo "[错误] 当前目录找不到可执行文件 mini-kanban" >&2
    echo "       请确保本脚本与可执行文件放在同一个目录下。" >&2
    exit 1
fi

# 版本号由打包脚本写入（见 build.sh 的 -v 选项），直接跑源码时显示 unknown
version=unknown
if [ -f ./version.txt ]; then
    version=$(head -n 1 ./version.txt)
fi

# 页面目录：缺失时服务仍能启动（接口可用），但网页打不开，这里提前说清楚
web_dist=./web-dist
if [ ! -f "$web_dist/index.html" ]; then
    if [ -n "${WEB_DIST_DIR:-}" ]; then
        web_dist=$WEB_DIST_DIR
    fi
    if [ ! -f "$web_dist/index.html" ]; then
        web_dist=""
    fi
fi

echo
echo "  =========================================="
echo "    mini-kanban"
echo "    版本: $version"
echo "    目录: $PWD"
echo "  =========================================="
echo

echo "正在启动，停止服务请按 Ctrl+C。"
if [ -n "$web_dist" ]; then
    echo "页面入口: http://localhost:8281      （页面由本进程从 $web_dist 提供）"
else
    echo "页面入口: http://localhost:8281"
    echo
    echo "  [警告] 没找到 web-dist 目录，网页将无法打开（接口不受影响）。"
    echo "         请确认压缩包完整解压：web-dist/ 必须与本程序在同一目录下，"
    echo "         或用 WEB_DIST_DIR 指定它的位置。"
    echo
fi
if [ ! -f config.yaml ]; then
    echo "未检测到 config.yaml，将进入初始化向导：请在浏览器打开"
    echo "    http://localhost:8281/setup"
else
    echo "配置文件: $PWD/config.yaml"
    echo "健康检查: http://localhost:8281/healthz"
fi
echo

set +e
./mini-kanban
code=$?
set -e

if [ "$code" -ne 0 ]; then
    cat >&2 <<EOF

  ======== 服务已退出，退出码 $code ========
  常见原因：
    1) 数据库连不上：config.yaml 里 database 段的地址 / 库名 / 账号密码不对，
       或数据库还没创建（服务只建表，不建库）
    2) Redis 连不上：不需要缓存就把 config.yaml 里 redis.host 留空
    3) 8281 端口被占用
    4) config.yaml 有必填项缺失或格式错误

  配置文件: $PWD/config.yaml
  想重新配置: SETUP_FORCE=1 ./start.sh 会再次进入向导
EOF
    exit "$code"
fi
