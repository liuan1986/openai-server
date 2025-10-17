# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o server ./cmd/server

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/server /usr/local/bin/openai-server
COPY config.sample.json /app/config.json
ENV CONFIG_PATH=/app/config.json
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/openai-server"]
