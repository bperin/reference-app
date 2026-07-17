#!/bin/sh
set -eu

if ! command -v curl >/dev/null || ! command -v jq >/dev/null || ! command -v psql >/dev/null; then
	echo "curl, jq, and psql are required for the smoke test" >&2
	exit 1
fi

if [ -z "${DATABASE_URL:-}" ] && [ -f .env ]; then
	set -a
	. ./.env
	set +a
fi

: "${DATABASE_URL:?DATABASE_URL is required}"
base_url="${SMOKE_BASE_URL:-http://127.0.0.1:8080}"
log_file="$(mktemp)"
binary_file="$(mktemp)"

cleanup() {
	if [ -n "${app_pid:-}" ]; then
		kill "$app_pid" 2>/dev/null || true
		wait "$app_pid" 2>/dev/null || true
	fi
	if [ -n "${email:-}" ]; then
		psql "$DATABASE_URL" --set=email="$email" -c "DELETE FROM users WHERE email = :'email'" >/dev/null 2>&1 || true
	fi
	rm -f "$log_file"
	rm -f "$binary_file"
}
trap cleanup EXIT INT TERM

go build -o "$binary_file" ./cmd/api
"$binary_file" >"$log_file" 2>&1 &
app_pid=$!

attempt=0
until curl --fail --silent --show-error "$base_url/healthz" >/dev/null; do
	attempt=$((attempt + 1))
	if [ "$attempt" -ge 30 ]; then
		cat "$log_file" >&2
		echo "API health check did not become ready" >&2
		exit 1
	fi
	sleep 1
done

email="smoke-$(date +%s)-$$@example.com"
curl --fail --silent --show-error -X POST "$base_url/auth/register" \
	-H 'Content-Type: application/json' \
	-d "{\"email\":\"$email\",\"password\":\"smoke-password\",\"display_name\":\"Smoke User\"}" | jq -e --arg email "$email" '.email == $email' >/dev/null

tokens="$(curl --fail --silent --show-error -X POST "$base_url/auth/oauth/token" \
	-H 'Content-Type: application/json' \
	-d "{\"grant_type\":\"password\",\"username\":\"$email\",\"password\":\"smoke-password\"}")"
access_token="$(printf '%s' "$tokens" | jq -er .access_token)"
refresh_token="$(printf '%s' "$tokens" | jq -er .refresh_token)"

curl --fail --silent --show-error -H "Authorization: Bearer $access_token" \
	"$base_url/users/me" | jq -e --arg email "$email" '.email == $email' >/dev/null

curl --fail --silent --show-error -X POST "$base_url/posts/" \
	-H "Authorization: Bearer $access_token" \
	-H 'Content-Type: application/json' \
	-d '{"title":"Smoke post","content":"OAuth-protected post"}' | jq -e '.title == "Smoke post"' >/dev/null

curl --fail --silent --show-error -X POST "$base_url/auth/oauth/token" \
	-H 'Content-Type: application/json' \
	-d "{\"grant_type\":\"refresh_token\",\"refresh_token\":\"$refresh_token\"}" | jq -e '.access_token != "" and .refresh_token != ""' >/dev/null

test "$(psql "$DATABASE_URL" -tAc 'SELECT uuid_generate_v4() IS NOT NULL')" = "t"
echo "smoke test passed against $base_url"
