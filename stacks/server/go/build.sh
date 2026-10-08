#!/usr/bin/env bash
# Builds the container image. CI (.github/workflows/pipeline-v2.yml, via
# CruGlobal/.github's build-candidate.yml) runs this; you can run it locally
# too. $DOCKER_ARGS is how CI passes the registry tags, --push, and
# --build-arg VERSION — always forward it. The Go version comes from
# .tool-versions so the image matches your toolchain.

docker buildx build $DOCKER_ARGS \
  --build-arg GO_VERSION=$(awk '$1 == "golang" { print $2 }' .tool-versions) \
  .
