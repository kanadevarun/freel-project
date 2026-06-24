# Freel Backend Phase 1

This is the foundational Go backend for the Freel logistics platform.
Currently, it supports AWS Cognito-based authentication APIs.

## Setup

1. Make sure you have Go 1.22+ installed.
2. Clone or navigate to the `backend/` directory.
3. Run `go mod download` to fetch dependencies.
4. Copy `.env.example` to `.env` and fill out any missing secrets.
   
```bash
cp .env.example .env
```

## Running the Server

Run the development server using:
```bash
go run cmd/server/main.go
```
The server will start on port `8080` by default.

## Available Endpoints

### `GET /health`
Returns the health status of the backend.

### `POST /auth/signup`
Creates a new user in the AWS Cognito User Pool.

### `POST /auth/verify-email`
Verifies the user's email using the 6-digit confirmation code.

### `POST /auth/login`
Authenticates a user and returns Cognito tokens (access, id, refresh).

### `POST /auth/forgot-password`
Requests a password reset code.

### `POST /auth/reset-password`
Resets the user's password using the received code.

### `GET /auth/me`
Placeholder endpoint to fetch current user data from the Authorization token.

## Sample Curl Commands

### 1. Health Check
```bash
curl http://localhost:8080/health
```

### 2. Signup
```bash
curl -X POST http://localhost:8080/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "full_name": "Varun Kanade",
    "company_name": "Tesla India Logistics",
    "email": "varun@example.com",
    "password": "StrongPass@123",
    "role": "shipper"
  }'
```

### 3. Verify Email
```bash
curl -X POST http://localhost:8080/auth/verify-email \
  -H "Content-Type: application/json" \
  -d '{
    "email": "varun@example.com",
    "code": "123456"
  }'
```

### 4. Login
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "varun@example.com",
    "password": "StrongPass@123"
  }'
```
