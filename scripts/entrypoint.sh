#!/bin/sh
set -e

# Run migrations if MIGRATE is set to true
if [ "$MIGRATE" = "true" ]; then
    echo "Running migrations..."
    /app/migrate up
fi

# Determine which binary to run based on command or defaults
if [ "$1" = "api" ]; then
    echo "Starting API server..."
    exec /app/api
elif [ "$1" = "migrate" ]; then
    echo "Using migration tool..."
    shift
    exec /app/migrate "$@"
else
    # Default behavior: execute whatever command is passed
    exec "$@"
fi
