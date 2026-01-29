#!/bin/bash

# Usage: ./scripts/db-reset.sh [postgres://user:pass@host:port/dbname]

DB_URL=$1

if [ -z "$DB_URL" ]; then
    echo "Usage: ./scripts/db-reset.sh <DATABASE_URL>"
    echo "Example: ./scripts/db-reset.sh postgres://admin:password@localhost:5432/madabank?sslmode=disable"
    echo ""
    echo "⚠️  WARNING: This will DESTROY all data in the database!"
    exit 1
fi

echo "🛑 CAUTION: You are about to RESET the database at:"
echo "   $DB_URL"
echo ""
read -p "Are you sure? (Type 'yes' to confirm): " confirm

if [ "$confirm" != "yes" ]; then
    echo "Cancelled."
    exit 1
fi

echo "Running Migrate DOWN (Drop all tables)..."
export DATABASE_URL=$DB_URL
go run cmd/migrate/main.go -down

if [ $? -ne 0 ]; then
    echo "❌ Migrate Down failed!"
    exit 1
fi

echo "Running Migrate UP (Re-create tables)..."
go run cmd/migrate/main.go -up

if [ $? -ne 0 ]; then
    echo "❌ Migrate Up failed!"
    exit 1
fi

echo "✅ Database reset successfully!"
