#!/usr/bin/env bash
set -euo pipefail

BENCH_PACKAGE="${BENCH_PACKAGE:-./internal/transport/httpserver}"
BENCH_REGEX="${BENCH_REGEX:-^BenchmarkHTTPServer}"
BENCH_TIME="${BENCH_TIME:-1s}"
BENCH_COUNT="${BENCH_COUNT:-1}"

go test "${BENCH_PACKAGE}" \
  -run '^$' \
  -bench "${BENCH_REGEX}" \
  -benchmem \
  -benchtime "${BENCH_TIME}" \
  -count "${BENCH_COUNT}"
