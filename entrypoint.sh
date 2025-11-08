#!/bin/sh
set -e

echo "Waiting for PostgreSQL..."

until PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" -c '\q' 2>/dev/null; do
  echo "PostgreSQL unavailable - sleeping"
  sleep 2
done

echo "PostgreSQL is ready"
echo "Running migrations..."

for migration in /app/migrations/*.sql; do
  if [ -f "$migration" ]; then
    echo "Applying: $(basename $migration)"
    PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" -f "$migration"
  fi
done

echo "Migrations completed - starting application"

exec "$@"
