# 集成测试

一键起 mysql + redis + ppanel-server,跑端到端 smoke。

## 前置

- Docker / Docker Compose
- Make
- bash + curl(smoke 脚本依赖)

## 一键

```bash
# 起栈 → 跑 smoke → 关栈
make integration-cycle
```

或分步:

```bash
make integration-up      # 起 mysql + redis + ppanel-server
make integration-test    # 跑 smoke
make integration-down    # 关栈 + 删数据卷
```

## smoke 覆盖

[`smoke.sh`](smoke.sh) 验证 14 个关键端点的路由通畅:

- 公开:`/v1/subscribe/config`、`/v1/public/qr`
- 用户:`/v1/public/subscribe/{list, my}`、`/v1/portal/{messages, terms/status}`
- 管理:`/v1/admin/audit/list`、`/v1/admin/device/list`、`/v1/admin/sitecontent/list`、`/v1/admin/server/:id/direct_list`
- 节点:`/v1/server/{user, alivelist}`

通过条件:全部返回非 5xx + 非超时(401/403/400 等业务错误也算通过,因 smoke 不带认证)。

## 端到端业务流测试(手工)

smoke 之后想测真实业务流(注册→购买→设备→重置→加购),建议:

```bash
# 1. 进容器手工 INSERT 测试用户和 plan
docker exec -it ppanel-int-mysql mysql -uroot -pintegration ppanel <<EOF
INSERT INTO subscribe (id, name, unit_price, unit_price_per_device, unit_time, traffic, max_device_count, sell, show, created_at, updated_at)
VALUES (1, 'IntTestPlan', 0, 1000, 'Month', 107374182400, 20, 1, 1, NOW(3), NOW(3));
EOF

# 2. 跑 Go 单测验证 logic
go test -count=1 ./internal/logic/public/subscribe/...

# 3. 用 curl 模拟用户操作流(需配合 admin token)
curl -X POST http://localhost:38080/v1/public/portal/purchase \
  -H "Content-Type: application/json" \
  -d '{"auth_type":"email","identifier":"test@example.com","password":"x","payment":1,"subscribe_id":1,"device_count":3}'
```

## 配置覆写

[`ppanel-test.yaml`](ppanel-test.yaml) 提供集成测试用的配置(关闭邮件、内置 trial balance 等)。
若要测真实邮件/支付,改这个文件即可。

## 资源占用

| 容器 | CPU | 内存 |
| --- | --- | --- |
| mysql:8.0 | < 0.5 | ≈ 400MB |
| redis:7-alpine | < 0.05 | ≈ 10MB |
| ppanel-server | < 0.5 | ≈ 80MB |

整套关栈即销毁(mysql 用 tmpfs)。

## CI 接入

`.github/workflows/ci.yml` 的 `migration-dryrun` job 已有 mysql service container,
可作为 CI 跑 schema up→down→up 验证的样板。完整集成测试受 docker-in-docker 限制,
若要在 CI 跑建议用 self-hosted runner 或 GitHub `services:` 直接起 mysql/redis。
