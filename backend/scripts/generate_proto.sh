#!/bin/bash

# Generate Protocol Buffer code with Connect support

# Install protoc-gen-go and protoc-gen-connect-go if not already installed
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest

# Create output directory
mkdir -p gen/cinema/v1

# Generate Go code from proto files
protoc \
  --go_out=. \
  --go_opt=paths=source_relative \
  --connect-go_out=. \
  --connect-go_opt=paths=source_relative \
  proto/cinema/v1/*.proto

echo "✅ Protocol buffers generated successfully"
