#!/bin/bash
set -e

echo "Ensuring dependencies..."
go mod tidy

echo "Running Tests..."
go test -v ./internal/...

echo "All tests passed!"
