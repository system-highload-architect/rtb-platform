#!/usr/bin/env bash
set -euo pipefail

BASE_URL="http://localhost:8080"
PASS=0
FAIL=0

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

test_case() {
    local name="$1"; local method="$2"; local url="$3"; local data="$4"
    local expected_code="$5"; local expected_body_contains="$6"
    shift 6 || true
    local extra_args=("$@")

    echo -n "TEST: $name ... "
    http_code=$(curl -s -o /tmp/resp.txt -w "%{http_code}" -X "$method" "$url" -H "Content-Type: application/json" "${extra_args[@]}" -d "$data")
    body=$(cat /tmp/resp.txt)

    if [[ "$http_code" != "$expected_code" ]]; then
        echo -e "${RED}FAIL${NC} (expected code $expected_code, got $http_code, body: $body)"
        FAIL=$((FAIL+1))
        return
    fi

    if [[ -n "$expected_body_contains" ]] && ! echo "$body" | grep -q "$expected_body_contains"; then
        echo -e "${RED}FAIL${NC} (body missing '$expected_body_contains', got: $body)"
        FAIL=$((FAIL+1))
    else
        echo -e "${GREEN}PASS${NC} (code=$http_code)"
        PASS=$((PASS+1))
    fi
}

echo "======================================"
echo "Running integration tests for RTB Platform"
echo "NOTE: Please restart auction and accounting services before running this script."
echo "======================================"

# 1. JSON-RPC валидация
test_case "Missing method" "POST" "$BASE_URL/rpc" '{"jsonrpc":"2.0","id":1}' "200" "Invalid Request: missing method"

# 2. Фрод-детекция
test_case "Fraud detection" "POST" "$BASE_URL/rpc" '{"jsonrpc":"2.0","method":"auction.bid","params":{"device_id":"bad","ip":"192.168.1.1"},"id":1}' "200" '"error":"fraud"'

# 3. Аукцион с кампаниями (должен вернуть campaign_id > 0)
test_case "Auction with campaigns" "POST" "$BASE_URL/rpc" \
  '{"jsonrpc":"2.0","method":"auction.bid","params":{"device_id":"d1","ip":"100.200.300.400","lat":55.7558,"lng":37.6173},"id":3}' \
  "200" '"campaign_id":"100'

# 4. Получение баланса (ожидаем 10000, если сервисы перезапущены)
test_case "Get balance" "POST" "$BASE_URL/rpc" \
  '{"jsonrpc":"2.0","method":"accounting.getBalance","params":{"campaign_id":"campaign-1"},"id":10}' \
  "200" '"amount":10000'

# 5. Дебет первый раз
test_case "Debit first time" "POST" "$BASE_URL/rpc" \
  '{"jsonrpc":"2.0","method":"accounting.debit","params":{"campaign_id":"campaign-1","amount":{"amount":150,"scale":2},"bid_id":"bid-1"},"id":11}' \
  "200" '"success":true' \
  -H "Idempotency-Key: debit-key-1"

# 6. Дебет дубликат
test_case "Debit duplicate" "POST" "$BASE_URL/rpc" \
  '{"jsonrpc":"2.0","method":"accounting.debit","params":{"campaign_id":"campaign-1","amount":{"amount":150,"scale":2},"bid_id":"bid-1"},"id":12}' \
  "409" "duplicate request" \
  -H "Idempotency-Key: debit-key-1"

# 7. Баланс после списания (должен уменьшиться на 150)
test_case "Balance after debit" "POST" "$BASE_URL/rpc" \
  '{"jsonrpc":"2.0","method":"accounting.getBalance","params":{"campaign_id":"campaign-1"},"id":13}' \
  "200" '"amount":9850'

# 8. CORS preflight
test_case "CORS preflight" "OPTIONS" "$BASE_URL/rpc" "" "200" "" \
  -H "Origin: http://example.com" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: Content-Type"

# 9. Rate limiting (ищем 429)
echo -n "TEST: Rate limiting ... "
rate_limit_hit=0
for i in {1..15}; do
    code=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/rpc" -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"auction.bid","params":{"device_id":"rl"},"id":1}')
    if [[ "$code" == "429" ]]; then
        rate_limit_hit=1; break
    fi
done
if [[ $rate_limit_hit -eq 1 ]]; then
    echo -e "${GREEN}PASS${NC} (got 429)"; PASS=$((PASS+1))
else
    echo -e "${RED}FAIL${NC} (no 429 received)"; FAIL=$((FAIL+1))
fi

echo ""
echo "======================================"
echo "Results: ${GREEN}${PASS} passed${NC}, ${RED}${FAIL} failed${NC}"
echo "======================================"
[[ $FAIL -eq 0 ]]