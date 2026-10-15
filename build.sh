#!/usr/bin/env bash
#
# 交叉编译 mini-kanban 并打包为「开箱即用」的可分发归档，产物落在 dist/。
#
# 前端静态产物不打进二进制，而是作为 web-dist/ 目录随归档分发：可执行文件因此小一个
# 量级（六个平台交叉编译时也不用把静态资源反复写进每个产物），运行时仍由同一个进程、
# 同一个端口托管这些页面 —— 目标机不需要 Node、不需要额外的静态目录，
# 也不需要 nginx 与 CORS。
#
# 归档名带版本号（不传 -v 时取最近的 git 标签），一眼就能认出下载的是哪个版本：
#
#   mini-kanban-v1.0.0-windows-amd64.zip      Windows
#   mini-kanban-v1.0.0-linux-amd64.tar.gz      Linux / macOS
#   ...
#
# 归档内的顶层目录与归档同名，解压后直接可用：
#
#   mini-kanban-v1.0.0-linux-amd64/
#   ├── mini-kanban(.exe)     可执行文件（非 Windows 平台带 0755 权限）
#   ├── web-dist/              前端静态产物（页面与接口同源提供）
#   ├── start.sh / start.bat  启动脚本（Windows 为 start.bat）
#   ├── version.txt           本包的版本号，启动脚本会打印它
#   ├── config.example.yaml   配置模板：复制为 config.yaml 后按环境修改
#   ├── README.md / README.en.md
#   └── LICENSE
#
# web-dist/ 必须与可执行文件同目录（也可以用 WEB_DIST_DIR 指向别处）；缺了它服务仍能
# 启动，但只提供接口、非 /api 路径返回 404。
#
# 归档里刻意不带 config.yaml：首次启动会自动进入初始化向导
# （浏览器打开 http://localhost:8281/setup），由向导生成配置。
#
# 另有 checksums.txt 记录全部归档的 SHA256（LF 换行，Linux 下可直接 sha256sum -c）。
#
# 用法: ./build.sh [-o OutDir] [-v Version]

set -euo pipefail

usage() {
    cat <<'EOF'
交叉编译 mini-kanban 并打包为开箱即用的可分发归档（前端以 web-dist/ 目录随包分发）。

用法: ./build.sh [-o OutDir] [-v Version]

选项:
  -o OutDir   输出目录（默认: dist，相对本脚本所在目录）
  -v Version  版本号，写进归档名与包内 version.txt
              （默认取最近的 git 标签，形如 v1.0.0；无标签时用 dev）
  -h          显示本帮助

平台矩阵: windows/amd64 windows/arm64 linux/amd64 linux/arm64 darwin/amd64 darwin/arm64
每个平台一个归档（Windows 为 .zip，其余为 .tar.gz），外加 checksums.txt。

注意: 必须先构建前端静态产物（build-web-static.sh / build-web-static.bat），
      产物落在 server/web-dist，本脚本会把它拷进每个归档；该目录为空时直接报错退出。
EOF
}

out_dir=dist
version=""
while getopts ':ho:v:' opt; do
    case "$opt" in
        h) usage; exit 0 ;;
        o) out_dir=$OPTARG ;;
        v) version=$OPTARG ;;
        '?') usage >&2; exit 2 ;;
    esac
done

cd "$(dirname "$0")"

if [ -z "$out_dir" ]; then
    echo "OutDir 不能为空" >&2
    exit 2
