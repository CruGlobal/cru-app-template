`src/db.ts` connects to Cloud SQL Postgres. Call `getPool()` to get a
[`pg`](https://node-postgres.com) pool.

Nothing connects at startup. The app only talks to the database when your code
calls the helper, so an app with no database is unaffected and `/health` keeps
working.

```ts
// The ".ts" import works in `npm run dev` and in the build.
import { closeDb, getPool } from "./db.ts";

const pool = await getPool();
try {
  const { rows } = await pool.query("select now()");
  console.log(rows[0]);
} finally {
  await closeDb(); // a one-off script must call this or it never exits
}
```

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
  default credentials and uses it as the password, with a fresh token for every
  new connection because tokens last about an hour. The connection is encrypted
  but the server certificate is not checked, because a per-instance CA signs it.
- **Password mode** (works anywhere, including ECS): `DATABASE_URL` or
  `DATABASE_PASSWORD` is set. The helper uses it and does no IAM work. With
  `DATABASE_PASSWORD`, the host, database and user come from `DATABASE_HOST`,
  `DATABASE_NAME` and `DATABASE_USER`. This is also how you point the app at a
  local Postgres, for example
  `DATABASE_URL=postgres://user:pass@localhost:5432/mydb`. The helper reads
  `sslmode` from the URL, or from `PGSSLMODE`, the way `psql` does: `require`
  encrypts without checking the server certificate, `verify-ca` and
  `verify-full` check it, and `disable` or nothing turns TLS off.
- **Off**: neither is set. `getPool()` throws a clear error.

**Migrations.** A migrations job uses the same variables and the same helper, so
if you turn on `database_migrations` in the Terraform, call `getPool()` from your
migration script. No extra setup is needed. Call `await closeDb()` in a
`finally` block when the work is done. The open connection pool keeps the
process alive, and the job will not exit until you do.

**Adding it to an existing app.** Copy `stacks/server/nodejs/src/db.ts` from the
[template repo](https://github.com/CruGlobal/cru-app-template) into the app and
add `google-auth-library` and `pg` as dependencies, plus `@types/pg` as a dev
dependency. Also set `rewriteRelativeImportExtensions` to `true` in
`tsconfig.json` (TypeScript 5.7 or newer) if you import it as `./db.ts`.

Do not put a database password in the repo. If you must use password mode, get
the value from the platform's secrets (see QUICK_START.md).
