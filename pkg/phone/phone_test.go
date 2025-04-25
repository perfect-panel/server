package phone

import (
	"testing"

	"github.com/nyaruka/phonenumbers"
)

func TestPhoneNumber(t *testing.T) {
	parsedNumber, err := phonenumbers.Parse("+8615502505555", "")
	if err != nil {
		t.Fatalf("Failed to parse phone number: %v", err)
		return
	}
	// 检查手机号是否有效
	isValid := phonenumbers.IsValidNumber(parsedNumber)
	// 获取区域代码 (如 CN, US, IN)
	region := phonenumbers.GetRegionCodeForNumber(parsedNumber)
	t.Log(isValid)
	t.Log(region)
}

func TestCheck(t *testing.T) {
	var phone = "15502505555"
	if !Check("86", phone) {
		t.Fatalf("Check phone number failed: %s", phone)
	}
	t.Logf("Check phone number success: %s", phone)
}

func TestGetCountryCode(t *testing.T) {
	var phone = "14407941888"
	countryCode := GetCountryCode(phone)
	t.Logf("Country code: %s", countryCode)
}

func TestFormatToInternational(t *testing.T) {
	var phone = "8615502505555"
	international := FormatToInternational(phone)
	t.Logf("International format: %s", international)
}

func TestFormatToE164(t *testing.T) {
	var phone = "4407941888"
	e164, err := FormatToE164("1", phone)
	if err != nil {
		t.Fatalf("Failed to format phone number to E164: %v", err)
		return
	}
	t.Logf("E164 format: %s", e164)
}

func TestMask(t *testing.T) {
	var phone = "+14407941888"
	mask := MaskPhoneNumber(phone)
	t.Logf("Mask format: %s", mask)
}
