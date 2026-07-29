# RUBY_VERSION set by build.sh based on .tool-versions file
ARG RUBY_VERSION=3.4

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

# RUBY_VERSION set by build.sh based on .tool-versions file
ARG RUBY_VERSION=3.4
# Use AWS Lambda Ruby base image for the final image
FROM public.ecr.aws/lambda/ruby:${RUBY_VERSION}

# DataDog logs source
LABEL com.datadoghq.ad.logs='[{"source": "ruby"}]'

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

# DataDog's ddtrace (pulled in by datadog-lambda) has native extensions that
# must be compiled, so install the build toolchain before `bundle install`.
RUN dnf install -y gcc gcc-c++ gmp-devel make && dnf clean all

# Install gems (datadog-lambda + ddtrace). There's no Gemfile.lock in a fresh
# template, so bundler resolves and writes one here. Once you commit a
# Gemfile.lock, builds become reproducible — which is what build-once wants.
COPY Gemfile ./
RUN bundle install --jobs 4 --retry 2

# Setup DataDog metrics/logs: copy the lambda-extension layer.
COPY --from=public.ecr.aws/datadog/lambda-extension:latest /opt/. /opt/

# Copy the application code
COPY . ${LAMBDA_TASK_ROOT}

# Boot the AWS Ruby runtime at handler.rb -> def handler(event:, context:).
# DataDog instrumentation is applied in code via Datadog::Lambda.wrap inside
# handler.rb (the gem's documented Ruby approach), so the standard AWS CMD
# below is correct and DataDog still captures traces/metrics.
#
# Alternatively, to let DataDog's wrapper drive the handler instead of the
# in-code wrap, drop the Datadog::Lambda.wrap block from handler.rb and use:
#   ENV DD_LAMBDA_HANDLER=handler.handler
#   CMD ["datadog_lambda/handler.handler"]
CMD ["handler.handler"]

# Build identity -> Datadog version (D10 bare suffix, e.g. 2026-07-24-10058).
# Passed by build-candidate as --build-arg VERSION (build.sh forwards
# $DOCKER_ARGS); "dev" for local builds. Runtime-only and last on purpose: a
# new build number then invalidates no earlier layer.
ARG VERSION="dev"
ENV DD_VERSION=${VERSION}
