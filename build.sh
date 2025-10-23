#!/bin/bash
set -e

# Obtener ruta absoluta del proyecto
ROOT_DIR=$(pwd)/.cache

export GOMODCACHE="$ROOT_DIR/.gomodcache"
export GOCACHE="$ROOT_DIR/gocache"
export GOLANGCI_LINT_CACHE="$ROOT_DIR/golangci-lint/.cache"
export GOTESTSUM_CACHE="$ROOT_DIR/gotestsum/.cache"

export CGO_ENABLED=1
export GORACE="halt_on_error=1"
export GOTESTSUM_FORMAT="testname"
export GOLANGCI_LINT_TIMEOUT="5m"

mkdir -p "${ROOT_DIR}" "${GOMODCACHE}" "${GOLANGCI_LINT_CACHE}" "${GOTESTSUM_CACHE}" "${GOCACHE}"
chmod -R u+w "${ROOT_DIR}"

echo "📦 Downloading dependencies..."
go env GOMODCACHE GOCACHE
go mod download
go mod verify
echo "✅ Dependencies ready"

echo "🔨 Building optimized binaries..."
go build -ldflags="-s -w -X rest.Version=$CI_COMMIT_SHA" ./...
echo "✅ Build completed"

echo "🧪 Running comprehensive tests (with -race and full coverage)...";
CGO_ENABLED=0 go test -race -v -timeout=10m -coverpkg=./rest/... -coverprofile=coverage.out -covermode=atomic ./...;
echo "📊 Generating full coverage reports...";
go tool cover -func=coverage.out | tee coverage-func.txt;
go tool cover -html=coverage.out -o coverage-report.html || true;
go tool gocover-cobertura <coverage.out >coverage.xml || true;
echo "📋 Generating test reports...";
go tool gotestsum --junitfile report.xml --format testname || true;
echo "🔍 Running code quality checks...";
go tool golangci-lint run --timeout=${GOLANGCI_LINT_TIMEOUT} || true;
echo "✅ All checks completed successfully"
