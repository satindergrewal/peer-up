# Multi-stage Dockerfile for peer-up integration testing.
# Builds a coverage-instrumented binary so Docker tests contribute to
# overall coverage via GOCOVERDIR + `go tool covdata merge`.

FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -cover -o /peerup ./cmd/peerup

# --- Relay server image ---
FROM alpine:3.21 AS relay
RUN apk add --no-cache ca-certificates
COPY --from=builder /peerup /usr/local/bin/peerup
RUN mkdir -p /covdata
WORKDIR /data
ENTRYPOINT ["peerup"]

# --- Node (peerup client/daemon) image ---
FROM alpine:3.21 AS node
RUN apk add --no-cache ca-certificates
COPY --from=builder /peerup /usr/local/bin/peerup
RUN mkdir -p /root/.config/peerup /covdata
WORKDIR /root
ENTRYPOINT ["peerup"]
