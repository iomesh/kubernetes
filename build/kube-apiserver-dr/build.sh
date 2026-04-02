#!/usr/bin/env bash
# Build and push multi-arch kube-apiserver (amd64 + arm64).
# Run from repo root. Override when publishing:
#   IMAGE_REPO=registry.smtx.io/sfs/kube-apiserver IMAGE_TAG=v1.23.4-dr.0 ./build/kube-apiserver-dr/build.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$(cd "${SCRIPT_DIR}/../.." && pwd)"

make release-images

IMAGE="${IMAGE_REPO:-registry.smtx.io/backup-dr/kube-apiserver}:${IMAGE_TAG:-v1.23.4-temp}"

docker buildx build \
	--platform linux/amd64,linux/arm64 \
	--provenance=false \
	-f "${SCRIPT_DIR}/Dockerfile" \
	-t "${IMAGE}" \
	--push \
	.

echo "Built: ${IMAGE}"
