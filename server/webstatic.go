package main

import (
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"mini-kanban/internal/i18n"
)

// web-dist 是前端静态导出产物（build-web-static.sh 从 web/out 拷来，打包时随归档分发）。
//
// 刻意不再用 go:embed 把它编进二进制：那会让每个平台的可执行文件凭空多出几 MB，
// 六平台交叉编译时这部分要被反复写进每一个产物。改成随包分发后，可执行文件只剩
// 业务代码（体积降一个量级），页面文件也能被网关 / CDN 直接读盘缓存。
//
// 代价是 web-dist 必须与可执行文件放在一起：默认找「工作目录下的 web-dist/」，
// 换个位置可以用 WEB_DIST_DIR 指定。目录缺失或里面没有 index.html 时，服务自动
// 降级为「只提供接口」——非 /api 路径返回 404，接口行为完全不受影响。
const webDistDirName = "web-dist"

// 启动时由 ResolveWebDist 解析一次并缓存，避免每个请求都去 stat 磁盘：
// webRoot 静态根目录的绝对路径，webEnabled 表示里面确有可用的 index.html。
var (
	webRoot    string
	webEnabled bool
)

// ResolveWebDist 解析 web-dist 目录并探测是否可用，结果供 serveWebApp 与 HasWebApp 使用。
// 探测不到时不报错退出：页面缺失不该让接口一起不可用，只是启动日志里说清楚降级了。
func ResolveWebDist() {
	dir := strings.TrimSpace(os.Getenv("WEB_DIST_DIR"))
	if dir == "" {
		dir = webDistDirName
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		log.Printf("web-dist 目录 %q 无法解析（前端将不可用，仅提供接口）：%v", dir, err)
		return
	}
	webRoot = abs

	if st, err := os.Stat(webPath("index.html")); err != nil || st.IsDir() {
		log.Printf("未找到 %s：非 /api 路径返回 404，服务仅提供接口。"+
			"（发布包里 web-dist/ 与可执行文件同目录；不在当前位置时用 WEB_DIST_DIR 指定）",
			webPath("index.html"))
		return
	}
	webEnabled = true
}

// HasWebApp 报告 web-dist 目录里是否有可用的前端。
// 它同时决定 /setup 完成页上是否给出首页入口（见 setup.DoneHandler）。
func HasWebApp() bool {
	return webEnabled
}

// webPath 把 URL 路径拼成磁盘路径（URL 用 "/" 分隔，磁盘用分隔符）。
func webPath(name string) string {
	return filepath.Join(webRoot, filepath.FromSlash(path.Clean("/"+name)))
}

// webFileExists 判断 web-dist 里是否有该路径（目录也算存在）。
func webFileExists(name string) bool {
	_, err := os.Stat(webPath(name))
	return err == nil
}

// boardPlaceholderID 是前端静态导出里看板动态段（app/projects/[id]）的占位 id，
// 必须与 web/app/projects/[id]/page.tsx 的 generateStaticParams() 返回值保持一致。
//
// 为什么要有占位：Next 的静态导出要求动态段至少能预渲染出一条路径 —— 它用
// 「预渲染路径数 > 0」判断页面有没有提供 generateStaticParams，返回空数组会被判定为缺失、
// 直接构建失败。而看板页对任意 id 渲染的内容完全一样（真实 id 由前端从地址栏解析），
// 所以预渲染 projects/_/ 一份就够，真实 id 的深链由 isBoardDeepLink 映射过去。
const boardPlaceholderID = "_"

// boardPlaceholderPath 占位页在静态产物里的目录（output: "export" + trailingSlash 时产物是 projects/_/index.html）。
// 结尾的斜杠是必需的：交给 FileServer 去补 index.html。若直指 .../index.html，
// net/http 的 serveFile 会把它 301 到 "./"，地址栏里的真实 id 就被换成占位 id 了。
const boardPlaceholderPath = "projects/" + boardPlaceholderID + "/"

// boardPlaceholderPayload 同一份占位页的 RSC 载荷。
//
// 站内跳转（点击项目卡）时 Next 取的不是 HTML 而是这个载荷：静态导出把载荷写成
// <路由>/index.txt，并且只认 text/plain（见 next/dist/client/components/router-reducer/
// fetch-server-response.js），拿不到就退化成整页刷新。这里同样映射到占位那份 —— 载荷里的
// 路由树带的仍是占位参数，真实 id 一样由前端从地址栏解析。
const boardPlaceholderPayload = "projects/" + boardPlaceholderID + "/index.txt"

