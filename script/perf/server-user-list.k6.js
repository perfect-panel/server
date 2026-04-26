// V4.3 §10 DoD #19 — /v1/server/user p99 < 500ms (1 万用户场景)
//
// 用法:
//   k6 run --vus 50 --duration 60s -e BASE=https://staging.example.com -e SECRET=xxx -e SERVER=1 server-user-list.k6.js
//
// SLO 阈值在 thresholds 里,通不过会以非 0 退出码结束(CI 友好)。
//
// 测试覆盖三个相关端点:
//   1. GET  /v1/server/user        — 节点拉用户列表(决策 19,主测点)
//   2. GET  /v1/server/alivelist   — 节点拉 alive IP 集
//   3. POST /v1/server/push        — 节点上报流量(决策 32 验聚合)

import http from "k6/http";
import { check, sleep } from "k6";
import { Trend } from "k6/metrics";

const BASE = __ENV.BASE || "http://localhost:8080";
const SECRET = __ENV.SECRET || ""; // node secret_key
const SERVER_ID = __ENV.SERVER || "1";

const userListLatency = new Trend("user_list_latency_ms");
const aliveLatency = new Trend("alive_latency_ms");
const pushLatency = new Trend("push_latency_ms");

export const options = {
  scenarios: {
    user_list: {
      executor: "constant-arrival-rate",
      rate: 30, // 每秒 30 次 — 与节点真实 pull(1 节点 / 60s)对齐放大 30 倍
      timeUnit: "1s",
      duration: "1m",
      preAllocatedVUs: 50,
      maxVUs: 100,
      exec: "userListScenario",
    },
    alive_list: {
      executor: "constant-arrival-rate",
      rate: 10,
      timeUnit: "1s",
      duration: "1m",
      preAllocatedVUs: 20,
      maxVUs: 50,
      exec: "aliveScenario",
      startTime: "5s",
    },
    push: {
      executor: "constant-arrival-rate",
      rate: 5,
      timeUnit: "1s",
      duration: "1m",
      preAllocatedVUs: 20,
      maxVUs: 50,
      exec: "pushScenario",
      startTime: "10s",
    },
  },
  thresholds: {
    // 决策 §10 DoD #19
    "user_list_latency_ms": ["p(99)<500", "p(95)<300", "p(50)<100"],
    "alive_latency_ms": ["p(99)<300"],
    "push_latency_ms": ["p(99)<500"],
    "http_req_failed": ["rate<0.01"], // <1% 失败
  },
};

function commonParams() {
  return {
    secret_key: SECRET,
    server_id: SERVER_ID,
    protocol: "vless", // 任选一个常驻协议
  };
}

export function userListScenario() {
  const params = commonParams();
  const url = `${BASE}/v1/server/user?secret_key=${encodeURIComponent(params.secret_key)}&server_id=${params.server_id}&protocol=${params.protocol}`;
  const t0 = Date.now();
  const res = http.get(url, { tags: { name: "user_list" } });
  userListLatency.add(Date.now() - t0);
  check(res, {
    "user_list 200/304": (r) => r.status === 200 || r.status === 304,
    "user_list has users key on 200": (r) =>
      r.status !== 200 || (r.body && r.body.includes("users")),
  });
  // ETag 头存在意味着缓存命中路径走通了(决策 19 节流 304 路径)
  if (res.headers["Etag"]) {
    // 第二次带 If-None-Match,期望 304
    const res2 = http.get(url, {
      headers: { "If-None-Match": res.headers["Etag"] },
      tags: { name: "user_list_etag" },
    });
    check(res2, { "etag returns 304": (r) => r.status === 304 });
  }
  sleep(0.1);
}

export function aliveScenario() {
  const p = commonParams();
  const url = `${BASE}/v1/server/alivelist?secret_key=${encodeURIComponent(p.secret_key)}&server_id=${p.server_id}`;
  const t0 = Date.now();
  const res = http.get(url, { tags: { name: "alive_list" } });
  aliveLatency.add(Date.now() - t0);
  check(res, { "alive 200": (r) => r.status === 200 });
}

export function pushScenario() {
  const p = commonParams();
  const url = `${BASE}/v1/server/push?secret_key=${encodeURIComponent(p.secret_key)}&server_id=${p.server_id}&protocol=${p.protocol}`;
  // 模拟 100 设备一次上报
  const traffic = [];
  for (let i = 0; i < 100; i++) {
    const deviceId = 1_000_000 + Math.floor(Math.random() * 30_000);
    traffic.push({ uid: deviceId, upload: 1024, download: 4096 });
  }
  const body = JSON.stringify({ traffic });
  const t0 = Date.now();
  const res = http.post(url, body, {
    headers: { "Content-Type": "application/json" },
    tags: { name: "push" },
  });
  pushLatency.add(Date.now() - t0);
  check(res, { "push 200": (r) => r.status === 200 });
}

export function handleSummary(data) {
  return {
    stdout: textSummary(data),
    "perf-summary.json": JSON.stringify(data, null, 2),
  };
}

function textSummary(data) {
  const m = data.metrics;
  const lines = [
    "",
    "===== V4.3 perf summary =====",
    `user_list p50:  ${fmt(m.user_list_latency_ms?.values?.["p(50)"])}ms`,
    `user_list p95:  ${fmt(m.user_list_latency_ms?.values?.["p(95)"])}ms`,
    `user_list p99:  ${fmt(m.user_list_latency_ms?.values?.["p(99)"])}ms  (target < 500)`,
    `alive p99:      ${fmt(m.alive_latency_ms?.values?.["p(99)"])}ms`,
    `push p99:       ${fmt(m.push_latency_ms?.values?.["p(99)"])}ms`,
    `http_req_failed rate: ${fmt(m.http_req_failed?.values?.rate, 4)}`,
    `total requests: ${m.http_reqs?.values?.count}`,
    "",
  ];
  return lines.join("\n");
}

function fmt(v, digits = 1) {
  if (v === undefined || v === null) return "-";
  return Number(v).toFixed(digits);
}
