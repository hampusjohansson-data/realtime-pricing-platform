#!/usr/bin/env bash
set -euo pipefail

DB_DSN="${POSTGRES_DSN:-postgres://pricing:pricing@localhost:5432/pricing?sslmode=disable}"

echo "Running migrations against $DB_DSN"

for file in infra/migrations/*.sql; do
  echo "Applying $file"
  PGPASSWORD=pricing psql "$DB_DSN" -f "$file"
done

echo "Done."

