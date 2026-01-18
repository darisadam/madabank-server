#!/bin/sh
set -e

# Run migrations
echo "Running database migrations..."
if [ -f "./migrate" ]; then
    ./migrate up
else
    echo "Migration binary not found, skipping migrations."
fi

# Start the application
echo "Starting application..."
exec "$@"
