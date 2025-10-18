#!/bin/bash

# GophKeeper TUI Client Startup Script

set -e

# Configuration
CLIENT_NAME="gophkeeper-client"
BUILD_DIR="bin"
SOURCE_DIR="cmd/client"
DEFAULT_SERVER="localhost:8082"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo
    echo "Options:"
    echo "  -s, --server ADDRESS    Set server address (default: $DEFAULT_SERVER)"
    echo "  -b, --build            Force rebuild before running"
    echo "  -h, --help             Show this help message"
    echo
    echo "Environment Variables:"
    echo "  GOPHKEEPER_SERVER      Server address to connect to"
    echo
    echo "Examples:"
    echo "  $0                                    # Run with default settings"
    echo "  $0 -s production.example.com:8082    # Connect to production server"
    echo "  $0 -b                                # Force rebuild and run"
    echo "  GOPHKEEPER_SERVER=dev.local:8082 $0  # Use environment variable"
}

# Parse command line arguments
SERVER_ADDR=""
FORCE_BUILD=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -s|--server)
            SERVER_ADDR="$2"
            shift 2
            ;;
        -b|--build)
            FORCE_BUILD=true
            shift
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Check if we're in the right directory
if [[ ! -f "go.mod" ]] || [[ ! -d "$SOURCE_DIR" ]]; then
    print_error "This script must be run from the GophKeeper root directory"
    exit 1
fi

# Create bin directory if it doesn't exist
if [[ ! -d "$BUILD_DIR" ]]; then
    print_info "Creating $BUILD_DIR directory..."
    mkdir -p "$BUILD_DIR"
fi

# Check if binary exists and if we need to build
BINARY_PATH="$BUILD_DIR/$CLIENT_NAME"
NEEDS_BUILD=false

if [[ ! -f "$BINARY_PATH" ]]; then
    print_info "Client binary not found, building..."
    NEEDS_BUILD=true
elif [[ "$FORCE_BUILD" == true ]]; then
    print_info "Force build requested, rebuilding..."
    NEEDS_BUILD=true
elif [[ "$SOURCE_DIR" -nt "$BINARY_PATH" ]]; then
    print_info "Source files are newer than binary, rebuilding..."
    NEEDS_BUILD=true
fi

# Build if needed
if [[ "$NEEDS_BUILD" == true ]]; then
    print_info "Building GophKeeper TUI client..."

    if go build -o "$BINARY_PATH" "./$SOURCE_DIR"; then
        print_success "Build completed successfully"
    else
        print_error "Build failed"
        exit 1
    fi
fi

# Set server address
if [[ -n "$SERVER_ADDR" ]]; then
    export GOPHKEEPER_SERVER="$SERVER_ADDR"
    print_info "Server address set to: $SERVER_ADDR"
elif [[ -n "$GOPHKEEPER_SERVER" ]]; then
    print_info "Using server address from environment: $GOPHKEEPER_SERVER"
else
    export GOPHKEEPER_SERVER="$DEFAULT_SERVER"
    print_info "Using default server address: $DEFAULT_SERVER"
fi

# Check if server is reachable (optional)
print_info "Checking server connectivity..."
if command -v nc >/dev/null 2>&1; then
    SERVER_HOST=$(echo "$GOPHKEEPER_SERVER" | cut -d':' -f1)
    SERVER_PORT=$(echo "$GOPHKEEPER_SERVER" | cut -d':' -f2)

    if nc -z "$SERVER_HOST" "$SERVER_PORT" 2>/dev/null; then
        print_success "Server is reachable at $GOPHKEEPER_SERVER"
    else
        print_warning "Server at $GOPHKEEPER_SERVER may not be available"
        print_warning "Make sure the GophKeeper server is running"
    fi
else
    print_warning "netcat (nc) not available, skipping connectivity check"
fi

# Run the client
print_info "Starting GophKeeper TUI client..."
echo
print_success "🔐 Welcome to GophKeeper!"
echo
print_info "Keyboard shortcuts:"
echo "  • Ctrl+R: Switch between login/register"
echo "  • a: Add new item"
echo "  • d: Delete item"
echo "  • /: Search items"
echo "  • h: Toggle help"
echo "  • q: Quit"
echo
echo "Connecting to server: $GOPHKEEPER_SERVER"
echo "----------------------------------------"

# Execute the client
exec "./$BINARY_PATH"
