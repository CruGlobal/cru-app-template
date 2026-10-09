"""Minimal Cru app — a tiny Flask app with a health check, so the container
builds and deploys as-is. Grow it into whatever you need (add routes, a
database, background jobs, etc.)."""
import json
import logging
import os

from cru_iap import dev_bypass, verify_request
from flask import Flask, g, request


class _JsonFormatter(logging.Formatter):
    """One JSON line per event; Cloud Run and Datadog read `severity`."""

    def format(self, record):
        fields = getattr(record, "fields", {})
        return json.dumps({"severity": record.levelname, "message": record.getMessage(), **fields})


_handler = logging.StreamHandler()
_handler.setFormatter(_JsonFormatter())
logging.basicConfig(level=logging.INFO, handlers=[_handler])
log = logging.getLogger(__name__)

app = Flask(__name__)


# IAP sign-in gate: 401 for all but /up. ECS has no IAP in front, so an ECS app
# must remove or replace it.
@app.before_request
def require_iap():
    # /up is the health probe: it calls the container directly, with no IAP assertion.
    if request.path == "/up":
        return None

    result = dev_bypass() or verify_request(request)
    if not result.ok:
        log.warning("iap_rejected", extra={"fields": {"reason": result.reason, "path": request.path}})
        return "Unauthorized", 401
    g.email = result.email
    return None


# Health check — the platform pings this to know the app is alive. Keep a 200
# here working or deploys will be marked unhealthy.
@app.get("/up")
def health():
    return {"status": "ok"}


@app.get("/")
def index():
    return f"Hello, {g.email} 👋", {"content-type": "text/plain; charset=utf-8"}


if __name__ == "__main__":
    # Local dev server; in the container gunicorn serves the app (see Dockerfile).
    # Only this local run reads .env.development. Variables already set win.
    if os.path.exists(".env.development"):
        with open(".env.development") as env_file:
            for line in env_file:
                key, sep, value = line.strip().partition("=")
                if sep and not key.startswith("#"):
                    os.environ.setdefault(key, value)
    app.run(host="0.0.0.0", port=int(os.environ.get("PORT", 8080)))
