#!/bin/bash

# GophKeeper Development Startup Script
# This script helps diagnose and fix common connection issues

set -e  # Exit on error

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

echo "🔐 GophKeeper Development Setup"
echo "================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default configuration
GRPC_PORT=${GRPC_PORT:-8082}
HTTP_PORT=${HTTP_PORT:-8080}
DATABASE_URI=${DATABASE_URI:-"postgres://postgres:password@localhost:5432/gophkeeper?sslmode=disable"}

echo -e "${BLUE}Configuration:${NC}"
echo "  GRPC_PORT: $GRPC_PORT"
echo "  HTTP_PORT: $HTTP_PORT"
echo "  DATABASE_URI: $DATABASE_URI"
echo ""

# Function to check if a port is in use
check_port() {
    local port=$1
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        return 0  # Port is in use
    else
        return 1  # Port is free
    fi
}

# Function to wait for a port to be available
wait_for_port() {
    local host=$1
    local port=$2
    local timeout=${3:-30}

    echo -e "${YELLOW}Waiting for $host:$port to be ready...${NC}"

    local count=0
    while ! nc -z "$host" "$port" 2>/dev/null; do
        if [ $count -ge $timeout ]; then
            echo -e "${RED}❌ Timeout waiting for $host:$port${NC}"
            return 1
        fi
        count=$((count + 1))
        sleep 1
    done

    echo -e "${GREEN}✅ $host:$port is ready${NC}"
    return 0
}

# Function to start database if not running
start_database() {
    echo -e "${BLUE}Checking database...${NC}"

    # Try to connect to database
    if ! pg_isready -h localhost -p 5432 >/dev/null 2>&1; then
        echo -e "${YELLOW}PostgreSQL not running. Attempting to start with Docker...${NC}"

        # Check if Docker is available
        if ! command -v docker &> /dev/null; then
            echo -e "${RED}❌ Docker not found. Please start PostgreSQL manually or install Docker.${NC}"
            echo "   You can start PostgreSQL with:"
            echo "   docker run --name gophkeeper-db -e POSTGRES_PASSWORD=password -e POSTGRES_DB=gophkeeper -p 5432:5432 -d postgres:15"
            exit 1
        fi

        # Start PostgreSQL container
        docker run --name gophkeeper-db \
            -e POSTGRES_PASSWORD=password \
            -e POSTGRES_DB=gophkeeper \
            -p 5432:5432 \
            -d postgres:15 >/dev/null 2>&1 || {

            # If container already exists, start it
            docker start gophkeeper-db >/dev/null 2>&1 || {
                echo -e "${RED}❌ Failed to start database container${NC}"
                exit 1
            }
        }

        # Wait for database to be ready
        if ! wait_for_port localhost 5432 30; then
            echo -e "${RED}❌ Database failed to start${NC}"
            exit 1
        fi
    fi

    echo -e "${GREEN}✅ Database is ready${NC}"
}

# Function to run database migrations
run_migrations() {
    echo -e "${BLUE}Running database migrations...${NC}"

    # Check if goose is installed
    if ! go list -m | grep "github.com/pressly/goose/v3" >/dev/null 2>&1; then
        echo -e "${YELLOW}Installing goose migration tool...${NC}"
        go install github.com/pressly/goose/v3/cmd/goose@latest
    fi

    # Run migrations
    if goose -dir migrations postgres "$DATABASE_URI" up; then
        echo -e "${GREEN}✅ Migrations completed${NC}"
    else
        echo -e "${RED}❌ Migration failed${NC}"
        exit 1
    fi
}

# Function to build the server
build_server() {
    echo -e "${BLUE}Building server...${NC}"

    if go build -o bin/server ./cmd/server; then
        echo -e "${GREEN}✅ Server built successfully${NC}"
    else
        echo -e "${RED}❌ Server build failed${NC}"
        exit 1
    fi
}

