#!/bin/bash

# Helper script to manually create databases in the running PostgreSQL container
# This is useful if you need to recreate databases or if the init script didn't run

set -e

# Load environment variables
if [ -f docker-compose.env ]; then
    export $(cat docker-compose.env | grep -v '^#' | xargs)
fi

USER_DB="${POSTGRES_DB_USER_SERVICES:-lms_user_services}"
ORDER_DB="${POSTGRES_DB_ORDER_SERVICES:-lms_order_serivecs}"
LESSON_DB="${POSTGRES_DB_LESSON_SERVICES:-lms_lesson_service}"

POSTGRES_USER="${POSTGRES_USER:-user}"
POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-password}"

# Use 'postgres' system database which always exists
CONNECT_DB="postgres"

echo "Creating databases in PostgreSQL container..."
echo "Connecting to: postgres container (using '$CONNECT_DB' system database)"
echo ""

# Check if container is running
if ! docker ps | grep -q "postgres"; then
    echo "❌ PostgreSQL container is not running!"
    echo "Please start it first with: docker-compose up -d postgres"
    exit 1
fi

# Function to create database if it doesn't exist
create_db() {
    local db_name=$1
    echo "Checking database: $db_name"
    
    # Check if database exists (connect to 'postgres' system database)
    exists=$(docker exec postgres psql -U "$POSTGRES_USER" -d "$CONNECT_DB" -tAc "SELECT 1 FROM pg_database WHERE datname='$db_name'" 2>/dev/null || echo "0")
    
    if [ "$exists" = "1" ]; then
        echo "  ✓ Database '$db_name' already exists"
    else
        echo "  → Creating database '$db_name'..."
        docker exec postgres psql -U "$POSTGRES_USER" -d "$CONNECT_DB" -c "CREATE DATABASE \"$db_name\";" 2>&1
        docker exec postgres psql -U "$POSTGRES_USER" -d "$CONNECT_DB" -c "GRANT ALL PRIVILEGES ON DATABASE \"$db_name\" TO \"$POSTGRES_USER\";" 2>&1
        echo "  ✓ Database '$db_name' created"
    fi
}

# Create databases
create_db "$USER_DB"
create_db "$ORDER_DB"
create_db "$LESSON_DB"

echo ""
echo "✅ Databases created/verified successfully:"
echo "   - $USER_DB (user-services)"
echo "   - $ORDER_DB (order-services)"
echo "   - $LESSON_DB (lesson-services)"
echo ""
echo "You can verify by running:"
echo "  docker exec -it postgres psql -U $POSTGRES_USER -c '\l'"

