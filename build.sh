#!/bin/bash
echo "Building mux-ssh..."
go build -o mux-ssh ./cmd/ssh-ogm/main.go
echo "Build complete! Run with: ./mux-ssh"
