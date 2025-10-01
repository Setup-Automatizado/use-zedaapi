#!/bin/bash
set -e

# Script to initialize multiple databases in PostgreSQL
# This runs automatically when the container starts for the first time

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    -- Create whatsmeow_store database
    CREATE DATABASE whatsmeow_store;
    GRANT ALL PRIVILEGES ON DATABASE whatsmeow_store TO $POSTGRES_USER;

    -- Connect to whatsmeow_store and create extensions
    \c whatsmeow_store;
    CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

    -- Connect back to api_core and create extensions
    \c $POSTGRES_DB;
    CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

    -- Log completion
    SELECT 'Databases initialized successfully' AS status;
EOSQL

echo "✅ PostgreSQL databases initialized: api_core, whatsmeow_store"
