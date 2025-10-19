#!/bin/bash

echo "Mogost Toolkit build script"
echo "======================"

# Check Go environment
if ! command -v go &> /dev/null; then
    echo "Error: Go environment not found. Please install Go first."
    exit 1
fi

echo "1. Downloading dependencies..."
go get -v

if [ $? -ne 0 ]; then
    echo "Error: failed to download dependencies."
    exit 1
fi

echo "2. Building binary..."
go build -v -o tools

if [ $? -ne 0 ]; then
    echo "Error: build failed."
    exit 1
fi

echo "3. Creating required directories..."
mkdir -p uploads/file-compare
mkdir -p uploads/csv
mkdir -p uploads/archive-compare
mkdir -p static
mkdir -p templates
mkdir -p temp

echo "4. Setting executable permissions..."
chmod +x tools

echo "Build complete!"
echo ""
echo "How to use:"
echo "1. Run the binary: ./tools"
echo "2. Open in your browser: http://localhost:8080"
echo ""
echo "Tool overview:"
echo "- Tool 1: File comparison - upload two files to compare their content."
echo "- Tool 2: CSV viewer - upload a CSV file to inspect its rows."
echo "- Tool 3: Archive trade comparison - upload a ZIP file to compare trade files."
