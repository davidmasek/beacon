from pathlib import Path
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker

from beacon.models import Base

db_path = Path.home() / "beacon.db"

# --- Database Configuration ---
# Use 'sqlite:///./sql_app.db' for a file-based SQLite database
# Use 'sqlite:///:memory:' for an in-memory database (data is lost when process ends)
SQLALCHEMY_DATABASE_URL = f"sqlite:///{db_path}"

engine = create_engine(
    SQLALCHEMY_DATABASE_URL,
    connect_args={"check_same_thread": False},  # Required for SQLite
)
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)


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


# --- Models ---


def create_tables():
    """Creates all defined tables in the database."""
    Base.metadata.create_all(bind=engine)


if __name__ == "__main__":
    create_tables()
    print(f"Database tables created at: {SQLALCHEMY_DATABASE_URL}")
