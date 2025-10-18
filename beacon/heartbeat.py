from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session
from beacon.db import get_db
from beacon.models import HealthCheck
from beacon.timestamp import Timestamp

router = APIRouter(
    prefix="/services",  # This sets the base path for all routes in this file
    tags=["Services Monitoring"],  # Groups routes in the auto-generated documentation
    responses={404: {"description": "Not found"}},
)


@router.post("/{service_id}/beat")
def beat(service_id: str, db: Session = Depends(get_db)):
    now = Timestamp.now()
    db.add(HealthCheck(service_id=service_id, timestamp=now.dt))
    db.commit()
    return {"service": service_id, "timestamp": now.format()}


@router.get("/{service_id}/status")
def status(service_id: str, db: Session = Depends(get_db)):
    service = (
        db.query(HealthCheck)
        .filter(HealthCheck.service_id == service_id)
        .order_by(HealthCheck.timestamp.desc())
        .first()
    )
    if not service:
        return {"service": service_id, "timestamp": "never"}

    return {
        "service": service.service_id,
        "timestamp": Timestamp(service.timestamp).format(),  # type: ignore
    }
