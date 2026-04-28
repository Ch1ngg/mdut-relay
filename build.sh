#!/bin/bash

echo "Building MDUT Relay for multiple platforms..."

# Linux
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/mdut-relay-linux-amd64 main.go
GOOS=linux GOARCH=386 go build -ldflags="-s -w" -o dist/mdut-relay-linux-386 main.go

# Windows
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/mdut-relay-windows-amd64.exe main.go
GOOS=windows GOARCH=386 go build -ldflags="-s -w" -o dist/mdut-relay-windows-386.exe main.go

# macOS
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o dist/mdut-relay-darwin-amd64 main.go
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/mdut-relay-darwin-arm64 main.go

echo "Done! Binaries are in the dist/ folder."
