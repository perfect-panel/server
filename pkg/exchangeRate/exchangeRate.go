package exchangeRate

import (
	"errors"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	Url = "https://api.exchangerate.host"
)

type Response struct {
	Success bool   `json:"success"`
	Terms   string `json:"terms"`
	Privacy string `json:"privacy"`
	Query   struct {
		From   string  `json:"from"`
		To     string  `json:"to"`
		Amount float64 `json:"amount"`
	} `json:"query"`
	Info struct {
		Timestamp int64   `json:"timestamp"`
		Quote     float64 `json:"quote"`
	} `json:"info"`
	Result float64 `json:"result"`
}

func GetExchangeRete(form, to, access string, amount float64) (float64, error) {
	client := resty.New()
	client.SetRetryCount(3)
	client.SetTimeout(5 * time.Second)
	client.SetBaseURL(Url)
	// amount  to string
	amountStr := strconv.FormatFloat(amount, 'f', -1, 64)

	client.SetQueryParams(map[string]string{
		"from":       form,
		"to":         to,
		"amount":     amountStr,
		"access_key": access,
	})
	resp := new(Response)
	_, err := client.R().SetResult(resp).Get("/convert")
	if err != nil {
		return 0, err
	}
	if !resp.Success {
		return 0, errors.New("exchange rate failed")
	}
	return resp.Result, nil
}
