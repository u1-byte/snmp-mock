# go-client/Dockerfile
# Stage 1: Build
FROM golang:1.21-alpine AS builder
WORKDIR /app

# Copy ALL files into the container first
COPY . .

# Run go mod tidy to automatically generate the missing go.sum file
RUN go mod tidy

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o snmp-client .

# Stage 2: Run
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/snmp-client .
CMD ["./snmp-client"]