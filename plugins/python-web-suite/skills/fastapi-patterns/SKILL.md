---
name: fastapi-patterns
description: >
  Production FastAPI patterns — routing, dependency injection, Pydantic models,
  async database access, authentication, background tasks, and testing.
  Triggers: "fastapi", "fastapi routes", "fastapi middleware", "pydantic",
  "fastapi authentication", "fastapi database", "fastapi testing".
  NOT for: Django projects (use django-patterns).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# FastAPI Production Patterns

## Project Structure

```
app/
  __init__.py
  main.py              # App factory
  config.py            # Settings with pydantic-settings
  database.py          # DB session management
  models/
    user.py            # SQLAlchemy models
  schemas/
    user.py            # Pydantic request/response models
  routers/
    users.py           # Route handlers
    auth.py
  services/
    user_service.py    # Business logic
  dependencies/
    auth.py            # Auth dependency
    pagination.py
  middleware/
    logging.py
tests/
  conftest.py
  test_users.py
```

## App Configuration

```python
# app/config.py
from pydantic_settings import BaseSettings
from functools import lru_cache

class Settings(BaseSettings):
    app_name: str = "MyAPI"
    debug: bool = False
    database_url: str
    redis_url: str = "redis://localhost:6379"
    jwt_secret: str
    jwt_algorithm: str = "HS256"
    jwt_expire_minutes: int = 30

    model_config = {"env_file": ".env"}

@lru_cache
def get_settings() -> Settings:
    return Settings()
```

```python
# app/main.py
from fastapi import FastAPI
from contextlib import asynccontextmanager
from app.database import engine, Base
from app.routers import users, auth

@asynccontextmanager
async def lifespan(app: FastAPI):
    # Startup
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)
    yield
    # Shutdown
    await engine.dispose()

app = FastAPI(
    title="My API",
    version="1.0.0",
    lifespan=lifespan,
)

app.include_router(auth.router, prefix="/auth", tags=["auth"])
app.include_router(users.router, prefix="/users", tags=["users"])

@app.get("/health")
async def health():
    return {"status": "ok"}
```

## Pydantic Schemas

```python
# app/schemas/user.py
from pydantic import BaseModel, EmailStr, Field, ConfigDict
from datetime import datetime
from uuid import UUID

class UserCreate(BaseModel):
    name: str = Field(min_length=2, max_length=100)
    email: EmailStr
    password: str = Field(min_length=8, max_length=128)

class UserUpdate(BaseModel):
    name: str | None = Field(None, min_length=2, max_length=100)
    email: EmailStr | None = None

class UserResponse(BaseModel):
    id: UUID
    name: str
    email: str
    created_at: datetime
    model_config = ConfigDict(from_attributes=True)

class UserList(BaseModel):
    data: list[UserResponse]
    total: int
    page: int
    pages: int
```

## Database (SQLAlchemy Async)

```python
# app/database.py
from sqlalchemy.ext.asyncio import create_async_engine, async_sessionmaker, AsyncSession
from sqlalchemy.orm import DeclarativeBase
from app.config import get_settings

settings = get_settings()
engine = create_async_engine(settings.database_url, echo=settings.debug)
AsyncSessionLocal = async_sessionmaker(engine, expire_on_commit=False)

class Base(DeclarativeBase):
    pass

async def get_db() -> AsyncSession:
    async with AsyncSessionLocal() as session:
        try:
            yield session
            await session.commit()
        except Exception:
            await session.rollback()
            raise
```

```python
# app/models/user.py
from sqlalchemy import String, DateTime, func
from sqlalchemy.orm import Mapped, mapped_column
from uuid import UUID, uuid4
from datetime import datetime
from app.database import Base

class User(Base):
    __tablename__ = "users"

    id: Mapped[UUID] = mapped_column(primary_key=True, default=uuid4)
    name: Mapped[str] = mapped_column(String(100))
    email: Mapped[str] = mapped_column(String(255), unique=True, index=True)
    password_hash: Mapped[str] = mapped_column(String(255))
    created_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), server_default=func.now())
    updated_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now())
```

## Routes with Dependencies

```python
# app/routers/users.py
from fastapi import APIRouter, Depends, HTTPException, status, Query
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select, func
from uuid import UUID
from app.database import get_db
from app.models.user import User
from app.schemas.user import UserCreate, UserUpdate, UserResponse, UserList
from app.dependencies.auth import get_current_user, require_admin

router = APIRouter()

@router.get("/", response_model=UserList)
async def list_users(
    page: int = Query(1, ge=1),
    limit: int = Query(20, ge=1, le=100),
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    offset = (page - 1) * limit
    query = select(User).offset(offset).limit(limit).order_by(User.created_at.desc())
    result = await db.execute(query)
    users = result.scalars().all()

    count_result = await db.execute(select(func.count(User.id)))
    total = count_result.scalar_one()

    return UserList(
        data=[UserResponse.model_validate(u) for u in users],
        total=total, page=page, pages=-(-total // limit),
    )

@router.get("/{user_id}", response_model=UserResponse)
async def get_user(user_id: UUID, db: AsyncSession = Depends(get_db)):
    result = await db.execute(select(User).where(User.id == user_id))
    user = result.scalar_one_or_none()
    if not user:
        raise HTTPException(status_code=404, detail="User not found")
    return user

@router.post("/", response_model=UserResponse, status_code=status.HTTP_201_CREATED)
async def create_user(data: UserCreate, db: AsyncSession = Depends(get_db)):
    # Check duplicate
    existing = await db.execute(select(User).where(User.email == data.email))
    if existing.scalar_one_or_none():
        raise HTTPException(status_code=409, detail="Email already registered")

    user = User(name=data.name, email=data.email, password_hash=hash_password(data.password))
    db.add(user)
    await db.flush()
    return user

@router.delete("/{user_id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete_user(
    user_id: UUID,
    db: AsyncSession = Depends(get_db),
    admin: User = Depends(require_admin),
):
    result = await db.execute(select(User).where(User.id == user_id))
    user = result.scalar_one_or_none()
    if not user:
        raise HTTPException(status_code=404, detail="User not found")
    await db.delete(user)
```

