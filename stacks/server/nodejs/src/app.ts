// The app's routes. index.ts serves them; app.test.ts tests them.
import type { IncomingMessage, ServerResponse } from "node:http";
import { devBypass, verifyRequest } from "@cruglobal/cru-iap";

// One JSON line per event; Cloud Run and Datadog read `severity`.
const warn = (message: string, fields: Record<string, unknown> = {}) =>
  console.log(JSON.stringify({ severity: "WARNING", message, ...fields }));
const logger = { warn: (message: string) => warn(message) };

// The sign-in gate. On Cloud Run, Google IAP signs people in with Okta before a
// request gets here, and the platform sets IAP_AUDIENCE: then every route but
// /up needs IAP's signed assertion, and anything else is a 401. Without
// IAP_AUDIENCE (ECS, or local dev) there is no gate;
// CRU_IAP_DEV_BYPASS_EMAIL=you@cru.org gives you a signed-in email locally.
// Sign out with /?gcp-iap-mode=CLEAR_LOGIN_COOKIE (logoutUrl()).
// Returns the email (null when there is no gate), or false to reject.
async function authenticate(req: IncomingMessage, path: string): Promise<string | null | false> {
  if (path === "/up") return null; // the health check; the load balancer lets it skip IAP

  // Gate on deploy config, never on the header being absent.
  if (!process.env.IAP_AUDIENCE) return devBypass({ log: logger })?.email ?? null;

  const result = await verifyRequest(req, { logger });
  if (result.ok) return result.email;
  // Fail closed: never fall back to a dev identity here.
  warn("iap_rejected", { reason: result.reason, path });
  return false;
}

export async function handler(req: IncomingMessage, res: ServerResponse) {
  const path = new URL(req.url ?? "/", "http://localhost").pathname;

  const email = await authenticate(req, path);
  if (email === false) {
    res.writeHead(401, { "content-type": "text/plain; charset=utf-8" });
    res.end("Unauthorized");
    return;
  }

  // Health check — the platform pings this to know the app is alive.
  // Keep a 200 here working or deploys will be marked unhealthy.
  if (path === "/health" || path === "/up") {
    res.writeHead(200, { "content-type": "application/json" });
    res.end(JSON.stringify({ status: "ok" }));
    return;
  }

  res.writeHead(200, { "content-type": "text/plain; charset=utf-8" });
  res.end(email ? `Hello, ${email} 👋` : "Hello from your Cru app 👋");
}
