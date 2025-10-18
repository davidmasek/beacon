from pathlib import Path
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker

from beacon.models import Base, SchemaVersion

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


# --- Database Initialization Function ---
def create_tables():
    """Creates all defined tables in the database."""
    Base.metadata.create_all(bind=engine)

    # Handle the initial INSERT for schema_version
    db = SessionLocal()
    try:
        count = db.query(SchemaVersion).count()
        if count == 0:
            initial_version = SchemaVersion(version=1)
            db.add(initial_version)
            db.commit()
    except Exception as e:
        db.rollback()
        print(f"Error during initial schema_version insert: {e}")
    finally:
        db.close()


if __name__ == "__main__":
    create_tables()
    print(f"Database tables created at: {SQLALCHEMY_DATABASE_URL}")
