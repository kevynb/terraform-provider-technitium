#!/usr/bin/env bash
set -euo pipefail

TECHNITIUM_API_URL="${TECHNITIUM_API_URL:-http://localhost:5380}"
TECHNITIUM_ADMIN_USER="${TECHNITIUM_ADMIN_USER:-admin}"
TECHNITIUM_ADMIN_PASSWORD="${TECHNITIUM_ADMIN_PASSWORD:-changeme}"
TECHNITIUM_SKIP_TLS_VERIFY="${TECHNITIUM_SKIP_TLS_VERIFY:-}"
TECHNITIUM_TOKEN_FILE="${TECHNITIUM_TOKEN_FILE:-tools/acceptance/token.env}"
export TECHNITIUM_API_URL TECHNITIUM_ADMIN_USER TECHNITIUM_ADMIN_PASSWORD

login_url="$(python3 - <<'PY'
import os
from urllib.parse import urlencode

base = os.environ["TECHNITIUM_API_URL"].rstrip("/")
if not base.endswith("/api"):
    base = f"{base}/api"
query = urlencode({
    "user": os.environ["TECHNITIUM_ADMIN_USER"],
    "pass": os.environ["TECHNITIUM_ADMIN_PASSWORD"],
    "includeInfo": "true",
})
print(f"{base}/user/login?{query}")
PY
)"

curl_opts=("--silent" "--show-error" "--fail")
if [[ -n "${TECHNITIUM_SKIP_TLS_VERIFY}" ]]; then
  curl_opts+=("--insecure")
fi

token=""
start_ts="$(date +%s)"
while :; do
  response="$(curl "${curl_opts[@]}" "${login_url}" 2>/dev/null || true)"
  if [[ -n "${response}" ]]; then
    token="$(printf "%s" "${response}" | python3 -c 'import json, sys; raw = sys.stdin.read(); payload = json.loads(raw) if raw.strip() else {}; print(payload["token"]) if payload.get("status") == "ok" and payload.get("token") else exit(1)')" || true
    if [[ -n "${token}" ]]; then
      break
    fi
  fi

  now_ts="$(date +%s)"
  if (( now_ts - start_ts >= 120 )); then
    echo "error: unable to obtain Technitium token" >&2
    exit 1
  fi
  sleep 1
done

mkdir -p "$(dirname "${TECHNITIUM_TOKEN_FILE}")"
printf "export TECHNITIUM_API_TOKEN=%s\n" "${token}" > "${TECHNITIUM_TOKEN_FILE}"
printf "%s\n" "${token}"
