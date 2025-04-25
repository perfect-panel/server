package exchangeRate

import "testing"

func TestGetExchangeRete(t *testing.T) {
	t.Skip("skip TestGetExchangeRete")
	result, err := GetExchangeRete("USD", "CNY", "90734e5af4f5353114cdaf3bb9c3f2e3", 1)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(result)
}
