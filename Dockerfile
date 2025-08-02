# Dockerfile for kubelogin
# This Dockerfile copies a pre-built binary into a minimal scratch image.
# The binary should be built before running docker build
FROM scratch

# Build arguments for multi-architecture support
ARG TARGETARCH=amd64

# Copy the pre-built binary from local build to /usr/local/bin
COPY bin/linux_${TARGETARCH}/azqr /usr/local/bin/azqr

# Set the entrypoint
ENTRYPOINT ["/usr/local/bin/azqr"]