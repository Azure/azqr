# This Dockerfile uses a multi-stage build to create a minimal, multi-architecture image for the azqr server.
# To build a multi-architecture image, use Docker Buildx:
# docker buildx build --platform linux/amd64,linux/arm64 -t <your-image-name>:latest --push .

# --- Build stage: build the Go binary statically for the target platform ---
FROM --platform=$BUILDPLATFORM golang:1.22-alpine AS builder

# Set necessary build arguments for cross-compilation
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

# Set the working directory inside the builder image
WORKDIR /src

# Copy the entire project source code
COPY ../../ .

# Download dependencies (improves build cache usage)
RUN go mod download

# Build the azqr server as a statically linked binary for the target platform
# - Trimpath removes file system paths for reproducible builds
# - CGO_ENABLED=0 ensures a static binary
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GOARM=${TARGETVARIANT#v} go build -trimpath -o /out/azqr ./cmd/azqr

# --- Final stage: minimal image using distroless ---
FROM gcr.io/distroless/static:nonroot

# Expose the server port
EXPOSE 8080/tcp

# Set the working directory in the final image
WORKDIR /app

# Copy the statically built binary from the builder stage
COPY --from=builder /out/azqr ./azqr

# Set the entrypoint and default command
ENTRYPOINT ["./azqr"]
CMD ["mcp"]