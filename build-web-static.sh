#!/usr/bin/env bash
#
# 前端静态导出（Next.js output: "export"），产物拷入 server/web-dist/，
# build.sh 会把整个目录拷进发布包；运行时由后端进程自己托管这些页面，
# 于是部署仍然是一个程序、一个端口，目标机不需要 Node。
# build.sh 会读这里的产物，所以本脚本必须在 build.sh 之前跑。
#
# 完整流程：
#   bash build-web-static.sh && bash build.sh        # 产物名带版本号，版本取自 git 标签
#   bash build-web-static.sh && bash build.sh -v v1.2.3
#
# 用法: ./build-web-static.sh

set -euo pipefail

cd "$(dirname "$0")"

api_base=${NEXT_PUBLIC_API_BASE:-/api}

echo "安装依赖"
(cd web && npm ci --no-audit --no-fund)

echo "静态导出（NEXT_PUBLIC_API_BASE=$api_base）"
(cd web && \
    NEXT_OUTPUT=export \
    NEXT_PUBLIC_STATIC=1 \
    NEXT_PUBLIC_API_BASE="$api_base" \
    npm run build)

if [ ! -f web/out/index.html ]; then
    echo "静态导出失败：web/out/index.html 不存在" >&2
    exit 1
fi

# 清空上次产物后拷入新的（目录里不该混入上次的残留文件）
mkdir -p server/web-dist
find server/web-dist -mindepth 1 -delete
cp -a web/out/. server/web-dist/

printf '前端已拷入 server/web-dist（%s），build.sh 会把它打进发布包\n' "$(du -sh server/web-dist | cut -f1)"
