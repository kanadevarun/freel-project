# ---------- Build Stage ----------
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN GOWORK=off CGO_ENABLED=0 GOOS=linux GOFLAGS="-mod=mod" go build \
    -ldflags="-w -s" \
    -o freel-server \
    ./cmd/server

# ---------- Runtime Stage ----------
FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/freel-server .
COPY --from=builder /app/internal/database/migrations ./internal/database/migrations

EXPOSE 8080

ENV PORT=8080
ENV APP_ENV=production

CMD ["./freel-server"]