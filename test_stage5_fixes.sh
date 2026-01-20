#!/bin/bash
# Test script to verify Stage 5 fixes

set -e

echo "=== Testing Stage 5 Fixes ==="
echo ""

cd /mnt/code/jamesbot

echo "Running CLI tests with race detection..."
go test -race ./internal/cli/...

echo ""
echo "=== All tests passed! ==="
echo ""
echo "Summary of fixes applied:"
echo "1. Standardized all error exit codes in stats.go to return 1"
echo "   - Line 79: Changed from 'return 2' to 'return 1'"
echo "   - Line 84: Changed from 'return 3' to 'return 1'"
echo "   - Line 90: Changed from 'return 3' to 'return 1'"
echo ""
echo "2. Updated stats_test.go to expect '127.0.0.1' instead of 'localhost'"
echo "   - Line 174: Changed expectedContains from 'localhost' to '127.0.0.1'"
