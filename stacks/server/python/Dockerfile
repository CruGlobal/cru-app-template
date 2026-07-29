# PYTHON_VERSION set by build.sh from .tool-versions
ARG PYTHON_VERSION=3.12
FROM public.ecr.aws/docker/library/python:${PYTHON_VERSION}-slim

# DataDog logs source
LABEL com.datadoghq.ad.logs='[{"source": "python"}]'

ENV PYTHONUNBUFFERED=1
WORKDIR /app

# Install dependencies first so this layer caches between code changes.
COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

# NOTE: pipeline v2 is build-once — ONE env-agnostic image is promoted from
# release-candidate to production, so nothing environment-specific may be baked
# in. Read configuration from environment variables at runtime; Terraform and
# the secrets injection supply them per environment.

# The app listens on $PORT (default 8080); gunicorn binds there and serves the
# `app` object in app/main.py. The platform routes traffic to that port.
EXPOSE 8080
CMD ["sh", "-c", "gunicorn --bind 0.0.0.0:${PORT:-8080} app.main:app"]

# Build identity -> Datadog version (D10 bare suffix, e.g. 2026-07-24-10058).
# Passed by build-candidate as --build-arg VERSION (build.sh forwards
# $DOCKER_ARGS); "dev" for local builds. Runtime-only and last on purpose: a
# new build number then invalidates no earlier layer.
ARG VERSION="dev"
ENV DD_VERSION=${VERSION}
