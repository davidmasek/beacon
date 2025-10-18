from fastapi import FastAPI
from beacon.db import create_tables
from beacon import heartbeat


app = FastAPI()

app.include_router(heartbeat.router)


@app.on_event("startup")
def on_startup():
    create_tables()


@app.get("/")
def read_root():
    return "hi"
