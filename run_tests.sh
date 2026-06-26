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

    if [[ -n "$expected_body_contains" ]] && ! echo "$body" | grep -Fq "$expected_body_contains"; then
        echo -e "${RED}FAIL${NC} (body missing '$expected_body_contains', got: $body)"
        FAIL=$((FAIL+1))
    else
        echo -e "${GREEN}PASS${NC} (code=$http_code)"
        PASS=$((PASS+1))
    fi
}

IDEMP_KEY="debit-key-$(date +%s)"

echo "======================================"
echo "Running full integration tests"
echo "======================================"

# ---------- Auth: получение токена ----------
echo -n "TEST: Auth login ... "
LOGIN_RESP=$(curl -s -X POST "$BASE_URL/rpc" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"auth.login","params":{"email":"admin@rtb-platform.local","password":"Admin123!"},"id":1}')
TOKEN=$(echo "$LOGIN_RESP" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
if [[ -z "$TOKEN" ]]; then
    echo -e "${RED}FAIL${NC} (could not get token, response: $LOGIN_RESP)"
    FAIL=$((FAIL+1))
else
    echo -e "${GREEN}PASS${NC} (token obtained)"
    PASS=$((PASS+1))
fi

AUTH_HEADER="Authorization: Bearer $TOKEN"

# 1. JSON-RPC валидация
test_case "Missing method" "POST" "$BASE_URL/rpc" '{"jsonrpc":"2.0","id":1}' "200" "Invalid Request: missing method"

# 2. Фрод-детекция
test_case "Fraud detection" "POST" "$BASE_URL/rpc" '{"jsonrpc":"2.0","method":"auction.bid","params":{"device_id":"bad","ip":"192.168.1.1"},"id":1}' "200" '"error":"fraud"'

# 3. Аукцион с кампаниями
test_case "Auction with campaigns" "POST" "$BASE_URL/rpc" '{"jsonrpc":"2.0","method":"auction.bid","params":{"device_id":"d1","ip":"100.200.300.400","lat":55.7558,"lng":37.6173},"id":3}' "200" '"accepted"'
# 4. Получение баланса
test_case "Get balance" "POST" "$BASE_URL/rpc" '{"jsonrpc":"2.0","method":"accounting.getBalance","params":{"campaign_id":"campaign-1"},"id":10}' "200" '"amount":9850'
# 5. Дебет первый раз
test_case "Debit first time" "POST" "$BASE_URL/rpc" \
  '{"jsonrpc":"2.0","method":"accounting.debit","params":{"campaign_id":"campaign-1","amount":{"amount":150,"scale":2},"bid_id":"bid-1"},"id":11}' \
  "200" '"success":true' -H "Idempotency-Key: $IDEMP_KEY"

# 6. Дебет дубликат
test_case "Debit duplicate" "POST" "$BASE_URL/rpc" \
  '{"jsonrpc":"2.0","method":"accounting.debit","params":{"campaign_id":"campaign-1","amount":{"amount":150,"scale":2},"bid_id":"bid-1"},"id":12}' \
  "409" "duplicate request" -H "Idempotency-Key: $IDEMP_KEY"

# 7. Баланс после списания
test_case "Balance after debit" "POST" "$BASE_URL/rpc" '{"jsonrpc":"2.0","method":"accounting.getBalance","params":{"campaign_id":"campaign-1"},"id":13}' "200" '"amount":9850'

# 8. Аналитика: отчёт (с токеном)
test_case "Analytics report" "GET" "$BASE_URL/api/report?start_date=2025-01-01&end_date=2025-01-31" "" "200" "[]" -H "$AUTH_HEADER"

# 9. Аналитика: прогноз (с токеном)
test_case "Analytics forecast" "GET" "$BASE_URL/api/forecast?history=1,2,3,4,5,6,7,8,9,10,11,12,13,14,15&horizon=3" "" "200" "forecast" -H "$AUTH_HEADER"

# 10. Аналитика: факторный анализ (с токеном)
test_case "Analytics factor analysis" "GET" "$BASE_URL/api/factor-analysis" "" "200" "explained_variance_ratio" -H "$AUTH_HEADER"

# 11. CORS preflight
test_case "CORS preflight" "OPTIONS" "$BASE_URL/rpc" "" "200" "" \
  -H "Origin: http://example.com" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: Content-Type"

echo ""
echo "======================================"
echo "Results: ${GREEN}${PASS} passed${NC}, ${RED}${FAIL} failed${NC}"
echo "======================================"
[[ $FAIL -eq 0 ]]