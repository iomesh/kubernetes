# 定制 kube-apiserver 镜像（DR / 内部发行）

本目录包含 **Dockerfile** 与 **`build.sh`**（真正执行 `make release-images` 与 `docker buildx` 的脚本）。Dockerfile 在官方 `k8s.gcr.io/kube-apiserver:v1.23.4` 镜像上替换 `kube-apiserver` 二进制。

**构建上下文必须是仓库根目录**（`_output/dockerized/bin/linux/${TARGETARCH}/kube-apiserver` 相对根路径）。

## 开发分支与合入

1. 从 `v1.23.4-dr` 检出新分支进行修改  
2. 向 `v1.23.4-dr` 提 PR  

## 构建与推送

在**仓库根目录**执行脚本（推荐）：

```bash
./build/kube-apiserver-dr/build.sh
```

发版时在命令前带上环境变量即可，**不必改脚本或 Dockerfile**：

```bash
IMAGE_REPO=registry.smtx.io/backup-dr/kube-apiserver IMAGE_TAG=v1.23.4-dr.0 ./build/kube-apiserver-dr/build.sh
```

脚本默认使用 `IMAGE_REPO=registry.smtx.io/backup-dr/kube-apiserver`、`IMAGE_TAG=v1.23.4-temp`（仅当对应变量未设置时）。

### 等价的手动命令

若不想用脚本，效果与 `build.sh` 相同：

```bash
export IMAGE_REPO="${IMAGE_REPO:-registry.smtx.io/backup-dr/kube-apiserver}"
export IMAGE_TAG="${IMAGE_TAG:-v1.23.4-temp}"

make release-images

docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --provenance=false \
  -f build/kube-apiserver-dr/Dockerfile \
  -t "${IMAGE_REPO}:${IMAGE_TAG}" \
  --push \
  .
```
