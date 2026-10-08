# NODE_VERSION set by build.sh based on .tool-versions file
ARG NODE_VERSION=latest
# Use standard Node.js image for building
FROM public.ecr.aws/docker/library/node:${NODE_VERSION} AS builder

# Build the lambda function
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build

# Download and extract the secrets-lambda-extension
FROM public.ecr.aws/docker/library/alpine:latest AS extension
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

# NODE_VERSION set by build.sh based on .tool-versions file
ARG NODE_VERSION=latest
# Use AWS Lambda Node.js base image for the final image
FROM public.ecr.aws/lambda/nodejs:${NODE_VERSION}

# NOTE: pipeline v2 is build-once — ONE env-agnostic image is promoted from
# release-candidate to production, so nothing environment-specific may be baked
# in. The v1 PROJECT_NAME / ENVIRONMENT / BUILD_NUMBER build args are gone; the
# aws/lambda/app Terraform module supplies PROJECT_NAME and ENVIRONMENT as
# function environment variables at deploy time.

# Set environment variables for the Lambda function
ENV NODE_OPTIONS=--enable-source-maps

# Set the Lambda task root directory
WORKDIR ${LAMBDA_TASK_ROOT}

# Copy the secrets-lambda-extension from the extension stage and setup wrapper
COPY --from=extension /opt/secrets-lambda-extension /opt/secrets-lambda-extension
ENV AWS_LAMBDA_EXEC_WRAPPER=/opt/secrets-lambda-extension/secrets-wrapper

# Setup DataDog metrics/logs
RUN npm install datadog-lambda-js dd-trace
COPY --from=public.ecr.aws/datadog/lambda-extension:latest /opt/. /opt/
CMD ["node_modules/datadog-lambda-js/dist/handler.handler"]

# Copy the built application from the builder stage
COPY --from=builder /app/dist/. ./

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
