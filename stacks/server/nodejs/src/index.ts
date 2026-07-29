// Minimal Cru app — a plain Node HTTP server with a health check, so the
// container builds and deploys as-is. Grow it into whatever you need (add a
// framework like Express or Fastify, routes, a database, etc.).
import { createServer } from "node:http";

// Listen on $PORT (defaults to 8080). The platform routes traffic to that
// port — an ALB target group on ECS, Cloud Run's own router on Cloud Run.
const port = Number(process.env.PORT) || 8080;

const server = createServer((req, res) => {
  // Health check — the platform pings this to know the app is alive.
  // Keep a 200 here working or deploys will be marked unhealthy.
  if (req.url === "/health" || req.url === "/up") {
    res.writeHead(200, { "content-type": "application/json" });
    res.end(JSON.stringify({ status: "ok" }));
    return;
  }

  res.writeHead(200, { "content-type": "text/plain; charset=utf-8" });
  res.end("Hello from your Cru app 👋");
});

server.listen(port, () => console.log(`listening on :${port}`));
