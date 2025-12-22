# ClassConnectRPC

A gRPC-based API server for managing educational institutions, providing services for students, teachers, and executives with JWT authentication, rate limiting, and MongoDB integration.

## Overview

ClassConnectRPC is a high-performance gRPC server built with Go that manages three core user types: Students, Teachers, and Executives (Execs). The system provides comprehensive CRUD operations, authentication, and user management capabilities with enterprise-grade security features.

## Features

- **gRPC API** - High-performance RPC communication using Protocol Buffers
- **JWT Authentication** - Secure token-based authentication with token blacklisting for logout
- **Rate Limiting** - Configurable request rate limiting (5 requests per minute default)
- **MongoDB Integration** - Persistent data storage with MongoDB
- **Response Time Tracking** - Built-in interceptor for monitoring request performance
- **User Management** - Complete CRUD operations for students, teachers, and executives
- **Password Management** - Update, reset, and forgot password functionality
- **User Deactivation** - Administrative capability to deactivate users

## Prerequisites

- Go 1.24.4 or higher
- MongoDB instance (local or remote)
- Protocol Buffer compiler (protoc) for regenerating proto files

## Installation

1. Clone the repository:
```bash
git clone https://github.com/vinyas-bharadwaj/ClassConnectRPC
cd ClassConnectRPC
```

2. Install dependencies:
```bash
go mod download
```

3. Create a `.env` file in the project root:
```env
SERVER_PORT=50051
MONGODB_URI=mongodb://localhost:27017
JWT_SECRET=your-secret-key-here
JWT_EXPIRES_IN=15m
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_PORT` | gRPC server port | 50051 |
| `MONGODB_URI` | MongoDB connection string | mongodb://localhost:27017 |
| `JWT_SECRET` | Secret key for JWT signing | Required |
| `JWT_EXPIRES_IN` | JWT token expiration duration | 15m |

### Rate Limiting

The server is configured with rate limiting of 5 requests per minute per client. This can be modified in `cmd/grpcapi/server.go`.

## Running the Server

Start the gRPC server:
```bash
go run cmd/grpcapi/server.go
```

The server will start on the configured port (default: 50051) and display:
```
gRPC server running on port :50051
```

## Project Structure

```
.
├── cmd/
│   └── grpcapi/
│       └── server.go              # Server entry point
├── internals/
│   ├── api/
│   │   ├── handlers/              # RPC method implementations
│   │   │   ├── execs.go
│   │   │   ├── students.go
│   │   │   ├── teachers.go
│   │   │   └── server_struct.go
│   │   └── interceptors/          # gRPC interceptors
│   │       ├── authentication.go   # JWT authentication
│   │       ├── rate_limiter.go    # Rate limiting
│   │       └── response_time.go   # Performance tracking
│   ├── models/                    # Data models
│   │   ├── exec.go
│   │   ├── student.go
│   │   └── teacher.go
│   └── repositories/
│       └── mongodb/               # MongoDB operations
│           ├── mongoconnect.go
│           ├── execs_crud.go
│           ├── students_crud.go
│           └── teachers_crud.go
├── pkg/
│   └── utils/                     # Utility functions
│       ├── jwt.go                 # JWT operations
│       ├── error_handler.go
│       └── verify_password.go
├── proto/                         # Protocol Buffer definitions
│   ├── execs.proto
│   ├── students.proto
│   ├── main.proto
│   └── gen/                       # Generated protobuf code
└── cert/                          # SSL/TLS certificates (optional)
```

## API Services

### ExecsService

- `Login` - Authenticate executives and receive JWT token
- `Logout` - Invalidate JWT token
- `GetExecs` - Retrieve executive records
- `AddExecs` - Create new executive accounts
- `UpdateExecs` - Update executive information
- `DeleteExecs` - Remove executive records
- `UpdatePassword` - Change password
- `ResetPassword` - Reset password with old password
- `ForgotPassword` - Reset password without old password
- `DeactivateUser` - Deactivate user accounts

### StudentsService

- `GetStudents` - Retrieve student records
- `AddStudents` - Create new student accounts
- `UpdateStudents` - Update student information
- `DeleteStudents` - Remove student records

### TeachersService

- `GetTeachers` - Retrieve teacher records
- `AddTeachers` - Create new teacher accounts
- `UpdateTeachers` - Update teacher information
- `DeleteTeachers` - Remove teacher records

## Authentication

The API uses JWT (JSON Web Tokens) for authentication. Protected endpoints require a valid JWT token in the metadata:

```
authorization: Bearer <token>
```

### Token Blacklisting

When users logout, their tokens are blacklisted and cannot be reused until expiration. A background cleanup process removes expired tokens from the blacklist every 2 minutes.

## Development

### Regenerating Protocol Buffers

After modifying `.proto` files, regenerate the Go code:

```bash
protoc --go_out=. --go-grpc_out=. proto/*.proto
```




