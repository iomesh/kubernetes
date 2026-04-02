# Replaces the upstream kube-apiserver binary in the official v1.23.4 image.
# Build context must be the Kubernetes repository root (see build.sh).

FROM k8s.gcr.io/kube-apiserver:v1.23.4

ARG TARGETARCH

COPY _output/dockerized/bin/linux/${TARGETARCH}/kube-apiserver /usr/local/bin/kube-apiserver
