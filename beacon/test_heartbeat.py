from fastapi.testclient import TestClient


def test_beat(client: TestClient):
    res = client.post("/services/foo/beat")
    assert res.is_success
    data = res.json()
    assert data["service"] == "foo"
    assert "timestamp" in data


def test_status(client: TestClient):
    res = client.get("/services/foo/status")
    assert res.is_success
    data = res.json()
    assert data["service"] == "foo"
    assert data["timestamp"] == "never"


def test_beat_status(client: TestClient):
    for _ in range(5):
        res = client.post("/services/foo/beat")
        assert res.is_success
    res = client.post("/services/foo/beat")
    assert res.is_success
    ts_written = res.json()["timestamp"]

    res = client.get("/services/foo/status")
    assert res.is_success
    ts_read = res.json()["timestamp"]

    assert ts_written == ts_read
