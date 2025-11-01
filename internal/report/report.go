package report

const (
	GatewayURL  = "http://127.0.0.1:%d" // 网关地址
	RegisterAPI = "/basic/register"     // 模块注册接口
)

// RegisterRequest 模块注册请求参数
type RegisterRequest struct {
	Secret         string `json:"secret"`          // 通讯密钥
	ProxyPath      string `json:"proxy_path"`      // 代理路径
	ServiceURL     string `json:"service_url"`     // 服务地址
	Repository     string `json:"repository"`      // 服务代码仓库
	ServiceName    string `json:"service_name"`    // 服务名称
	ServiceVersion string `json:"service_version"` // 服务版本
}

// RegisterResponse 模块注册响应参数
type RegisterResponse struct {
	Success bool   `json:"success"` // 注册是否成功
	Message string `json:"message"` // 返回信息
}
