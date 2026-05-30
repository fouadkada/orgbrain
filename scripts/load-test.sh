#!/usr/bin/env bash
# load-test.sh
# k6 load test: 10 concurrent queries, assert p95 < 8s.
# Runs on release/* branches only (CI gate added in Story 1.5).
# TODO (Story 1.5): implement k6 script targeting /v1/query.
set -euo pipefail
echo "load-test: not yet enabled"
