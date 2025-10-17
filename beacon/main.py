from fastapi import FastAPI, Depends
from sqlalchemy.orm import Session

from beacon.timestamp import Timestamp
from beacon.db import SessionLocal, create_tables, HealthCheck


def get_db():
    """
    Dependency that yields a new SQLAlchemy SessionLocal session and
    closes it after the response is sent.
    """
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()


app = FastAPI()


@app.on_event("startup")
def on_startup():
    create_tables()


@app.get("/")
def read_root():
    return {"Hello": "World"}


@app.post("/services/{service_id}/beat")
def beat(service_id: str, db: Session = Depends(get_db)):
    now = Timestamp.now()
    db.add(HealthCheck(service_id=service_id, timestamp=now.dt))
    db.commit()
    return {"service": service_id, "timestamp": now.format()}


@app.get("/services/{service_id}/status")
def status(service_id: str, db: Session = Depends(get_db)):
    service = (
        db.query(HealthCheck)
        .filter(HealthCheck.service_id == service_id)
        .order_by(HealthCheck.timestamp.desc())
        .first()
    )
    if not service:
        return {"service": service_id, "timestamp": "never"}

    return {"service": service.service_id, "timestamp": service.timestamp}
