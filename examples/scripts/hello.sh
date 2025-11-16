#!/bin/bash

# Simple test script
echo "Hello from OpenFroyo script module!"
echo "Hostname: $(hostname)"
echo "User: $(whoami)"
echo "Date: $(date)"

# If arguments provided, echo them
if [ -n "$1" ]; then
    echo "Arguments: $@"
fi
