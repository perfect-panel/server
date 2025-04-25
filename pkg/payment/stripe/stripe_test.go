package stripe

import (
	"testing"

	"github.com/stripe/stripe-go/v81"
)

func TestStripeAlipay(t *testing.T) {
	t.Skipf("Skip TestStripeAlipay test")
	client := NewClient(Config{
		WebhookSecret: "",
	})
	order := Order{
		OrderNo:   "JS20210719123456789",
		Subscribe: "测试",
		Amount:    100,
		Currency:  string(stripe.CurrencyGBP),
		Payment:   "alipay",
	}
	user := User{
		UserId: 1,
		Email:  "tension@ppanel.dev",
	}
	result, err := client.CreatePaymentSheet(&order, &user)
	if err != nil {
		t.Error(err.Error())
	}
	t.Logf("TradeNo: %s\n", result.ClientSecret)
}

func TestStripeWechat(t *testing.T) {
	t.Skipf("Skip TestStripeWechat test")
	client := NewClient(Config{
		SecretKey:     "SecretKey",
		PublicKey:     "PublicKey",
		WebhookSecret: "",
	})
	order := Order{
		OrderNo:   "JS20210719123456789",
		Subscribe: "测试",
		Amount:    100,
		Currency:  string(stripe.CurrencyGBP),
		Payment:   "wechat_pay",
	}
	user := User{
		UserId: 1,
		Email:  "tension@ppanel.dev",
	}
	result, err := client.CreatePaymentSheet(&order, &user)
	if err != nil {
		t.Error(err.Error())
	}
	t.Logf("TradeNo: %s\n", result.ClientSecret)
}
