package epay

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/tool"
)

type Client struct {
	Pid  string
	Url  string
	Key  string
	Type string
}

type Order struct {
	Name      string
	OrderNo   string
	Amount    float64
	SignType  string
	NotifyUrl string
	ReturnUrl string
}

type queryOrderStatusResponse struct {
	Code       int    `json:"code"`
	Msg        string `json:"msg"`
	TradeNo    string `json:"trade_no"`
	OutTradeNo string `json:"out_trade_no"`
	Type       string `json:"type"`
	Status     int    `json:"status"`
}

func NewClient(pid, url, key string, Type string) *Client {
	return &Client{
		Pid:  pid,
		Url:  url,
		Key:  key,
		Type: Type,
	}
}

func (c *Client) CreatePayUrl(order Order) string {
	// Prepare URL values
	params := url.Values{}
	params.Set("name", order.Name)
	params.Set("money", tool.FormatFloat(order.Amount, 2))
	params.Set("notify_url", order.NotifyUrl)
	params.Set("out_trade_no", order.OrderNo)
	params.Set("pid", c.Pid)
	params.Set("type", c.Type)
	params.Set("return_url", order.ReturnUrl)

	// Generate the sign using the CreateSign function
	sign := c.createSign(c.structToMap(order))
	params.Set("sign", sign)

	// Add sign_type manually
	params.Set("sign_type", "MD5")
	return c.Url + "/submit.php?" + params.Encode()
}

func (c *Client) createSign(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		if params[k] != "" && k != "sign" && k != "sign_type" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		parts = append(parts, k+"="+params[k])
	}
	queryString := strings.Join(parts, "&")
	text := queryString + c.Key
	return tool.Md5Encode(text, false)
}

func (c *Client) VerifySign(params map[string]string) bool {
	return c.createSign(params) == params["sign"]
}

func (c *Client) QueryOrderStatus(orderNo string) bool {
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get(c.Url + "/api.php" + "?act=order" + "&pid=" + c.Pid + "&key=" + c.Key + "&out_trade_no=" + orderNo)
	if err != nil {
		logger.Error("[Epay] QueryOrderStatus error", logger.Field("orderNo", orderNo), logger.Field("error", err.Error()))
		return false
	}
	defer resp.Body.Close()
	value, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("[Epay] QueryOrderStatus error", logger.Field("orderNo", orderNo), logger.Field("error", err.Error()))
		return false
	}
	var response queryOrderStatusResponse
	err = json.Unmarshal(value, &response)
	if err != nil {
		logger.Error("[Epay] QueryOrderStatus error", logger.Field("orderNo", orderNo), logger.Field("error", err.Error()))
		return false
	}
	return response.Status == 1
}

// StructToMap converts a struct to map[string]string
func (c *Client) structToMap(order Order) map[string]string {
	result := make(map[string]string)
	result["money"] = tool.FormatFloat(order.Amount, 2)
	result["name"] = order.Name
	result["notify_url"] = order.NotifyUrl
	result["out_trade_no"] = order.OrderNo
	result["pid"] = c.Pid
	result["type"] = c.Type
	result["return_url"] = order.ReturnUrl
	return result
}
