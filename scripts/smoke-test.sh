#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${1:-http://localhost:8080}"
PASS=0
FAIL=0

green() { printf "\033[0;32m✓ %s\033[0m\n" "$*"; }
red()   { printf "\033[0;31m✗ %s\033[0m\n" "$*"; }

check() {
  local label="$1"
  local expected_status="$2"
  local actual_status="$3"
  local body="$4"

  if [ "$actual_status" -eq "$expected_status" ]; then
    green "$label — HTTP $actual_status — $body"
    PASS=$((PASS + 1))
  else
    red "$label — expected HTTP $expected_status, got HTTP $actual_status — $body"
    FAIL=$((FAIL + 1))
  fi
}

echo "Smoke testing $BASE_URL"
echo "---"

# GET /health
response=$(curl -s -w "\n%{http_code}" "$BASE_URL/health")
body=$(echo "$response" | head -n1)
status=$(echo "$response" | tail -n1)
check "GET /health" 200 "$status" "$body"

# GET /hello
response=$(curl -s -w "\n%{http_code}" "$BASE_URL/hello")
body=$(echo "$response" | head -n1)
status=$(echo "$response" | tail -n1)
check "GET /hello" 200 "$status" "$body"

# POST /items — valid payload
response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/items" \
  -H "Content-Type: application/json" \
  -d '{"name":"widget","value":"42"}')
body=$(echo "$response" | head -n1)
status=$(echo "$response" | tail -n1)
check "POST /items (valid)" 201 "$status" "$body"

# POST /items — second item
response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/items" \
  -H "Content-Type: application/json" \
  -d '{"name":"gadget","value":"hello"}')
body=$(echo "$response" | head -n1)
status=$(echo "$response" | tail -n1)
check "POST /items (second)" 201 "$status" "$body"

# POST /items — bad content type (expect 400 or 415)
response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/items" \
  -H "Content-Type: text/plain" \
  -d 'not json')
body=$(echo "$response" | head -n1)
status=$(echo "$response" | tail -n1)
if [ "$status" -ge 400 ]; then
  green "POST /items (bad content-type) — HTTP $status (error as expected) — $body"
  PASS=$((PASS + 1))
else
  red "POST /items (bad content-type) — expected 4xx, got HTTP $status — $body"
  FAIL=$((FAIL + 1))
fi

echo "---"
echo "Results: $PASS passed, $FAIL failed"
[ "$FAIL" -eq 0 ]
