from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session
from pydantic import BaseModel
from beacon.db import get_db
from beacon.models import Service

router = APIRouter(
    prefix="/services/management",  # This sets the base path for all routes in this file
    tags=["Services Management"],  # Groups routes in the auto-generated documentation
    responses={404: {"description": "Not found"}},
)


class ServiceIn(BaseModel):
    name: str
    url: str | None = None


class ServiceOut(BaseModel):
    name: str
    url: str | None = None


@router.post("/")
def create(create_service: ServiceIn, db: Session = Depends(get_db)) -> ServiceOut:
    service = Service(service_name=create_service.name, url=create_service.url)
    db.add(service)
    db.commit()
    return ServiceOut(name=service.service_name, url=service.url)  # type: ignore


@router.delete("/{service_name}")
def delete(service_name: str, db: Session = Depends(get_db)) -> None:
    deleted_count = (
        db.query(Service).filter(Service.service_name == service_name).delete()
    )
    if deleted_count == 0:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Service with name '{service_name}' not found",
        )
    db.commit()
    return None


@router.get("/")
def list_services(db: Session = Depends(get_db)) -> list[ServiceOut]:
    services = db.query(Service).all()

    return [ServiceOut(name=s.service_name, url=s.url) for s in services]  # type: ignore
