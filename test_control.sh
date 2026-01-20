#!/bin/bash
set -e

echo "Running control package tests..."
cd /mnt/code/jamesbot

echo ""
echo "=== Running control tests with verbose output ==="
go test -v ./internal/control/... 2>&1 | grep -E "(PASS|FAIL|---)"

echo ""
echo "=== Running race detector on all packages ==="
go test -race ./...

echo ""
echo "All tests passed successfully!"
