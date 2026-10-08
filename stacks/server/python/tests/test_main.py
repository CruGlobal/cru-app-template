import pytest

from app.main import app


@pytest.fixture
def client(monkeypatch):
    monkeypatch.delenv("IAP_AUDIENCE", raising=False)
    monkeypatch.delenv("CRU_IAP_DEV_BYPASS_EMAIL", raising=False)
    return app.test_client()


def test_iap_audience_gates_everything_but_up(client, monkeypatch):
    monkeypatch.setenv("IAP_AUDIENCE", "/projects/1/global/backendServices/2")
    monkeypatch.setenv("CRU_IAP_DEV_BYPASS_EMAIL", "dev@example.com")  # must not open the gate

    assert client.get("/up").status_code == 200
    assert client.get("/").status_code == 401
    assert client.get("/health").status_code == 401


def test_no_gate_without_iap_audience(client):
    response = client.get("/")
    assert response.status_code == 200
    assert response.text == "Hello from your Cru app 👋"


def test_dev_bypass_names_the_user(client, monkeypatch):
    monkeypatch.setenv("CRU_IAP_DEV_BYPASS_EMAIL", "dev@example.com")
    assert client.get("/").text == "Hello, dev@example.com 👋"
