import type { IncomingMessage, ServerResponse } from "node:http";
import { devBypass, verifyRequest } from "@cruglobal/cru-iap";

// One JSON line per event; Cloud Run and Datadog read `severity`.
const warn = (message: string, fields: Record<string, unknown> = {}) =>
  console.log(JSON.stringify({ severity: "WARNING", message, ...fields }));
const logger = { warn: (message: string) => warn(message) };

// IAP sign-in gate: 401 for all but /up. ECS has no IAP in front, so an ECS app
// must remove or replace it.
async function authenticate(req: IncomingMessage, path: string): Promise<string | null | false> {
  // /up is the health probe: it calls the container directly, with no IAP assertion.
  if (path === "/up") return null;

  const result = devBypass({ log: logger }) ?? (await verifyRequest(req, { logger }));
  if (result.ok) return result.email;
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

  // Health check: keep a 200 here or deploys are marked unhealthy.
  if (path === "/health" || path === "/up") {
    res.writeHead(200, { "content-type": "application/json" });
    res.end(JSON.stringify({ status: "ok" }));
    return;
  }

  res.writeHead(200, { "content-type": "text/plain; charset=utf-8" });
  res.end(`Hello, ${email} 👋`);
}
