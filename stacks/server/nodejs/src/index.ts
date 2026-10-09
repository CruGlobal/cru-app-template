// Minimal Cru app — a plain Node HTTP server with a health check, so the
// container builds and deploys as-is. Grow it into whatever you need (add a
// framework like Express or Fastify, routes, a database, etc.). The routes,
// and the IAP sign-in gate, are in app.ts.
import { createServer } from "node:http";
import { handler } from "./app.ts";

// Listen on $PORT (defaults to 8080). The platform routes traffic to that
// port — an ALB target group on ECS, Cloud Run's own router on Cloud Run.
const port = Number(process.env.PORT) || 8080;

const server = createServer((req, res) => {
  handler(req, res).catch((err) => {
    console.error(err);
    if (!res.headersSent) res.writeHead(500);
    res.end();
  });
});

server.listen(port, () => console.log(`listening on :${port}`));
