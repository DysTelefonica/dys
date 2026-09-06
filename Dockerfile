# syntax=docker/dockerfile:1.7

# Runtime-only stage. nixpacks (the active build_pack) already produced
# /build/output/dys in stage 0; this Dockerfile just copies that binary
# into a minimal alpine image and runs it. Using `--from=0` because nixpacks
# does not name its build stage "build".
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=0 /build/output/dys /usr/local/bin/dys
ENTRYPOINT ["/usr/local/bin/dys"]
