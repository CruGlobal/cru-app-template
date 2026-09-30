"""Postgres connection helper for Cloud SQL. Nothing connects until you call
connect(), so an app with no database is unaffected.

Modes, chosen from environment variables:
  password: DATABASE_URL or DATABASE_PASSWORD is set. Plain pg8000 connection.
  iam:      DATABASE_HOST, DATABASE_NAME, DATABASE_USER and
            INSTANCE_CONNECTION_NAME are set. The Cloud SQL connector logs in as
            the service account over the private IP, no password needed.
  off:      anything else. connect() raises.
"""
import atexit
import os
import ssl
import threading
from urllib.parse import parse_qs, unquote, urlparse

IAM_VARS = ("DATABASE_HOST", "DATABASE_NAME", "DATABASE_USER", "INSTANCE_CONNECTION_NAME")

_connector = None
_lock = threading.Lock()


def db_mode(env=None):
    """Return "password", "iam" or "off" for the given environment."""
    env = os.environ if env is None else env
    if env.get("DATABASE_URL") or env.get("DATABASE_PASSWORD"):
        return "password"
    if all(env.get(name) for name in IAM_VARS):
        return "iam"
    return "off"


def _get_connector():
    # One connector per process: it keeps its own background thread.
    global _connector
    with _lock:
        if _connector is None:
            # Loaded here so apps in password mode never load the connector.
            from google.cloud.sql.connector import Connector, IPTypes

            # Lazy refresh suits Cloud Run, where CPU is throttled between requests.
            _connector = Connector(ip_type=IPTypes.PRIVATE, refresh_strategy="lazy")
            atexit.register(close_db)
        return _connector


def _ssl_context(sslmode):
    if sslmode == "require":
        # Encrypt without checking the server certificate, like libpq.
        context = ssl.create_default_context()
        context.check_hostname = False
        context.verify_mode = ssl.CERT_NONE
        return context
    if sslmode in ("verify-ca", "verify-full"):
        return ssl.create_default_context()
    return None


def _password_args():
    url = os.environ.get("DATABASE_URL")
    if url:
        parts = urlparse(url)
        sslmode = parse_qs(parts.query).get("sslmode", [None])[0] or os.environ.get("PGSSLMODE")
        return {
            "user": unquote(parts.username or ""),
            "password": unquote(parts.password) if parts.password else None,
            "host": parts.hostname or "localhost",
            "port": parts.port or 5432,
            "database": unquote(parts.path.lstrip("/")) or None,
            "ssl_context": _ssl_context(sslmode),
        }
    return {
        "user": os.environ.get("DATABASE_USER", ""),
        "password": os.environ["DATABASE_PASSWORD"],
        "host": os.environ.get("DATABASE_HOST", "localhost"),
        "database": os.environ.get("DATABASE_NAME"),
        "ssl_context": _ssl_context(os.environ.get("PGSSLMODE")),
    }


def connect():
    """Open a new pg8000 (DB-API) connection. The caller closes it."""
    mode = db_mode()
    if mode == "iam":
        # DATABASE_USER is the IAM database user exactly as Terraform sets it.
        # The connector finds the private IP from INSTANCE_CONNECTION_NAME.
        return _get_connector().connect(
            os.environ["INSTANCE_CONNECTION_NAME"],
            "pg8000",
            user=os.environ["DATABASE_USER"],
            db=os.environ["DATABASE_NAME"],
            enable_iam_auth=True,
        )
    if mode == "password":
        import pg8000.dbapi

        return pg8000.dbapi.connect(**_password_args())
    missing = [name for name in IAM_VARS if not os.environ.get(name)]
    raise RuntimeError(
        "Database is not configured. Set DATABASE_URL or DATABASE_PASSWORD, "
        f"or set {', '.join(missing)}."
    )


def close_db():
    """Stop the connector's background work. Safe to call more than once."""
    global _connector
    with _lock:
        if _connector is not None:
            _connector.close()
            _connector = None
