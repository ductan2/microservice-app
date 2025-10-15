#!/bin/bash

# Database Migration Helper Script
# Usage: ./migrate.sh [command]

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

show_help() {
    echo "Database Migration Helper"
    echo ""
    echo "Usage: ./migrate.sh [command]"
    echo ""
    echo "Commands:"
    echo "  status      - Show current migration status"
    echo "  history     - Show migration history"
    echo "  upgrade     - Apply all pending migrations"
    echo "  downgrade   - Rollback one migration"
    echo "  create      - Create a new migration (requires message)"
    echo "  sql         - Show SQL that would be executed"
    echo "  reset       - Reset database to base (DEVELOPMENT ONLY)"
    echo ""
    echo "Examples:"
    echo "  ./migrate.sh status"
    echo "  ./migrate.sh upgrade"
    echo "  ./migrate.sh create \"Add new column to users table\""
}

case "$1" in
    status)
        echo -e "${GREEN}Current migration status:${NC}"
        alembic current
        ;;
    
    history)
        echo -e "${GREEN}Migration history:${NC}"
        alembic history --verbose
        ;;
    
    upgrade)
        echo -e "${GREEN}Applying migrations...${NC}"
        alembic upgrade head
        echo -e "${GREEN}✓ Migrations applied successfully${NC}"
        ;;
    
    downgrade)
        echo -e "${YELLOW}Rolling back one migration...${NC}"
        read -p "Are you sure? (y/N) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            alembic downgrade -1
            echo -e "${GREEN}✓ Rollback completed${NC}"
        else
            echo -e "${RED}Cancelled${NC}"
        fi
        ;;
    
    create)
        if [ -z "$2" ]; then
            echo -e "${RED}Error: Migration message required${NC}"
            echo "Usage: ./migrate.sh create \"your migration message\""
            exit 1
        fi
        echo -e "${GREEN}Creating new migration...${NC}"
        alembic revision -m "$2"
        echo -e "${GREEN}✓ Migration file created${NC}"
        echo -e "${YELLOW}Don't forget to implement upgrade() and downgrade() functions${NC}"
        ;;
    
    sql)
        echo -e "${GREEN}SQL that would be executed:${NC}"
        alembic upgrade head --sql
        ;;
    
    reset)
        echo -e "${RED}⚠️  WARNING: This will reset the database to base state${NC}"
        echo -e "${RED}This should ONLY be used in development!${NC}"
        read -p "Are you absolutely sure? (type 'yes' to confirm) " -r
        echo
        if [[ $REPLY == "yes" ]]; then
            echo -e "${YELLOW}Resetting database...${NC}"
            alembic downgrade base
            echo -e "${GREEN}✓ Database reset to base${NC}"
        else
            echo -e "${RED}Cancelled${NC}"
        fi
        ;;
    
    help|--help|-h|"")
        show_help
        ;;
    
    *)
        echo -e "${RED}Error: Unknown command '$1'${NC}"
        echo ""
        show_help
        exit 1
        ;;
esac

