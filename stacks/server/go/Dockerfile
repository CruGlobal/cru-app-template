# GO_VERSION set by build.sh from .tool-versions
ARG GO_VERSION=1.27.1
FROM public.ecr.aws/docker/library/golang:${GO_VERSION}-trixie AS build

WORKDIR /src

# A static binary, so the final image needs no C library.
ENV CGO_ENABLED=0

# Download modules first so this layer caches between code changes.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# distroless/static holds the binary and little else: no shell, no package
# manager, and it runs as a non-root user.
FROM gcr.io/distroless/static-debian13:nonroot

# DataDog logs source
LABEL com.datadoghq.ad.logs='[{"source": "go"}]'

COPY --from=build /out/server /server

# NOTE: pipeline v2 is build-once — ONE env-agnostic image is promoted from
# release-candidate to production, so nothing environment-specific may be baked
# in. Read configuration from environment variables at runtime; Terraform and
# the secrets injection supply them per environment.

# The app listens on $PORT (default 8080); the platform routes traffic there.
# CMD rather than ENTRYPOINT, so a Terraform `command` (a migrations job, a
# scheduled task) replaces it on ECS as well as on Cloud Run: ECS's `command`
# only ever replaces CMD.
EXPOSE 8080
CMD ["/server"]

# Build identity -> Datadog version (D10 bare suffix, e.g. 2026-07-24-10058).
# Passed by build-candidate as --build-arg VERSION (build.sh forwards
# $DOCKER_ARGS); "dev" for local builds. Runtime-only and last on purpose: a
# new build number then invalidates no earlier layer.
ARG VERSION="dev"
ENV DD_VERSION=${VERSION}
