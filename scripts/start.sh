#!/bin/sh
set -e

echo "waiting for database..."

until pg_isready -d "$DATABASE_URL"; do
  echo "Postgres is unavailable - sleeping"
  sleep 2
done

echo "running migrations..."
migrate -path /app/migrations \
  -database "$DATABASE_URL" up

echo "running seed..."
psql "$DATABASE_URL" -f ./migrations/seed.sql || true

echo "starting app..."
exec ./wallet-service