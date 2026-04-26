# V4.3 Perf 测试

对应方案 §10 DoD #19:**1 万用户场景下 `/v1/server/user` p99 < 500ms**。

## 准备

1. **隔离环境** — staging,数据库与生产隔离。
2. 跑过 V4.3 全部 migration(02126 + 02127 + 02128)。
3. 至少有一个 `server` 行(对应节点)和一个 `subscribe` 行;脚本会再追加一个 plan id=10000。

## Step 1:数据填充

```bash
mysql -h $DB_HOST -u $DB_USER -p$DB_PASS $DB_NAME < seed.sql
```

执行约 30s,产生:
- 1 万 `user`(id 1000000-1009999)
- 1 万 `user_subscribe`(每个用户 1 个,绑 plan 10000)
- 3 万 `user_subscribe_device`(每订阅 3 槽,id 1000000-1029999)

## Step 2:配置 plan 关联节点

种子脚本不会动 `subscribe.nodes` / `subscribe.node_tags`,
你需要手工(或用 admin 接口)把 plan 10000 关联到要压测的节点列表,
否则 `/v1/server/user` 看到 0 个匹配 subscribe,就只返回占位用户。

```sql
-- 例:plan 10000 绑定全部节点(node_id 1..N)
UPDATE subscribe SET nodes = '1,2,3,4,5' WHERE id = 10000;
```

## Step 3:跑 k6

```bash
# 安装 k6:https://k6.io/docs/getting-started/installation/
k6 run \
  -e BASE=https://staging.example.com \
  -e SECRET=xxxxxx \
  -e SERVER=1 \
  server-user-list.k6.js
```

通过条件(thresholds 在脚本里):

| 指标 | 阈值 |
| --- | --- |
| `user_list_latency_ms` p99 | < 500ms |
| `user_list_latency_ms` p95 | < 300ms |
| `user_list_latency_ms` p50 | < 100ms |
| `alive_latency_ms` p99 | < 300ms |
| `push_latency_ms` p99 | < 500ms |
| `http_req_failed` rate | < 1% |

任一不达标 k6 以非 0 退出,可挂 CI。

## Step 4:查看结果

控制台会打 summary,同时 dump `perf-summary.json` 供后续 diff:

```
===== V4.3 perf summary =====
user_list p50:  18.0ms
user_list p95:  42.5ms
user_list p99:  88.3ms  (target < 500)
alive p99:      35.1ms
push p99:       72.4ms
http_req_failed rate: 0.0000
total requests: 2700
```

## 调参方向(若 SLO 不达标)

1. **首跑超 500ms**:看 server log + DB slow log,可能是:
   - `subscribe.nodes` 没建立 FK 导致全表扫描 → 加 nodes 字段索引
   - `FindUsersSubscribeBySubscribeId` N+1 查询 → 改 IN 批量
2. **缓存命中率低**:检查 Redis `server:user:*` key 是否被频繁 DEL —
   重置/加购等操作会失活缓存,压测时要避免并发触发。
3. **节点数 ×订阅数 ×3 设备 = 行数膨胀**:用 `EXPLAIN` 看 query plan,
   `idx_user_subscribe_id` 命中是 P1。

## 清理

```sql
DELETE FROM user_subscribe_device WHERE id BETWEEN 1000000 AND 1029999;
DELETE FROM user_subscribe        WHERE id BETWEEN 1000000 AND 1009999;
DELETE FROM user                  WHERE id BETWEEN 1000000 AND 1009999;
DELETE FROM subscribe             WHERE id = 10000;
```
