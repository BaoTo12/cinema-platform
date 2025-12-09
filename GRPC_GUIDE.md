# 📡 gRPC Guide for Go Developers

> A beginner-friendly explanation of gRPC and Protocol Buffers

---

## 📋 Table of Contents

1. [What is gRPC?](#what-is-grpc)
2. [gRPC vs REST](#grpc-vs-rest)
3. [When to Use gRPC](#when-to-use-grpc)
4. [Protocol Buffers (.proto files)](#protocol-buffers)
5. [gRPC in This Project](#grpc-in-this-project)
6. [How to Add gRPC to Your Project](#how-to-add-grpc)

---

## What is gRPC?

**gRPC = Google Remote Procedure Call**

It's a way for programs to call functions on **other computers** as if they were local.

```
┌─────────────────┐                    ┌─────────────────┐
│   Your App      │   ──── gRPC ────►  │  Another App    │
│  (Client)       │                    │   (Server)      │
│                 │                    │                 │
│  GetUser(123)   │   ───────────────► │  Finds user 123 │
│                 │   ◄─────────────── │  Returns data   │
└─────────────────┘                    └─────────────────┘
```

### Key Concepts

| Term | Meaning |
|------|---------|
| **RPC** | Remote Procedure Call - calling a function on another machine |
| **Protobuf** | Protocol Buffers - binary format for data (smaller than JSON) |
| **Service** | A collection of RPC methods (like an API) |
| **Message** | Data structure definition (like a struct) |
| **Stub** | Auto-generated code that handles network calls |

---

## gRPC vs REST

| Feature | REST (HTTP/JSON) | gRPC (HTTP/2/Protobuf) |
|---------|------------------|------------------------|
| **Data format** | JSON (text, human-readable) | Protobuf (binary, compact) |
| **Speed** | Slower | 2-10x faster |
| **Message size** | Larger | Smaller (compressed) |
| **Contract** | OpenAPI/Swagger (optional) | `.proto` files (required) |
| **Code generation** | Optional | Built-in |
| **Browser support** | ✅ Direct | ❌ Needs gRPC-Web proxy |
| **Streaming** | Limited (SSE, WebSocket) | Native bidirectional |
| **Best for** | Web/Mobile clients | Service-to-service |

### Data Size Comparison

```json
// JSON (REST) - 82 bytes
{"id":"abc123","email":"john@example.com","firstName":"John","lastName":"Doe"}
```

```
// Protobuf (gRPC) - ~45 bytes (binary, not shown)
```

---

## When to Use gRPC

### Use gRPC When:
- ✅ **Multiple microservices** talking to each other
- ✅ You need **high performance** (low latency)
- ✅ You want **streaming** (real-time data)
- ✅ You need **strict API contracts**
- ✅ You're building **internal services**

### Use REST When:
- ✅ **Web browsers** are your clients
- ✅ You want **simple debugging** (curl, Postman)
- ✅ You have a **monolith** application
- ✅ You need **easy human readability**
- ✅ **Public APIs** for third-party developers

### Architecture Example

```
                    ┌─────────────┐
                    │   Frontend  │
                    │  (Browser)  │
                    └──────┬──────┘
                           │ REST ← Browsers need REST/JSON
                           ▼
                    ┌─────────────┐
                    │  API Gateway│
                    │   (Go)      │
                    └──────┬──────┘
                           │ gRPC ← Fast internal communication
              ┌────────────┼────────────┐
              ▼            ▼            ▼
        ┌──────────┐ ┌──────────┐ ┌──────────┐
        │  Auth    │ │  Movie   │ │ Booking  │
        │ Service  │ │ Service  │ │ Service  │
        └──────────┘ └──────────┘ └──────────┘
```

---

## Protocol Buffers

Protocol Buffers (protobuf) is the language gRPC uses to define services and messages.

### Basic Syntax

```protobuf
// api/proto/cinema/v1/auth.proto

syntax = "proto3";  // Use Protocol Buffers version 3

package cinema.v1;  // Namespace (prevents naming conflicts)

option go_package = "github.com/cinemaos/backend/gen/cinema/v1;cinemav1";

// Service = Collection of RPC methods (like an interface)
service AuthService {
  rpc Register(RegisterRequest) returns (RegisterResponse);
  rpc Login(LoginRequest) returns (LoginResponse);
  rpc GetCurrentUser(GetCurrentUserRequest) returns (GetCurrentUserResponse);
}

// Message = Data structure (like a Go struct)
message RegisterRequest {
  string email = 1;         // Field number 1
  string password = 2;      // Field number 2
  string first_name = 3;    // Field number 3
  string last_name = 4;     // Field number 4
  optional string phone = 5; // Optional field
}

message RegisterResponse {
  bool success = 1;
  string message = 2;
  User user = 3;  // Nested message
}

message User {
  string id = 1;
  string email = 2;
  string first_name = 3;
  string last_name = 4;
  string role = 5;
}
```

### Field Numbers

Each field has a unique number (1, 2, 3...) that identifies it in the binary format.

**Rules:**
- Numbers 1-15 take 1 byte (use for common fields)
- Numbers 16-2047 take 2 bytes
- Once assigned, **never change** them (breaks compatibility)

### Data Types

| Protobuf Type | Go Type | Description |
|---------------|---------|-------------|
| `string` | `string` | UTF-8 text |
| `int32` | `int32` | 32-bit integer |
| `int64` | `int64` | 64-bit integer |
| `float` | `float32` | 32-bit floating point |
| `double` | `float64` | 64-bit floating point |
| `bool` | `bool` | True/false |
| `bytes` | `[]byte` | Raw bytes |
| `repeated` | `[]T` | Array/slice |
| `optional` | `*T` | Pointer (nullable) |

### Streaming Types

```protobuf
service MovieService {
  // Unary - Single request, single response
  rpc GetMovie(GetMovieRequest) returns (GetMovieResponse);
  
  // Server streaming - Single request, stream of responses
  rpc WatchUpdates(WatchRequest) returns (stream MovieUpdate);
  
  // Client streaming - Stream of requests, single response  
  rpc UploadImages(stream ImageChunk) returns (UploadResult);
  
  // Bidirectional streaming - Both sides stream
  rpc Chat(stream ChatMessage) returns (stream ChatMessage);
}
```

---

## gRPC in This Project

### Current Status

| Component | Uses gRPC? | Notes |
|-----------|-----------|-------|
| API endpoints | ❌ REST | Gin HTTP handlers |
| `.proto` files | ✅ Defined | Not compiled yet |
| Tracer (Jaeger) | ✅ gRPC | Sends traces via gRPC |

**The `.proto` files are prepared for future use** - if you need to:
- Split into microservices
- Add high-performance internal APIs
- Implement real-time streaming

### Files in This Project

```
api/proto/cinema/v1/
├── auth.proto       # Authentication service
├── movies.proto     # Movie service
├── showtimes.proto  # Showtime service
├── bookings.proto   # Booking service
└── pricing.proto    # Pricing service
```

---

## How to Add gRPC

### Step 1: Install Tools

```bash
# Install protoc (Protocol Buffer compiler)
# Windows: Download from https://github.com/protocolbuffers/protobuf/releases

# Install Go plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### Step 2: Generate Go Code

```bash
protoc --go_out=. --go-grpc_out=. api/proto/cinema/v1/*.proto
```

This generates:
- `*_pb.go` - Message structs
- `*_grpc.pb.go` - gRPC service interfaces

### Step 3: Implement Server

```go
// internal/grpc/auth_server.go
package grpc

import (
    "context"
    pb "cinemaos-backend/gen/cinema/v1"
)

type AuthServer struct {
    pb.UnimplementedAuthServiceServer
    authService *auth.Service
}

func (s *AuthServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
    // Call your existing service
    result, err := s.authService.Register(ctx, auth.RegisterRequest{
        Email:     req.Email,
        Password:  req.Password,
        FirstName: req.FirstName,
        LastName:  req.LastName,
    })
    if err != nil {
        return nil, err
    }
    
    return &pb.RegisterResponse{
        Success: true,
        User:    toProtoUser(result.User),
    }, nil
}
```

### Step 4: Start Server

```go
// cmd/api/main.go
import "google.golang.org/grpc"

func main() {
    // ... existing code ...
    
    // Start gRPC server on different port
    grpcServer := grpc.NewServer()
    pb.RegisterAuthServiceServer(grpcServer, &AuthServer{})
    
    go func() {
        lis, _ := net.Listen("tcp", ":9090")
        grpcServer.Serve(lis)
    }()
}
```

---

## Summary

| Concept | One-Line Explanation |
|---------|---------------------|
| **gRPC** | Remote function calls using binary protocol |
| **Protobuf** | Language for defining data structures and services |
| **Service** | Collection of RPC methods (like an API) |
| **Message** | Data structure definition |
| **Streaming** | Continuous data flow (real-time) |

**This project currently uses REST** for the API but has `.proto` files ready for when you need gRPC (microservices, high-performance internal communication).

---

## Further Reading

- [gRPC Official Docs](https://grpc.io/docs/)
- [Protocol Buffers Guide](https://protobuf.dev/)
- [gRPC-Go Tutorial](https://grpc.io/docs/languages/go/quickstart/)
