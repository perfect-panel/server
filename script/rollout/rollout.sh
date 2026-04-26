#!/usr/bin/env bash
# V4.3 灰度切换脚本(对齐 docs/V4.3-rollout-checklist.md 三阶段)。
#
# 用法:
#   ./rollout.sh status              查看当前 rollout 状态
#   ./rollout.sh enable <subscribeId> [unit_price_per_device_cents]
#                                    把指定套餐切到 V4.3 device-billing
#   ./rollout.sh disable <subscribeId>
#                                    退回旧时长计费(unit_price_per_device=0)
#   ./rollout.sh stage1 <subscribeId> <serverId>
#                                    阶段 1:1 套餐 + 1 节点(灰度 10%)
#   ./rollout.sh stage2              阶段 2:把 V4.3 套餐绑定 50% 节点
#   ./rollout.sh stage3              阶段 3:全量 — 所有套餐启用 device-billing
#                                    (但不强制 unit_price_per_device > 0,需手工设)
#   ./rollout.sh invalidate-cache    DEL server:user:list:* 强制节点重拉
#
# 环境变量(必填):
#   DB_HOST  DB_USER  DB_PASS  DB_NAME
#   REDIS_HOST  REDIS_PORT (默认 6379)
#
# 安全:
#   - 所有 SQL 都打印 dry-run,确认后用 --confirm 真正执行
#   - 任何更新都先备份对应行到 ./rollout-backup-<timestamp>.sql

set -euo pipefail

DB_HOST="${DB_HOST:?DB_HOST required}"
DB_USER="${DB_USER:?DB_USER required}"
DB_PASS="${DB_PASS:?DB_PASS required}"
DB_NAME="${DB_NAME:?DB_NAME required}"
REDIS_HOST="${REDIS_HOST:-127.0.0.1}"
REDIS_PORT="${REDIS_PORT:-6379}"

CONFIRM=0
for arg in "$@"; do
  if [ "$arg" = "--confirm" ]; then CONFIRM=1; fi
done

mysql_run() {
  mysql --protocol=TCP -h "$DB_HOST" -u "$DB_USER" -p"$DB_PASS" "$DB_NAME" -B -e "$1"
}

backup_subscribe() {
  local sid="$1"
  local f="rollout-backup-$(date +%Y%m%d-%H%M%S)-sub-$sid.sql"
  mysql --protocol=TCP -h "$DB_HOST" -u "$DB_USER" -p"$DB_PASS" "$DB_NAME" \
    -e "SELECT * FROM subscribe WHERE id = $sid \G" > "$f"
  echo "[backup] subscribe id=$sid → $f"
}

cmd_status() {
  echo "=== V4.3 rollout status ==="
  mysql_run "
    SELECT
      id, name,
      unit_price_per_device,
      max_device_count,
      traffic_addon_unit_price,
      traffic_addon_unit_size,
      commission_rate,
      CASE WHEN unit_price_per_device > 0 THEN 'V4.3 device-billing'
           ELSE 'legacy time-billing' END AS mode
    FROM subscribe
    ORDER BY id;
  "
  echo ""
  echo "=== Active V4.3 user_subscribes (status=1, expire>now) ==="
  mysql_run "
    SELECT COUNT(*) AS active_subs,
           SUM(device_count) AS total_device_slots
    FROM user_subscribe
    WHERE status = 1 AND expire_time > NOW();
  "
  echo ""
  echo "=== Devices by status ==="
  mysql_run "
    SELECT status, COUNT(*) AS n FROM user_subscribe_device GROUP BY status;
  "
}

cmd_enable() {
  local sid="${1:?subscribeId required}"
  local price="${2:-1000}" # default 10.00 元/device
  echo "[enable] Will set subscribe.id=$sid:"
  echo "  unit_price_per_device = $price (cents)"
  echo "  max_device_count = 20 (if currently 0)"
  echo "  traffic_addon_unit_price = 100 (cents/GB) (if currently 0)"
  echo "  traffic_addon_unit_size = 1073741824 (1 GiB) (if currently 0)"
  echo "  commission_rate = 10 (if currently 0)"
  if [ "$CONFIRM" -eq 0 ]; then
    echo "(dry-run; rerun with --confirm to apply)"; return 0
  fi
  backup_subscribe "$sid"
  mysql_run "
    UPDATE subscribe SET
      unit_price_per_device = $price,
      max_device_count = IF(max_device_count > 0, max_device_count, 20),
      traffic_addon_unit_price = IF(traffic_addon_unit_price > 0, traffic_addon_unit_price, 100),
      traffic_addon_unit_size  = IF(traffic_addon_unit_size  > 0, traffic_addon_unit_size,  1073741824),
      commission_rate = IF(commission_rate > 0, commission_rate, 10)
    WHERE id = $sid;
  "
  cmd_invalidate_cache
  echo "[enable] OK. Verify with: ./rollout.sh status"
}

