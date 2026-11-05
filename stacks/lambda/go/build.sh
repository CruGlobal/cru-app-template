#!/usr/bin/env bash
# Builds the Lambda container image. CI (.github/workflows/pipeline-v2.yml, via
# CruGlobal/.github's build-candidate.yml) runs this; you can run it locally
# too. $DOCKER_ARGS is how CI passes the registry tags, --push,
# --provenance=false, and --build-arg VERSION. Always forward it. The Go
# version comes from .tool-versions so the image matches your toolchain.
#
# --platform linux/amd64: the function runs on x86_64, Lambda's default, so the
# image is built for it even on an arm64 Mac.

docker buildx build $DOCKER_ARGS \
  --platform linux/amd64 \
  --build-arg GO_VERSION="$(awk '$1 == "golang" { print $2 }' .tool-versions)" \
  .
