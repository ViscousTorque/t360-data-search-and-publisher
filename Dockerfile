FROM --platform=linux/amd64 golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

# Build Go binary for Linux amd64 - TODO : need to enable a parameter for other archs
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/main .

# Build Go binary for Linux amd64
# Use minimal Debian image - TODO : need to enable a parameter for other archs
FROM --platform=linux/amd64 debian:bullseye-slim

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /app/main /app/main
RUN chmod +x /app/main

CMD ["/app/main"]
