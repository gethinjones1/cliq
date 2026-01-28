#!/bin/bash
set -e

# Build script for Cliq
# Builds binaries for all supported platforms

VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE=$(date -u '+%Y-%m-%dT%H:%M:%SZ')

LDFLAGS="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}"

# Output directory
DIST_DIR="dist"
mkdir -p "$DIST_DIR"

# Platforms to build for
PLATFORMS=(
    "darwin/amd64"
    "darwin/arm64"
    "linux/amd64"
    "linux/arm64"
)

echo "Building Cliq version ${VERSION}"
echo "Commit: ${COMMIT}"
echo "Date: ${DATE}"
echo ""

for platform in "${PLATFORMS[@]}"; do
    GOOS=${platform%/*}
    GOARCH=${platform#*/}

    output="cliq-${GOOS}-${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
        output="${output}.exe"
    fi

    echo "Building for ${GOOS}/${GOARCH}..."

    CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH go build \
        -trimpath \
        -ldflags="$LDFLAGS" \
        -o "${DIST_DIR}/${output}" \
        .

    # Create tarball
    if [ "$GOOS" != "windows" ]; then
        tar -czf "${DIST_DIR}/${output}.tar.gz" -C "$DIST_DIR" "$output"
    else
        (cd "$DIST_DIR" && zip -q "${output}.zip" "$output")
    fi

    echo "  -> ${DIST_DIR}/${output}"
done

echo ""
echo "Build complete! Binaries are in ${DIST_DIR}/"
ls -la "$DIST_DIR"
