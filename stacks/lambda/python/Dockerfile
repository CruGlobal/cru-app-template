# PYTHON_VERSION set by build.sh based on .tool-versions file
ARG PYTHON_VERSION=3.12

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

# PYTHON_VERSION set by build.sh based on .tool-versions file
ARG PYTHON_VERSION=3.12
# Use AWS Lambda Python base image for the final image
FROM public.ecr.aws/lambda/python:${PYTHON_VERSION}

# DataDog logs source
LABEL com.datadoghq.ad.logs='[{"source": "python"}]'

# NOTE: pipeline v2 is build-once — ONE env-agnostic image is promoted from
# release-candidate to production, so nothing environment-specific may be baked
# in. The v1 PROJECT_NAME / ENVIRONMENT / BUILD_NUMBER build args are gone; the
# aws/lambda/app Terraform module supplies PROJECT_NAME and ENVIRONMENT as
# function environment variables at deploy time.

# Set the Lambda task root directory
WORKDIR ${LAMBDA_TASK_ROOT}

# Copy the secrets-lambda-extension from the extension stage and setup wrapper
COPY --from=extension /opt/secrets-lambda-extension /opt/secrets-lambda-extension
ENV AWS_LAMBDA_EXEC_WRAPPER=/opt/secrets-lambda-extension/secrets-wrapper

# Install application + DataDog dependencies (datadog-lambda, ddtrace).
COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt

# Setup DataDog metrics/logs: copy the lambda-extension layer, then have the
# DataDog handler wrap our handler. CMD points at DataDog's wrapper, and
# DD_LAMBDA_HANDLER names our actual handler (handler.py -> def handler()).
COPY --from=public.ecr.aws/datadog/lambda-extension:latest /opt/. /opt/
ENV DD_LAMBDA_HANDLER=handler.handler
CMD ["datadog_lambda.handler.handler"]

# Copy the application code
COPY . ${LAMBDA_TASK_ROOT}

# Build identity -> Datadog version (D10 bare suffix, e.g. 2026-07-24-10058).
# Passed by build-candidate as --build-arg VERSION (build.sh forwards
# $DOCKER_ARGS); "dev" for local builds. Runtime-only and last on purpose: a
# new build number then invalidates no earlier layer.
ARG VERSION="dev"
ENV DD_VERSION=${VERSION}
