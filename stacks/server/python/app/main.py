"""Minimal Cru app — a tiny Flask app with a health check, so the container
builds and deploys as-is. Grow it into whatever you need (add routes, a
database, background jobs, etc.)."""
import os

from flask import Flask

app = Flask(__name__)


# Health check — the platform pings this to know the app is alive. Keep a 200
# here working or deploys will be marked unhealthy.
@app.get("/health")
@app.get("/up")
def health():
    return {"status": "ok"}


@app.get("/")
def index():
    return "Hello from your Cru app 👋"


if __name__ == "__main__":
    # Local dev server; in the container gunicorn serves the app (see Dockerfile).
    app.run(host="0.0.0.0", port=int(os.environ.get("PORT", 8080)))
