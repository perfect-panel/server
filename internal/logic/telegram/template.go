package telegram

const BindNotify = `🤖 **尊敬的用户，您已成功绑定 Bot！**

**绑定账号**: {{.Id}} 
**绑定时间**: {{.Time}}

感谢您的支持！🎉  
您现在可以通过该 Bot 随时管理您的账户和服务。如有任何问题，请联系客服。💬
`

const PurchaseNotify = `🎉 **尊敬的用户，您已成功购买服务！**

**订单编号**: {{.OrderNo}}
**订阅名称**: {{.SubscribeName}}
**订单金额**: **{{.OrderAmount}}**  
**到期时间**: {{.ExpireTime}}

感谢您的支持！💖  
您的服务已成功激活，随时为您提供高速、稳定、安全的网络体验。  
如有疑问，请联系客服，我们将竭诚为您服务！💬`

const RenewalNotify = `🎉 **尊敬的用户，您已成功续费服务！**

**订单编号**: {{.OrderNo}}
**订阅名称**: {{.SubscribeName}}
**订单金额**: **{{.OrderAmount}}**  
**到期时间**: {{.ExpireTime}}

感谢您的支持！💖  
您的服务已成功激活，随时为您提供高速、稳定、安全的网络体验。  
如有疑问，请联系客服，我们将竭诚为您服务！💬`

// RechargeNotify 充值通知
const RechargeNotify = `💳 **尊敬的用户，您的账户充值已完成！**

💰 **充值金额**: {{.OrderAmount}} 
🏦 **充值方式**: _{{.PaymentMethod}}_
⏰ **充值时间**: {{.Time}}  
📊 **当前账户余额**: **{{.Balance}}**

感谢您的支持！🎉  
余额可用于购买套餐或其他服务。  
如有疑问，请联系客服，我们将竭诚为您服务！💬`

// AdminOrderNotify 管理员订单通知
const AdminOrderNotify = `
📦 **订单通知**

🆔 **系统订单号**: {{.OrderNo}}
🔖 **商户订单号**: {{.TradeNo}}
👤 **用户账号**: {{.UserEmail}}
💰 **订单金额**: **{{.OrderAmount}}**
📋 **订单状态**: **{{.OrderStatus}}**
📦 **订阅名称**: _{{.SubscribeName}}_
⏰ **下单时间**: {{.OrderTime}}
💳 **支付方式**: _{{.PaymentMethod}}_
`

// AdminOrderDaily 管理员每日订单统计
const AdminOrderDaily = `
📊 **每日流水统计**

**统计日期**: {{.Date}}  
**总订单数**: **{{.Orders}}**  
**总成交金额**: **{{.Amount}}**

**按套餐分类：**
{{.Subscribe}}

**按支付方式：**
{{.Payment}}

**总览：**
 **当日退款数**: {{.RefundOrders}} 单，退款金额：**{{RefundAmount}}**
 **实际入账金额**: **{{ActualAmount}}**

**请注意**:
以上数据为系统自动统计，仅供参考。如需详细数据或对账，请查看管理后台。
`

// SubscribeExpireNotify 订阅到期通知
const SubscribeExpireNotify = `尊敬的用户，您的订阅即将到期。

📦 **订阅名称**: _{{.SubscribeName}}_
⏰ **到期时间**: {{.ExpiredAt}} 
💰 **续费金额**: **{{.RenewalAmount}}**

为确保服务不受影响，请尽快续费。
如有疑问，请联系客服，我们将竭诚为您服务！💬`

// UnbindNotify 解绑通知
const UnbindNotify = `🤖 尊敬的用户，您好！

您的账户已成功解绑：

**用户ID**：{{.Id}}
**解绑时间**：{{.Time}}

解绑后，您将无法通过该Bot进行账户相关操作。
如需重新绑定，请访问[绑定页面](#)完成操作。

如有任何疑问，请随时联系客服，我们将竭诚为您服务！
感谢您的理解与支持！`

// ResetTrafficNotify 重置流量通知
const ResetTrafficNotify = `📊 尊敬的用户，您好！

您的账户流量已成功重置：

**用户邮箱**：{{.Email}}
**套餐名称**：{{.SubscribeName}}
**重置时间**：{{.ResetTime}}
**到期时间**：{{.ExpireTime}}

新的流量额度已生效，感谢您的支持！
如有任何问题，请随时联系客服，我们将竭诚为您服务！💬`
