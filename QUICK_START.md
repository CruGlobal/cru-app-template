# Quick Start — a Cru App on pipeline v2

How a Cru application built from this template is provisioned, built, and
deployed. If you're a coding agent, read **[AGENTS.md](./AGENTS.md)** first — it
covers day-to-day work; this file covers the platform around it.

## 1. Activate a stack

This template isn't tied to a runtime or a language. Pick both:

```bash
bin/use-stack <ecs|cloudrun|lambda> <nodejs|ruby|python>
```

That copies the chosen stack to the repo root — a minimal app with a
`Dockerfile`, `build.sh`, `.tool-versions`, and a dependency manifest — writes
the type into `.github/workflows/pipeline-v2.yml`, generates
`.github/dependabot.yml` and `.github/workflows/ci.yml`, and removes the rest.
Commit the result, then build your app on top of it. (See
[AGENTS.md](./AGENTS.md) for the per-stack run commands.)

<!-- CRU:STACK -->
**No stack is activated yet.** Run the command above; this section then names the
stack you chose and the Terraform module that provisions it.
<!-- /CRU:STACK -->

## 2. Provision the application (TerraBloks)

All Cru cloud infrastructure is managed as code in the
[`cru-terraform`](https://github.com/CruGlobal/cru-terraform) repo. You don't
create cloud resources by hand — you generate them with **TerraBloks**, Cru's
Terraform templating engine. The module you want depends on the type you
activated:

| type       | Terraform module                                                                                     |
| ---------- | ---------------------------------------------------------------------------------------------------- |
| `ecs`      | [aws/ecs/app](https://github.com/CruGlobal/cru-terraform-modules/blob/main/aws/ecs/app/README.md)         |
| `cloudrun` | [gcp/cloudrun/app](https://github.com/CruGlobal/cru-terraform-modules/blob/main/gcp/cloudrun/app/README.md) |
| `lambda`   | [aws/lambda/app](https://github.com/CruGlobal/cru-terraform-modules/blob/main/aws/lambda/app/README.md)   |

TerraBloks is available as an MCP server (`terrabloks`) that a coding agent can
drive directly:

```
list_templates → get_template → preview_pr → create_pr
```

It opens a Pull Request against `cru-terraform` containing your app's
infrastructure (the runtime, any database, secrets, and the GitHub deploy
permissions), plus its `CruApplicationInfo` record — which is what tells the
pipeline your app's provider and type at deploy time. A maintainer reviews and
applies it; applying creates the real resources and wires up the permissions this
repo's CI uses to build.

Under pipeline v2 the surfaces are **release-candidate** (the stage surface) and
**production**.

**The GitHub repository itself is provisioned this way too** — TerraBloks'
`github-repository` template creates it from **this** template repo and sets up
the branch ruleset (squash-only, required checks `lint-and-build` and
**Validate PR Title**). If you generated this repo by hand from the template,
still run that template so the ruleset and permissions exist.

## 3. Build & deploy (pipeline v2)

This app is on **pipeline v2: build once, then promote.** One
environment-agnostic image is built from `main`, deployed to release-candidate,
and — if it's good — promoted byte-for-byte to production. There is **no
`staging` branch, no `On Staging` label, and no merge-bot.**

**The flow:**

1. Work on a branch off `main` and open a Pull Request. The title must be a
   Conventional Commit (`feat: …`, `fix: …`) — the **Validate PR Title** check
   enforces it. PRs are **squash-merged with auto-merge** once the required
   checks (**lint-and-build**, **Validate PR Title**) pass.
2. [`.github/workflows/pipeline-v2.yml`](./.github/workflows/pipeline-v2.yml)
   builds on a **nightly-if-changed cron at 05:00 UTC** (midnight EST / 1am
   EDT), or on demand via **workflow_dispatch**. Merging to `main` does *not*
   trigger a build. The build runs `build.sh`, pushes a candidate image tagged
   `candidate-<yyyy-mm-dd>-<n>` to the shared registry, and dispatches a deploy
   to **release-candidate**. If `main` hasn't moved since the last build, the
   candidate is reused and the deploy no-ops.
3. **Promote to production** — run the Promote Action in
   [`cru-deploy`](https://github.com/CruGlobal/cru-deploy), or `cru deploy` from
   the CLI. It deploys the exact digest that has been running on
   release-candidate; nothing is rebuilt, and the release is tagged
   `release-<yyyy-mm-dd>-<n>`.
4. **Roll back** — run the Rollback Action in `cru-deploy` against an earlier
   release. Because the artifact is immutable, a rollback restores the exact
   bytes that were running before.

All deployments happen in `cru-deploy`, never in this repo. **Builds fail until
step 2's Terraform is applied** — that's expected, so don't chase it.

## 4. The Cru CLI

The `cru` CLI talks to Cru's platform — use it to run commands against a real
environment's injected secrets without ever copying secret values locally:

```bash
# Run a command with staging's secrets injected as environment variables:
cru application impersonate -e staging -- <command>

# Read specific secret keys (values are never committed):
cru application secrets read --keys DATABASE_URL -e staging
```

Run `cru --help` for the full command set.

## Reference

- **[AGENTS.md](./AGENTS.md)** — how coding agents should work in this repo.
- [`cru-terraform`](https://github.com/CruGlobal/cru-terraform) — infrastructure as code.
- [`cru-deploy`](https://github.com/CruGlobal/cru-deploy) — where deployments, promotions and rollbacks run.
- [`cru-app-template`](https://github.com/CruGlobal/cru-app-template) — the template this repo came from.
- **deploys.cru.org** — the fleet dashboard: what is running where, and each app's deploy timeline.
