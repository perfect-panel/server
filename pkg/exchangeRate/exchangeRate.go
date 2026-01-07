package exchangeRate

import (
	"errors"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/perfect-panel/server/pkg/logger"
)

const (
	Url = "https://api.apilayer.com"
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
		"from":   form,
		"to":     to,
		"amount": amountStr,
	})
	result := new(Response)
	resp, err := client.R().SetHeader("apikey", access).SetResult(result).Get("/currency_data/convert")

	if err != nil {

		return 0, err
	}
	if !result.Success {
		logger.Info("Exchange Rate Response: ", resp.String())
		return 0, errors.New("exchange rate failed")
	}
	return result.Result, nil
}
