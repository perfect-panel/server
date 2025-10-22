package epay

import "testing"

func TestEpay(t *testing.T) {
	client := NewClient("", "http://127.0.0.1", "", "")
	order := Order{
		Name:      "测试",
		OrderNo:   "123456789",
		Amount:    1000,
		SignType:  "md5",
		NotifyUrl: "http://127.0.0.1",
		ReturnUrl: "http://127.0.0.1",
	}
	url := client.CreatePayUrl(order)
	t.Logf("PayUrl: %s\n", url)

}

func TestQueryOrderStatus(t *testing.T) {
	t.Skipf("Skip TestQueryOrderStatus test")
	client := NewClient("Pid", "Url", "Key", "Type")
	orderNo := "123456789"
	status := client.QueryOrderStatus(orderNo)
	t.Logf("OrderNo: %s, Status: %v\n", orderNo, status)
}

func TestVerifySign(t *testing.T) {
	t.Skipf("Skip TestVerifySign test")
	params := map[string]string{
		"pid":          "1654",
		"trade_no":     "2024121521150860990",
		"out_trade_no": "202412152115078262977262254",
		"type":         "alipay",
		"name":         "product",
		"money":        "10",
		"trade_status": "TRADE_SUCCESS",
		"sign":         "d3181f18ebdf9821f0ab6ee93faa82d1",
		"sign_type":    "MD5",
	}

	key := "LbTabbB580zWyhXhyyww7wwvy5u8k0wl"
	c := NewClient("Pid", "Url", key, "Type")
	if c.VerifySign(params) {
		t.Logf("Sign verification success!")
	} else {
		t.Error("Sign verification failed!")
	}
}
