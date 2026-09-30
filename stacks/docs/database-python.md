`app/db.py` connects to Cloud SQL Postgres. Call `connect()` to get a
[`pg8000`](https://github.com/tlocke/pg8000) connection.

Nothing connects at startup. The app only talks to the database when your code
calls the helper, so an app with no database is unaffected and `/health` keeps
working.

```python
# Run locally with `python -m app.main` so `app.db` resolves.
from app.db import connect

conn = connect()
try:
    cur = conn.cursor()
    cur.execute("select now()")
    print(cur.fetchone())
finally:
    conn.close()
```

The helper does no pooling. Open a connection when you need one and close it, or
wrap `connect` in the pool your framework provides.

New apps on the shared Cloud SQL setup log in with **IAM** on Cloud Run. The
Cloud Run service account is the database user, so there is no password to
store. The platform sets these variables for you:

| Variable                   | Meaning                                                 |
| -------------------------- | ------------------------------------------------------- |
| `DATABASE_HOST`            | Address of the instance                                 |
| `DATABASE_NAME`            | Database to open                                        |
| `DATABASE_USER`            | Database user, exactly as Terraform created it          |
| `INSTANCE_CONNECTION_NAME` | Instance name, read by the Cloud SQL Python connector   |
| `PGSSLMODE`                | TLS setting for password mode. IAM mode always encrypts |

The helper picks a mode from those variables:

- **IAM mode** (the default on Cloud Run): `DATABASE_HOST`, `DATABASE_NAME`,
  `DATABASE_USER` and `INSTANCE_CONNECTION_NAME` are set. The helper uses the
  Cloud SQL Python connector with IAM login over the private IP. The connector
  finds the address from `INSTANCE_CONNECTION_NAME`, so it does not use
  `DATABASE_HOST`. It refreshes its credentials lazily, which suits Cloud Run,
  where the CPU is throttled between requests.
- **Password mode** (works anywhere, including ECS): `DATABASE_URL` or
  `DATABASE_PASSWORD` is set. The helper uses it and does no IAM work. With
  `DATABASE_PASSWORD`, the host, database and user come from `DATABASE_HOST`,
  `DATABASE_NAME` and `DATABASE_USER`. This is also how you point the app at a
  local Postgres, for example
  `DATABASE_URL=postgres://user:pass@localhost:5432/mydb`. The helper reads
  `sslmode` from the URL, or from `PGSSLMODE`, the way `psql` does: `require`
  encrypts without checking the server certificate, `verify-ca` and
  `verify-full` check it, and `disable` or nothing turns TLS off.
- **Off**: neither is set. `connect()` raises a clear error.

**Migrations.** A migrations job uses the same variables and the same helper, so
if you turn on `database_migrations` in the Terraform, call `connect()` from
your migration script. No extra setup is needed.

**Adding it to an existing app.** Copy `stacks/server/python/app/db.py` from the
[template repo](https://github.com/CruGlobal/cru-app-template) into the app and
add `cloud-sql-python-connector[pg8000]` and `pg8000` to `requirements.txt`.

Do not put a database password in the repo. If you must use password mode, get
the value from the platform's secrets (see QUICK_START.md).
