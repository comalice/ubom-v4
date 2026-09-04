#!/usr/bin/env bash

set -u
set -o pipefail

iterations=${1:-100}
if ! [[ "$iterations" =~ ^[1-9][0-9]*$ ]]; then
  printf 'usage: %s [iterations]\n' "$0" >&2
  exit 2
fi

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
tmpdir=$(mktemp -d "${TMPDIR:-/tmp}/ubom-dev-spam.XXXXXX")
db="$tmpdir/dev.db"
log="$tmpdir/make-dev.log"
dev_pid=''

cleanup() {
  status=$?
  trap - INT TERM EXIT
  if [[ -n "$dev_pid" ]] && kill -0 "$dev_pid" 2>/dev/null; then
    kill -TERM "$dev_pid" 2>/dev/null || true
    wait "$dev_pid" 2>/dev/null || true
  fi
  if ss -ltn '( sport = :5173 or sport = :8080 )' 2>/dev/null | grep -q LISTEN; then
    printf 'warning: a dev listener remains after cleanup\n' >&2
  fi
  rm -rf "$tmpdir"
  exit "$status"
}
trap cleanup INT TERM EXIT

wait_for() {
  url=$1
  for _ in $(seq 1 120); do
    if curl -fsS --max-time 1 -o /dev/null "$url" 2>/dev/null; then
      return 0
    fi
    sleep 0.25
  done
  printf 'timed out waiting for %s\n--- make dev log ---\n' "$url" >&2
  cat "$log" >&2
  return 1
}

cd "$root" || exit 1
make dev DB="$db" >"$log" 2>&1 &
dev_pid=$!

wait_for 'http://127.0.0.1:8080/api/version'
wait_for 'http://127.0.0.1:5173/settings'

endpoints=(
  '/'
  '/settings'
  '/catalog'
  '/api/version'
  '/api/taxonomy-defs'
  '/api/catalog/root'
  '/api/parts?page=1&pageSize=50'
  '/api/parts?taxonomyDef=sample-taxonomy-v1&taxonomyNode=resistors&includeArchived=false'
  '/api/taxonomies/sample-taxonomy-v1/nodes/resistors'
  '/api/catalog/facets?taxonomyDef=sample-taxonomy-v1&taxonomyNode=resistors&includeArchived=false'
)

printf 'iterations=%s endpoints=%s\n' "$iterations" "${#endpoints[@]}"
for iteration in $(seq 1 "$iterations"); do
  for endpoint in "${endpoints[@]}"; do
    result=$(curl -sS --max-time 30 -o /dev/null \
      -w '%{http_code} %{time_connect} %{time_starttransfer} %{time_total}' \
      "http://127.0.0.1:5173${endpoint}" 2>/dev/null)
    curl_status=$?
    if [[ "$curl_status" -ne 0 ]]; then
      printf 'iteration=%s endpoint=%s curl_status=%s status=000 connect=- start=- total=-\n' \
        "$iteration" "$endpoint" "$curl_status"
    else
      read -r http_status connect_time start_time total_time <<<"$result"
      printf 'iteration=%s endpoint=%s curl_status=0 status=%s connect=%ss start=%ss total=%ss\n' \
        "$iteration" "$endpoint" "$http_status" "$connect_time" "$start_time" "$total_time"
    fi
  done
done
