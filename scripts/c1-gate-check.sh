#!/usr/bin/env bash
# C1 gate check — Phase 9.1 inbound MCP endpoint.
# Run from the repo root:  bash scripts/c1-gate-check.sh
# Requires: docker compose, go, jq, psql, golang-migrate, and a configured .env
# (JWT_SECRET + DATABASE_URL), same as the README local-dev setup.
set -euo pipefail

API_URL="${API_URL:-http://localhost:8080}"
MCP_URL="$API_URL/mcp"
REQ_ID=0
PASS=0
FAIL=0
API_PID=""

pass() { PASS=$((PASS+1)); echo "  PASS  $1"; }
fail() { FAIL=$((FAIL+1)); echo "  FAIL  $1"; }

cleanup() {
  if [ -n "$API_PID" ] && kill -0 "$API_PID" 2>/dev/null; then
    kill "$API_PID" 2>/dev/null || true
  fi
}
trap cleanup EXIT

# Streamable HTTP may answer as SSE frames; unwrap to bare JSON-RPC if so.
extract_jsonrpc() {
  local body
  body=$(cat)
  if printf '%s' "$body" | grep -q '^data: '; then
    printf '%s\n' "$body" | sed -n 's/^data: //p' | tail -n 1
  else
    printf '%s\n' "$body"
  fi
}

# mcp_call <method> <params-json> [token]
mcp_call() {
  local method="$1" params="$2" token="${3:-}"
  REQ_ID=$((REQ_ID+1))
  local args=(
    -sS -X POST "$MCP_URL"
    -H 'Content-Type: application/json'
    -H 'Accept: application/json, text/event-stream'
    -d "{\"jsonrpc\":\"2.0\",\"id\":$REQ_ID,\"method\":\"$method\",\"params\":$params}"
  )
  if [ -n "$token" ]; then
    args+=( -H "Authorization: Bearer $token" )
  fi
  curl "${args[@]}" | extract_jsonrpc
}

login() { # email password -> token on stdout
  curl -sS -X POST "$API_URL/api/v1/auth/login" \
    -H 'Content-Type: application/json' \
    -d "{\"email\":\"$1\",\"password\":\"$2\"}" | jq -r .accessToken
}

echo "==> Loading .env"
set -a; source .env; set +a

echo "==> Starting postgres"
docker compose up -d postgres >/dev/null
for i in $(seq 1 30); do
  if docker compose exec -T postgres pg_isready -q 2>/dev/null; then break; fi
  sleep 1
done

echo "==> Migrations + seed users"
migrate -path backend/migrations -database "$DATABASE_URL" up
psql "$DATABASE_URL" -f scripts/seed_users.sql >/dev/null

echo "==> Seeding C1 catalog fixture (idempotent)"
psql "$DATABASE_URL" >/dev/null <<'SQL'
INSERT INTO mcp_servers (name, description, base_url, transport_type, status, owner_team)
SELECT 'demo-tools', 'C1 gate fixture (mock executor)', 'http://localhost:1/mock', 'streamable_http', 'active', 'platform'
WHERE NOT EXISTS (SELECT 1 FROM mcp_servers WHERE name = 'demo-tools');

INSERT INTO mcp_servers (name, description, base_url, transport_type, status, owner_team)
SELECT 'inactive-server', 'C1 gate fixture, must stay hidden', 'http://localhost:1/mock', 'streamable_http', 'inactive', 'platform'
WHERE NOT EXISTS (SELECT 1 FROM mcp_servers WHERE name = 'inactive-server');

INSERT INTO mcp_tools (server_id, name, title, description, input_schema, risk_level, enabled)
SELECT id, 'echo', 'Echo', 'Echoes the message back',
       '{"type":"object","properties":{"message":{"type":"string"}},"required":["message"]}',
       'low', true
FROM mcp_servers WHERE name = 'demo-tools'
AND NOT EXISTS (SELECT 1 FROM mcp_tools t WHERE t.server_id = mcp_servers.id AND t.name = 'echo');

INSERT INTO mcp_tools (server_id, name, title, description, input_schema, risk_level, enabled)
SELECT id, 'disabled_tool', 'Disabled', 'Disabled tool, must stay hidden',
       '{"type":"object","properties":{}}', 'low', false
FROM mcp_servers WHERE name = 'demo-tools'
AND NOT EXISTS (SELECT 1 FROM mcp_tools t WHERE t.server_id = mcp_servers.id AND t.name = 'disabled_tool');

INSERT INTO mcp_tools (server_id, name, title, description, input_schema, risk_level, enabled)
SELECT id, 'risky_tool', 'Risky', 'Medium-risk tool, must stay hidden',
       '{"type":"object","properties":{}}', 'medium', true
FROM mcp_servers WHERE name = 'demo-tools'
AND NOT EXISTS (SELECT 1 FROM mcp_tools t WHERE t.server_id = mcp_servers.id AND t.name = 'risky_tool');

INSERT INTO mcp_tools (server_id, name, title, description, input_schema, risk_level, enabled)
SELECT id, 'ghost_tool', 'Ghost', 'On an inactive server, must stay hidden',
       '{"type":"object","properties":{}}', 'low', true
