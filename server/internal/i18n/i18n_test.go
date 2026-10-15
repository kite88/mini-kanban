package i18n

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func TestTranslateExact(t *testing.T) {
	cases := []struct{ in, want string }{
		{"邮箱或密码错误", "Incorrect email or password"},
		{"参数错误", "Invalid request"},
		{"已彻底删除", "Permanently deleted"},
		{"登录已失效", "Your session has expired"},
	}
	for _, c := range cases {
		if got := translate(EN, c.in); got != c.want {
			t.Errorf("translate(en, %q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestTranslatePrefixAndDynamic(t *testing.T) {
	cases := []struct{ in, want string }{
		// 前缀 + 动态错误详情
		{"保存失败: pq: duplicate key", "Failed to save: pq: duplicate key"},
		{"恢复失败：dial tcp: connection refused", "Restore failed: dial tcp: connection refused"},
		{"彻底删除失败：boom", "Purge failed: boom"},
		// 前缀 + 数量 + 尾缀
		{"一次最多上传 10 个文件", "Up to 10 files at a time"},
		{"该部门下还有 3 个用户，请先调整归属后再删除", "This department still has 3 users; reassign them before deleting"},
		// 前缀 + 「动态内容」 + 尾缀
		{"附件「报表.png」不存在", `Attachment "报表.png" does not exist`},
		// 「动态内容」 + 尾缀（无前缀）
		{"「a.zip」超过大小上限", `"a.zip" exceeds the size limit`},
		// storage 的错误串
		{"不支持的文件类型 .exe", "Unsupported file type .exe"},
		{"无法识别文件类型：x", "Unable to determine the file type: x"},
		{"文件超过大小上限 10 MB", "The file exceeds the limit of 10 MB"},
	}
	for _, c := range cases {
		if got := translate(EN, c.in); got != c.want {
			t.Errorf("translate(en, %q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestTranslateZhAndUnknownPassthrough(t *testing.T) {
	// 中文语言原样返回
	if got := translate(ZhCN, "邮箱或密码错误"); got != "邮箱或密码错误" {
		t.Errorf("zh 应该原样返回，得到 %q", got)
	}
	// 未收录的原文原样返回（例如数据库返回的英文错误）
	if got := translate(EN, "pq: relation \"tasks\" does not exist"); got != `pq: relation "tasks" does not exist` {
		t.Errorf("未收录的原文应该原样返回，得到 %q", got)
	}
	if got := translate(EN, ""); got != "" {
		t.Errorf("空串应该原样返回，得到 %q", got)
	}
}

func TestLangOf(t *testing.T) {
	cases := []struct{ in, want string }{
		{"zh-CN", "zh-CN"},
		{"zh", "zh-CN"},
		{"zh-Hans-CN;q=0.9", "zh-CN"},
		{"en", "en"},
		{"en-US,en;q=0.9", "en"},
		{"en-GB", "en"},
		{"ja", ""},
		{"", ""},
	}
	for _, c := range cases {
		if got := string(LangOf(c.in)); got != c.want {
			t.Errorf("LangOf(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestLangOfHeader(t *testing.T) {
	if got := LangOfHeader("en-US,en;q=0.9,zh-CN;q=0.8"); got != EN {
		t.Errorf("按 Accept-Language 的顺序取第一个受支持项，得到 %q", got)
	}
	if got := LangOfHeader("fr-FR,ja;q=0.9"); got != "" {
		t.Errorf("没有受支持的语言时返回空，得到 %q", got)
	}
	if got := LangOfHeader(""); got != "" {
		t.Errorf("空头返回空，得到 %q", got)
	}
}

// TestHandlerMessagesHaveTranslation 保证 handlers 里写死的中文串都有英文译文。
// 这是「中文原文即键」这套方案的护栏：改了原文 / 新增了文案，这里会立刻失败。
func TestHandlerMessagesHaveTranslation(t *testing.T) {
	entries, err := os.ReadDir(filepath.Join("..", "handlers"))
	if err != nil {
		t.Fatalf("读取 handlers 目录失败：%v", err)
	}
	// 只在这些调用形态里找文案
	callRe := regexp.MustCompile(`(?:fail|badRequest|okMessage)\(c,\s*(?:[^,()]*,\s*)?("(?:[^"\\]|\\.)*")|return ("(?:[^"\\]|\\.)*"),\s*false|errors\.New\("((?:[^"\\]|\\.)*)"\)`)
	checked := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		data, err := os.ReadFile(filepath.Join("..", "handlers", e.Name()))
		if err != nil {
			t.Fatalf("读取 %s 失败：%v", e.Name(), err)
		}
		for _, m := range callRe.FindAllStringSubmatch(string(data), -1) {
			raw := m[1] + m[2] + m[3]
			if raw == "" {
				continue
			}
			msg, err := strconv.Unquote(raw)
			if err != nil {
				continue
			}
			// 只校验中文文案：英文原文（比如透传的数据库错误）不需要目录
			if !containsHan(msg) {
				continue
			}
			checked++
			if got := translate(EN, msg); got == msg {
				t.Errorf("%s：文案 %q 没有英文译文（整句 / 前缀 / 尾缀都没命中）", e.Name(), msg)
			}
		}
	}
	if checked == 0 {
		t.Fatal("一条文案都没匹配到，正则需要修一下")
	}
	t.Logf("校验了 %d 条中文文案", checked)
}

// containsHan 是否包含中日韩统一表意文字（中文文案的粗略判断）。
func containsHan(s string) bool {
	for _, r := range s {
		if r >= 0x4e00 && r <= 0x9fff {
			return true
		}
	}
	return false
}
