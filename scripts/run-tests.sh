#!/bin/bash

# GophKeeper Test Runner Script
# Runs unit tests and integration tests for the client package

set -e

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

echo "🧪 GophKeeper Test Runner"
echo "========================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configuration
UNIT_TESTS_ONLY=${UNIT_TESTS_ONLY:-false}
INTEGRATION_TESTS_ONLY=${INTEGRATION_TESTS_ONLY:-false}
COVERAGE=${COVERAGE:-true}
VERBOSE=${VERBOSE:-false}
RACE=${RACE:-true}
TIMEOUT=${TIMEOUT:-30s}

# Function to print usage
print_usage() {
    echo "Usage: $0 [options]"
    echo ""
    echo "Options:"
    echo "  -u, --unit-only        Run only unit tests"
    echo "  -i, --integration-only Run only integration tests"
    echo "  -c, --no-coverage      Disable coverage reporting"
    echo "  -v, --verbose          Enable verbose output"
    echo "  -r, --no-race          Disable race detection"
    echo "  -t, --timeout DURATION Set test timeout (default: 30s)"
    echo "  -h, --help             Show this help message"
    echo ""
    echo "Environment variables:"
    echo "  GOPHKEEPER_TEST_SERVER - Test server address (default: localhost:8082)"
    echo "  RUN_INTEGRATION_TESTS  - Set to 1 to enable integration tests"
    echo ""
    echo "Examples:"
    echo "  $0                     # Run all tests with coverage"
    echo "  $0 -u                  # Run only unit tests"
    echo "  $0 -i                  # Run only integration tests"
    echo "  $0 -v -c               # Run all tests verbose without coverage"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -u|--unit-only)
            UNIT_TESTS_ONLY=true
            shift
            ;;
        -i|--integration-only)
            INTEGRATION_TESTS_ONLY=true
            shift
            ;;
        -c|--no-coverage)
            COVERAGE=false
            shift
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -r|--no-race)
            RACE=false
            shift
            ;;
        -t|--timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        -h|--help)
            print_usage
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            print_usage
            exit 1
            ;;
    esac
done

# Build test flags
TEST_FLAGS=""
if [ "$VERBOSE" = true ]; then
    TEST_FLAGS="$TEST_FLAGS -v"
fi

if [ "$RACE" = true ]; then
    TEST_FLAGS="$TEST_FLAGS -race"
fi

TEST_FLAGS="$TEST_FLAGS -timeout $TIMEOUT"

# Coverage flags
COVERAGE_FLAGS=""
if [ "$COVERAGE" = true ]; then
    COVERAGE_FLAGS="-coverprofile=coverage.out -covermode=atomic"
fi

# Function to run unit tests
run_unit_tests() {
    echo -e "${BLUE}Running unit tests...${NC}"
    echo ""

    # Run tests excluding integration tests
    go test $TEST_FLAGS $COVERAGE_FLAGS -tags="!integration" ./client/

    local exit_code=$?
    if [ $exit_code -eq 0 ]; then
        echo -e "${GREEN}✅ Unit tests passed${NC}"
    else
        echo -e "${RED}❌ Unit tests failed${NC}"
        return $exit_code
    fi

    echo ""
}

# Function to check if server is available
check_server() {
    local server_addr=${GOPHKEEPER_TEST_SERVER:-"localhost:8082"}

    echo -e "${BLUE}Checking test server at $server_addr...${NC}"

    if timeout 5 bash -c "</dev/tcp/${server_addr%:*}/${server_addr#*:}" 2>/dev/null; then
        echo -e "${GREEN}✅ Test server is available${NC}"
        return 0
    else
        echo -e "${YELLOW}⚠️  Test server not available at $server_addr${NC}"
        return 1
    fi
}

# Function to run integration tests
run_integration_tests() {
    echo -e "${BLUE}Running integration tests...${NC}"
    echo ""

    # Check if server is available
    if ! check_server; then
        echo -e "${YELLOW}Integration tests require a running server.${NC}"
        echo "Start the server with:"
        echo "  ./scripts/start-dev.sh"
        echo ""
        echo "Or set GOPHKEEPER_TEST_SERVER to point to an existing server:"
        echo "  export GOPHKEEPER_TEST_SERVER=your-server:8082"
        echo ""
        return 1
    fi

    # Set environment variable to enable integration tests
    export RUN_INTEGRATION_TESTS=1

    # Run integration tests
    go test $TEST_FLAGS -tags="integration" ./client/

    local exit_code=$?
    if [ $exit_code -eq 0 ]; then
        echo -e "${GREEN}✅ Integration tests passed${NC}"
    else
        echo -e "${RED}❌ Integration tests failed${NC}"
        return $exit_code
    fi

    echo ""
}

# Function to generate coverage report
generate_coverage_report() {
    if [ "$COVERAGE" = true ] && [ -f coverage.out ]; then
        echo -e "${BLUE}Generating coverage report...${NC}"

        # Generate HTML coverage report
        go tool cover -html=coverage.out -o coverage.html

        # Show coverage summary
        go tool cover -func=coverage.out | tail -1

        echo -e "${GREEN}✅ Coverage report generated: coverage.html${NC}"
        echo ""
    fi
}

# Function to run benchmarks
run_benchmarks() {
    echo -e "${BLUE}Running benchmarks...${NC}"
    echo ""

    go test -bench=. -benchmem ./client/

    echo -e "${GREEN}✅ Benchmarks completed${NC}"
    echo ""
}

# Main execution
main() {
    # Check prerequisites
    echo -e "${BLUE}Checking prerequisites...${NC}"

    if ! command -v go &> /dev/null; then
        echo -e "${RED}❌ Go is not installed${NC}"
        exit 1
    fi

    echo -e "${GREEN}✅ Prerequisites check passed${NC}"
    echo ""

    # Ensure we're in the right directory
    if [ ! -f "go.mod" ]; then
        echo -e "${RED}❌ Not in project root directory${NC}"
        exit 1
    fi

    # Clean previous coverage data
    rm -f coverage.out coverage.html

    # Determine which tests to run
    if [ "$INTEGRATION_TESTS_ONLY" = true ]; then
        run_integration_tests
    elif [ "$UNIT_TESTS_ONLY" = true ]; then
        run_unit_tests
    else
        # Run both unit and integration tests
        run_unit_tests
        if [ $? -eq 0 ]; then
            run_integration_tests
        else
            echo -e "${RED}❌ Unit tests failed, skipping integration tests${NC}"
            exit 1
        fi
    fi

    # Generate coverage report
    generate_coverage_report

    echo -e "${GREEN}🎉 All tests completed successfully!${NC}"
}

# Error handling
trap 'echo -e "${RED}Test run interrupted${NC}"; exit 1' INT TERM

# Run main function
main "$@"
