# Device Billing Grafana 仪表盘

## 导入

1. Grafana → Dashboards → New → Import
2. 上传 `device-billing-dashboard.json`(或粘贴 JSON)
3. 选 Prometheus 数据源(变量 `datasource`)
4. 保存,UID:`ppanel-device-billing`

## 4 大模块

| 行 | 关注点 | SLO |
| --- | --- | --- |
| Server User List | p50/p95/p99 + 304 命中率 + QPS | p99 < 500ms (DoD #19) |
| Throttle State Machine | throttled/cut-off 计数 + 通知派发 + 邮件 p99 | 通知积压 < 60s |
| Devices & Reset | 在线设备 / reset rate / limiter 桶基数 | limiter 桶基数稳定 |
| Commerce | device vs legacy 收入对比 + 加购 + 佣金 | 灰度后 device 占比上升 |

## 服务端需要暴露的 metrics

仪表盘引用以下 metric 名,**需要在 ppanel-server 里加埋点**(决策 §10 DoD 未要求,
但 Phase 10 上线后建议补上以便可观测):

| Metric | 类型 | Label | 触发点 |
| --- | --- | --- | --- |
| `http_request_duration_seconds_bucket` | histogram | `path,method` | gin middleware(已有? 需检查) |
| `http_requests_total` | counter | `path,status` | 同上 |
| `ppanel_user_subscribe_throttled_total` | gauge | — | trafficStatusLogic.markCutOff 反向计 |
| `ppanel_user_subscribe_cutoff_total` | gauge | — | 同上 |
| `ppanel_notice_dispatch_total` | counter | `template` | noticeDispatchLogic.dispatchOne |
| `ppanel_email_send_duration_seconds_bucket` | histogram | — | sendEmailLogic |
| `ppanel_device_active_total` | gauge | — | 定时 / DB query |
| `ppanel_device_reset_total` | counter | — | resetDevice 成功路径 |
| `ppanel_device_reset_blocked_total` | counter | — | applyResetFrequencyCheck 拒绝路径 |
| `ppanel_node_speed_limiter_buckets` | gauge | `node` | ppanel-node 内 sync.Map.Range 计数 |
| `ppanel_order_amount_total` | counter | `mode` | activateOrderLogic 完成时 |
| `ppanel_traffic_addon_total` | counter | — | addTrafficAddonLogic 成功路径 |
| `ppanel_commission_amount_total` | counter | — | commission.go grantCommission 成功 |

埋点最低 cost 实现:用 `prometheus/client_golang` 在每个写路径加 `.Inc()` / `.Observe()`,~30 行代码。
未埋点的指标在仪表盘上会显示空,不影响其他面板。

## 告警建议(Grafana Alerting)

| 告警 | 条件 | 严重度 |
| --- | --- | --- |
| user list p99 超 SLO | `histogram_quantile(0.99, …) > 0.5` 持续 5min | P1 |
| 邮件派发积压 | `ppanel_email_send_duration_seconds_bucket{le="+Inf"}` 增长 < 派发率 | P2 |
| reset 拒绝率突增 | `rate(reset_blocked) / rate(reset_total) > 0.5` 持续 10min | P3 |
| limiter 桶突增(决策 32 倍增 bug) | `ppanel_node_speed_limiter_buckets` 突增 ≥ 3x | P1 |
