// Package i18n 接口错误信息 / 提示文案的多语言支持。
//
// 设计：源码里写死的中文串本身就是「中文文案」，因此英文目录直接以中文原文为键
// （见 messages.go），调用方无需改动 —— 只要经过 fail / badRequest / okMessage
// 这几个出口，返回给前端的文案就会按请求语言切换。
//
// 语言协商（middleware.Language）：
//  1. Accept-Language 头：前端 fetch 会带上当前界面语言，优先级最高；
//  2. mini-kanban-lang cookie：附件下载（?ticket=）由浏览器直接发起、带不了自定义头，靠它兜底；
//  3. 都没有：默认中文。
//
// 注意：操作日志的 summary / 变更项是「写入时落库的文本」，不在这里翻译，
// 否则同一条日志会随着查看者语言变化，详见 README。
package i18n

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// Lang 界面语言，与前端 web/lib/i18n/index.ts 的 Lang 一一对应。
type Lang string

const (
	// ZhCN 简体中文（默认语言，也是源码文案的语言）
	ZhCN Lang = "zh-CN"
	// EN English
	EN Lang = "en"
)

// DefaultLang 协商不出结果时的兜底语言。
const DefaultLang = ZhCN

// LangCookie 前端写入的语言 cookie 名，与前端 LANG_COOKIE 保持一致。
const LangCookie = "mini-kanban-lang"

// CtxLang 语言在 gin.Context 里的键。
const CtxLang = "lang"

// LangOf 解析单个语言标签（Accept-Language 的一项或 cookie 值）。
// 只认主子标签：zh / zh-CN / zh-Hans → 中文，en / en-GB → 英文，其余返回 ""。
func LangOf(value string) Lang {
	v := strings.TrimSpace(value)
	if v == "" {
		return ""
	}
	// 去掉 q 值（;q=0.9）与可能的空白
	if i := strings.IndexByte(v, ';'); i >= 0 {
		v = strings.TrimSpace(v[:i])
	}
	primary := strings.ToLower(strings.SplitN(v, "-", 2)[0])
	switch primary {
	case "zh":
		return ZhCN
	case "en":
		return EN
	}
	return ""
}

// LangOfHeader 解析 Accept-Language 头（可带多个语言与 q 值），取第一个受支持的。
func LangOfHeader(header string) Lang {
	if strings.TrimSpace(header) == "" {
		return ""
	}
	for _, part := range strings.Split(header, ",") {
		if l := LangOf(part); l != "" {
			return l
		}
	}
	return ""
}

// FromContext 取本次请求协商出的语言；中间件没跑过（如向导服务）时按默认中文。
func FromContext(c *gin.Context) Lang {
	if v, ok := c.Get(CtxLang); ok {
		if l, ok := v.(Lang); ok && l != "" {
			return l
		}
	}
	return DefaultLang
}

// Middleware 语言协商中间件：Accept-Language → cookie → 默认中文。
// 挂在最前面，保证后续处理器 / 鉴权中间件都能拿到语言。
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := LangOfHeader(c.GetHeader("Accept-Language"))
		if lang == "" {
			// 附件下载（?ticket=）由浏览器直接发起、带不上自定义头，只能靠 cookie
			if v, err := c.Cookie(LangCookie); err == nil {
				lang = LangOf(v)
			}
		}
		if lang == "" {
			lang = DefaultLang
		}
		c.Set(CtxLang, lang)
		c.Next()
	}
}

// Msg 把中文原文翻译成目标语言。
// 中文、以及未收录的原文（例如数据库返回的英文错误）一律原样返回，
// 因此新增文案忘记翻译时表现为「还是中文」，不会出现空串或 key 名。
func Msg(lang Lang, msg string) string {
	return translate(lang, msg)
}
