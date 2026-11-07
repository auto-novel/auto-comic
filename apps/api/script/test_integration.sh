#!/usr/bin/env bash

set -euo pipefail

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
source "$script_dir/postgres.sh"
start_postgres integration "${TEST_DB_PORT:-}"

cd -- "$script_dir/.."
TEST_DB_HOST=127.0.0.1 \
TEST_DB_PORT="$POSTGRES_PORT" \
TEST_DB_USER=auto_comic \
TEST_DB_PASSWORD=auto-comic-temporary-password \
TEST_DB_NAME=auto_comic_temp \
go test -tags=integration ./tests -count=1 -timeout=60s "$@"
