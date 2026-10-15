package utils

import (
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

// richTextPolicy 富文本描述的标签白名单：只放行编辑器能产出的标签与属性。
// 说明：
//   - 不调用 AllowStandardURLs，href 单独用 Matching 收紧到 http/https/mailto；
//   - img 的 src 只允许本站附件地址（/api/files/<随机名>），避免外链 / javascript: 注入；
//   - script / style / iframe / on* 事件属性等一律被丢弃。
var richTextPolicy = buildRichTextPolicy()

func buildRichTextPolicy() *bluemonday.Policy {
	p := bluemonday.NewPolicy()
	p.AllowElements(
		"p", "br", "hr", "div", "span",
		"strong", "b", "em", "i", "u", "s", "strike", "del", "mark",
		"h1", "h2", "h3", "h4",
		"ul", "ol", "li", "blockquote", "pre", "code",
	)
	// URL 策略：整体要求可解析，链接只允许 http/https/mailto，图片允许本站相对地址
	p.RequireParseableURLs(true)
	p.AllowURLSchemes("http", "https", "mailto")
	p.AllowRelativeURLs(true)

	p.AllowAttrs("href").Matching(regexp.MustCompile(`^(https?://|mailto:)[^\s"'<>]+$`)).OnElements("a")
	p.AllowAttrs("title").OnElements("a")
	p.RequireNoFollowOnLinks(true)
	p.RequireNoReferrerOnLinks(true)

	// 图片只接受本站上传的附件地址
	p.AllowAttrs("src").Matching(regexp.MustCompile(`^/api/files/[0-9a-f]{32}\.[a-z0-9]{1,8}$`)).OnElements("img")
	p.AllowAttrs("alt", "title", "width", "height").OnElements("img")

	return p
}

// SanitizeRichText 净化富文本描述，返回可安全直接渲染的 HTML。
func SanitizeRichText(html string) string {
	if strings.TrimSpace(html) == "" {
		return ""
	}
	return strings.TrimSpace(richTextPolicy.Sanitize(html))
}
