# Working in this repo (for coding agents)

This repository was created from Cru's **app template**
([`cru-app-template`](https://github.com/CruGlobal/cru-app-template)) and runs
on **pipeline v2** — build once, then promote the artifact. The person you're
helping may not be a developer: your job is to make good, safe defaults and
explain what you're doing in plain language. This file tells you how the repo is
wired and what to do when you're unsure.

> New here? Do **"Start here"** below before writing app code.

---

## Start here: activate a stack

A fresh repo from this template is **not tied to a runtime or a language**. It
ships starter "stacks" for every combination of the two, each a minimal app that
already builds and works:

| type         | what it gives you                                        |
| ------------ | -------------------------------------------------------- |
| **ecs**      | a web server on AWS ECS, listening on `$PORT`            |
| **cloudrun** | a web server on Google Cloud Run, listening on `$PORT`   |
| **lambda**   | a handler on AWS Lambda, invoked by an event (no port)   |

| language     | what it gives you                                        |
| ------------ | -------------------------------------------------------- |
| **nodejs**   | TypeScript on Node (Lambda bundles with esbuild)         |
| **ruby**     | a minimal Rack app (one command upgrades it to full Rails — see `stacks/server/ruby/config.ru`), or a Ruby handler on Lambda |
| **python**   | Flask, or a Python handler on Lambda                     |

**Your first steps on a new app:**

1. Ask the user **where it should run** and **what language**. If they don't
   know:
   - **type** — a request/response web app or API → **`ecs`** (Cru's default and
     best-supported surface). Event-driven, scheduled, or bursty-and-idle work
     → **`lambda`**. Choose **`cloudrun`** when the app belongs next to other
     GCP resources.
   - **language** — recommend **`nodejs`**; it's the most common here.
2. Activate it: `bin/use-stack <ecs|cloudrun|lambda> <nodejs|ruby|python>`
   That copies the stack to the repo root, writes the type into
   `.github/workflows/pipeline-v2.yml`, generates `.github/dependabot.yml` and
   `.github/workflows/ci.yml`, fills in the docs, and removes everything it no
   longer needs (including itself). It only needs to run once.
3. Commit the result (`chore: use the <type>/<language> stack`).

After that, the root holds a normal app — build on it.

<!-- CRU:STACK -->
**No stack is activated yet** — `stacks/` and `bin/use-stack` are still here, so
step 2 above hasn't run. Do that first; this section then describes the
activated stack and its run commands.
<!-- /CRU:STACK -->

## Project layout (after activation)

```
.
├── AGENTS.md          # this file
├── CLAUDE.md          # points here
├── README.md          # the app's own readme — customize it for the app
├── QUICK_START.md     # how this app is provisioned & deployed at Cru
├── Dockerfile         # how the app is containerized
├── build.sh           # builds the image (CI runs this; you can too)
├── .tool-versions     # pinned language version (asdf / mise)
├── .github/
│   ├── dependabot.yml                       # grouped weekly dependency updates
│   └── workflows/
│       ├── pipeline-v2.yml                  # nightly build + release-candidate deploy
│       ├── ci.yml                           # PR check: lint-and-build
│       ├── conventional-commits.yml         # PR check: Validate PR Title
│       └── dependabot-auto-merge.yml        # auto-merges safe dependency PRs
└── <your app files>   # src/ (node), app/ or handler.py (python), config.ru or handler.rb (ruby)
```

## The Dockerfile

Two rules, both load-bearing:

- **Never bake anything environment-specific into the image.** Pipeline v2 is
  **build-once**: one image is built from `main`, deployed to
  release-candidate, and promoted *byte-for-byte* to production. Anything
  baked in would be wrong in one of those two places. Read configuration from
  environment variables at runtime — Terraform and the secrets injection supply
  them per environment. (This is why the v1 `PROJECT_NAME` / `ENVIRONMENT` /
  `BUILD_NUMBER` build args are gone.)
- **Keep the last two lines.** Every Cru Dockerfile ends with:

  ```dockerfile
  ARG VERSION="dev"
  ENV DD_VERSION=${VERSION}
  ```

  `build-candidate` passes the build identity as `--build-arg VERSION`, and
  that becomes the app's Datadog version. It is last so a new build number
  invalidates no earlier layer, and it is the **only** identity value baked in.

On **Lambda**, also leave Cru's wiring in place: the **secrets-lambda-extension**
(set as `AWS_LAMBDA_EXEC_WRAPPER`, which injects secrets at runtime) and the
**DataDog lambda-extension** plus DataDog instrumentation.

`build.sh` must always forward `$DOCKER_ARGS` — that's how CI passes the registry
tags, `--push`, and `--build-arg VERSION`.

## How this app ships

This repo is on **pipeline v2**: build-once, then promote. There is **no
`staging` branch, no `On Staging` label, and no merge-bot** — those belong to
pipeline v1, and this repo has never had them. If you see them referenced
anywhere, the reference is stale.

1. **Work on a branch** off `main` and open a **Pull Request** back to `main`.
   PRs are **squash-merged with auto-merge** once the required checks pass.
2. **The PR title must be a Conventional Commit** (`feat: …`, `fix(api): …`) —
   it becomes the squash commit subject, and the **Validate PR Title** check
   enforces it.
3. **Builds do not happen on push.** `.github/workflows/pipeline-v2.yml` runs on
   a **nightly-if-changed cron at 05:00 UTC** (midnight EST / 1am EDT) and on
   manual **`workflow_dispatch`**. A build produces a candidate image tagged
   `candidate-<yyyy-mm-dd>-<n>`, which **auto-deploys to release-candidate**
   (the stage surface). If `main` hasn't moved, the build reuses the existing
   candidate and the deploy no-ops — a quiet night is a true end-to-end no-op.
4. **Promotion to production and rollbacks** are Actions in
   [`cru-deploy`](https://github.com/CruGlobal/cru-deploy) (or `cru deploy`
   subcommands in the Cru CLI) — not workflows in this repo. Merging to `main`
   does not deploy anything by itself. Because the artifact is immutable, a
   rollback restores the exact bytes that were running before.

**Nothing builds or deploys until the app's Terraform has been applied** (see
QUICK_START.md). Until then a build fails by design — that's expected, not a bug
to chase.

Watch a run with the GitHub CLI: `gh run watch` (or check the Actions tab).

## Feature flags

The pipeline runs a **feature-flag service**, and it is a first-class part of
this platform — it is what replaces the `staging` branch v2 retired. Unfinished
work merges to `main` **dark** behind a flag, a demo on stage is a flag flip, and
launching is a flag flip in production.

- **`CRU_FLAGS_URL` is injected by the app's Terraform module** in
  release-candidate and production. It is absent locally, in CI and in `lab` — a
  supported state meaning *every flag is off*, never a misconfiguration.
- **Unknown flags read off**, so code can merge before its flag exists. Anything
  behind a flag must behave with the flag off.
- **Never build your own flag system** — no env-var toggles, no config table, no
  hand-rolled fetch of `CRU_FLAGS_URL` — and never wrap the official client in
  retries, caching or timeouts of your own: it already polls with `ETag`
  revalidation and serves last-known-good through an outage. **App code never
  writes a flag.**
- **The CLI is the only writer**: `cru application flags create <name> -n <app>
  --description "…"`, then `enable`/`disable` it `-e <stage|production>`. A flip
  is live within about a minute — no build, no deploy.
- **Use the client that matches this app's language:**

  | language | client | reading a flag |
  | --- | --- | --- |
  | **nodejs** | `npm install @cruglobal/flags` | `flags.enabled("name")` — server side only |
  | **python** | `pip install cru-flags` | `flags.enabled("name")` |
  | **ruby** | stock Flipper — `flipper` + `flipper-active_support_cache_store`, wired by the canonical initializer in the pipeline guide (copy it exactly; the `Failsafe` adapter and the Marshal-safe read wrapper are load-bearing) | `Flipper.enabled?(:name)` |

The full story — the CLI, the wire format, the Rails initializer — is the
[pipeline guide's Feature flags section](https://github.com/CruGlobal/cru-deploy/blob/main/docs/pipeline-v2.md#feature-flags).
An app's flags are visible on the dashboard at <https://deploys.cru.org>.

## Tests & CI

`.github/workflows/ci.yml` runs on every pull request as the **`lint-and-build`**
check: it installs dependencies, builds / syntax-checks the app, and runs your
tests **if there are any** (so a fresh repo stays green). Add real tests as the
app grows — the workflow picks them up automatically, and you should then drop
the `--if-present` / `hashFiles` guard so they're actually enforced.

To *block* merges (and Dependabot auto-merge) until checks pass, mark
**`lint-and-build`** and **Validate PR Title** required in the repo's branch
ruleset — that's configured via TerraBloks / `cru-terraform`, not in this repo.
Don't rename the `lint-and-build` job without updating the ruleset, or merges
will block on a check that never reports.

## Infrastructure & secrets

- **Provisioning** (the runtime — ECS service, Cloud Run service, or Lambda
  function — plus databases, secrets, and the GitHub repo's deploy permissions)
  is generated by **TerraBloks**, Cru's Terraform templating engine, available
  to you as the `terrabloks` MCP server (`list_templates` → `get_template` →
  `preview_pr` → `create_pr`). It opens a Pull Request against the
  `cru-terraform` repo. A maintainer reviews and applies it. See
  **QUICK_START.md**. (Don't hand-write cloud infrastructure.)
- **Secrets** are injected by the platform at runtime — **never commit secrets**
  to this repo. Read configuration from environment variables. To run or debug
  against a real environment's secrets, use the **Cru CLI**:
  `cru application impersonate -e staging -- <command>` (see QUICK_START.md).
- **Build-time** secrets, if the image genuinely needs one, are repo secrets
  named `BUILD_<NAME>`; `pipeline-v2.yml` passes them through with
  `secrets: inherit` and the reusable workflow exports only the `BUILD_*` keys
  into the build environment, with the prefix stripped.

## Database access

<!-- CRU:DATABASE -->
**No stack is activated yet.** The `ecs` and `cloudrun` stacks ship a database
helper; after activation this section names it and says how to use it.
<!-- /CRU:DATABASE -->

## If you're not sure what to do

- **Keep changes small and on a branch.** Open a PR; don't push straight to
  `main`.
- **Don't invent infrastructure.** If the app needs a database, queue, bucket,
  or new secret, that's a TerraBloks/`cru-terraform` change — say so and use the
  `terrabloks` MCP rather than configuring cloud resources by hand.
- **Don't build a feature-flag system.** The pipeline already has one — see
  [Feature flags](#feature-flags). Use the official client for the language and
  flip flags with `cru application flags`.
- **Never paste secrets** (API keys, passwords, tokens) into files. Use env
  vars; fetch real values through the Cru CLI.
- **Never bake environment-specific values into the image** — see "The
  Dockerfile" above. Build-once makes this a correctness rule, not a style one.
- **Ask the user about intent, not plumbing.** "What should this app do?" is a
  great question; make the technical calls yourself with sensible defaults.
- **Confirm before anything outward-facing or hard to undo** — pushing,
  opening PRs, deleting things, promoting to production, sending real
  notifications.
