#!/bin/bash

# Build script for Monitor System

set -e

echo "Building Monitor System..."

# Create bin directory
mkdir -p bin

# Build Server
echo "Building API Server..."
go build -o bin/monitor-server ./cmd/server
echo "✓ Server built successfully"

# Build Agent
echo "Building Agent..."
go build -o bin/monitor-agent ./cmd/agent
echo "✓ Agent built successfully"

echo ""
echo "Build complete! Binaries are in the bin/ directory:"
echo "  - bin/monitor-server (API Server)"
echo "  - bin/monitor-agent (Agent)"
echo ""
echo "To run:"
echo "  Server: ./bin/monitor-server -config ./configs/server-config.yaml"
echo "  Agent:  ./bin/monitor-agent -config ./configs/agent-config.yaml"