FROM mcp_servers WHERE name = 'inactive-server'
AND NOT EXISTS (SELECT 1 FROM mcp_tools t WHERE t.server_id = mcp_servers.id AND t.name = 'ghost_tool');
SQL

echo "==> Starting API"
(cd backend && go build -o /tmp/mcp-gateway-api ./cmd/api)
/tmp/mcp-gateway-api &
API_PID=$!
for i in $(seq 1 30); do
  if curl -sf "$API_URL/health" >/dev/null 2>&1; then break; fi
  sleep 1
done
curl -sf "$API_URL/health" >/dev/null || { echo "API failed to start"; exit 1; }

echo
echo "== T1: unauthenticated /mcp is rejected with a Bearer challenge"
HEADERS=$(curl -s -D - -o /dev/null -X POST "$MCP_URL" \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json, text/event-stream' \
  -d '{"jsonrpc":"2.0","id":0,"method":"tools/list","params":{}}')
printf '%s' "$HEADERS" | grep -q '^HTTP/.* 401' \
  && pass "no token -> 401" || fail "no token -> expected 401"
printf '%s' "$HEADERS" | grep -qi '^WWW-Authenticate: Bearer' \
  && pass "WWW-Authenticate: Bearer present" || fail "challenge header missing"

echo "== T2: developer login"
DEV_TOKEN=$(login developer@mcp-gateway.local DeveloperPass123)
[ -n "$DEV_TOKEN" ] && [ "$DEV_TOKEN" != "null" ] && pass "developer token" || { fail "developer login"; exit 1; }

echo "== T3: initialize"
INIT=$(mcp_call initialize '{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"c1-gate","version":"0.1.0"}}' "$DEV_TOKEN")
printf '%s' "$INIT" | jq -e '.result.serverInfo.name == "mcp-gateway"' >/dev/null \
  && pass "initialize -> mcp-gateway" || fail "initialize -> $INIT"

echo "== T4: tools/list exposure rules (D3/D7)"
LIST=$(mcp_call tools/list '{}' "$DEV_TOKEN")
NAMES=$(printf '%s' "$LIST" | jq -r '.result.tools[].name' | sort)
echo "    advertised: $(echo $NAMES | tr '\n' ' ')"
printf '%s\n' "$NAMES" | grep -qx 'demo-tools__echo'        && pass "namespaced tool advertised" || fail "demo-tools__echo missing"
! printf '%s\n' "$NAMES" | grep -q 'disabled_tool'          && pass "disabled tool hidden"       || fail "disabled tool leaked"
! printf '%s\n' "$NAMES" | grep -q 'risky_tool'             && pass "medium-risk tool hidden"    || fail "medium-risk tool leaked"
! printf '%s\n' "$NAMES" | grep -q 'ghost_tool'             && pass "inactive server hidden"     || fail "inactive server leaked"
printf '%s' "$LIST" | jq -e '.result.tools[0].inputSchema.type == "object"' >/dev/null \
  && pass "input schema passed through" || fail "input schema missing"

echo "== T5: tools/call success path + audit row"
CALL=$(mcp_call tools/call '{"name":"demo-tools__echo","arguments":{"message":"c1-gate"}}' "$DEV_TOKEN")
printf '%s' "$CALL" | jq -e '.result.isError | not' >/dev/null \
  && pass "call succeeded" || fail "call failed -> $CALL"
printf '%s' "$CALL" | jq -e '.result.content[0].text | contains("c1-gate")' >/dev/null \
  && pass "echo result content" || fail "unexpected content -> $CALL"
AUDIT=$(psql "$DATABASE_URL" -tAc "SELECT status FROM tool_invocations ORDER BY created_at DESC LIMIT 1")
[ "$AUDIT" = "succeeded" ] && pass "audit row = succeeded" || fail "audit row = $AUDIT"

echo "== T6: viewer is denied by the policy chain"
VIEW_TOKEN=$(login viewer@mcp-gateway.local ViewerPass123)
VIEW=$(mcp_call tools/call '{"name":"demo-tools__echo","arguments":{"message":"nope"}}' "$VIEW_TOKEN")
printf '%s' "$VIEW" | jq -e '.result.isError == true' >/dev/null \
  && pass "viewer call -> isError" || fail "viewer call -> $VIEW"

echo "== T7: unknown tool is rejected at the protocol layer"
# The SDK validates the name against registered tools first, so unknown tools
# get JSON-RPC -32602 (invalid params) before our handler runs. An isError
# result is also acceptable per MCP convention.
UNK=$(mcp_call tools/call '{"name":"demo-tools__nope","arguments":{}}' "$DEV_TOKEN")
printf '%s' "$UNK" | jq -e '(.error.code == -32602) or (.result.isError == true)' >/dev/null \
  && pass "unknown tool -> -32602 or isError" || fail "unknown tool -> $UNK"

echo
echo "==================================="
echo "C1 gate: $PASS passed, $FAIL failed"
echo "==================================="
[ "$FAIL" -eq 0 ]
