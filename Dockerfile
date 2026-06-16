# Dockerfile for Azure Quick Review MCP Server
# This Dockerfile copies a pre-built binary into a minimal scratch image.
# The binary should be built before running docker build
ARG BUILDPLATFORM=linux/amd64

# Extract CA certificates from a known-good base image
FROM --platform=$BUILDPLATFORM alpine:3.21 AS certs
RUN apk --no-cache add ca-certificates

FROM --platform=$BUILDPLATFORM scratch

# Build arguments for multi-architecture support
ARG TARGETARCH=amd64

# Include CA certificates so TLS verification works for Azure endpoints
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# Copy the pre-built binary from local build to /usr/local/bin
COPY bin/linux_${TARGETARCH}/azqr /usr/local/bin/azqr

# Expose HTTP/SSE port for MCP server
EXPOSE 8080

# Set the entrypoint with mcp server in HTTP mode as default
ENTRYPOINT ["/usr/local/bin/azqr"]
CMD ["mcp", "--mode", "http", "--addr", ":8080"]