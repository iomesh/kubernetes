# customize kube-apiserver

## how to edit

1. checkout a new branch from `v1.23.4-sfs`
2. PR to branch `v1.23.4-sfs1`

## how to build

run `./build-apisever-image.sh` to build docker image, it will push image(both x86-64 and aarch64) to registry.smtx.io/sfs/kube-apiserver
