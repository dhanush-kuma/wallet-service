#!/bin/sh
set -e

echo "waiting for database..."

sleep 2

echo "running migrations..."
migrate -path /app/migrations \
  -database "$DATABASE_URL" up

echo "running seed..."
psql "$DATABASE_URL" -f /app/migrations/seed.sql || true

echo "starting app..."
exec ./wallet-service