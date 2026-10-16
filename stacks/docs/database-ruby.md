`lib/db.rb` connects to Cloud SQL Postgres. Call `Db.connect` to get a
`PG::Connection` from the [`pg`](https://github.com/ged/ruby-pg) gem.

Nothing connects at startup. The app only talks to the database when your code
calls the helper, so an app with no database is unaffected and `/up` keeps
working.

```ruby
require_relative "lib/db"

Db.with_connection do |conn|
  conn.exec("select now()").first
end
```

The helper does no pooling. Open a connection when you need one and close it, or
wrap `Db.connect` in the pool your framework provides.

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
  `DATABASE_USER` are set. The helper connects straight to the private IP. It
  gets a Google login token (scope `sqlservice.login`) from the default
  credentials with the `googleauth` gem and uses it as the password, with a
  fresh token for every new connection because tokens last about an hour. The
  connection is encrypted but the server certificate is not checked, because a
  per-instance CA signs it.
- **Password mode** (works anywhere, including ECS): `DATABASE_URL` or
  `DATABASE_PASSWORD` is set. The helper uses it and does no IAM work. With
  `DATABASE_PASSWORD`, the host, database and user come from `DATABASE_HOST`,
  `DATABASE_NAME` and `DATABASE_USER`. This is also how you point the app at a
  local Postgres, for example
  `DATABASE_URL=postgres://user:pass@localhost:5432/mydb`. libpq reads
  `sslmode` from the URL, or from `PGSSLMODE`, the way `psql` does.
- **Off**: neither is set. `Db.connect` raises a clear error.

**Migrations.** A migrations job uses the same variables and the same helper, so
if you turn on `database_migrations` in the Terraform, call `Db.connect` from
your migration script. No extra setup is needed.

**Adding it to an existing app.** Copy `stacks/server/ruby/lib/db.rb` from the
[template repo](https://github.com/CruGlobal/cru-app-template) into the app and
add the `pg` gem (version 1.6.1 or newer) and `googleauth` to the `Gemfile`.
Those `pg` versions install prebuilt on the Alpine base image, so the Dockerfile
needs no extra packages.

Do not put a database password in the repo. If you must use password mode, get
the value from the platform's secrets (see QUICK_START.md).
