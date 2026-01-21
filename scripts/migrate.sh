#!/bin/bash

set -e

if [ -z "$DATABASE_URL" ]; then
    echo "Error: DATABASE_URL environment variable is not set"
    exit 1
fi

echo "Running database migrations..."

for file in sql/migrations/*.sql; do
    if [ -f "$file" ]; then
        echo "Running $file..."
        psql "$DATABASE_URL" -f "$file"
    fi
done

echo "Migrations completed successfully"
