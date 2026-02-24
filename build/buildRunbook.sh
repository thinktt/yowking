#!/usr/bin/env bash
set -euo pipefail

LOG="/build/logs/runbook.log"
: > "$LOG"

cd /build/runbook

if ! pkg runbook.js -o /build/dist/runbook -t node18-alpine-x64 >"$LOG" 2>&1; then
  echo "ERROR: runbook build failed"
  cat $LOG
  exit 1
fi

# Smoke test compiled binary in Alpine build env.
test_input='{"moves": ["d2d4"], "book": "EarlyQueen.bin"}'
echo "Testing runbook"
echo "Sending command: /build/dist/runbook '$test_input'"
if ! test_output=$(/build/dist/runbook "$test_input" 2>&1); then
  echo "ERROR: runbook binary smoke test failed"
  echo "Output is: $test_output"
  exit 1
fi
echo "Output is: $test_output"
