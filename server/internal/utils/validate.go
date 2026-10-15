package utils

import (
	"regexp"
	"strings"
)

// EmailMaxLength 邮箱最大长度，与 models.User.Email 的列宽（size:128）保持一致。
const EmailMaxLength = 128

// emailRe 与前端 web/lib/validate.ts 中的 emailRe 保持一致：
// 本地部分允许字母数字与 ._%+-，域名至少包含一个点，每段以字母数字开头 / 结尾。
var emailRe = regexp.MustCompile(`^[A-Za-z0-9._%+-]+@[A-Za-z0-9](?:[A-Za-z0-9-]*[A-Za-z0-9])?(?:\.[A-Za-z0-9](?:[A-Za-z0-9-]*[A-Za-z0-9])?)+$`)

// NormalizeEmail 去掉邮箱首尾空白，便于入库与查重。
func NormalizeEmail(email string) string {
	return strings.TrimSpace(email)
}

// IsEmail 校验邮箱格式是否合法（只判断格式与长度，不校验域名是否真实存在）。
func IsEmail(email string) bool {
	email = NormalizeEmail(email)
	return email != "" && len(email) <= EmailMaxLength && emailRe.MatchString(email)
}
