#!/usr/bin/env bash
# Shared implementation for disposable integration/code-generation databases.
# Source this file; call start_postgres <purpose> <optional-host-port>.

start_postgres() {
  local purpose="$1" requested_port="$2" status
  local script_dir project_root
  script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
  project_root="$(cd -- "$script_dir/../../.." && pwd)"

  if [[ -n "$requested_port" ]] && { [[ ! "$requested_port" =~ ^[0-9]{1,5}$ ]] || (( 10#$requested_port > 65535 )); }; then
    echo "Database port must be empty (dynamic) or between 0 and 65535" >&2
    return 1
  fi
  command -v docker >/dev/null
  command -v go >/dev/null
  local container_name="auto-comic-${purpose}-postgresql-$$-${RANDOM}"
  # Keep the name available to the EXIT trap after this function returns.
  POSTGRES_CONTAINER="$container_name"
  trap 'docker rm --force "$POSTGRES_CONTAINER" >/dev/null 2>&1 || true' EXIT
  trap 'exit 130' INT
  trap 'exit 143' TERM

  docker run --rm --detach \
    --name "$container_name" \
    --publish "127.0.0.1:${requested_port}:5432" \
    --env POSTGRES_USER=auto_comic \
    --env POSTGRES_PASSWORD=auto-comic-temporary-password \
    --env POSTGRES_DB=auto_comic_temp \
    --tmpfs /var/lib/postgresql/data \
    --volume "$project_root/sql/init.sql:/docker-entrypoint-initdb.d/init.sql:ro" \
    --health-cmd 'pg_isready -h 127.0.0.1 -U auto_comic -d auto_comic_temp' \
    --health-interval 1s --health-timeout 3s --health-retries 30 \
    postgres:17.7-alpine3.23 >/dev/null

  # TCP health checks do not accept the socket-only initialization server:
  # init.sql must finish before the final PostgreSQL server becomes healthy.
  for ((attempt = 0; attempt < 60; attempt++)); do
    status="$(docker inspect --format '{{.State.Health.Status}}' "$container_name")"
    case "$status" in
      healthy) break ;;
      unhealthy)
        docker logs "$container_name" >&2
        echo "Temporary PostgreSQL became unhealthy" >&2
        return 1
        ;;
    esac
    sleep 1
  done
  if [[ "$status" != healthy ]]; then
    docker logs "$container_name" >&2
    echo "Timed out waiting for temporary PostgreSQL (60 attempts)" >&2
    return 1
  fi

  POSTGRES_PORT="$(docker inspect --format '{{(index (index .NetworkSettings.Ports "5432/tcp") 0).HostPort}}' "$container_name")"
  echo "Temporary PostgreSQL ready on 127.0.0.1:$POSTGRES_PORT"
}
