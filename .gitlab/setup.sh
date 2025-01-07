#!/bin/bash
set -e  # Stop the script if an error occurs

# Function to build the project
build() {
  echo "Building the project..."
  go build -v ./...
}

# Function to run tests with JUnit report
test() {
  echo "Running tests with gotestsum..."
  go install gotest.tools/gotestsum@latest
  gotestsum --junitfile report.xml --format testname

  echo "Generating coverage report..."
  CGO_ENABLED=0 go test ./... -coverprofile=coverage-report.out
  go tool cover -html=coverage-report.out -o coverage-report.html
  go tool cover -func=coverage-report.out

  echo "Converting coverage to Cobertura format..."
  go test ./... -coverprofile=coverage.txt -covermode count
  go install github.com/boumenot/gocover-cobertura@latest
  gocover-cobertura <coverage.txt >coverage.xml
}

# Function to run static code analysis
lint() {
  echo "Running linter (golangci-lint)..."
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
  golangci-lint run --issues-exit-code 0 --print-issued-lines=false --out-format code-climate:gl-code-quality-report.json,line-number
}

# Help message
help() {
  echo "Usage: $0 {build|test|lint}"
  echo "  build   Build the project"
  echo "  test    Run tests and generate coverage reports"
  echo "  lint    Run static code analysis"
}

# Main script logic
case "$1" in
  build)
    build
    ;;
  test)
    test
    ;;
  lint)
    lint
    ;;
  *)
    echo "Error: Invalid command"
    help
    exit 1
    ;;
esac
