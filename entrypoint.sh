#!/bin/sh
set -e

echo "Running database migrations..."

# Wait for database to be ready
until pg_isready -h "${POSTGRES_HOST:-db}" -p "${POSTGRES_PORT:-5432}" -U "${POSTGRES_USER}" > /dev/null 2>&1; do
  echo "Waiting for database to be ready..."
  sleep 2
done

echo "Database is ready. Running migrations..."

# Run goose migrations
goose up

echo "Migrations completed successfully!"

# Start the server
echo "Starting server..."
cd /app
exec ./server
