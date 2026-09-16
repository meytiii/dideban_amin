# Test Suite

This directory contains the automated test suite for the Herasat Backend.

The test files were moved here from the project root to keep the main workspace clean and focused on production runtime files.

## Running Tests

Because the tests belong to `package main`, to run the tests against the root package files, either:

### Option 1: Temporary restore to root
```powershell
# Move tests back to root
Move-Item test\*_test.go .

# Run all tests
go test -v .

# Move back to test directory
Move-Item *_test.go test\
```

### Option 2: Run specific test with root source files
Copy the target test file to the root or copy root files to a staging build.
