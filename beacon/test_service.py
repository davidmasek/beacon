from fastapi.testclient import TestClient


def test_list(client: TestClient):
    res = client.get("/services/management")
    assert res.is_success
    data = res.json()
    assert len(data) == 0


def test_create(client: TestClient):
    res = client.post("/services/management", json={"name": "foobar"})
    assert res.is_success

    res = client.get("/services/management")
    assert res.is_success
    data = res.json()
    assert len(data) == 1


def test_delete(client: TestClient):
    res = client.delete("/services/management/foobar")
    assert res.status_code == 404

    res = client.post("/services/management", json={"name": "foobar"})
    assert res.is_success

    res = client.delete("/services/management/foobar")
    assert res.is_success

    res = client.get("/services/management")
    assert res.is_success
    data = res.json()
    assert len(data) == 0
