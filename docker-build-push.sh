#!/bin/sh

VERSION=$(cat ./version.txt)
IMAGE="securitybunker/databunker"

ARCH=$(uname -m)
case "$ARCH" in
  x86_64)  ARCH_TAG="amd64" ;;
  aarch64|arm64) ARCH_TAG="arm64" ;;
  *) ARCH_TAG="$ARCH" ;;
esac

# Build and push arch-specific image
BUILDKIT_PROGRESS=plain docker build -t $IMAGE:$VERSION-$ARCH_TAG .
docker push $IMAGE:$VERSION-$ARCH_TAG

# Create and push multi-arch manifests (only if both arch images exist)
if docker manifest inspect $IMAGE:$VERSION-amd64 > /dev/null 2>&1 && \
   docker manifest inspect $IMAGE:$VERSION-arm64 > /dev/null 2>&1; then
  # Purge local manifest cache so we don't accumulate stale entries across releases
  docker manifest rm $IMAGE:$VERSION 2>/dev/null || true
  docker manifest rm $IMAGE:latest  2>/dev/null || true
  docker manifest create $IMAGE:$VERSION $IMAGE:$VERSION-amd64 $IMAGE:$VERSION-arm64
  docker manifest push $IMAGE:$VERSION
  docker manifest create $IMAGE:latest $IMAGE:$VERSION-amd64 $IMAGE:$VERSION-arm64
  docker manifest push $IMAGE:latest
  echo "Multi-arch manifest pushed for $IMAGE:$VERSION and $IMAGE:latest"
else
  echo "Skipping manifest: waiting for both amd64 and arm64 images to be pushed."
fi
