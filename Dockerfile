# syntax=docker/dockerfile:1.7

# Build stage
# Go 1.22 (LTS) for arm64 toolchain availability on the Coolify VPS.
# 1.25 lacks an arm64 build in the upstream toolchain cache at the time
# of writing (2026-09); bumping back when 1.25+ ships arm64.
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/dys ./cmd/dys

# Runtime stage
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /out/dys /usr/local/bin/dys
ENTRYPOINT ["/usr/local/bin/dys"]
