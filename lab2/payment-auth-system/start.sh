#!/bin/sh
set -e

if [ ! -f /app/data/app.db ]; then
    touch /app/data/app.db
fi

/app/goose -dir /app/migrations sqlite3 /app/data/app.db up

exec /app/server
