from sqlalchemy import (
    create_engine,
    Column,
    Integer,
    String,
    DateTime,
    ForeignKey,
    Index,
    Text,
)
from sqlalchemy.orm import sessionmaker, relationship, declarative_base
from sqlalchemy.sql import func

# --- Database Configuration ---
# Use 'sqlite:///./sql_app.db' for a file-based SQLite database
# Use 'sqlite:///:memory:' for an in-memory database (data is lost when process ends)
SQLALCHEMY_DATABASE_URL = "sqlite:///./app_database.db"

engine = create_engine(
    SQLALCHEMY_DATABASE_URL,
    connect_args={"check_same_thread": False},  # Required for SQLite
)
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)

Base = declarative_base()

# --- Models ---


class User(Base):
    __tablename__ = "users"
    __table_args__ = (Index("idx_users_email", "email", unique=True),)

    id = Column(Integer, primary_key=True, index=True)
    email = Column(String, unique=True, nullable=False)
    password_hash = Column(String, nullable=False)
    created_at = Column(DateTime, default=func.now())

    # Relationships (Optional but helpful)
    health_checks = relationship("HealthCheck", back_populates="owner")


class HealthCheck(Base):
    __tablename__ = "health_checks"
    __table_args__ = (Index("idx_health_checks_timestamp", "timestamp"),)

    id = Column(Integer, primary_key=True, index=True)
    user_id = Column(Integer, ForeignKey("users.id", ondelete="CASCADE"))
    service_id = Column(String, nullable=False)
    timestamp = Column(DateTime, default=func.now())
    meta_data = Column("metadata", Text)

    # Relationships
    owner = relationship("User", back_populates="health_checks")


class TaskLog(Base):
    __tablename__ = "task_logs"
    __table_args__ = (Index("idx_task_logs_timestamp", "timestamp"),)

    id = Column(Integer, primary_key=True, index=True)
    user_id = Column(Integer)  # No FK constraint based on your SQL
    task_name = Column(String, nullable=False)
    timestamp = Column(DateTime, default=func.now())
    status = Column(String, nullable=False)
    details = Column(Text)


class ServiceState(Base):
    __tablename__ = "service_state"
    __table_args__ = (Index("idx_service_state_service_id", "service_id", unique=True),)

    id = Column(Integer, primary_key=True, index=True)
    user_id = Column(Integer)  # No FK constraint based on your SQL
    service_id = Column(String, unique=True, nullable=False)
    status = Column(String, nullable=False)
    last_reported_status = Column(String)
    updated_at = Column(DateTime, default=func.now(), onupdate=func.now())


class SchemaVersion(Base):
    __tablename__ = "schema_version"

    # NOTE: Since your original SQL doesn't define a PK, a composite or a
    # single non-nullable column must be chosen for SQLAlchemy.
    # I'll use `version` as the primary key for simplicity, though this might need adjustment.
    version = Column(Integer, primary_key=True, nullable=False)
    applied_at = Column(DateTime, default=func.now())


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
