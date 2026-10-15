import json
import logging
import sys

import pytest

from app.main import _JsonFormatter, app


@pytest.fixture
def client(monkeypatch):
    monkeypatch.delenv("IAP_AUDIENCE", raising=False)
    monkeypatch.delenv("CRU_IAP_DEV_BYPASS_EMAIL", raising=False)
    return app.test_client()


def test_up_is_always_open(client):
    assert client.get("/up").status_code == 200


def test_everything_else_needs_an_assertion_or_the_dev_bypass(client):
    assert client.get("/").status_code == 401
    assert client.get("/nope").status_code == 401


def test_dev_bypass_names_the_user(client, monkeypatch):
    monkeypatch.setenv("CRU_IAP_DEV_BYPASS_EMAIL", "dev@example.com")
    response = client.get("/")
    assert response.status_code == 200
    assert response.text == "Hello, dev@example.com 👋"


def test_iap_audience_ignores_the_dev_bypass(client, monkeypatch):
    monkeypatch.setenv("IAP_AUDIENCE", "/projects/1/global/backendServices/2")
    monkeypatch.setenv("CRU_IAP_DEV_BYPASS_EMAIL", "dev@example.com")
    assert client.get("/").status_code == 401
    assert client.get("/up").status_code == 200


def test_log_lines_keep_the_traceback():
    try:
        raise ValueError("boom")
    except ValueError:
        record = logging.makeLogRecord({"levelname": "ERROR", "msg": "failed", "exc_info": sys.exc_info()})
    message = json.loads(_JsonFormatter().format(record))["message"]
    assert message.startswith("failed\nTraceback")
    assert "ValueError: boom" in message
