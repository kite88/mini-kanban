// 邮箱校验规则与后端 server/internal/utils/validate.go 中的 emailRe 保持一致，
// 前端先拦一道，避免明显非法的邮箱发起请求（后端仍会再校验一次）。

/** 邮箱最大长度，与后端 EmailMaxLength / users.email 列宽保持一致 */
export const EMAIL_MAX_LENGTH = 128;

/**
 * 邮箱格式：本地部分允许字母数字与 ._%+-，域名至少包含一个点，
 * 每段以字母数字开头 / 结尾（不允许 a@-b.com、a@b..com 这类写法）。
 */
export const emailRe =
  /^[A-Za-z0-9._%+-]+@[A-Za-z0-9](?:[A-Za-z0-9-]*[A-Za-z0-9])?(?:\.[A-Za-z0-9](?:[A-Za-z0-9-]*[A-Za-z0-9])?)+$/;

/** 校验邮箱是否合法（先去掉首尾空白，再判断长度与格式） */
export function isValidEmail(value: string): boolean {
  const email = value.trim();
  return email.length > 0 && email.length <= EMAIL_MAX_LENGTH && emailRe.test(email);
}

// 邮箱格式提示文案见 i18n 的 validate.emailInvalid（前后端同一句，
// 后端按 Accept-Language 返回对应语言，见 server/internal/i18n）。
