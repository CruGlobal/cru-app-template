# RUBY_VERSION set by build.sh from .tool-versions
ARG RUBY_VERSION=3.4
FROM public.ecr.aws/docker/library/ruby:${RUBY_VERSION}-alpine

# DataDog logs source
LABEL com.datadoghq.ad.logs='[{"source": "ruby"}]'

# build-base is needed to compile native gem extensions (e.g. puma).
RUN apk upgrade --no-cache && apk --no-cache add build-base

WORKDIR /app

# Install gems first so this layer caches between code changes. There's no
# Gemfile.lock in a fresh template, so bundler resolves and writes one here.
# Once you commit a Gemfile.lock, builds become reproducible — which is what
# build-once wants.
COPY Gemfile ./
RUN bundle install --jobs 4 --retry 2

COPY . .

ENV RACK_ENV=production

# NOTE: pipeline v2 is build-once — ONE env-agnostic image is promoted from
# release-candidate to production, so nothing environment-specific may be baked
# in. Read configuration from environment variables at runtime; Terraform and
# the secrets injection supply them per environment.

# The app listens on $PORT (default 8080); puma binds there and serves
# config.ru. The platform routes traffic to that port.
EXPOSE 8080
CMD ["sh", "-c", "bundle exec puma -p ${PORT:-8080}"]

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
