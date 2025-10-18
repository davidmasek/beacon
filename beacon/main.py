from fastapi import FastAPI
from beacon.db import create_tables
from beacon import heartbeat
from beacon import service


app = FastAPI()

app.include_router(heartbeat.router)
app.include_router(service.router)


@app.on_event("startup")
def on_startup():
    create_tables()


@app.get("/")
def read_root():
    return "hi"
