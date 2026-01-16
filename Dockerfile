FROM joseluisq/docker-osxcross:1.0.0-beta.2

LABEL version="0.0.1" \
      description="Docker image for cross-compiling Go programs for macOS using osxcross + o64-clang" \
      maintainer="Reda Drissi <reda@example.com>"

# Go version
ARG GO_VERSION=1.22.2
ENV GO_VERSION=${GO_VERSION} \
    PATH=/usr/local/go/bin:/usr/local/osxcross/bin:$PATH \
    CGO_ENABLED=1

# Install Go
RUN set -eux; \
    curl -fsSL https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz | tar -C /usr/local -xzf -

# Set default CC/CXX to osxcross clang
ENV CC=o64-clang
ENV CXX=o64-clang++

# Workdir
WORKDIR /root/src

# Default command
CMD ["bash"]
