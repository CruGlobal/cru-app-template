// Postgres connection helper for Cloud SQL. Nothing connects until you call
// getPool(), so an app with no database is unaffected.
//
// Modes, chosen from environment variables:
//   password: DATABASE_URL or DATABASE_PASSWORD is set. Plain pg connection.
//   iam:      DATABASE_HOST, DATABASE_NAME and DATABASE_USER are set. Connects
//             to the private IP and uses a Google login token as the password.
//   off:      anything else. getPool() throws.
import type { Pool as PgPool, PoolConfig } from "pg";

export type DbMode = "off" | "password" | "iam";

const IAM_VARS = ["DATABASE_HOST", "DATABASE_NAME", "DATABASE_USER"];

const LOGIN_SCOPE = "https://www.googleapis.com/auth/sqlservice.login";

type Env = Record<string, string | undefined>;

export function dbMode(env: Env = process.env): DbMode {
  if (env.DATABASE_URL || env.DATABASE_PASSWORD) return "password";
  if (IAM_VARS.every((name) => env[name])) return "iam";
  return "off";
}

// Turns a libpq sslmode into a pg ssl option. pg 8 treats "require" as full
// certificate checking, but libpq only encrypts. Cloud SQL certificates are
// signed by a per-instance CA that Node does not trust, so match libpq here.
export function sslFromMode(sslmode: string | undefined): PoolConfig["ssl"] {
  switch (sslmode) {
    case "require":
      return { rejectUnauthorized: false };
    case "verify-ca":
    case "verify-full":
      return true;
    default:
      return false;
  }
}

// Reads DATABASE_URL into separate fields. Passing it to pg as a connection
// string would make pg apply its own sslmode rules over ours.
function configFromUrl(value: string, env: Env): PoolConfig {
  const url = new URL(value);
  const decode = (part: string) => decodeURIComponent(part);
  return {
    host: decode(url.hostname) || undefined,
    port: url.port ? Number(url.port) : undefined,
    user: url.username ? decode(url.username) : undefined,
    password: url.password ? decode(url.password) : undefined,
    database: decode(url.pathname.replace(/^\//, "")) || undefined,
    ssl: sslFromMode(url.searchParams.get("sslmode") || env.PGSSLMODE),
  };
}

function passwordConfig(env: Env): PoolConfig {
  if (env.DATABASE_URL) return configFromUrl(env.DATABASE_URL, env);
  return {
    host: env.DATABASE_HOST,
    database: env.DATABASE_NAME,
    user: env.DATABASE_USER,
    password: env.DATABASE_PASSWORD,
    ssl: sslFromMode(env.PGSSLMODE),
  };
}

let poolPromise: Promise<PgPool> | undefined;

// A pooled connection that the server drops while idle emits "error" on the
// pool. Without a listener Node treats that as fatal and the container exits.
function withErrorLog(pool: PgPool): PgPool {
  pool.on("error", (err) => console.error("database pool error:", err.message));
  return pool;
}

async function createPool(): Promise<PgPool> {
  const { default: pg } = await import("pg");
  const mode = dbMode();

  if (mode === "password") {
    return withErrorLog(new pg.Pool(passwordConfig(process.env)));
  }

  if (mode === "iam") {
    // Loaded here so apps without a database never load the Google library.
    const { GoogleAuth } = await import("google-auth-library");
    const auth = new GoogleAuth({ scopes: [LOGIN_SCOPE] });
    // The library caches the token and renews it near expiry. pg calls this for
    // each new connection, so a connection opened after a long idle gap works.
    const password = async (): Promise<string> => {
      const token = await auth.getAccessToken();
      if (!token) throw new Error("Could not get a Google login token for the database.");
      return token;
    };
    return withErrorLog(
      new pg.Pool({
        host: process.env.DATABASE_HOST,
        port: 5432,
        user: process.env.DATABASE_USER,
        database: process.env.DATABASE_NAME,
        password,
        // Encrypts but does not check the certificate, which a per-instance CA signs.
        ssl: { rejectUnauthorized: false },
      }),
    );
  }

  const missing = IAM_VARS.filter((name) => !process.env[name]);
  throw new Error(
    `Database is not configured. Set DATABASE_URL or DATABASE_PASSWORD, or set ${missing.join(", ")}.`,
  );
}

// Returns the shared pool, creating it on first use.
export function getPool(): Promise<PgPool> {
  if (!poolPromise) {
    poolPromise = createPool().catch((err) => {
      poolPromise = undefined; // let the next call try again
      throw err;
    });
  }
  return poolPromise;
}

// Call when the work is done. A one-off script (such as a migration) must call
// it, or the open pool keeps the process alive and the job never exits. A web
// server calls it on shutdown.
export async function closeDb(): Promise<void> {
  const pending = poolPromise;
  poolPromise = undefined;
  if (pending) {
    const pool = await pending.catch(() => undefined);
    await pool?.end();
  }
}
