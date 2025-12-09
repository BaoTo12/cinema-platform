# 📡 gRPC Complete Guide for Go Developers

> From absolute beginner to advanced concepts with real code examples

---

## 📋 Table of Contents

1. [Part 1: The Basics](#part-1-the-basics)
2. [Part 2: Protocol Buffers Deep Dive](#part-2-protocol-buffers-deep-dive)
3. [Part 3: Building Your First gRPC Service](#part-3-building-your-first-grpc-service)
4. [Part 4: Advanced Patterns](#part-4-advanced-patterns)
5. [Part 5: Real-World Examples](#part-5-real-world-examples)
6. [Part 6: gRPC in This Project](#part-6-grpc-in-this-project)

---

# Part 1: The Basics

## What Problem Does gRPC Solve?

Imagine you have two programs that need to talk to each other:

```
Program A: "Hey, I need user data for ID 123"
Program B: "Here's the data: {name: John, email: john@example.com}"
```

**The old way (REST API):**
```
Program A → HTTP POST /api/users/123 → Program B
Program A ← JSON response ← Program B
```

**The gRPC way:**
```
Program A → GetUser(123) → Program B  (like calling a local function!)
Program A ← UserData ← Program B
```

## Why is gRPC Faster?

### 1. Binary Format (Protobuf vs JSON)

```
JSON (REST):  {"name":"John","email":"john@example.com","age":25}
              └── 52 bytes, text, needs parsing

Protobuf:     [binary data]
              └── ~20 bytes, binary, no parsing needed
```

### 2. HTTP/2 (Multiplexing)

```
HTTP/1.1 (REST):
Request 1  ─────────────────►  Response 1
Request 2  ─────────────────►  Response 2
Request 3  ─────────────────►  Response 3
└── Each request waits for previous to complete

HTTP/2 (gRPC):
Request 1  ─────►  Response 1
Request 2  ─────►  Response 2
Request 3  ─────►  Response 3
└── All requests happen simultaneously on ONE connection
```

### 3. Code Generation

gRPC automatically generates client/server code from `.proto` files:

```
           ┌─────────────────┐
           │   auth.proto    │
           │  (definition)   │
           └────────┬────────┘
                    │ protoc compiler
        ┌───────────┼───────────┐
        ▼           ▼           ▼
  ┌──────────┐ ┌──────────┐ ┌──────────┐
  │ Go code  │ │ Python   │ │ Java     │
  │ (client) │ │ (server) │ │ (client) │
  └──────────┘ └──────────┘ └──────────┘
```

---

# Part 2: Protocol Buffers Deep Dive

## Basic Syntax

```protobuf
// File: user.proto

syntax = "proto3";  // Always use proto3 (latest version)

package myapp.v1;   // Namespace to avoid naming conflicts

option go_package = "myapp/gen/v1;myappv1";  // Go import path

// Message = Like a struct in Go
message User {
  string id = 1;          // Field number 1
  string name = 2;        // Field number 2
  string email = 3;       // Field number 3
  int32 age = 4;          // Field number 4
  bool is_active = 5;     // Field number 5
}
```

**Generated Go code:**
```go
type User struct {
    Id       string `protobuf:"bytes,1,opt,name=id,proto3"`
    Name     string `protobuf:"bytes,2,opt,name=name,proto3"`
    Email    string `protobuf:"bytes,3,opt,name=email,proto3"`
    Age      int32  `protobuf:"varint,4,opt,name=age,proto3"`
    IsActive bool   `protobuf:"varint,5,opt,name=is_active,proto3"`
}
```

## Field Numbers Explained

```protobuf
message Example {
  string name = 1;   // Number 1
  int32 age = 2;     // Number 2
  // SKIP NUMBER 3 (maybe deleted in past)
  string email = 4;  // Number 4
}
```

**Rules:**
- Each field must have a UNIQUE number
- Numbers 1-15 use 1 byte (most efficient)
- Numbers 16-2047 use 2 bytes
- Once used, NEVER reuse a number (for backward compatibility)

## All Data Types

### Scalar Types

```protobuf
message AllTypes {
  // Integer types
  int32 small_int = 1;      // -2^31 to 2^31-1
  int64 big_int = 2;        // -2^63 to 2^63-1
  uint32 unsigned_int = 3;  // 0 to 2^32-1
  uint64 big_unsigned = 4;  // 0 to 2^64-1
  sint32 signed_int = 5;    // Better for negative numbers
  sint64 big_signed = 6;
  
  // Fixed-size (faster but always use full bytes)
  fixed32 fixed_small = 7;  // Always 4 bytes
  fixed64 fixed_big = 8;    // Always 8 bytes
  
  // Floating point
  float decimal = 9;        // 32-bit
  double precise = 10;      // 64-bit
  
  // Other
  bool flag = 11;           // true/false
  string text = 12;         // UTF-8 string
  bytes raw_data = 13;      // Raw bytes []byte
}
```

### Repeated (Arrays/Slices)

```protobuf
message MovieList {
  repeated string genres = 1;        // []string in Go
  repeated int32 ratings = 2;        // []int32 in Go
  repeated Movie movies = 3;         // []*Movie in Go
}
```

**Go usage:**
```go
list := &MovieList{
    Genres:  []string{"Action", "Comedy"},
    Ratings: []int32{5, 4, 3},
    Movies:  []*Movie{{Title: "Movie 1"}, {Title: "Movie 2"}},
}
```

### Optional Fields

```protobuf
message User {
  string name = 1;                    // Required (has default "")
  optional string nickname = 2;       // Truly optional (can be nil)
  optional int32 age = 3;             // Can distinguish 0 from "not set"
}
```

**Go usage:**
```go
user := &User{
    Name:     "John",
    Nickname: proto.String("Johnny"),  // *string
    Age:      proto.Int32(25),         // *int32
}

// Check if set
if user.Nickname != nil {
    fmt.Println(*user.Nickname)
}
```

### Maps

```protobuf
message Settings {
  map<string, string> string_settings = 1;   // map[string]string
  map<string, int32> int_settings = 2;       // map[string]int32
  map<int32, User> users_by_id = 3;          // map[int32]*User
}
```

**Go usage:**
```go
settings := &Settings{
    StringSettings: map[string]string{
        "theme": "dark",
        "lang":  "en",
    },
}
```

### Enums

```protobuf
enum UserRole {
  USER_ROLE_UNSPECIFIED = 0;  // Always have 0 as default
  USER_ROLE_ADMIN = 1;
  USER_ROLE_EDITOR = 2;
  USER_ROLE_VIEWER = 3;
}

message User {
  string name = 1;
  UserRole role = 2;
}
```

**Go usage:**
```go
user := &User{
    Name: "John",
    Role: myappv1.UserRole_USER_ROLE_ADMIN,
}

switch user.Role {
case myappv1.UserRole_USER_ROLE_ADMIN:
    fmt.Println("Admin user")
case myappv1.UserRole_USER_ROLE_VIEWER:
    fmt.Println("Viewer user")
}
```

### Nested Messages

```protobuf
message Order {
  string id = 1;
  Customer customer = 2;      // Nested message
  repeated Item items = 3;    // Array of nested messages
  
  message Item {              // Defined inside Order
    string product_id = 1;
    int32 quantity = 2;
    double price = 3;
  }
}

message Customer {
  string name = 1;
  Address address = 2;
}

message Address {
  string street = 1;
  string city = 2;
  string country = 3;
}
```

### OneOf (Union Types)

```protobuf
message Payment {
  string id = 1;
  double amount = 2;
  
  // Only ONE of these can be set
  oneof method {
    CreditCard credit_card = 3;
    BankTransfer bank_transfer = 4;
    Crypto crypto = 5;
  }
}

message CreditCard {
  string number = 1;
  string expiry = 2;
}

message BankTransfer {
  string account = 1;
  string routing = 2;
}

message Crypto {
  string wallet_address = 1;
  string coin_type = 2;
}
```

**Go usage:**
```go
// Credit card payment
payment := &Payment{
    Id:     "pay_123",
    Amount: 99.99,
    Method: &Payment_CreditCard{
        CreditCard: &CreditCard{
            Number: "4111111111111111",
            Expiry: "12/25",
        },
    },
}

// Check which payment method
switch m := payment.Method.(type) {
case *Payment_CreditCard:
    fmt.Println("Card:", m.CreditCard.Number)
case *Payment_BankTransfer:
    fmt.Println("Bank:", m.BankTransfer.Account)
case *Payment_Crypto:
    fmt.Println("Crypto:", m.Crypto.WalletAddress)
}
```

---

# Part 3: Building Your First gRPC Service

## Step 1: Define the Service

```protobuf
// api/proto/greeter/v1/greeter.proto

syntax = "proto3";

package greeter.v1;

option go_package = "myapp/gen/greeter/v1;greeterv1";

// Service definition
service GreeterService {
  // Unary RPC - single request, single response
  rpc SayHello(SayHelloRequest) returns (SayHelloResponse);
  
  // Server streaming - single request, multiple responses
  rpc SayHelloStream(SayHelloRequest) returns (stream SayHelloResponse);
}

message SayHelloRequest {
  string name = 1;
}

message SayHelloResponse {
  string message = 1;
  string timestamp = 2;
}
```

## Step 2: Generate Go Code

```bash
# Install protoc plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate code
protoc --go_out=. --go-grpc_out=. api/proto/greeter/v1/greeter.proto
```

This generates:
- `gen/greeter/v1/greeter.pb.go` - Message structs
- `gen/greeter/v1/greeter_grpc.pb.go` - Service interface

## Step 3: Implement the Server

```go
// internal/grpc/greeter_server.go

package grpc

import (
    "context"
    "fmt"
    "time"
    
    pb "myapp/gen/greeter/v1"
)

// GreeterServer implements the GreeterService interface
type GreeterServer struct {
    pb.UnimplementedGreeterServiceServer  // Embed for forward compatibility
}

// NewGreeterServer creates a new server
func NewGreeterServer() *GreeterServer {
    return &GreeterServer{}
}

// SayHello - Unary RPC
func (s *GreeterServer) SayHello(ctx context.Context, req *pb.SayHelloRequest) (*pb.SayHelloResponse, error) {
    // Validate input
    if req.Name == "" {
        return nil, status.Error(codes.InvalidArgument, "name is required")
    }
    
    return &pb.SayHelloResponse{
        Message:   fmt.Sprintf("Hello, %s!", req.Name),
        Timestamp: time.Now().Format(time.RFC3339),
    }, nil
}

// SayHelloStream - Server Streaming RPC
func (s *GreeterServer) SayHelloStream(req *pb.SayHelloRequest, stream pb.GreeterService_SayHelloStreamServer) error {
    // Send 5 messages over time
    for i := 1; i <= 5; i++ {
        response := &pb.SayHelloResponse{
            Message:   fmt.Sprintf("Hello #%d, %s!", i, req.Name),
            Timestamp: time.Now().Format(time.RFC3339),
        }
        
        if err := stream.Send(response); err != nil {
            return err
        }
        
        time.Sleep(1 * time.Second)  // Delay between messages
    }
    
    return nil
}
```

## Step 4: Start the gRPC Server

```go
// cmd/grpc/main.go

package main

import (
    "log"
    "net"
    
    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"
    
    pb "myapp/gen/greeter/v1"
    grpcserver "myapp/internal/grpc"
)

func main() {
    // Create TCP listener
    lis, err := net.Listen("tcp", ":9090")
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }
    
    // Create gRPC server
    server := grpc.NewServer()
    
    // Register our service
    pb.RegisterGreeterServiceServer(server, grpcserver.NewGreeterServer())
    
    // Enable reflection (for grpcurl and debugging)
    reflection.Register(server)
    
    log.Println("gRPC server listening on :9090")
    
    // Start serving
    if err := server.Serve(lis); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}
```

## Step 5: Create a Client

```go
// cmd/client/main.go

package main

import (
    "context"
    "io"
    "log"
    "time"
    
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    
    pb "myapp/gen/greeter/v1"
)

func main() {
    // Connect to server
    conn, err := grpc.Dial("localhost:9090", grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("failed to connect: %v", err)
    }
    defer conn.Close()
    
    // Create client
    client := pb.NewGreeterServiceClient(conn)
    
    // === Unary Call ===
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    resp, err := client.SayHello(ctx, &pb.SayHelloRequest{Name: "World"})
    if err != nil {
        log.Fatalf("SayHello failed: %v", err)
    }
    log.Printf("Unary response: %s", resp.Message)
    
    // === Streaming Call ===
    stream, err := client.SayHelloStream(ctx, &pb.SayHelloRequest{Name: "World"})
    if err != nil {
        log.Fatalf("SayHelloStream failed: %v", err)
    }
    
    for {
        resp, err := stream.Recv()
        if err == io.EOF {
            break  // Stream finished
        }
        if err != nil {
            log.Fatalf("stream error: %v", err)
        }
        log.Printf("Stream response: %s", resp.Message)
    }
}
```

---

# Part 4: Advanced Patterns

## 1. Interceptors (Middleware for gRPC)

```go
// Unary interceptor (like middleware)
func loggingInterceptor(
    ctx context.Context,
    req interface{},
    info *grpc.UnaryServerInfo,
    handler grpc.UnaryHandler,
) (interface{}, error) {
    start := time.Now()
    
    // Call the actual handler
    resp, err := handler(ctx, req)
    
    // Log after handling
    log.Printf("Method: %s, Duration: %v, Error: %v",
        info.FullMethod,
        time.Since(start),
        err,
    )
    
    return resp, err
}

// Use it
server := grpc.NewServer(
    grpc.UnaryInterceptor(loggingInterceptor),
)
```

## 2. Authentication Interceptor

```go
func authInterceptor(
    ctx context.Context,
    req interface{},
    info *grpc.UnaryServerInfo,
    handler grpc.UnaryHandler,
) (interface{}, error) {
    // Get metadata (like HTTP headers)
    md, ok := metadata.FromIncomingContext(ctx)
    if !ok {
        return nil, status.Error(codes.Unauthenticated, "no metadata")
    }
    
    // Check authorization header
    tokens := md.Get("authorization")
    if len(tokens) == 0 {
        return nil, status.Error(codes.Unauthenticated, "no token")
    }
    
    // Validate token
    userID, err := validateToken(tokens[0])
    if err != nil {
        return nil, status.Error(codes.Unauthenticated, "invalid token")
    }
    
    // Add user ID to context
    ctx = context.WithValue(ctx, "user_id", userID)
    
    return handler(ctx, req)
}
```

## 3. Error Handling with Status Codes

```go
import (
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

func (s *UserServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
    // Validation error
    if req.Id == "" {
        return nil, status.Error(codes.InvalidArgument, "user ID is required")
    }
    
    // Not found
    user, err := s.repo.GetByID(ctx, req.Id)
    if err != nil {
        if errors.Is(err, ErrNotFound) {
            return nil, status.Error(codes.NotFound, "user not found")
        }
        return nil, status.Error(codes.Internal, "internal error")
    }
    
    // Permission denied
    if !canAccess(ctx, user) {
        return nil, status.Error(codes.PermissionDenied, "access denied")
    }
    
    return toProtoUser(user), nil
}
```

**gRPC Status Codes:**

| Code | When to Use |
|------|-------------|
| `OK` | Success |
| `InvalidArgument` | Bad input from client |
| `NotFound` | Resource doesn't exist |
| `AlreadyExists` | Duplicate resource |
| `PermissionDenied` | No permission |
| `Unauthenticated` | Not logged in |
| `ResourceExhausted` | Rate limited |
| `FailedPrecondition` | State conflict |
| `Internal` | Server error |
| `Unavailable` | Service down |
| `DeadlineExceeded` | Timeout |

## 4. Bidirectional Streaming (Chat Example)

```protobuf
service ChatService {
  rpc Chat(stream ChatMessage) returns (stream ChatMessage);
}

message ChatMessage {
  string user = 1;
  string text = 2;
  string timestamp = 3;
}
```

```go
func (s *ChatServer) Chat(stream pb.ChatService_ChatServer) error {
    for {
        // Receive message from client
        msg, err := stream.Recv()
        if err == io.EOF {
            return nil
        }
        if err != nil {
            return err
        }
        
        log.Printf("Received: %s says: %s", msg.User, msg.Text)
        
        // Send response back
        response := &pb.ChatMessage{
            User:      "Server",
            Text:      fmt.Sprintf("Echo: %s", msg.Text),
            Timestamp: time.Now().Format(time.RFC3339),
        }
        
        if err := stream.Send(response); err != nil {
            return err
        }
    }
}
```

## 5. Deadlines and Timeouts

```go
// Client side - set deadline
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

resp, err := client.SlowOperation(ctx, req)
if err != nil {
    if status.Code(err) == codes.DeadlineExceeded {
        log.Println("Operation timed out")
    }
}

// Server side - check deadline
func (s *Server) SlowOperation(ctx context.Context, req *pb.Request) (*pb.Response, error) {
    // Check if we have enough time
    deadline, ok := ctx.Deadline()
    if ok && time.Until(deadline) < 1*time.Second {
        return nil, status.Error(codes.DeadlineExceeded, "not enough time")
    }
    
    // Do work...
    select {
    case <-ctx.Done():
        return nil, ctx.Err()  // Client cancelled
    case result := <-doWork():
        return result, nil
    }
}
```

---

# Part 5: Real-World Examples

## Cinema Booking Service

```protobuf
// api/proto/cinema/v1/bookings.proto

syntax = "proto3";

package cinema.v1;

service BookingService {
  // Create a new booking
  rpc CreateBooking(CreateBookingRequest) returns (CreateBookingResponse);
  
  // Get booking by ID
  rpc GetBooking(GetBookingRequest) returns (Booking);
  
  // Stream seat availability updates (real-time)
  rpc WatchSeatAvailability(WatchSeatsRequest) returns (stream SeatUpdate);
  
  // Batch cancel bookings
  rpc CancelBookings(stream CancelBookingRequest) returns (CancelBookingsResponse);
}

message CreateBookingRequest {
  string user_id = 1;
  string showtime_id = 2;
  repeated string seat_ids = 3;
  PaymentMethod payment = 4;
}

message Booking {
  string id = 1;
  string reference = 2;
  string user_id = 3;
  string showtime_id = 4;
  repeated Seat seats = 5;
  BookingStatus status = 6;
  double total_amount = 7;
  string created_at = 8;
}

enum BookingStatus {
  BOOKING_STATUS_UNSPECIFIED = 0;
  BOOKING_STATUS_PENDING = 1;
  BOOKING_STATUS_CONFIRMED = 2;
  BOOKING_STATUS_CANCELLED = 3;
}

message Seat {
  string id = 1;
  string row = 2;
  int32 number = 3;
  SeatType type = 4;
  double price = 5;
}

message SeatUpdate {
  string showtime_id = 1;
  string seat_id = 2;
  bool is_available = 3;
  string updated_at = 4;
}
```

## Implementation

```go
// Real-time seat availability using server streaming
func (s *BookingServer) WatchSeatAvailability(
    req *pb.WatchSeatsRequest,
    stream pb.BookingService_WatchSeatAvailabilityServer,
) error {
    ctx := stream.Context()
    
    // Subscribe to seat updates
    updates := s.seatService.Subscribe(req.ShowtimeId)
    defer s.seatService.Unsubscribe(req.ShowtimeId)
    
    for {
        select {
        case <-ctx.Done():
            return nil  // Client disconnected
            
        case update := <-updates:
            if err := stream.Send(&pb.SeatUpdate{
                ShowtimeId:  update.ShowtimeID,
                SeatId:      update.SeatID,
                IsAvailable: update.IsAvailable,
                UpdatedAt:   update.UpdatedAt.Format(time.RFC3339),
            }); err != nil {
                return err
            }
        }
    }
}
```

---

# Part 6: gRPC in This Project

## Current Implementation Status

| Component | Status | Notes |
|-----------|--------|-------|
| `.proto` files | ✅ Defined | `api/proto/cinema/v1/` |
| Generated code | ❌ Not yet | Need to run `protoc` |
| gRPC server | ❌ Not yet | Use REST handlers currently |
| gRPC client | ❌ N/A | Frontend uses REST |

## Files in This Project

```
api/proto/cinema/v1/
├── auth.proto       # Login, register, tokens
├── movies.proto     # Movie CRUD operations
├── showtimes.proto  # Showtime listings
├── bookings.proto   # Seat booking
└── pricing.proto    # Dynamic pricing
```

## When to Activate gRPC

Enable gRPC when you need:

1. **Microservices** - Split into auth-service, booking-service, etc.
2. **Real-time features** - Live seat updates during booking
3. **High performance** - Internal service-to-service calls
4. **Mobile apps** - Faster than REST for native apps

## Migration Path

```
Current (Monolith + REST):
┌─────────────────────────────────────┐
│           Go Backend                │
│  ┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐   │
│  │Auth │ │Movie│ │Show │ │Book │   │
│  └──┬──┘ └──┬──┘ └──┬──┘ └──┬──┘   │
│     └───────┴───────┴───────┘       │
│              REST API                │
└─────────────────────────────────────┘

Future (Microservices + gRPC):
┌─────────────────────────────────────┐
│           API Gateway               │
│              REST                    │
└───────────────┬─────────────────────┘
                │ gRPC
    ┌───────────┼───────────┬───────────┐
    ▼           ▼           ▼           ▼
┌───────┐  ┌───────┐  ┌───────┐  ┌───────┐
│ Auth  │  │ Movie │  │ Show  │  │ Book  │
│Service│  │Service│  │Service│  │Service│
└───────┘  └───────┘  └───────┘  └───────┘
```

---

## Quick Reference

### Commands

```bash
# Install tools
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate Go code
protoc --go_out=. --go-grpc_out=. api/proto/**/*.proto

# Test with grpcurl
grpcurl -plaintext localhost:9090 list
grpcurl -plaintext -d '{"name":"World"}' localhost:9090 greeter.v1.GreeterService/SayHello
```

### Common Imports

```go
import (
    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "google.golang.org/grpc/metadata"
    "google.golang.org/grpc/credentials"
    "google.golang.org/grpc/credentials/insecure"
)
```

---

## Summary

| Concept | Description |
|---------|-------------|
| **gRPC** | Fast RPC framework using HTTP/2 + Protobuf |
| **Protobuf** | Binary serialization format (smaller than JSON) |
| **Service** | Collection of RPC methods |
| **Unary** | Single request → Single response |
| **Server Streaming** | Single request → Multiple responses |
| **Client Streaming** | Multiple requests → Single response |
| **Bidirectional** | Both sides stream simultaneously |

**Key takeaway:** This project uses REST for the web API, but the `.proto` definitions are ready for when you need high-performance internal communication or want to split into microservices.
