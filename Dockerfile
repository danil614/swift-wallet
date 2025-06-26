FROM golang:1.24-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /swiftwallet ./cmd/server

FROM alpine:latest
WORKDIR /app
COPY --from=builder /swiftwallet .
ENTRYPOINT ["./swiftwallet"]