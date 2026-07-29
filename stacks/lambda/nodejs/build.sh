#!/usr/bin/env bash
# Builds the Lambda container image. CI (.github/workflows/pipeline-v2.yml, via
# CruGlobal/.github's build-candidate.yml) runs this; you can run it locally
# too. $DOCKER_ARGS is how CI passes the registry tags, --push,
# --provenance=false, and --build-arg VERSION — always forward it. The Node
# version comes from .tool-versions so the image matches your toolchain.

docker buildx build $DOCKER_ARGS \
  --build-arg NODE_VERSION=$(grep nodejs .tool-versions | awk '{ print $NF }' | cut -d'.' -f1) \
  .
