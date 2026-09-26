# --- Build stage ---
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy the full source first so go mod tidy can see all imports
COPY . .

# No go.sum in this repo — resolve/lock dependencies now that source is present
RUN go mod tidy

# Build a static binary
RUN CGO_ENABLED=0 GOOS=linux go build -o edgeanalyzer ./cmd

# --- Runtime stage ---
FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/edgeanalyzer ./edgeanalyzer
COPY --from=builder /app/config.yaml ./config.yaml
COPY --from=builder /app/web ./web

# Render provides $PORT at runtime; the app already reads it (falls back to config.yaml's port)
EXPOSE 8080

CMD ["./edgeanalyzer"]
