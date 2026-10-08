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


# The sign-in gate. On Cloud Run, Google IAP signs people in with Okta before a
# request gets here, and the platform sets IAP_AUDIENCE: then every route but
# /up needs IAP's signed assertion, and anything else is a 401. Without
# IAP_AUDIENCE (ECS, or local dev) there is no gate;
# CRU_IAP_DEV_BYPASS_EMAIL=you@cru.org gives you a signed-in email locally.
# Sign out with /?gcp-iap-mode=CLEAR_LOGIN_COOKIE (cru_iap.logout_url()).
@app.before_request
def require_iap():
    g.email = None
    if request.path == "/up":  # the health check; the load balancer lets it skip IAP
        return None

    # Gate on deploy config, never on the header being absent.
    if not os.environ.get("IAP_AUDIENCE"):
        bypass = dev_bypass()
        g.email = bypass.email if bypass else None
        return None

    result = verify_request(request)
    if not result.ok:
        # Fail closed: never fall back to a dev identity here.
        log.warning("iap_rejected", extra={"fields": {"reason": result.reason, "path": request.path}})
        return "Unauthorized", 401
    g.email = result.email
    return None


# Health check — the platform pings this to know the app is alive. Keep a 200
# here working or deploys will be marked unhealthy.
@app.get("/health")
@app.get("/up")
def health():
    return {"status": "ok"}


@app.get("/")
def index():
    greeting = f"Hello, {g.email} 👋" if g.email else "Hello from your Cru app 👋"
    return greeting, {"content-type": "text/plain; charset=utf-8"}


if __name__ == "__main__":
    # Local dev server; in the container gunicorn serves the app (see Dockerfile).
    app.run(host="0.0.0.0", port=int(os.environ.get("PORT", 8080)))
