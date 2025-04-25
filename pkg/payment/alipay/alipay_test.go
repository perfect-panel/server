package alipay

import (
	"context"
	"testing"
)

func TestClientPreCreateTrade(t *testing.T) {
	t.Skipf("Skip TestClientPreCreateTrade")
	cfg := Config{
		InvoiceName: "XrayR",
		NotifyURL:   "https://example.com/alipay/notify",
		Sandbox:     true,
	}
	c := NewClient(cfg)
	order := Order{
		OrderNo: "20210701000001",
		Amount:  100,
	}
	qr, err := c.PreCreateTrade(context.Background(), order)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(qr)
}
