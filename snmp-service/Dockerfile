# build stage
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64
RUN go build -o snmp-service ./cmd/

# runtime stage
FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/snmp-service .
EXPOSE 8080
ENTRYPOINT ["./snmp-service"]

