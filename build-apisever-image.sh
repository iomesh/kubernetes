make release-images
IMAGE_REPO=registry.smtx.io/sfs/kube-apiserver
IMAGE_TAG="v1.23.4-sfs-b1"
docker buildx build --platform linux/amd64,linux/arm64 --provenance=false --push -t ${IMAGE_REPO}:${IMAGE_TAG} -f Dockerfile.apiserver  .
