# My Cru App

> Replace this with a sentence describing what your app does.

A Cru application on **pipeline v2** — build once, promote the artifact.

## Getting started

This repo starts from [`cru-app-template`](https://github.com/CruGlobal/cru-app-template)
and isn't tied to a stack yet. A stack is two choices:

```bash
bin/use-stack <ecs|cloudrun|lambda> <nodejs|ruby|python>
```

|                | `nodejs`             | `ruby`                      | `python` |
| -------------- | -------------------- | --------------------------- | -------- |
| **`ecs`**      | TypeScript on Node   | Rack (upgradable to Rails)  | Flask    |
| **`cloudrun`** | TypeScript on Node   | Rack (upgradable to Rails)  | Flask    |
| **`lambda`**   | TypeScript handler   | Ruby handler                | Python handler |

- **type** — where it runs. `ecs` and `cloudrun` give you a web server that
  listens on `$PORT`; `lambda` gives you a handler invoked by an event.
- **language** — what it's written in. If you have no preference, pick
  **`nodejs`**: it's the most common here and the easiest to grow.

Run it once, review, and commit. It copies your stack to the repo root, wires up
CI and dependency updates, and removes everything it no longer needs — including
itself.

<!-- CRU:STACK -->
**No stack is activated yet.** Run `bin/use-stack` above; this section then
describes the stack you chose, and how to install and run it.
<!-- /CRU:STACK -->

## Database (Cloud SQL)

<!-- CRU:DATABASE -->
**No stack is activated yet.** After `bin/use-stack`, this section describes the
database helper that ships with the `ecs` and `cloudrun` stacks in your language,
including how it logs in to Cloud SQL and which variables it reads.
<!-- /CRU:DATABASE -->

## Deploying

Merging to `main` does not deploy. A nightly build produces a candidate image
that auto-deploys to **release-candidate**; you **promote** that exact artifact
to production from [`cru-deploy`](https://github.com/CruGlobal/cru-deploy). See
**[QUICK_START.md](./QUICK_START.md)** for provisioning (TerraBloks), the Cru
CLI, and the full flow.

## For coding agents

See **[AGENTS.md](./AGENTS.md)** — it explains how this repo is wired and how to
work in it safely.
