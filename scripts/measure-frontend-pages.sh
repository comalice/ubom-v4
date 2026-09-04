#!/usr/bin/env bash

set -u
set -o pipefail

iterations=${1:-1}
base_url=${2:-http://127.0.0.1:5173}
max_time=${CURL_MAX_TIME:-30}

if ! [[ "$iterations" =~ ^[1-9][0-9]*$ ]]; then
  printf 'usage: %s [iterations] [base-url]\n' "$0" >&2
  printf 'example: %s 20 http://127.0.0.1:5173\n' "$0" >&2
  exit 2
fi

if ! [[ "$max_time" =~ ^[1-9][0-9]*(\.[0-9]+)?$ ]]; then
  printf 'CURL_MAX_TIME must be a positive number of seconds\n' >&2
  exit 2
fi

base_url=${base_url%/}
pages=(
  '/'
  '/catalog'
  '/parts'
  '/artifacts'
  '/settings'
  '/settings/sequence-definitions'
  '/settings/taxonomies'
)

printf 'iterations=%s pages=%s base_url=%s max_time=%ss\n' \
  "$iterations" "${#pages[@]}" "$base_url" "$max_time"

for iteration in $(seq 1 "$iterations"); do
  for page in "${pages[@]}"; do
    result=$(curl -sS --max-time "$max_time" -o /dev/null \
      -w '%{http_code} %{time_namelookup} %{time_connect} %{time_starttransfer} %{time_total}' \
      "$base_url$page" 2>/dev/null)
    curl_status=$?

    if [[ "$curl_status" -ne 0 ]]; then
      printf 'iteration=%s page=%s curl_status=%s status=000 dns=- connect=- start=- total=-\n' \
        "$iteration" "$page" "$curl_status"
      continue
    fi

    read -r http_status dns_time connect_time start_time total_time <<<"$result"
    printf 'iteration=%s page=%s curl_status=0 status=%s dns=%ss connect=%ss start=%ss total=%ss\n' \
      "$iteration" "$page" "$http_status" "$dns_time" "$connect_time" "$start_time" "$total_time"
  done
done
