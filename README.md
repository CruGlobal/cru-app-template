# My Cru App

> Replace this with a sentence describing what your app does.

A Cru application on **pipeline v2** — build once, promote the artifact.

## Getting started

This repo starts from [`cru-app-template`](https://github.com/CruGlobal/cru-app-template)
and isn't tied to a stack yet. A stack is two choices:

```bash
bin/use-stack <ecs|cloudrun|lambda> <nodejs|ruby|python|go>
```

|                | `nodejs`             | `ruby`                      | `python`       | `go`              |
| -------------- | -------------------- | --------------------------- | -------------- | ----------------- |
| **`ecs`**      | TypeScript on Node   | Rack (upgradable to Rails)  | Flask          | `net/http` server |
| **`cloudrun`** | TypeScript on Node   | Rack (upgradable to Rails)  | Flask          | `net/http` server |
| **`lambda`**   | TypeScript handler   | Ruby handler                | Python handler | n/a               |

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

## Sign-in (Google IAP)

New **`cloudrun`** apps from TerraBloks sit behind **Google Identity-Aware
Proxy**: people sign in with Okta at the load balancer, before a request ever
reaches the app. The `ecs` and `cloudrun` stacks check IAP's signed assertion
with [`cru-iap`](https://github.com/CruGlobal/cru-iap):

- **`IAP_AUDIENCE` set** (the platform sets it when IAP is on): every route
  except the `/up` health check needs a verified assertion. Anything else gets
  a 401 and an `iap_rejected` log line with the reason. `/` greets the
  signed-in email.
- **`IAP_AUDIENCE` unset** (ECS, an app that opted out of IAP, or your laptop):
  no gate. To act as someone locally, run with
  `CRU_IAP_DEV_BYPASS_EMAIL=you@cru.org`. The library refuses it whenever
  `IAP_AUDIENCE` is set or the app is running on Cloud Run.
- **Sign out** by linking to `/?gcp-iap-mode=CLEAR_LOGIN_COOKIE`. No app route
  is needed.
- **A path that must skip sign-in** (a webhook, a cron call) needs two things:
  an entry in the Terraform's `iap.bypass_paths` (the default is `["/up"]`),
  and an exemption in the app's gate, plus its own credential check.
- **Opting out:** pick an Okta option other than IAP in TerraBloks. No
  `IAP_AUDIENCE` is set, so the gate stays off.

The gate lives in `src/app.ts` (Node), `app/main.py` (Python), `lib/app.rb`
(Ruby) or `cmd/server/iap.go` (Go).

## Deploying

Merging to `main` does not deploy. A nightly build produces a candidate image
that auto-deploys to **release-candidate**; you **promote** that exact artifact
to production from [`cru-deploy`](https://github.com/CruGlobal/cru-deploy). See
**[QUICK_START.md](./QUICK_START.md)** for provisioning (TerraBloks), the Cru
CLI, and the full flow.

## For coding agents

See **[AGENTS.md](./AGENTS.md)** — it explains how this repo is wired and how to
work in it safely.
