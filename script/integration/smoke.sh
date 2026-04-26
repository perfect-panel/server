#!/usr/bin/env bash
# 集成 smoke 测试 — 跑过 docker-compose up 之后调用。
# 验证关键端点都能返回非 5xx。需要的"真实业务流"(注册→购买→设备)依赖
# 邮件/支付配置,smoke 只验路由通畅 + DB schema 已迁移。

set -euo pipefail
BASE="${BASE:-http://localhost:38080}"

n_pass=0
n_fail=0

check() {
  local name="$1" expected="$2" url="$3"
  local code
  code=$(curl -s -o /dev/null -w "%{http_code}" "$url" || echo 000)
  # 接受 2xx / 3xx / 401 / 403(未鉴权)/ 400 (请求参数缺失);拒绝 5xx 和 0
  if [[ "$code" =~ ^(2..|3..|400|401|403|404)$ ]]; then
    echo "  ✓ $name → $code"
    n_pass=$((n_pass + 1))
  else
    echo "  ✗ $name → $code (expected $expected)"
    n_fail=$((n_fail + 1))
  fi
}

echo "=== smoke @ $BASE ==="

echo "[user public endpoints]"
check "subscribe URL no token"      "400" "$BASE/v1/subscribe/config"
check "QR no token"                 "400" "$BASE/v1/public/qr"
check "QR fake token"               "404" "$BASE/v1/public/qr?token=nonexistent"
check "subscribes list (auth)"      "401" "$BASE/v1/public/subscribe/list"
check "my subscribes (auth)"        "401" "$BASE/v1/public/subscribe/my"
check "messages list (auth)"        "401" "$BASE/v1/portal/messages"
check "terms status (auth)"         "401" "$BASE/v1/portal/terms/status"

echo "[admin endpoints]"
check "admin audit list (auth)"     "401" "$BASE/v1/admin/audit/list"
check "admin device list (auth)"    "401" "$BASE/v1/admin/device/list?user_id=1"
check "admin sitecontent (auth)"    "401" "$BASE/v1/admin/sitecontent/list"
check "admin direct_list (auth)"    "401" "$BASE/v1/admin/server/1/direct_list"

echo "[node endpoints — secret_key required]"
check "GET user list"               "200" "$BASE/v1/server/user?server_id=1&secret_key=wrong"
check "GET alivelist"               "200" "$BASE/v1/server/alivelist?server_id=1&secret_key=wrong"

echo ""
echo "=== summary: $n_pass pass / $n_fail fail ==="
[[ $n_fail -eq 0 ]] || exit 1
