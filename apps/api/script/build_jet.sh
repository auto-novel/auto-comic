#!/usr/bin/env bash

set -euo pipefail

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
source "$script_dir/postgres.sh"
start_postgres jet "${JET_DB_PORT:-}"

# Read only the real schema created by root sql/init.sql; never create fake tables.
table_count="$(docker exec "$POSTGRES_CONTAINER" psql -X -U auto_comic -d auto_comic_temp -At -v ON_ERROR_STOP=1 \
  -c "SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE'")"
if [[ "$table_count" == 0 ]]; then
  echo "public schema has no business tables; skipping Jet code generation (no fake tables created)."
  exit 0
fi

cd -- "$script_dir/.."
# Version-pinned go run avoids depending on a globally installed Jet binary.
go run github.com/go-jet/jet/v2/cmd/jet@v2.13.0 \
  -dsn="postgresql://auto_comic:auto-comic-temporary-password@127.0.0.1:$POSTGRES_PORT/auto_comic_temp?sslmode=disable" \
  -schema=public -path=./.gen
