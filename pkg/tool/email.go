package tool

import (
	"strings"
)

func MaskEmail(email string) string {
	atIndex := strings.Index(email, "@")
	if atIndex == -1 || atIndex == 0 || atIndex == len(email)-1 {
		return email
	}
	localPart := email[:atIndex]
	domainPart := email[atIndex+1:]

	// 本地部分需要至少保留首字符和末字符
	if len(localPart) < 2 {
		return email
	}
	// 替换本地部分中间字符为星号
	maskedLocal := string(localPart[0]) + strings.Repeat("*", len(localPart)-2) + string(localPart[len(localPart)-1])
	// 返回处理后的邮箱地址
	return maskedLocal + "@" + domainPart
}
