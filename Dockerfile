# syntax=docker/dockerfile:1.7

# Build stage — Go 1.24 for bubbletea 1.3.x compatibility.
FROM golang:1.24-alpine AS build
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
CMD ["version"]
