#!/bin/bash

# GophKeeper Client Debug Script
# This script helps diagnose client connection issues

set -e

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

echo "🔐 GophKeeper Client Diagnostics"
echo "================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default configuration
GRPC_PORT=${GRPC_PORT:-8082}
SERVER_ADDR=${GOPHKEEPER_SERVER:-"localhost:$GRPC_PORT"}

echo -e "${BLUE}Configuration:${NC}"
echo "  Server Address: $SERVER_ADDR"
echo ""

# Function to check if a port is reachable
check_connectivity() {
    local host_port=$1
    local host=$(echo "$host_port" | cut -d: -f1)
    local port=$(echo "$host_port" | cut -d: -f2)

    echo -e "${BLUE}Testing connectivity to $host:$port...${NC}"

    # Test basic TCP connection
    if timeout 5 bash -c "</dev/tcp/$host/$port" 2>/dev/null; then
        echo -e "${GREEN}✅ TCP connection successful${NC}"
        return 0
    else
        echo -e "${RED}❌ Cannot connect to $host:$port${NC}"
        return 1
    fi
}

# Function to check if server is running locally
check_local_server() {
    echo -e "${BLUE}Checking for local server process...${NC}"

    if pgrep -f "cmd/server" > /dev/null || pgrep -f "bin/server" > /dev/null; then
        echo -e "${GREEN}✅ Server process found${NC}"
        return 0
    else
        echo -e "${YELLOW}⚠️  No server process found locally${NC}"
        return 1
    fi
}

# Function to build client
build_client() {
    echo -e "${BLUE}Building client...${NC}"

    if go build -o bin/client ./cmd/client; then
        echo -e "${GREEN}✅ Client built successfully${NC}"
        return 0
    else
        echo -e "${RED}❌ Client build failed${NC}"
        return 1
    fi
}

# Function to run diagnostics
run_diagnostics() {
    echo -e "${BLUE}Running connection diagnostics...${NC}"
    echo ""

    # Check if server is running locally
    if ! check_local_server; then
        echo -e "${YELLOW}Suggestion: Start the server first with:${NC}"
        echo "  ./scripts/start-dev.sh"
        echo "  or"
        echo "  go run ./cmd/server"
        echo ""
    fi

    # Check connectivity
    if ! check_connectivity "$SERVER_ADDR"; then
        echo ""
        echo -e "${RED}Connection failed!${NC}"
        echo ""
        echo -e "${YELLOW}Troubleshooting steps:${NC}"
        echo "1. Make sure the server is running:"
        echo "   ./scripts/start-dev.sh"
        echo ""
        echo "2. Check if the server is listening on the correct port:"
        echo "   netstat -tlnp | grep $GRPC_PORT"
        echo ""
        echo "3. Verify server address configuration:"
        echo "   export GOPHKEEPER_SERVER=\"localhost:$GRPC_PORT\""
        echo ""
        echo "4. Check firewall settings (if using a remote server)"
        echo ""
        return 1
    fi

    echo ""
    echo -e "${GREEN}✅ Connection diagnostics passed${NC}"
    return 0
}

# Function to run client with detailed logging
run_client() {
    echo -e "${BLUE}Starting client...${NC}"
    echo ""
    echo -e "${YELLOW}Note: Server logs will help diagnose authentication issues${NC}"
    echo -e "${YELLOW}Enable verbose logging with: export LOG_LEVEL=DEBUG${NC}"
    echo ""

    # Set environment for better error reporting
    export GOPHKEEPER_SERVER="$SERVER_ADDR"

    # Run client
    if [ -f bin/client ]; then
        ./bin/client
    else
        go run ./cmd/client
    fi
}

# Main execution
main() {
    case "${1:-run}" in
        "run")
            build_client
            run_diagnostics
            if [ $? -eq 0 ]; then
                run_client
            else
                exit 1
            fi
            ;;

        "diag" | "diagnostics")
            run_diagnostics
            ;;

        "build")
            build_client
            ;;

        "help")
            echo "Usage: $0 [run|diag|build|help]"
            echo ""
            echo "Commands:"
            echo "  run (default) - Build and run client with diagnostics"
            echo "  diag         - Run connection diagnostics only"
            echo "  build        - Build client only"
            echo "  help         - Show this help message"
            echo ""
            echo "Environment variables:"
            echo "  GOPHKEEPER_SERVER - Server address (default: localhost:8082)"
            echo "  GRPC_PORT        - Server port (default: 8082)"
            echo ""
            echo "Common issues and solutions:"
            echo ""
            echo "1. 'context canceled' error:"
            echo "   - Server is not running or not reachable"
            echo "   - Start server: ./scripts/start-dev.sh"
            echo ""
            echo "2. 'connection refused' error:"
            echo "   - Server is not listening on the expected port"
            echo "   - Check server logs and port configuration"
            echo ""
            echo "3. TLS/certificate errors:"
            echo "   - Development uses self-signed certificates"
            echo "   - This is normal and expected for local development"
            ;;

        *)
            echo -e "${RED}Unknown command: $1${NC}"
            echo "Use '$0 help' for usage information"
            exit 1
            ;;
    esac
}

# Check prerequisites
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go is not installed${NC}"
    exit 1
fi

echo ""

# Run main function
main "$@"
