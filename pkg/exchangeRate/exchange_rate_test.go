package exchangeRate

import "testing"

func TestGetExchangeRete(t *testing.T) {
	t.Skip("skip TestGetExchangeRete")
	result, err := GetExchangeRete("USD", "CNY", "", 1)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(result)
}
