#!/bin/bash

# Multi-platform release build script
# Builds frontend once, then cross-compiles Go binary for all target platforms.

set -e

# Project name
PROJECT="open-website-defender"
VERSION=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}
OUTPUT_DIR=${OUTPUT_DIR:-"dist"}

# Target platforms (GOOS/GOARCH)
PLATFORMS=(
  "linux/amd64"
  "linux/arm64"
  "darwin/amd64"
  "darwin/arm64"
  "windows/amd64"
)

# Read .env if exists
if [ -f .env ]; then
  echo "Loading .env file..."
  export $(grep -v '^#' .env | xargs)
fi

# Path config
ROOT_PATH=${ROOT_PATH:-"/wall"}
ADMIN_PATH=${ADMIN_PATH:-"/admin"}
GUARD_PATH=${GUARD_PATH:-"/guard"}

LDFLAGS="-s -w \
  -X 'main.RootPath=$ROOT_PATH' \
  -X 'main.AdminPath=$ADMIN_PATH' \
  -X 'main.GuardPath=$GUARD_PATH'"

echo "========================================="
echo "Release Build: $PROJECT $VERSION"
echo "ROOT_PATH: $ROOT_PATH"
echo "ADMIN_PATH: $ADMIN_PATH"
echo "GUARD_PATH: $GUARD_PATH"
echo "Platforms: ${#PLATFORMS[@]}"
echo "========================================="

# 1. Build frontend (once, shared by all platforms)
echo ""
echo "[1/3] Building frontend..."

cd ui/guard
export VITE_ROOT_PATH=$ROOT_PATH
export VITE_ADMIN_PATH=$ADMIN_PATH
export VITE_GUARD_PATH=$GUARD_PATH
npm run build
cd ../..

cd ui/admin
export VITE_ROOT_PATH=$ROOT_PATH
export VITE_ADMIN_PATH=$ADMIN_PATH
export VITE_GUARD_PATH=$GUARD_PATH
npm run build
cd ../..

echo "Frontend build complete."

# 2. Create output directory
echo ""
echo "[2/3] Preparing output directory..."
rm -rf "$OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR"

# 3. Cross-compile for each platform
echo ""
echo "[3/3] Building binaries..."

for PLATFORM in "${PLATFORMS[@]}"; do
  GOOS="${PLATFORM%/*}"
  GOARCH="${PLATFORM#*/}"

  BINARY="${PROJECT}-${GOOS}-${GOARCH}"
  if [ "$GOOS" = "windows" ]; then
    BINARY="${BINARY}.exe"
  fi

  echo "  Building ${GOOS}/${GOARCH}..."
  CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH go build \
    -ldflags "$LDFLAGS" \
    -o "${OUTPUT_DIR}/${BINARY}" \
    main.go

  # Create archive with config
  ARCHIVE_DIR="${OUTPUT_DIR}/release/${PROJECT}-${GOOS}-${GOARCH}"
  mkdir -p "${ARCHIVE_DIR}/config"
  cp "${OUTPUT_DIR}/${BINARY}" "${ARCHIVE_DIR}/"
  cp config/config.yaml "${ARCHIVE_DIR}/config/"

  if [ "$GOOS" = "windows" ]; then
    (cd "${OUTPUT_DIR}/release" && zip -rq "${PROJECT}-${GOOS}-${GOARCH}.zip" "${PROJECT}-${GOOS}-${GOARCH}/")
    mv "${OUTPUT_DIR}/release/${PROJECT}-${GOOS}-${GOARCH}.zip" "${OUTPUT_DIR}/"
  else
    tar -czf "${OUTPUT_DIR}/${PROJECT}-${GOOS}-${GOARCH}.tar.gz" -C "${OUTPUT_DIR}/release" "${PROJECT}-${GOOS}-${GOARCH}/"
  fi

  rm -rf "${ARCHIVE_DIR}"
done

echo ""
echo "========================================="
echo "Release build complete!"
echo "Version: $VERSION"
echo "Output:  $OUTPUT_DIR/"
echo ""
ls -lh "$OUTPUT_DIR/"
echo "========================================="
