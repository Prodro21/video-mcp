# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /video-mcp ./cmd/server

# Runtime stage (minimal for MCP stdio server)
FROM alpine:3.20

RUN apk add --no-cache ca-certificates

RUN adduser -D -g '' appuser

COPY --from=builder /video-mcp /usr/local/bin/

USER appuser

ENTRYPOINT ["video-mcp"]
