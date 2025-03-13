#!/bin/bash

# Load environment variables from .env file
if [ -f .env ]; then
    export $(grep -v '^#' .env | xargs)
    echo "Loaded environment variables from .env file"
else
    echo "No .env file found. Using default environment variables."
fi

# Check if the input directory exists
if [ ! -d "input" ]; then
    mkdir -p input
    echo "Created input directory"
fi

# Check if the BBC screenshot exists
if [ ! -f "input/12_march_2025_bbc.png" ]; then
    echo "Error: BBC screenshot not found at input/12_march_2025_bbc.png"
    echo "Please place a screenshot of the BBC website at this location and try again."
    exit 1
fi

# Check if the output directory exists
if [ ! -d "output" ]; then
    mkdir -p output
    echo "Created output directory"
fi

# Build the scene generator if it doesn't exist
if [ ! -f "bin/scene-generator" ]; then
    echo "Building scene generator..."
    mkdir -p bin
    go build -o bin/scene-generator cmd/scene-generator/main.go
fi

# Run the scene generator
echo "Running scene generator..."
./bin/scene-generator "$@"

# Check if the scenes were generated successfully
if [ $? -eq 0 ]; then
    echo "Scenes generated successfully!"
    echo "Output saved to output/scenes.json"
else
    echo "Error: Failed to generate scenes"
    exit 1
fi 