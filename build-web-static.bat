@echo off
setlocal EnableExtensions
REM ============================================================
REM mini-kanban 前端静态导出（build-web-static.sh 的 Windows 版）
REM
REM 做两件事：把前端按 output: "export" 构建到 web\out，再把产物拷进
REM server\web-dist —— 之后 go build 会用 go:embed 把它内嵌进
REM mini-kanban.exe，一个二进制同时提供接口与页面，目标机不需要 Node。
REM
REM 本脚本必须在 go build 之前跑，否则二进制里没有前端。
REM
REM 用法（双击或命令行都行）：
REM   build-web-static.bat                 完整流程：npm ci + 静态导出 + 拷入 web-dist
REM   build-web-static.bat --skip-install  跳过 npm ci，直接用现有的 node_modules
REM
REM 接口前缀由环境变量 NEXT_PUBLIC_API_BASE 决定（默认 /api，即与后端同源），
REM 换地址必须重新构建。
REM ============================================================
REM ENCODING: this file is GBK (cp936), NOT UTF-8. Keep it that way --
REM cmd.exe mis-tracks line offsets in a UTF-8 batch file that contains
REM multi-byte characters, and starts executing comments as commands.
REM ============================================================

cd /d "%~dp0"

set "SKIP_INSTALL="
if /i "%~1"=="--skip-install" set "SKIP_INSTALL=1"
if not defined NEXT_PUBLIC_API_BASE set "NEXT_PUBLIC_API_BASE=/api"

echo.
echo   ==========================================
echo     前端静态导出 -^> server\web-dist
echo     目录: %CD%
echo     接口前缀: %NEXT_PUBLIC_API_BASE%
echo   ==========================================
echo.

where npm >nul 2>&1
if errorlevel 1 (
    echo [错误] 找不到 npm。请先安装 Node.js 18.17 或更高版本，并确保它在 PATH 里。
    goto FAILED
)

if not exist "web\package.json" (
    echo [错误] 找不到 web\package.json，请把本脚本放在仓库根目录运行。
    goto FAILED
)

if defined SKIP_INSTALL goto BUILD

echo 安装依赖
pushd web
call npm ci --no-audit --no-fund
set "EXITCODE=%ERRORLEVEL%"
popd
if not "%EXITCODE%"=="0" (
    echo [错误] npm ci 失败，退出码 %EXITCODE%
    goto FAILED
)

:BUILD
echo 静态导出（NEXT_OUTPUT=export NEXT_PUBLIC_STATIC=1）
pushd web
REM NEXT_PUBLIC_STATIC=1 时 layout 不再读 cookie（静态产物没有请求上下文），
REM 少了这个变量 next build 会因为 output: export 下用了 cookies() 直接失败。
set "NEXT_OUTPUT=export"
set "NEXT_PUBLIC_STATIC=1"
call npm run build
set "EXITCODE=%ERRORLEVEL%"
popd
if not "%EXITCODE%"=="0" (
    echo [错误] 前端构建失败，退出码 %EXITCODE%
    goto FAILED
)

if not exist "web\out\index.html" (
    echo [错误] 静态导出失败：web\out\index.html 不存在。
    goto FAILED
)

REM 清空上次产物但保留 .gitkeep（没有它、又没跑过本脚本时 go:embed 会因目录为空而编译失败）。
REM 这里用 dir 列名字而不是 for 通配：通配没有匹配时会退化成字面量 "*"，反而会报错。
if not exist "server\web-dist" mkdir "server\web-dist"
for /f "delims=" %%D in ('dir /b /ad "server\web-dist" 2^>nul') do rd /s /q "server\web-dist\%%D"
for /f "delims=" %%F in ('dir /b /a-d "server\web-dist" 2^>nul') do if /i not "%%F"==".gitkeep" del /q "server\web-dist\%%F"

echo 拷贝 web\out -^> server\web-dist
robocopy "web\out" "server\web-dist" /E /NFL /NDL /NJH /NJS /NP >nul
REM robocopy 的退出码 0~7 都算成功（1 = 有文件被复制），8 以上才是失败
if errorlevel 8 (
    echo [错误] 拷贝失败，robocopy 退出码 %ERRORLEVEL%
    goto FAILED
)

echo.
echo   完成：静态产物已拷入 server\web-dist
echo.
echo   下一步（内嵌进二进制）：
echo       cd server
echo       go build -o mini-kanban.exe .
echo.
echo   运行 mini-kanban.exe 后，页面与接口同在 http://localhost:8281
echo.
pause
endlocal
exit /b 0

:FAILED
echo.
pause
endlocal
exit /b 1