cmd_disable() {
  local sid="${1:?subscribeId required}"
  echo "[disable] Will set subscribe.id=$sid unit_price_per_device = 0 (legacy mode)"
  if [ "$CONFIRM" -eq 0 ]; then echo "(dry-run; rerun --confirm)"; return 0; fi
  backup_subscribe "$sid"
  mysql_run "UPDATE subscribe SET unit_price_per_device = 0 WHERE id = $sid;"
  cmd_invalidate_cache
}

cmd_stage1() {
  local sid="${1:?subscribeId required}"
  local nodeId="${2:?nodeId required}"
  echo "[stage1] V4.3 灰度 10% — 把 plan $sid 限制到 node $nodeId"
  if [ "$CONFIRM" -eq 0 ]; then echo "(dry-run; rerun --confirm)"; return 0; fi
  backup_subscribe "$sid"
  mysql_run "
    UPDATE subscribe SET nodes = '$nodeId' WHERE id = $sid;
  "
  cmd_enable "$sid"
}

cmd_stage2() {
  echo "[stage2] V4.3 灰度 50% — 把所有 V4.3 套餐扩到当前 enabled nodes 的 50%"
  echo "(此操作复杂,建议手工:在 admin 把 plan.nodes 改写)"
  if [ "$CONFIRM" -eq 0 ]; then echo "(dry-run)"; return 0; fi
  echo "Listing currently V4.3 plans:"
  mysql_run "SELECT id, name, nodes FROM subscribe WHERE unit_price_per_device > 0;"
  echo "请在 admin 的 /dashboard/servers 节点编辑页同步配置 direct_list,"
  echo "并按 docs/V4.3-rollout-checklist.md §阶段 2 校核监控指标。"
}

cmd_stage3() {
  echo "[stage3] V4.3 全量 — 标记所有 plan 启用 device-billing(不强制改价)"
  echo "默认价格 1000 cents/device(¥10/台);通过环境变量 DEFAULT_PRICE 覆写"
  local default_price="${DEFAULT_PRICE:-1000}"
  if [ "$CONFIRM" -eq 0 ]; then
    echo "(dry-run) 会执行:UPDATE subscribe SET unit_price_per_device = $default_price WHERE unit_price_per_device = 0;"
    return 0
  fi
  mysql --protocol=TCP -h "$DB_HOST" -u "$DB_USER" -p"$DB_PASS" "$DB_NAME" \
    -e "SELECT * FROM subscribe \G" > "rollout-backup-$(date +%Y%m%d-%H%M%S)-stage3.sql"
  mysql_run "UPDATE subscribe SET unit_price_per_device = $default_price WHERE unit_price_per_device = 0;"
  cmd_invalidate_cache
}

cmd_invalidate_cache() {
  echo "[cache] DEL server:user:list:* (节点最快 ≤60s 拉取最新 user list)"
  redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" --scan --pattern "server:user:*" | while read k; do
    redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" DEL "$k" >/dev/null
  done
  echo "[cache] done."
}

# ---- main ----

case "${1:-status}" in
  status)              cmd_status ;;
  enable)              cmd_enable "${2:?}" "${3:-}" ;;
  disable)             cmd_disable "${2:?}" ;;
  stage1)              cmd_stage1 "${2:?}" "${3:?}" ;;
  stage2)              cmd_stage2 ;;
  stage3)              cmd_stage3 ;;
  invalidate-cache)    cmd_invalidate_cache ;;
  *)
    echo "usage: $0 {status|enable|disable|stage1|stage2|stage3|invalidate-cache} [args] [--confirm]"
    exit 2
    ;;
esac
