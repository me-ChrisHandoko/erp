#!/bin/bash

# Database seeding script for ERP system
# Usage: ./scripts/seed.sh [dev|test]

set -e

# Configuration
SEEDS_DIR="db/seeds"
DB_URL="${DATABASE_URL:?DATABASE_URL environment variable is required}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print colored message
print_message() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Execute SQL file (PostgreSQL only)
execute_sql() {
    local sql_file=$1

    if [ ! -f "$sql_file" ]; then
        print_message "$YELLOW" "Warning: Seed file not found: $sql_file"
        return
    fi

    print_message "$BLUE" "  Executing: $sql_file"
    psql "$DB_URL" < "$sql_file"
}

# Seed development data
seed_dev() {
    local seed_dir="$SEEDS_DIR/dev"

    print_message "$BLUE" "Seeding development data..."

    if [ ! -d "$seed_dir" ]; then
        print_message "$YELLOW" "No development seed data found"
        return
    fi

    # Execute seed files in order
    for sql_file in "$seed_dir"/*.sql; do
        if [ -f "$sql_file" ]; then
            execute_sql "$sql_file"
        fi
    done

    print_message "$GREEN" "✓ Development data seeded successfully"
}

# Seed test data
seed_test() {
    local seed_dir="$SEEDS_DIR/test"

    print_message "$BLUE" "Seeding test data..."

    if [ ! -d "$seed_dir" ]; then
        print_message "$YELLOW" "No test seed data found"
        return
    fi

    # Execute seed files in order
    for sql_file in "$seed_dir"/*.sql; do
        if [ -f "$sql_file" ]; then
            execute_sql "$sql_file"
        fi
    done

    print_message "$GREEN" "✓ Test data seeded successfully"
}

# Show help
show_help() {
    echo "Database Seeding Script"
    echo ""
    echo "Usage: ./scripts/seed.sh [environment]"
    echo ""
    echo "Environments:"
    echo "  dev    Seed development data"
    echo "  test   Seed test data"
    echo "  help   Show this help message"
    echo ""
    echo "Examples:"
    echo "  ./scripts/seed.sh dev"
    echo "  ./scripts/seed.sh test"
    echo ""
}

# Main script
main() {
    local environment=${1:-dev}

    case "$environment" in
        dev)
            seed_dev
            ;;
        test)
            seed_test
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            print_message "$RED" "Error: Unknown environment '$environment'"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

# Run main function
main "$@"
