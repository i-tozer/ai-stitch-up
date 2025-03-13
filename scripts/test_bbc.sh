#!/bin/bash

# Load environment variables from .env file
if [ -f .env ]; then
    export $(grep -v '^#' .env | xargs)
    echo "Loaded environment variables from .env file"
else
    echo "No .env file found. Using default environment variables."
fi

# Run the real BBC test
REAL_TEST=true go test ./pkg/1_contentextraction -run TestRealBBCExtraction -v

# Exit with the test's exit code
exit $? 