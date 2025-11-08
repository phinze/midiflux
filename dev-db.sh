#!/usr/bin/env bash
# Development database helper script for Miniflux

set -e

CONTAINER_NAME="miniflux-dev-db"
DB_NAME="miniflux2"
DB_USER="postgres"
DB_PASSWORD="postgres"
DB_PORT="5432"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

function print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

function print_error() {
    echo -e "${RED}✗${NC} $1"
}

function print_info() {
    echo -e "${YELLOW}ℹ${NC} $1"
}

function start_db() {
    if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
        if docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
            print_info "Database is already running"
            show_connection_info
            return 0
        else
            print_info "Starting existing container..."
            docker start "${CONTAINER_NAME}"
            print_success "Database started"
            show_connection_info
            return 0
        fi
    fi

    print_info "Creating new database container..."
    docker run -d \
        --name "${CONTAINER_NAME}" \
        -p "${DB_PORT}:5432" \
        -e POSTGRES_DB="${DB_NAME}" \
        -e POSTGRES_USER="${DB_USER}" \
        -e POSTGRES_PASSWORD="${DB_PASSWORD}" \
        postgres:latest

    print_success "Database container created and started"
    print_info "Waiting for database to be ready..."
    sleep 3
    show_connection_info
}

function stop_db() {
    if ! docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
        print_info "Database is not running"
        return 0
    fi

    print_info "Stopping database..."
    docker stop "${CONTAINER_NAME}"
    print_success "Database stopped"
}

function destroy_db() {
    if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
        print_info "Removing database container (this will delete all data)..."
        docker rm -f "${CONTAINER_NAME}"
        print_success "Database container removed"
    else
        print_info "No database container to remove"
    fi
}

function status_db() {
    if docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
        print_success "Database is running"
        show_connection_info
    elif docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
        print_info "Database container exists but is not running"
        echo "  Run: $0 start"
    else
        print_info "Database container does not exist"
        echo "  Run: $0 start"
    fi
}

function show_connection_info() {
    echo ""
    echo "Database Connection Info:"
    echo "  Host:     localhost"
    echo "  Port:     ${DB_PORT}"
    echo "  Database: ${DB_NAME}"
    echo "  User:     ${DB_USER}"
    echo "  Password: ${DB_PASSWORD}"
    echo ""
    echo "DATABASE_URL:"
    echo "  postgres://${DB_USER}:${DB_PASSWORD}@localhost:${DB_PORT}/${DB_NAME}?sslmode=disable"
    echo ""
    echo "Connect with psql:"
    echo "  docker exec -it ${CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME}"
}

function logs_db() {
    if ! docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
        print_error "Database is not running"
        return 1
    fi

    docker logs -f "${CONTAINER_NAME}"
}

function psql_db() {
    if ! docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
        print_error "Database is not running"
        return 1
    fi

    docker exec -it "${CONTAINER_NAME}" psql -U "${DB_USER}" -d "${DB_NAME}"
}

function show_help() {
    cat << EOF
Development Database Helper for Miniflux

Usage: $0 <command>

Commands:
  start       Start the development database (creates if needed)
  stop        Stop the development database
  restart     Restart the development database
  destroy     Remove the database container (deletes all data!)
  status      Show database status
  logs        Show database logs (follow mode)
  psql        Connect to database with psql
  info        Show connection information
  help        Show this help message

Examples:
  $0 start           # Start the database
  $0 stop            # Stop the database
  $0 psql            # Connect to the database
  $0 destroy         # Remove everything and start fresh

EOF
}

# Main command dispatcher
case "${1:-}" in
    start)
        start_db
        ;;
    stop)
        stop_db
        ;;
    restart)
        stop_db
        start_db
        ;;
    destroy)
        destroy_db
        ;;
    status)
        status_db
        ;;
    logs)
        logs_db
        ;;
    psql)
        psql_db
        ;;
    info)
        show_connection_info
        ;;
    help|--help|-h)
        show_help
        ;;
    "")
        print_error "No command specified"
        echo ""
        show_help
        exit 1
        ;;
    *)
        print_error "Unknown command: $1"
        echo ""
        show_help
        exit 1
        ;;
esac
