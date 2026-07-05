#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)
cd "$ROOT_DIR"

FAILED=0

run_rule() {
  local title="$1"
  local pattern="$2"
  shift 2
  local paths=("$@")

  local output
  output=$(rg -n --pcre2 "$pattern" "${paths[@]}" --glob '**/*.go' --glob '!**/*_test.go' || true)
  if [[ -n "$output" ]]; then
    echo "[FAIL] $title"
    echo "$output"
    echo
    FAILED=1
  else
    echo "[OK]   $title"
  fi
}

# Main flow must not depend on plugin-owned data structures.
run_rule \
  "main flow must not reference plugin data structures" \
  'check_items|branch_task|inspection_results|chain_steps' \
  backend/internal/service

# Main flow must not branch on concrete plugin kind values.
run_rule \
  "main flow must not hardcode concrete kind values" \
  '"simple"|"inspection"|"chain"' \
  backend/internal/service

# Main flow must not lookup handlers by hardcoded literal kind.
run_rule \
  "main flow must not lookup handlers by literal kind" \
  'LookupHandler\("|taskManager\.Get\("' \
  backend/internal/service

# Plugin boundaries: chain plugin should not reference inspection-owned data structures.
run_rule \
  "chain plugin must not reference inspection private structures" \
  'check_items|branch_task|inspection_results' \
  backend/pkg/taskkind/chain

# Plugin boundaries: inspection plugin should not reference chain-owned table structures.
run_rule \
  "inspection plugin must not reference chain private structures" \
  'chain_steps' \
  backend/pkg/taskkind/inspection

if [[ "$FAILED" -ne 0 ]]; then
  echo "plugin isolation check failed"
  exit 1
fi

echo "plugin isolation check passed"
