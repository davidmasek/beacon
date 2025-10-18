import pytest
from fastapi.testclient import TestClient
from sqlalchemy.orm import Session
from sqlalchemy.pool import StaticPool
from beacon.main import app
from beacon.db import create_engine, get_db
from beacon.models import Base


@pytest.fixture(name="session", scope="function")
def session_fixture():
    engine = create_engine(
        "sqlite://",
        connect_args={"check_same_thread": False},
        poolclass=StaticPool,
        echo=False,
    )
    Base.metadata.create_all(engine)
    with Session(engine) as session:
        yield session


@pytest.fixture(name="client")
def client_fixture(session: Session):
    def get_session_override():
        return session

    app.dependency_overrides[get_db] = get_session_override

    client = TestClient(app)
    yield client
    app.dependency_overrides.clear()


def test_beat(client: TestClient, session: Session):
    res = client.post("/services/foo/beat")
    assert res.status_code == 200
    data = res.json()
    assert data["service"] == "foo"
    assert "timestamp" in data


def test_status(client: TestClient):
    res = client.get("/services/foo/status")
    assert res.status_code == 200
    data = res.json()
    assert data["service"] == "foo"
    assert data["timestamp"] == "never"


def test_beat_status(client: TestClient):
    for _ in range(5):
        res = client.post("/services/foo/beat")
        assert res.status_code == 200
    res = client.post("/services/foo/beat")
    assert res.status_code == 200
    ts_written = res.json()["timestamp"]

    res = client.get("/services/foo/status")
    assert res.status_code == 200
    ts_read = res.json()["timestamp"]

    assert ts_written == ts_read
