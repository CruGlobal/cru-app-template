# Starter stacks

`bin/use-stack <type> <language>` copies one of these to the repo root and
deletes the rest. Two axes:

| Axis         | Values                          |
| ------------ | ------------------------------- |
| **type**     | `ecs`, `cloudrun`, `lambda`     |
| **language** | `nodejs`, `ruby`, `python`      |

## Why there is no `stacks/ecs/` or `stacks/cloudrun/`

There are two *shapes* of app here, not three:

- **A server** that listens on `$PORT` and answers a health check. This is what
  both **ECS** and **Cloud Run** run, and the scaffold is byte-for-byte the
  same — the platform difference (an ALB target group vs. Cloud Run's own
  router) is entirely on the infrastructure side, not in the app. So both types
  share `stacks/server/<language>/`, and `bin/use-stack` maps `ecs` and
  `cloudrun` onto it.
- **A handler** invoked by an event, with no port and no health URL. That's
  **Lambda**: `stacks/lambda/<language>/`. It needs its own scaffold because
  the base image, the entry point, and Cru's Lambda wiring (the
  secrets-lambda-extension exec wrapper, the DataDog lambda-extension) are all
  different.

The three predecessor templates (`cru-app-ecs-template`,
`cru-app-cloudrun-template`, `cru-app-lambda-template`) proved this: the ECS and
Cloud Run stacks differed only in code comments and one Flask patch version.
Duplicating them here would have been two copies to keep in sync for no
behavioural difference.

The **type** still matters, and it is a real input — `bin/use-stack` writes it
into `.github/workflows/pipeline-v2.yml` as the `type:` passed to
`build-candidate.yml`, which selects the build recipe (ECR vs. Artifact
Registry, the Lambda-specific `--provenance=false`, and so on). Everything
*after* the build — deploy, promote, rollback — resolves the type server-side
from the app's `CruApplicationInfo` record, which is why those workflows are
type-agnostic and this repo ships one pipeline for all three types.

## What every stack contains

- `Dockerfile` — ends with the fleet's `ARG VERSION="dev"` / `ENV
  DD_VERSION=${VERSION}` pair, which `build-candidate` fills with the build
  identity. Nothing environment-specific is ever baked in: pipeline v2 is
  build-once, and the same bytes run on release-candidate and production.
- `build.sh` — what CI runs. It forwards `$DOCKER_ARGS` (tags, `--push`,
  `--build-arg VERSION`) and adds the language version read from
  `.tool-versions`.
- `.tool-versions` — the pinned language version (asdf / mise).
- A dependency manifest, and a minimal app that already builds and works.

## The database helper (server stacks only)

Each `stacks/server/<language>/` also ships a small Cloud SQL connection helper
(`src/db.ts`, `app/db.py` or `lib/db.rb`). ECS and Cloud Run share it. It does
nothing until the app calls it, and it uses IAM login only when the `DATABASE_*`
variables are set, so an ECS app is unaffected. Lambda stacks have none.

Its docs live in `stacks/docs/database-<language>.md`. `bin/use-stack` copies the
one for the chosen language into the `CRU:DATABASE` blocks of README.md,
AGENTS.md and QUICK_START.md. Change the helper and its fragment together.