# Function to start the server
start_server() {
    echo -e "${BLUE}Starting server...${NC}"

    # Check if server port is already in use
    if check_port $GRPC_PORT; then
        echo -e "${YELLOW}⚠️  Port $GRPC_PORT is already in use${NC}"

        # Try to find what's using the port
        local pid=$(lsof -ti:$GRPC_PORT)
        if [ -n "$pid" ]; then
            echo "   Process $pid is using port $GRPC_PORT"
            echo "   You may need to stop it with: kill $pid"
        fi

        read -p "Continue anyway? (y/N) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi

    # Set environment variables
    export GRPC_PORT
    export HTTP_PORT
    export DATABASE_URI
    export LOG_LEVEL=${LOG_LEVEL:-DEBUG}
    export SALT_SECRET=${SALT_SECRET:-"dev_salt_secret_change_in_production"}
    export JWT_SECRET=${JWT_SECRET:-"dev_jwt_secret_change_in_production"}

    # Start server in background
    ./bin/server &
    SERVER_PID=$!

    # Wait for server to be ready
    if wait_for_port localhost $GRPC_PORT 30; then
        echo -e "${GREEN}✅ Server started successfully (PID: $SERVER_PID)${NC}"
        echo "   gRPC server listening on port $GRPC_PORT"
        echo "   HTTP gateway listening on port $HTTP_PORT"

        # Create PID file for easy cleanup
        echo $SERVER_PID > .server.pid

        return 0
    else
        echo -e "${RED}❌ Server failed to start${NC}"
        kill $SERVER_PID 2>/dev/null || true
        exit 1
    fi
}

# Function to test the connection
test_connection() {
    echo -e "${BLUE}Testing connection...${NC}"

    # Test basic TCP connection
    if nc -z localhost $GRPC_PORT; then
        echo -e "${GREEN}✅ TCP connection to localhost:$GRPC_PORT successful${NC}"
    else
        echo -e "${RED}❌ Cannot connect to localhost:$GRPC_PORT${NC}"
        return 1
    fi

    # Test HTTP gateway if available
    if nc -z localhost $HTTP_PORT; then
        echo -e "${GREEN}✅ HTTP gateway available at localhost:$HTTP_PORT${NC}"
    fi
}

# Function to cleanup on exit
cleanup() {
    echo -e "\n${YELLOW}Cleaning up...${NC}"

    # Kill server if running
    if [ -f .server.pid ]; then
        local pid=$(cat .server.pid)
        if kill -0 $pid 2>/dev/null; then
            echo "Stopping server (PID: $pid)"
            kill $pid
            # Wait for graceful shutdown
            sleep 2
            # Force kill if still running
            if kill -0 $pid 2>/dev/null; then
                kill -9 $pid
            fi
        fi
        rm -f .server.pid
    fi

    echo -e "${GREEN}✅ Cleanup completed${NC}"
}

# Set trap for cleanup
trap cleanup EXIT INT TERM

# Main execution flow
main() {
    case "${1:-start}" in
        "start")
            start_database
            run_migrations
            build_server
            start_server
            test_connection

            echo ""
            echo -e "${GREEN}🎉 Development environment is ready!${NC}"
            echo ""
            echo "You can now run the client with:"
            echo "  go run ./cmd/client"
            echo ""
            echo "Or build and run the client separately:"
            echo "  go build -o bin/client ./cmd/client"
            echo "  ./bin/client"
            echo ""
            echo "Press Ctrl+C to stop the server and cleanup"

            # Wait for interrupt
            wait $SERVER_PID
            ;;

        "stop")
            if [ -f .server.pid ]; then
                local pid=$(cat .server.pid)
                if kill -0 $pid 2>/dev/null; then
                    kill $pid
                    echo -e "${GREEN}✅ Server stopped${NC}"
                else
                    echo -e "${YELLOW}Server not running${NC}"
                fi
                rm -f .server.pid
            else
                echo -e "${YELLOW}No server PID file found${NC}"
            fi
            ;;

        "test")
            test_connection
            ;;

        "help")
            echo "Usage: $0 [start|stop|test|help]"
            echo ""
            echo "Commands:"
            echo "  start (default) - Start the development environment"
            echo "  stop           - Stop the server"
            echo "  test           - Test connection to server"
            echo "  help           - Show this help message"
            echo ""
            echo "Environment variables:"
            echo "  GRPC_PORT      - gRPC server port (default: 8082)"
            echo "  HTTP_PORT      - HTTP gateway port (default: 8080)"
            echo "  DATABASE_URI   - PostgreSQL connection string"
            ;;

        *)
            echo -e "${RED}Unknown command: $1${NC}"
            echo "Use '$0 help' for usage information"
            exit 1
            ;;
    esac
}

# Check prerequisites
echo -e "${BLUE}Checking prerequisites...${NC}"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go is not installed${NC}"
    exit 1
fi

# Check if netcat is available for port checking
if ! command -v nc &> /dev/null; then
    echo -e "${YELLOW}⚠️  netcat not found, some checks may be limited${NC}"
fi

echo -e "${GREEN}✅ Prerequisites check passed${NC}"
echo ""

# Run main function
main "$@"
