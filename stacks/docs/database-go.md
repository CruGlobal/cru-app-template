`internal/db/db.go` connects to Cloud SQL Postgres. Call `db.Open(ctx)` to get a
[`pgx`](https://github.com/jackc/pgx) connection pool.

Nothing connects at startup. The pool opens connections only when your code runs
a query, so an app with no database is unaffected and `/up` keeps working.

```go
// Import it as "<module>/internal/db", where <module> is the path after
// `module` in go.mod.
pool, err := db.Open(ctx)
if err != nil {
	return err
}
defer pool.Close()

var now time.Time
if err := pool.QueryRow(ctx, "select now()").Scan(&now); err != nil {
	return err
}
```

Open the pool once, share it, and close it on shutdown. A pool is safe to use
from many goroutines at once.

New apps on the shared Cloud SQL setup log in with **IAM** on Cloud Run. The
Cloud Run service account is the database user, so there is no password to
store. The platform sets these variables for you:

| Variable        | Meaning                                                 |
| --------------- | ------------------------------------------------------- |
| `DATABASE_HOST` | Address of the instance                                 |
| `DATABASE_NAME` | Database to open                                        |
| `DATABASE_USER` | Database user, exactly as Terraform created it          |
| `PGSSLMODE`     | TLS setting for password mode. IAM mode always encrypts |

The helper picks a mode from those variables:

- **IAM mode** (the default on Cloud Run): `DATABASE_HOST`, `DATABASE_NAME` and
  `DATABASE_USER` are set. The helper connects straight to the private IP on
  port 5432. It gets a Google login token (scope `sqlservice.login`) from the
  default credentials and uses it as the password. The token is cached and
  renewed before it expires, and each new connection asks for it, so a
  connection opened after a long idle gap still works. The connection is
  encrypted but the server certificate is not checked, because a per-instance CA
  signs it.
- **Password mode** (works anywhere, including ECS): `DATABASE_URL` or
  `DATABASE_PASSWORD` is set. The helper uses it and does no IAM work. With
  `DATABASE_PASSWORD`, the host, database and user come from `DATABASE_HOST`,
  `DATABASE_NAME` and `DATABASE_USER`. This is also how you point the app at a
  local Postgres, for example
  `DATABASE_URL=postgres://user:pass@localhost:5432/mydb`. pgx reads `sslmode`
  from the URL, or from `PGSSLMODE`, the way `psql` does: `require` encrypts
  without checking the server certificate, `verify-ca` and `verify-full` check
  it, `disable` turns TLS off, and nothing at all means `prefer`, which tries TLS
  and falls back to a plain connection.
- **Off**: neither is set. `db.Open` returns an error that names the missing
  variables.

**Migrations.** A migrations job uses the same variables and the same helper, so
if you turn on `database_migrations` in the Terraform, call `db.Open` from your
migration code. The job runs the app's image with the Terraform's `command` in
place of the image's `CMD`, and the image holds only `/server`. So either give
the server a `migrate` subcommand and set `command = ["/server", "migrate"]`, or
build a second binary (say `cmd/migrate`) into the image next to `/server` and
set `command = ["/migrate"]`. Keep `CMD`, not `ENTRYPOINT`, in the Dockerfile:
on ECS the job's `command` replaces only `CMD`, so with an `ENTRYPOINT` the job
would start a second web server instead of migrating.

**Adding it to an existing app.** Copy `stacks/server/go/internal/db/db.go` from
the [template repo](https://github.com/CruGlobal/cru-app-template) into the app,
then run `go get github.com/jackc/pgx/v5 golang.org/x/oauth2` and `go mod tidy`.

Do not put a database password in the repo. If you must use password mode, get
the value from the platform's secrets (see QUICK_START.md).
