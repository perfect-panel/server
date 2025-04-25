package phone

import (
	"fmt"
	"regexp"

	"github.com/nyaruka/phonenumbers"
)

func Check(areaCode, telephone string) bool {
	parsedNumber, err := phonenumbers.Parse(fmt.Sprintf("+%s%s", areaCode, telephone), areaCode)
	if err != nil {
		return false
	}
	// 检查手机号是否有效
	return phonenumbers.IsValidNumber(parsedNumber)
}

func CheckPhone(telephone string) bool {
	parsedNumber, err := phonenumbers.Parse(fmt.Sprintf("+%s", telephone), "")
	if err != nil {
		return false
	}
	// 检查手机号是否有效
	return phonenumbers.IsValidNumber(parsedNumber)
}

func GetCountryCode(telephone string) string {
	parsedNumber, err := phonenumbers.Parse(fmt.Sprintf("+%s", telephone), "")
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%d", *parsedNumber.CountryCode)
}

func FormatToInternational(telephone string) string {
	parsedNumber, err := phonenumbers.Parse(fmt.Sprintf("+%s", telephone), "")
	if err != nil {
		return ""
	}
	return phonenumbers.Format(parsedNumber, phonenumbers.INTERNATIONAL)
}

func FormatToE164(area, phone string) (string, error) {
	parsedNumber, err := phonenumbers.Parse(fmt.Sprintf("+%s%s", area, phone), "")
	if err != nil {
		return "", err
	}
	return phonenumbers.Format(parsedNumber, phonenumbers.E164), nil
}

// MaskPhoneNumber 解析并脱敏电话号码
func MaskPhoneNumber(phone string) string {
	// 解析电话号码
	num, err := phonenumbers.Parse(phone, "")
	if err != nil {
		return ""
	}
	// 获取国际格式，如 "+1 512-345-6789"
	formatted := phonenumbers.Format(num, phonenumbers.INTERNATIONAL)

	// 使用正则匹配国家代码和号码部分
	re := regexp.MustCompile(`(\+\d{1,3})\s*(.*)`)
	matches := re.FindStringSubmatch(formatted)
	if len(matches) < 3 {
		return formatted // 如果格式不匹配，返回原格式
	}

	countryCode := matches[1] // 国家代码（如 "+1"）
	numberPart := matches[2]  // 本地号码部分（如 "512-345-6789"）

	// 根据不同国家的号码格式进行脱敏
	maskedNumber := maskDigits(numberPart, countryCode)

	// 组合脱敏后的号码
	return fmt.Sprintf("%s %s", countryCode, maskedNumber)
}

// maskDigits 替换部分数字，保持位数和分隔符
func maskDigits(number string, countryCode string) string {
	// 统计数字个数
	digitCount := 0
	for _, r := range number {
		if r >= '0' && r <= '9' {
			digitCount++
		}
	}

	// 处理不同国家的号码格式
	runes := []rune(number)
	digitIndex := 0

	for i, r := range runes {
		if r >= '0' && r <= '9' {
			digitIndex++
			if countryCode == "+1" { // 美国号码格式：+1 (512) ***-1278
				if digitIndex > 3 && digitIndex <= digitCount-4 { // 只替换中间部分
					runes[i] = '*'
				}
			} else if countryCode == "+86" { // 中国号码格式：+86 138 **** 5678
				if digitIndex > 3 && digitIndex <= digitCount-4 {
					runes[i] = '*'
				}
			} else { // 其他国家号码，采用类似规则
				if digitIndex > 3 && digitIndex <= digitCount-4 {
					runes[i] = '*'
				}
			}
		}
	}

	return string(runes)
}
