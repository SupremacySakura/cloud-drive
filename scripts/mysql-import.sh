#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 1 || $# -gt 2 ]]; then
  echo "Usage: $0 <sql-file-path> [database]"
  echo "Example: $0 ./ops/mysql/import/sample.sql cloud-drive"
  exit 1
fi

SQL_FILE="$1"
TARGET_DB="${2:-cloud-drive}"
MYSQL_ROOT_PASSWORD="${MYSQL_ROOT_PASSWORD:-123456123456}"

if [[ ! -f "$SQL_FILE" ]]; then
  echo "SQL file not found: $SQL_FILE"
  exit 1
fi

if ! docker compose ps --status running mysql >/dev/null 2>&1; then
  echo "MySQL container is not running. Start it first:"
  echo "  docker compose up -d mysql"
  exit 1
fi

echo "Importing '$SQL_FILE' into database '$TARGET_DB'..."
docker compose exec -T mysql \
  mysql -uroot -p"${MYSQL_ROOT_PASSWORD}" "${TARGET_DB}" < "$SQL_FILE"

echo "Import completed."
