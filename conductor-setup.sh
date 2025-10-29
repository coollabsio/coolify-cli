#!/bin/bash
set -e

echo "🔧 Setting up Coolify CLI workspace..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Error: Go is not installed"
    echo "Please install Go 1.24+ from https://go.dev/dl/"
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
MAJOR_MINOR=$(echo $GO_VERSION | cut -d. -f1,2)

# Compare version (must be 1.24 or higher)
if [ $(echo "$MAJOR_MINOR" | awk -F. '{print ($1 * 100) + $2}') -lt 124 ]; then
    echo "❌ Error: Go version 1.24+ is required"
    echo "Current version: $GO_VERSION"
    echo "Please upgrade Go from https://go.dev/dl/"
    exit 1
fi

echo "✅ Go version $GO_VERSION detected"

# Download dependencies
echo "📦 Downloading dependencies..."
if ! go mod download; then
    echo "❌ Error: Failed to download dependencies"
    exit 1
fi

echo "✅ Dependencies downloaded"

# Install air if not already installed
if ! command -v air &> /dev/null; then
    echo "📦 Installing air (Go file watcher)..."
    if ! go install github.com/air-verse/air@latest; then
        echo "⚠️  Warning: Failed to install air, but continuing..."
    else
        echo "✅ air installed successfully"
    fi
else
    echo "✅ air already installed"
fi

# Build the binary
echo "🔨 Building coolify binary..."
if ! go build -o coolify ./coolify; then
    echo "❌ Error: Build failed"
    exit 1
fi

echo "✅ Binary built successfully: ./coolify/coolify"
echo "🎉 Workspace setup complete!"
echo "🔥 Use the run script for hot reload during development"