fi
case "$out_dir" in
    /*) ;;
    *) out_dir=$PWD/$out_dir ;;
esac

# 版本号：-v 优先，否则取最近的 git 标签；都没有就用 dev（源码归档等场景）。
if [ -z "$version" ]; then
    version=$(git describe --tags --abbrev=0 2>/dev/null || true)
fi
version=${version#v}
if [ -z "$version" ]; then
    echo "提示: 没能从 git 标签取到版本号，按 dev 打包（发版请显式 -v vX.Y.Z）" >&2
    version=dev
fi
# 版本号会拼进文件名与包内路径，只放行安全字符，避免拼出目录穿越或带空格的名字
case "$version" in
    *[!A-Za-z0-9.+-]*)
        echo "版本号含非法字符: $version" >&2
        exit 2
        ;;
esac
version="v$version"

# 前端静态产物要作为 web-dist/ 目录打进包里。这里不做兜底构建、也不放行：
# 少了它就发出去的包只有接口没有页面，正是要避免的分开发布。
if [ ! -f server/web-dist/index.html ]; then
    cat >&2 <<'EOF'
[错误] server/web-dist/index.html 不存在：归档必须带上 web-dist/ 页面目录。
       请先构建前端静态产物，再重新执行本脚本：
         bash build-web-static.sh          # Windows: build-web-static.bat
EOF
    exit 1
fi

# 目标矩阵，'os/arch' 形式；增删一行即可调整发布平台。
targets='windows/amd64 windows/arm64 linux/amd64 linux/arm64 darwin/amd64 darwin/arm64'

# 交叉编译不需要 cgo（纯 Go 依赖，含 PostgreSQL / Redis 驱动）
export CGO_ENABLED=0

echo "版本: $version"
mkdir -p "$out_dir"
# 只删除输出目录顶层的文件，不递归，避免误伤目录外的东西
find "$out_dir" -maxdepth 1 -type f -delete

staging=$(mktemp -d)
trap 'rm -rf "$staging"' EXIT

# Windows 归档用 zip。Git for Windows 的 bash 默认不带 zip 命令，
# 所以没有 zip 时退回 Python 的 zipfile，打出来的结构完全一致，
# 让 Windows 用户在本地也能直接 bash build.sh。
make_zip() {
    local archive=$1 top=$2
    if command -v zip >/dev/null 2>&1; then
        (cd "$staging" && zip -qr "$archive" "$top")
        return
    fi
    # 判据是「能不能真的跑起来」，不能只看 command -v：Windows 上 python3 常是
    # 微软商店的应用执行别名（stub），存在但一执行就失败。
    local py=""
    for cand in python3 python py; do
        if command -v "$cand" >/dev/null 2>&1 && "$cand" -c 'import zipfile' >/dev/null 2>&1; then
            py=$cand
            break
        fi
    done
    if [ -z "$py" ]; then
        echo "[错误] 打 Windows 归档需要 zip 或可用的 python，当前环境两者都没有" >&2
        exit 1
    fi
    echo "（未找到 zip，改用 $py 打包）"
    (cd "$staging" && "$py" - "$archive" "$top" <<'PY'
import os
import sys
import zipfile

archive, top = sys.argv[1], sys.argv[2]
with zipfile.ZipFile(archive, "w", zipfile.ZIP_DEFLATED) as zf:
    for root, dirs, files in os.walk(top):
        dirs.sort()
        for name in sorted(files):
            path = os.path.join(root, name)
            zf.write(path, path.replace(os.sep, "/"))
PY
    )
}

count=0
for target in $targets; do
    goos=${target%%/*}
    goarch=${target#*/}

    # 归档内的可执行文件用短名（启动脚本按短名找它），版本信息由 version.txt 承载
    bin_name=mini-kanban
    if [ "$goos" = windows ]; then
        bin_name=mini-kanban.exe
    fi
    # 归档名与顶层目录名都带版本号，避免多个版本的包混在一起时分不清
    pkg_name="mini-kanban-$version-$goos-$goarch"
    if [ "$goos" = windows ]; then
        archive_name="$pkg_name.zip"
    else
        archive_name="$pkg_name.tar.gz"
    fi

    pkg_dir="$staging/$pkg_name"
    rm -rf "$pkg_dir"
    mkdir -p "$pkg_dir"

    # Go 模块在 server/ 下，构建要切到模块根目录
    (cd server && GOOS="$goos" GOARCH="$goarch" go build -trimpath -ldflags '-s -w' -o "$pkg_dir/$bin_name" .)
    cp server/config.example.yaml LICENSE README.md README.en.md "$pkg_dir/"
    printf '%s\n' "$version" > "$pkg_dir/version.txt"

    # 前端静态产物随包分发：运行时要与可执行文件同目录（或用 WEB_DIST_DIR 指过去），
    # 这样可执行文件里不含任何静态资源，只有业务代码。
    cp -a server/web-dist "$pkg_dir/web-dist"
    # 构建中间文件不该出现在发布包里（.gitkeep 只为占位，删掉更干净）
    rm -f "$pkg_dir/web-dist/.gitkeep"

    # 启动脚本：脚本自身会切到所在目录，因此可在任意位置调用
    if [ "$goos" = windows ]; then
        cp tools/release/start.bat "$pkg_dir/start.bat"
    else
        cp tools/release/start.sh "$pkg_dir/start.sh"
    fi

    if [ "$goos" = windows ]; then
        # zip 不保存可执行位，Windows 也不需要
        make_zip "$out_dir/$archive_name" "$pkg_name"
    else
        chmod 0755 "$pkg_dir/$bin_name" "$pkg_dir/start.sh"
        # 固定 mtime，保证同一份源码重复构建产物一致
        find "$pkg_dir" -exec touch -t 202601010000.00 {} +
        (cd "$staging" && tar -czf "$out_dir/$archive_name" "$pkg_name")
    fi

    size=$(du -h "$out_dir/$archive_name" | cut -f1)
    printf '  build %-16s -> %s  (%s)\n' "$target" "$archive_name" "$size"
    count=$((count + 1))
done

# 归档的 SHA256；只统计要发布的归档，LF 换行以便 sha256sum -c
hash_of() {
    if command -v sha256sum >/dev/null 2>&1; then
        sha256sum "$1" | cut -d' ' -f1
    else
        shasum -a 256 "$1" | cut -d' ' -f1   # macOS
    fi
}
find "$out_dir" -maxdepth 1 -type f \( -name '*.zip' -o -name '*.tar.gz' \) | LC_ALL=C sort | while IFS= read -r file; do
    printf '%s  %s\n' "$(hash_of "$file")" "$(basename "$file")"
done > "$out_dir/checksums.txt"

echo
echo "$count 个归档 + checksums.txt -> $out_dir"