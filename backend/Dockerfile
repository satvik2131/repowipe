# Go Backend Dockerfile
FROM golang:1.24.6-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o repowipe

FROM alpine:latest
RUN apk add --no-cache git ca-certificates
WORKDIR /app
COPY --from=builder /app/repowipe ./
EXPOSE 8080
ENV PORT=8080
CMD ["./repowipe"]
