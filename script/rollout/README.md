# V4.3 灰度切换工具

对应 [docs/V4.3-rollout-checklist.md](../../docs/V4.3-rollout-checklist.md) 的三阶段灰度。

## 准备

```bash
export DB_HOST=db.staging.example.com
export DB_USER=ppanel
export DB_PASS=...
export DB_NAME=ppanel
export REDIS_HOST=redis.staging.example.com
chmod +x rollout.sh
```

## 常用命令

```bash
./rollout.sh status                       # 查看当前所有 plan 的 V4.3 模式
./rollout.sh enable 5 1500 --confirm      # 把 plan 5 切到 V4.3,每设备 ¥15
./rollout.sh disable 5 --confirm          # 退回旧时长计费
./rollout.sh stage1 5 3 --confirm         # plan 5 限制到 node 3(灰度 10%)
./rollout.sh stage3 --confirm             # 全量切换(慎用)
./rollout.sh invalidate-cache             # 强制节点立即重拉 user list
```

## 安全机制

- 所有写操作默认是 dry-run,需 `--confirm` 才真正执行
- 每次写入前自动备份对应 `subscribe` 行到 `rollout-backup-*.sql`
- 只在 staging / 灰度环境运行;生产环境建议先复盘备份

## 阶段对照

| 阶段 | 命令 | 监控指标(对照 Grafana Device Billing 看板) |
| --- | --- | --- |
| Stage 0 准备 | `status` 看 baseline | `server_user_list_p99` 基线 |
| Stage 1 灰度 10% | `stage1 <plan> <node>` | 同节点错误率 < 0.1% |
| Stage 2 灰度 50% | 手工绑节点 + `enable` 多 plan | p99 < 500ms |
| Stage 3 全量 | `stage3` | `notice_dispatch_total{tpl="throttle_*"}` 不暴增 |

## 紧急回滚

```bash
# 立即把所有 V4.3 plan 退回旧时长计费
mysql -h $DB_HOST -u $DB_USER -p$DB_PASS $DB_NAME \
  -e "UPDATE subscribe SET unit_price_per_device = 0 WHERE unit_price_per_device > 0;"

./rollout.sh invalidate-cache
```

完整回滚脚本详见 [docs/V4.3-rollout-checklist.md §紧急回滚](../../docs/V4.3-rollout-checklist.md)。
