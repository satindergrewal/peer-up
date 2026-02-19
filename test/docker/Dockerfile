# Multi-stage Dockerfile for peer-up integration testing.
# Builds both binaries, produces two minimal targets: relay and node.

FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /peerup ./cmd/peerup
RUN CGO_ENABLED=0 go build -o /relay-server ./cmd/relay-server

# --- Relay server image ---
FROM alpine:3.21 AS relay
RUN apk add --no-cache ca-certificates
COPY --from=builder /relay-server /usr/local/bin/relay-server
WORKDIR /data
ENTRYPOINT ["relay-server"]

# --- Node (peerup client/daemon) image ---
FROM alpine:3.21 AS node
RUN apk add --no-cache ca-certificates
COPY --from=builder /peerup /usr/local/bin/peerup
RUN mkdir -p /root/.config/peerup
WORKDIR /root
ENTRYPOINT ["peerup"]
