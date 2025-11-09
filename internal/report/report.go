package report

const (
	RegisterAPI = "/basic/register" // 模块注册接口
)

// RegisterResponse 模块注册响应参数
type RegisterResponse struct {
	Success bool   `json:"success"` // 注册是否成功
	Message string `json:"message"` // 返回信息
}
