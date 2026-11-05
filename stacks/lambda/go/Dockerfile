# GO_VERSION set by build.sh from .tool-versions
ARG GO_VERSION=1.27.1
# The builder runs on the build machine's own platform and cross-compiles for
# the image's platform (build.sh asks for linux/amd64), so it is fast on an
# arm64 Mac too.
FROM --platform=$BUILDPLATFORM public.ecr.aws/docker/library/golang:${GO_VERSION}-trixie AS build

WORKDIR /src

# A static binary, so it runs on Lambda's OS-only runtime with nothing else.
ENV CGO_ENABLED=0

# Download modules first so this layer caches between code changes.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Lambda's OS-only runtime runs a binary named bootstrap. The lambda.norpc tag
# leaves out the RPC mode only the retired go1.x runtime used.
ARG TARGETOS
ARG TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -tags lambda.norpc -trimpath -ldflags="-s -w" -o /out/bootstrap ./cmd/handler

# Download and extract the secrets-lambda-extension. This stage only downloads
# files, so it runs on the build machine's own platform. The tarball is the
# linux-amd64 build, which matches the image build.sh asks for.
FROM --platform=$BUILDPLATFORM public.ecr.aws/docker/library/alpine:latest AS extension
# Deliberately tracks the LATEST extension release (pinned versions rotted in
# Dockerfiles fleet-wide; under build-once the artifact itself is the pin and a
# bad release is caught on the release-candidate surface). Resolve `latest`
# once, record what was baked, then fetch that exact version.
RUN apk add --no-cache curl && \
    mkdir -p /opt/secrets-lambda-extension && \
    version=$(curl -fsIL -o /dev/null -w '%{url_effective}' https://github.com/CruGlobal/secrets-lambda-extension/releases/latest | sed 's|.*/tag/||') && \
    echo "secrets-lambda-extension: ${version}" && \
    echo "${version}" > /opt/secrets-lambda-extension/VERSION && \
    wget "https://github.com/CruGlobal/secrets-lambda-extension/releases/download/${version}/secrets-lambda-extension-linux-amd64.tar.gz" -q -O - |tar -xzC /opt/secrets-lambda-extension/

# Lambda's OS-only base image, the one for Go and other compiled languages.
FROM public.ecr.aws/lambda/provided:al2023

# NOTE: pipeline v2 is build-once. ONE env-agnostic image is promoted from
# release-candidate to production, so nothing environment-specific may be baked
# in. The aws/lambda/app Terraform module supplies PROJECT_NAME, ENVIRONMENT
# and the DD_* settings as function environment variables at deploy time.

# Copy the secrets-lambda-extension from the extension stage
COPY --from=extension /opt/secrets-lambda-extension /opt/secrets-lambda-extension

# Setup DataDog metrics/logs: copy the lambda-extension. The Datadog wrapper in
# cmd/handler/main.go sends it metrics and traces, and it forwards the logs.
COPY --from=public.ecr.aws/datadog/lambda-extension:latest /opt/. /opt/

# The function itself. The base image's runtime looks for it here.
COPY --from=build /out/bootstrap ${LAMBDA_TASK_ROOT}/bootstrap

# Secrets. The OS-only runtime ignores AWS_LAMBDA_EXEC_WRAPPER, the way the
# other Lambda stacks start the secrets wrapper, so this image starts it
# itself. The wrapper reads the app's SSM parameters into the environment, then
# runs the base image's own entry point, which runs bootstrap. If the wrapper
# can't read them, it exits and the function fails to start, as in the other
# stacks. Keep this line: without it the function starts with no secrets.
ENTRYPOINT ["/opt/secrets-lambda-extension/secrets-wrapper", "/lambda-entrypoint.sh"]
# The base image's entry point needs a handler name and only passes it on as
# _HANDLER, which bootstrap ignores. The Terraform module sets its own value,
# and that is fine.
CMD ["bootstrap"]

# The commit this image is built from, passed by build-candidate as
# --build-arg GIT_SHA. Error reports send it as their code_version, so each
# error lines up with the release that shipped it. Empty for local builds, and
# placed near the end for the same reason as VERSION below.
ARG GIT_SHA=""
ENV GIT_SHA=${GIT_SHA}

# Build identity -> Datadog version (D10 bare suffix, e.g. 2026-07-24-10058).
# Passed by build-candidate as --build-arg VERSION (build.sh forwards
# $DOCKER_ARGS); "dev" for local builds. Runtime-only and last on purpose: a
# new build number then invalidates no earlier layer.
ARG VERSION="dev"
ENV DD_VERSION=${VERSION}
