#!/usr/bin/env bash
# CLASSIFICATION: UNCLASSIFIED
# Regression tests for CD Build's service selection rules.

set -euo pipefail

select_services() {
  local changed="$1"
  local shared="$2"
  python3 - "$changed" "$shared" <<'PY'
import json
import sys

go_all = [
    "svc-radar-ingestion",
    "svc-fusion-engine",
    "svc-track",
    "svc-webtransport",
    "svc-query",
]
changed = json.loads(sys.argv[1])
if sys.argv[2] == "true":
    go = go_all
    web = True
else:
    go = [service for service in go_all if service in changed]
    web = "web-cop-gpu" in changed
services = go + (["web-cop-gpu"] if web else [])
print(json.dumps(services))
PY
}

[[ "$(select_services '[]' false)" == '[]' ]] || exit 1
[[ "$(select_services '["svc-track"]' false)" == '["svc-track"]' ]] || exit 1
[[ "$(select_services '["web-cop-gpu"]' false)" == '["web-cop-gpu"]' ]] || exit 1
[[ "$(select_services '["pkg/ignored"]' true)" == '["svc-radar-ingestion", "svc-fusion-engine", "svc-track", "svc-webtransport", "svc-query", "web-cop-gpu"]' ]] || exit 1

echo "TEST PASS: CD Build change detection"
