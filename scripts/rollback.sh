#!/bin/bash

set -e

if [ -z "$DATABASE_URL" ]; then
    echo "Error: DATABASE_URL environment variable is not set"
    exit 1
fi

echo "Dropping all tables..."
psql "$DATABASE_URL" -c "DROP TABLE IF EXISTS transactions CASCADE;"
psql "$DATABASE_URL" -c "DROP TABLE IF EXISTS categories CASCADE;"
psql "$DATABASE_URL" -c "DROP TABLE IF EXISTS users CASCADE;"

echo "Tables dropped successfully"
