#!/bin/bash
echo "Setting up Go environment for China..."
echo ""

# Set Go proxy to China mirror
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOSUMDB=sum.golang.google.cn

echo "Go proxy configured!"
echo ""
echo "Downloading dependencies..."
go mod download

echo ""
echo "Setup complete! You can now run:"
echo "  go run main.go"
echo ""
