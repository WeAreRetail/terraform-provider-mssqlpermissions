#!/bin/bash

# Test runner script for terraform-provider-mssqlpermissions
# This script provides easy commands to run different types of tests

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
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
show_help() {
  echo "Test runner for terraform-provider-mssqlpermissions"
  echo ""
  echo "Usage: $0 [COMMAND]"
  echo ""
  echo "Commands:"
  echo "  unit          Run only unit tests (fast, no database required)"
  echo "  integration   Run only integration tests (requires database setup)"
  echo "  all           Run all tests (unit + integration)"
  echo "  coverage      Run unit tests with coverage report"
  echo "  help          Show this help message"
  echo ""
  echo "Examples:"
  echo "  $0 unit                    # Run unit tests"
  echo "  $0 integration            # Run integration tests"
  echo "  $0 coverage               # Unit tests with coverage"
  echo ""
  echo "Test Architecture:"
  echo "  Unit tests: Fast tests with no external dependencies"
  echo "    - Validation functions"
  echo "    - Business logic with mocks"
  echo "    - Parameter mutation prevention"
  echo ""
  echo "  Integration tests: Full end-to-end tests"
  echo "    - Require SQL Server database connection"
  echo "    - Test complete workflows"
  echo "    - Environment setup needed"
}

# Change to project directory
cd "$(dirname "$0")"

case "${1:-help}" in
  "unit")
    print_status "Running unit tests..."
    print_status "These tests run quickly without database dependencies"
    echo ""
    go test -v ./internal/queries -run "Test.*_Unit" || (print_error "Unit tests failed" && exit 1)
    echo ""
    print_success "Unit tests completed successfully!"
    test_count=$(go test ./internal/queries -run "Test.*_Unit" 2>/dev/null | grep -o "PASS" | wc -l || echo "0")
    print_status "$test_count unit tests passed"
    ;;

  "integration")
    print_status "Running integration tests..."
    print_warning "These tests require a SQL Server database connection"
    print_status "Make sure your test environment is properly configured"
    echo ""
    go test -tags=integration -v ./internal/queries || (print_error "Integration tests failed" && exit 1)
    echo ""
    print_success "Integration tests completed successfully!"
    ;;

  "all")
    print_status "Running all tests (unit + integration)..."
    echo ""

    print_status "Step 1: Running unit tests..."
    go test -v ./internal/queries -run "Test.*_Unit" || (print_error "Unit tests failed" && exit 1)
    print_success "Unit tests passed!"
    echo ""

    print_status "Step 2: Running integration tests..."
    print_warning "Integration tests require database connection"
    go test -tags=integration -v ./internal/queries || (print_error "Integration tests failed" && exit 1)
    print_success "Integration tests passed!"
    echo ""

    print_success "All tests completed successfully!"
    ;;

  "coverage")
    print_status "Running unit tests with coverage report..."
    echo ""
    go test -v -coverprofile=coverage.out ./internal/queries -run "Test.*_Unit"

    if [ -f coverage.out ]; then
      echo ""
      print_status "Coverage report:"
      go tool cover -func=coverage.out | tail -1
      echo ""
      print_status "Generate HTML coverage report with:"
      print_status "go tool cover -html=coverage.out -o coverage.html"
    fi
    ;;

  "help")
    show_help
    ;;

  *)
    print_error "Unknown command: $1"
    echo ""
    show_help
    exit 1
    ;;
esac