## Authentication

```python
# app/dependencies/auth.py
from fastapi import Depends, HTTPException, status
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from jose import jwt, JWTError
from app.config import get_settings
from app.database import get_db

security = HTTPBearer()

async def get_current_user(
    credentials: HTTPAuthorizationCredentials = Depends(security),
    db: AsyncSession = Depends(get_db),
) -> User:
    settings = get_settings()
    try:
        payload = jwt.decode(credentials.credentials, settings.jwt_secret, algorithms=[settings.jwt_algorithm])
        user_id = payload.get("sub")
        if not user_id:
            raise HTTPException(status_code=401, detail="Invalid token")
    except JWTError:
        raise HTTPException(status_code=401, detail="Invalid token")

    result = await db.execute(select(User).where(User.id == user_id))
    user = result.scalar_one_or_none()
    if not user:
        raise HTTPException(status_code=401, detail="User not found")
    return user

def require_admin(current_user: User = Depends(get_current_user)) -> User:
    if current_user.role != "admin":
        raise HTTPException(status_code=403, detail="Admin access required")
    return current_user
```

## Background Tasks

```python
from fastapi import BackgroundTasks

@router.post("/users/", response_model=UserResponse)
async def create_user(data: UserCreate, background_tasks: BackgroundTasks, db: AsyncSession = Depends(get_db)):
    user = User(name=data.name, email=data.email, password_hash=hash_password(data.password))
    db.add(user)
    await db.flush()

    # Runs after response is sent
    background_tasks.add_task(send_welcome_email, user.email, user.name)
    background_tasks.add_task(track_signup, user.id)

    return user

async def send_welcome_email(email: str, name: str):
    # This runs after the 201 response is sent to the client
    await email_service.send(to=email, template="welcome", context={"name": name})
```

## Testing

```python
# tests/conftest.py
import pytest
from httpx import AsyncClient, ASGITransport
from sqlalchemy.ext.asyncio import create_async_engine, async_sessionmaker
from app.main import app
from app.database import Base, get_db

TEST_DATABASE_URL = "sqlite+aiosqlite:///./test.db"
engine = create_async_engine(TEST_DATABASE_URL)
TestSession = async_sessionmaker(engine, expire_on_commit=False)

@pytest.fixture(autouse=True)
async def setup_db():
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)
    yield
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.drop_all)

@pytest.fixture
async def db():
    async with TestSession() as session:
        yield session

@pytest.fixture
async def client(db):
    app.dependency_overrides[get_db] = lambda: db
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as c:
        yield c
    app.dependency_overrides.clear()

# tests/test_users.py
@pytest.mark.anyio
async def test_create_user(client: AsyncClient):
    response = await client.post("/users/", json={
        "name": "Jane Doe", "email": "jane@test.com", "password": "securepass123",
    })
    assert response.status_code == 201
    data = response.json()
    assert data["name"] == "Jane Doe"
    assert data["email"] == "jane@test.com"
    assert "password" not in data
```

## Gotchas

1. **`Depends(get_db)` creates a NEW session per request** — Don't pass the session to background tasks. Background tasks run after the response, and the session is closed by then. Create a new session inside the background task.

2. **Response model filters output automatically** — If `UserResponse` doesn't include `password_hash`, FastAPI strips it. But if you return a dict instead of a model instance, no filtering happens. Always use `response_model` for security.

3. **Pydantic v2 uses `model_validate` not `from_orm`** — `UserResponse.from_orm(user)` is Pydantic v1. In v2, use `UserResponse.model_validate(user)` with `model_config = ConfigDict(from_attributes=True)`.

4. **`async def` vs `def` route handlers** — `async def` runs on the event loop (good for I/O). Plain `def` runs in a thread pool (good for CPU-bound or sync libraries). Using sync DB calls inside `async def` blocks the event loop.

5. **Startup events replaced by lifespan** — `@app.on_event("startup")` is deprecated. Use `lifespan` context manager for setup/teardown. The lifespan pattern properly handles async resource management.

6. **HTTPException is not a Python exception in middleware** — FastAPI's `HTTPException` is only caught by FastAPI's exception handler, not by generic try/except in middleware. Use `RequestValidationError` handler for custom validation error formatting.