// isBoardDeepLink 判断 p 是需要映射到占位页的看板深链，即「单个 id 段的 projects/<id>」。
// 形如 projects/<id>/notes 的多段路径不在此列（当前没有这种路由，有的话再按需放开）。
func isBoardDeepLink(p string) bool {
	rest, ok := strings.CutPrefix(p, "projects/")
	if !ok || rest == "" || rest == boardPlaceholderID {
		return false
	}
	return !strings.Contains(rest, "/")
}

// isBoardPayload 判断 p 是看板深链的 RSC 载荷请求：projects/<id>/index.txt。
// 用 CutSuffix 精确切出 id 段（而不是 TrimSuffix），免得把静态路由自己的载荷
// projects/index.txt 也当成深链载荷。
func isBoardPayload(p string) bool {
	rest, ok := strings.CutPrefix(p, "projects/")
	if !ok {
		return false
	}
	id, ok := strings.CutSuffix(rest, "/index.txt")
	if !ok || id == "" || id == boardPlaceholderID {
		return false
	}
	return !strings.Contains(id, "/")
}

// serveWebApp 托管前端静态站点，挂在 r.NoRoute 上：
// /api、/healthz、/setup 由已有路由优先匹配，不会走到这里。
//
// 解析顺序：
//  1. /api/ 前缀 —— 接口不存在，返回 JSON 404（绝不回退成页面，否则前端会拿到 HTML 当 JSON 解析）
//  2. web-dist 缺失 —— 返回 JSON 404 并说明原因，接口不受影响
//  3. 能命中具体文件（/_next/static/... 等）、目录（补 / 后取它的 index.html）或 xxx.html 直接产物 —— 直出
//  4. 看板深链 /projects/<id> —— 直出预渲染的占位页 projects/_/（兜底页是首页，会把地址栏的 id 丢掉）
//  5. 看板深链的 RSC 载荷 /projects/<id>/index.txt —— 同样给占位页那份，站内跳转才不会退化成整页刷新
//  6. 其余路径 —— 兜底产物根目录的 index.html，交给前端路由处理（地址栏不变）
func serveWebApp(c *gin.Context) {
	if strings.HasPrefix(c.Request.URL.Path, "/api/") {
		c.AbortWithStatusJSON(http.StatusNotFound,
			gin.H{"error": i18n.Msg(i18n.FromContext(c), "接口不存在")})
		return
	}
	if !webEnabled {
		c.AbortWithStatusJSON(http.StatusNotFound,
			gin.H{"error": i18n.Msg(i18n.FromContext(c), "前端静态资源未构建")})
		return
	}

	// 交给 FileServer 的路径一律用「目录形式」（以 / 结尾），由它去取目录里的 index.html。
	// 不能填成 .../index.html：net/http 的 serveFile 会把这种路径 301 到 "./"，
	// 而浏览器对 "./" 的解析又落回同一个地址，形成无限重定向（整站除静态文件外全部打不开）。
	p := strings.TrimPrefix(path.Clean("/"+c.Request.URL.Path), "/")
	switch {
	case p == "":
		// 根路径：FileServer 直接交出产物根目录的 index.html
	case webFileExists(p):
		if info, err := os.Stat(webPath(p)); err == nil && info.IsDir() {
			p += "/"
		}
	case !strings.HasSuffix(p, "/") && webFileExists(p+".html"):
		// 没有 trailingSlash 的产物：login.html 这类直接产物
		p += ".html"
	case isBoardDeepLink(p):
		// 看板深链：静态产物里只有占位页，真实 id 由前端从地址栏解析
		// （见 web/app/projects/[id]/ProjectBoard.tsx 的 resolveProjectId）
		p = boardPlaceholderPath
	case isBoardPayload(p):
		p = boardPlaceholderPayload
	default:
		// 其余前端路由：兜底到产物的 index.html，交给前端路由处理
		p = ""
	}

	// Next 的静态资源文件名带内容哈希，可以放心长期缓存；页面与其他产物一律回源校验，
	// 否则换版本后浏览器会拿着旧 index.html 去请求已经不存在的 chunk。
	if strings.HasPrefix(p, "_next/static/") {
		c.Header("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		c.Header("Cache-Control", "no-cache")
	}

	c.Request.URL.Path = "/" + p
	http.FileServer(http.Dir(webRoot)).ServeHTTP(c.Writer, c.Request)
}

// webAppDesc 启动日志里的一句话说明：找到 web-dist 说明页面同源提供，否则说明只有接口。
func webAppDesc() string {
	if webEnabled {
		return webRoot + "（页面同源提供）"
	}
	return "未找到（仅提供接口）"
}
