@echo off
setlocal EnableExtensions
REM ============================================================
REM mini-kanban 后端启动脚本（Windows）
REM
REM 双击即可运行：脚本会先切到自身所在目录，附件目录（storage.dir 默认
REM data/uploads）、config.yaml 与前端页面目录（web-dist，可用 WEB_DIST_DIR 改）
REM 都按该目录解析。
REM
REM 首次运行目录下没有 config.yaml，程序会自动进入初始化向导，
REM 按提示在浏览器打开 http://localhost:8281/setup 填写数据库 / Redis 即可；
REM 想重新配置，先设置环境变量 SETUP_FORCE=1 再运行本脚本。
REM
REM 服务异常退出时窗口会停住并列出常见原因，不会一闪而过。
REM ============================================================

cd /d "%~dp0"
title mini-kanban

REM 本文件是 UTF-8 编码，而中文 Windows 的控制台默认代码页是 936（GBK），
REM 不切换的话下面这些中文提示会显示成乱码。
chcp 65001 >nul 2>&1

echo.
REM 版本号由打包脚本写入（见 build.sh 的 -v 选项），直接跑源码时显示 unknown
set "VERSION=unknown"
if exist "version.txt" set /p VERSION=<version.txt

echo.
echo   ==========================================
echo     mini-kanban
echo     版本: %VERSION%
echo     目录: %CD%
echo   ==========================================
echo.

if not exist "mini-kanban.exe" (
    echo [错误] 当前目录找不到 mini-kanban.exe
    echo        请确保本脚本与可执行文件放在同一个目录下。
    echo.
    pause
    exit /b 1
)

REM 页面目录：缺失时服务仍能启动（接口可用），但网页打不开，这里提前说清楚
set "WEB_DIST=.\web-dist"
if not exist "%WEB_DIST%\index.html" (
    if defined WEB_DIST_DIR set "WEB_DIST=%WEB_DIST_DIR%"
    if not exist "%WEB_DIST%\index.html" set "WEB_DIST="
)

echo 正在启动，停止服务请按 Ctrl+C。
if defined WEB_DIST (
    echo 页面入口: http://localhost:8281      （页面由本进程从 %WEB_DIST% 提供）
) else (
    echo 页面入口: http://localhost:8281
    echo.
    echo   [警告] 没找到 web-dist 目录，网页将无法打开（接口不受影响）。
    echo          请确认压缩包完整解压：web-dist\ 必须与本程序在同一目录下，
    echo          或用 WEB_DIST_DIR 指定它的位置。
    echo.
)
if not exist "config.yaml" (
    echo 未检测到 config.yaml，将进入初始化向导：请在浏览器打开
    echo     http://localhost:8281/setup
) else (
    echo 配置文件: %CD%\config.yaml
    echo 健康检查: http://localhost:8281/healthz
)
echo.

mini-kanban.exe
set "EXITCODE=%ERRORLEVEL%"
if errorlevel 1 goto FAILED
goto END

:FAILED
echo.
echo   ======== 服务已退出，退出码 %EXITCODE% ========
echo   常见原因：
echo     1) 数据库连不上：config.yaml 里 database 段的地址 / 库名 / 账号密码不对，
echo        或数据库还没创建（服务只建表，不建库）
echo     2) Redis 连不上：不需要缓存就把 config.yaml 里 redis.host 留空
echo     3) 8281 端口被占用
echo     4) config.yaml 有必填项缺失或格式错误
echo.
echo   配置文件: %CD%\config.yaml
echo   想重新配置: 设置 SETUP_FORCE=1 后重新运行本脚本，会再次进入向导

:END
echo.
pause
endlocal
