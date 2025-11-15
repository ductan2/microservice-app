#!/bin/bash
set -e

# This script creates the databases for each microservice
# It runs automatically when PostgreSQL container starts for the first time
# Note: This script only runs when the data directory is empty (first initialization)

USER_DB="${POSTGRES_DB_USER_SERVICES:-lms_english_app}"
ORDER_DB="${POSTGRES_DB_ORDER_SERVICES:-lms_order_serivecs}"
LESSON_DB="${POSTGRES_DB_LESSON_SERVICES:-lms_lesson_service}"

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    -- Create database for user-services
    CREATE DATABASE "$USER_DB";
    
    -- Create database for order-services
    CREATE DATABASE "$ORDER_DB";
    
    -- Create database for lesson-services
    CREATE DATABASE "$LESSON_DB";
    
    -- Grant all privileges to the postgres user
    GRANT ALL PRIVILEGES ON DATABASE "$USER_DB" TO "$POSTGRES_USER";
    GRANT ALL PRIVILEGES ON DATABASE "$ORDER_DB" TO "$POSTGRES_USER";
    GRANT ALL PRIVILEGES ON DATABASE "$LESSON_DB" TO "$POSTGRES_USER";
EOSQL

echo "✅ Databases created successfully:"
echo "   - $USER_DB (user-services)"
echo "   - $ORDER_DB (order-services)"
echo "   - $LESSON_DB (lesson-services)"

