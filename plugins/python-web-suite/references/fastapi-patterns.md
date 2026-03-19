# FastAPI Patterns Reference

Quick-reference guide for FastAPI patterns including async operations, dependency injection, Pydantic v2 models, background tasks, and application lifespan management. Consult this when building or reviewing FastAPI applications.

---

## Table of Contents

1. [Async Patterns](#1-async-patterns)
2. [Dependency Injection Patterns](#2-dependency-injection-patterns)
3. [Pydantic v2 Model Patterns](#3-pydantic-v2-model-patterns)
4. [Background Tasks & Lifespan](#4-background-tasks--lifespan)
5. [Router Organization](#5-router-organization)
6. [Error Handling](#6-error-handling)
7. [Middleware Patterns](#7-middleware-patterns)
8. [WebSocket Patterns](#8-websocket-patterns)

---

## 1. Async Patterns

### async def vs def endpoints

Use `async def` when the handler awaits anything: DB queries, HTTP calls, file I/O.
Use plain `def` only when the work is truly CPU-bound — FastAPI runs sync endpoints in a thread pool automatically, so blocking there is safe.

```python
from fastapi import FastAPI
import httpx
import asyncio

app = FastAPI()

# GOOD: async for I/O
@app.get("/users/{user_id}")
async def get_user(user_id: int) -> dict:
    async with httpx.AsyncClient() as client:
        response = await client.get(f"https://api.example.com/users/{user_id}")
        return response.json()

# GOOD: sync for CPU-bound (runs in thread pool)
@app.post("/compute")
def heavy_computation(data: list[int]) -> dict:
    result = sum(x ** 2 for x in data)  # CPU-bound, not I/O
    return {"result": result}

# BAD: sync blocking call inside async endpoint — stalls the event loop
@app.get("/bad-example")
async def bad_endpoint() -> dict:
    import time
    time.sleep(2)  # NEVER do this in async def
    return {"status": "done"}
```

### asyncio.gather for parallel operations

Use `asyncio.gather` to fan out multiple independent async calls and collect results simultaneously.

```python
from fastapi import FastAPI, HTTPException
import asyncio
import httpx

app = FastAPI()

async def fetch_user(client: httpx.AsyncClient, user_id: int) -> dict:
    r = await client.get(f"https://api.example.com/users/{user_id}")
    r.raise_for_status()
    return r.json()

async def fetch_posts(client: httpx.AsyncClient, user_id: int) -> list:
    r = await client.get(f"https://api.example.com/users/{user_id}/posts")
    r.raise_for_status()
    return r.json()

async def fetch_follows(client: httpx.AsyncClient, user_id: int) -> list:
    r = await client.get(f"https://api.example.com/users/{user_id}/followers")
    r.raise_for_status()
    return r.json()

@app.get("/users/{user_id}/full-profile")
async def get_full_profile(user_id: int) -> dict:
    async with httpx.AsyncClient(timeout=10.0) as client:
        user, posts, followers = await asyncio.gather(
            fetch_user(client, user_id),
            fetch_posts(client, user_id),
            fetch_follows(client, user_id),
        )
    return {"user": user, "posts": posts, "followers": followers}

# gather with return_exceptions=True for partial failures
@app.get("/users/{user_id}/tolerant-profile")
async def get_tolerant_profile(user_id: int) -> dict:
    async with httpx.AsyncClient(timeout=10.0) as client:
        results = await asyncio.gather(
            fetch_user(client, user_id),
            fetch_posts(client, user_id),
            fetch_follows(client, user_id),
            return_exceptions=True,
        )
    user, posts, followers = results
    return {
        "user": user if not isinstance(user, Exception) else None,
        "posts": posts if not isinstance(posts, Exception) else [],
        "followers": followers if not isinstance(followers, Exception) else [],
    }
```

### Async database queries with SQLAlchemy 2.0 AsyncSession

```python
from collections.abc import AsyncGenerator
from fastapi import FastAPI, Depends
from sqlalchemy.ext.asyncio import (
    AsyncSession,
    async_sessionmaker,
    create_async_engine,
)
from sqlalchemy.orm import DeclarativeBase, Mapped, mapped_column
from sqlalchemy import select, String, Integer

DATABASE_URL = "postgresql+asyncpg://user:pass@localhost/dbname"

engine = create_async_engine(DATABASE_URL, pool_size=20, max_overflow=10)
AsyncSessionLocal = async_sessionmaker(engine, expire_on_commit=False)

class Base(DeclarativeBase):
    pass

class User(Base):
    __tablename__ = "users"
    id: Mapped[int] = mapped_column(Integer, primary_key=True)
    name: Mapped[str] = mapped_column(String(100))
    email: Mapped[str] = mapped_column(String(200), unique=True)

async def get_db() -> AsyncGenerator[AsyncSession, None]:
    async with AsyncSessionLocal() as session:
        yield session

app = FastAPI()

@app.get("/users/{user_id}")
async def get_user(user_id: int, db: AsyncSession = Depends(get_db)) -> dict:
    result = await db.execute(select(User).where(User.id == user_id))
    user = result.scalar_one_or_none()
    if user is None:
        raise HTTPException(status_code=404, detail="User not found")
    return {"id": user.id, "name": user.name, "email": user.email}

@app.get("/users")
async def list_users(db: AsyncSession = Depends(get_db)) -> list[dict]:
    result = await db.execute(select(User).order_by(User.name))
    users = result.scalars().all()
    return [{"id": u.id, "name": u.name, "email": u.email} for u in users]
```

### Async HTTP calls with httpx.AsyncClient

Prefer a single shared client (via lifespan) over creating one per request.

```python
from contextlib import asynccontextmanager
from fastapi import FastAPI, Request
import httpx

@asynccontextmanager
async def lifespan(app: FastAPI):
    # Startup: create a shared client
    app.state.http_client = httpx.AsyncClient(
        timeout=httpx.Timeout(10.0, connect=5.0),
        limits=httpx.Limits(max_connections=100, max_keepalive_connections=20),
    )
    yield
    # Shutdown: close the client
    await app.state.http_client.aclose()

app = FastAPI(lifespan=lifespan)

@app.get("/proxy/{path:path}")
async def proxy(path: str, request: Request) -> dict:
    client: httpx.AsyncClient = request.app.state.http_client
    response = await client.get(f"https://upstream.example.com/{path}")
    return response.json()
```

### Semaphores for concurrency limiting

Prevent overwhelming downstream services by capping concurrent requests.

```python
import asyncio
import httpx
from fastapi import FastAPI

app = FastAPI()
_semaphore = asyncio.Semaphore(10)  # max 10 concurrent outbound calls

async def limited_fetch(client: httpx.AsyncClient, url: str) -> dict:
    async with _semaphore:
        response = await client.get(url)
        return response.json()

@app.get("/batch-fetch")
async def batch_fetch(urls: list[str]) -> list[dict]:
    async with httpx.AsyncClient() as client:
        tasks = [limited_fetch(client, url) for url in urls]
        return await asyncio.gather(*tasks)
```

### Async generators and streaming responses

```python
import asyncio
from fastapi import FastAPI
from fastapi.responses import StreamingResponse

app = FastAPI()

async def event_stream(topic: str):
    """Server-Sent Events generator."""
    for i in range(100):
        yield f"data: {topic} event {i}\n\n"
        await asyncio.sleep(1)

@app.get("/stream/{topic}")
async def stream_events(topic: str) -> StreamingResponse:
    return StreamingResponse(
        event_stream(topic),
        media_type="text/event-stream",
        headers={"Cache-Control": "no-cache", "X-Accel-Buffering": "no"},
    )

# Chunked file streaming
async def file_chunks(file_path: str, chunk_size: int = 65536):
    import aiofiles
    async with aiofiles.open(file_path, "rb") as f:
        while chunk := await f.read(chunk_size):
            yield chunk

@app.get("/files/{filename}")
async def download_file(filename: str) -> StreamingResponse:
    return StreamingResponse(
        file_chunks(f"/data/{filename}"),
        media_type="application/octet-stream",
    )
```

### Task groups (Python 3.11+ TaskGroup)

`TaskGroup` cancels all sibling tasks on first failure — better error semantics than `gather`.

```python
import asyncio
import httpx
from fastapi import FastAPI

app = FastAPI()

@app.get("/users/{user_id}/dashboard")
async def get_dashboard(user_id: int) -> dict:
    results: dict = {}

    async def fetch_orders(client: httpx.AsyncClient) -> None:
        r = await client.get(f"https://api.example.com/users/{user_id}/orders")
        results["orders"] = r.json()

    async def fetch_invoices(client: httpx.AsyncClient) -> None:
        r = await client.get(f"https://api.example.com/users/{user_id}/invoices")
        results["invoices"] = r.json()

    async with httpx.AsyncClient() as client:
        async with asyncio.TaskGroup() as tg:
            tg.create_task(fetch_orders(client))
            tg.create_task(fetch_invoices(client))

    return results
```

### Async context managers

```python
from contextlib import asynccontextmanager
from fastapi import FastAPI
import asyncio

@asynccontextmanager
async def managed_lock(name: str):
    """Distributed lock pattern placeholder."""
    print(f"Acquiring lock: {name}")
    try:
        yield
    finally:
        print(f"Releasing lock: {name}")

@app.post("/critical-section")
async def critical_section() -> dict:
    async with managed_lock("my-resource"):
        await asyncio.sleep(0.1)  # protected work
        return {"status": "done"}
```

### Common async pitfalls

```python
# PITFALL 1: Blocking call inside async endpoint
import time
import requests  # sync library

# BAD
async def bad_sync_http() -> dict:
    r = requests.get("https://example.com")  # blocks event loop!
    return r.json()

# GOOD
async def good_async_http() -> dict:
    async with httpx.AsyncClient() as client:
        r = await client.get("https://example.com")
    return r.json()

# PITFALL 2: Using sync SQLAlchemy Session in async endpoint
# BAD: from sqlalchemy.orm import Session — will block
# GOOD: from sqlalchemy.ext.asyncio import AsyncSession

# PITFALL 3: Forgetting to await
async def missing_await(db: AsyncSession) -> None:
    db.execute(select(User))  # BAD: coroutine never awaited
    await db.execute(select(User))  # GOOD

# PITFALL 4: Creating a new event loop inside an async context
# BAD: asyncio.run(some_coroutine())  inside an async def
# GOOD: await some_coroutine()
```

---

## 2. Dependency Injection Patterns

### Simple function dependencies

```python
from fastapi import FastAPI, Depends, Query, Header, HTTPException
from typing import Annotated

app = FastAPI()

def common_pagination(
    page: int = Query(1, ge=1),
    page_size: int = Query(20, ge=1, le=100),
) -> dict:
    return {"offset": (page - 1) * page_size, "limit": page_size}

Pagination = Annotated[dict, Depends(common_pagination)]

@app.get("/items")
async def list_items(pagination: Pagination) -> dict:
    return {"offset": pagination["offset"], "limit": pagination["limit"]}

@app.get("/products")
async def list_products(pagination: Pagination) -> dict:
    return {"offset": pagination["offset"], "limit": pagination["limit"]}
```

### Class-based dependencies with `__call__`

```python
from fastapi import FastAPI, Depends, Request
from typing import Annotated

class RateLimiter:
    def __init__(self, max_calls: int, window_seconds: int) -> None:
        self.max_calls = max_calls
        self.window_seconds = window_seconds
        self._store: dict[str, list[float]] = {}

    async def __call__(self, request: Request) -> None:
        import time
        client_ip = request.client.host if request.client else "unknown"
        now = time.time()
        window_start = now - self.window_seconds
        calls = [t for t in self._store.get(client_ip, []) if t > window_start]
        if len(calls) >= self.max_calls:
            raise HTTPException(status_code=429, detail="Rate limit exceeded")
        calls.append(now)
        self._store[client_ip] = calls

strict_limiter = RateLimiter(max_calls=10, window_seconds=60)
loose_limiter = RateLimiter(max_calls=100, window_seconds=60)

app = FastAPI()

@app.get("/api/sensitive", dependencies=[Depends(strict_limiter)])
async def sensitive_endpoint() -> dict:
    return {"data": "sensitive"}

@app.get("/api/public", dependencies=[Depends(loose_limiter)])
async def public_endpoint() -> dict:
    return {"data": "public"}
```

### Parameterized dependencies (factories)

```python
from fastapi import FastAPI, Depends, HTTPException
from typing import Annotated

def require_role(*allowed_roles: str):
    """Dependency factory that returns a dependency checking for specific roles."""
    async def role_checker(
        current_user: Annotated[dict, Depends(get_current_user)],
    ) -> dict:
        if current_user.get("role") not in allowed_roles:
            raise HTTPException(status_code=403, detail="Insufficient permissions")
        return current_user
    return role_checker

async def get_current_user() -> dict:
    # Placeholder: decode JWT, look up user, etc.
    return {"id": 1, "role": "admin"}

app = FastAPI()

@app.delete("/admin/users/{user_id}")
async def delete_user(
    user_id: int,
    user: Annotated[dict, Depends(require_role("admin", "superadmin"))],
) -> dict:
    return {"deleted": user_id, "by": user["id"]}

@app.get("/reports")
async def get_reports(
    user: Annotated[dict, Depends(require_role("admin", "analyst", "viewer"))],
) -> dict:
    return {"reports": [], "viewer": user["id"]}
```

### Yield dependencies for DB sessions and transactions

```python
from collections.abc import AsyncGenerator
from fastapi import FastAPI, Depends
from sqlalchemy.ext.asyncio import AsyncSession, async_sessionmaker, create_async_engine

engine = create_async_engine("postgresql+asyncpg://user:pass@localhost/db")
SessionLocal = async_sessionmaker(engine, expire_on_commit=False)

async def get_db() -> AsyncGenerator[AsyncSession, None]:
    """Plain session — auto-rollback on exception, commit must be explicit."""
    async with SessionLocal() as session:
        try:
            yield session
        except Exception:
            await session.rollback()
            raise

async def get_db_transaction() -> AsyncGenerator[AsyncSession, None]:
    """Session with an explicit transaction — auto-commit on success."""
    async with SessionLocal() as session:
        async with session.begin():
            yield session
            # session.begin() auto-commits here if no exception was raised

app = FastAPI()

@app.post("/users")
async def create_user(
    name: str,
    db: AsyncSession = Depends(get_db_transaction),
) -> dict:
    user = User(name=name, email=f"{name}@example.com")
    db.add(user)
    # No explicit commit needed — get_db_transaction handles it
    return {"name": name}
```

### Dependency chains and sub-dependencies

```python
from fastapi import FastAPI, Depends, Header, HTTPException
from typing import Annotated

async def get_token(authorization: str = Header(...)) -> str:
    if not authorization.startswith("Bearer "):
        raise HTTPException(status_code=401, detail="Invalid auth scheme")
    return authorization.removeprefix("Bearer ")

async def get_current_user(token: str = Depends(get_token)) -> dict:
    # Validate token, look up user in DB
    if token == "invalid":
        raise HTTPException(status_code=401, detail="Invalid token")
    return {"id": 1, "name": "Alice", "role": "admin"}

async def get_active_user(
    user: dict = Depends(get_current_user),
) -> dict:
    if not user.get("active", True):
        raise HTTPException(status_code=403, detail="Account disabled")
    return user

app = FastAPI()

@app.get("/me")
async def get_me(user: Annotated[dict, Depends(get_active_user)]) -> dict:
    return user
```

### Global dependencies on app and router

```python
from fastapi import FastAPI, APIRouter, Depends, Request

async def verify_api_key(request: Request) -> None:
    key = request.headers.get("X-API-Key")
    if key != "secret-key":
        raise HTTPException(status_code=403, detail="Invalid API key")

# Apply to every route in the app
app = FastAPI(dependencies=[Depends(verify_api_key)])

# Apply to every route in a router
router = APIRouter(
    prefix="/internal",
    dependencies=[Depends(verify_api_key)],
)

@router.get("/status")
async def internal_status() -> dict:
    return {"ok": True}

app.include_router(router)
```

### Override dependencies for testing

```python
from fastapi import FastAPI, Depends
from fastapi.testclient import TestClient

async def get_current_user() -> dict:
    # Real implementation talks to DB
    ...

app = FastAPI()

@app.get("/me")
async def me(user: dict = Depends(get_current_user)) -> dict:
    return user

# In tests:
def override_get_current_user() -> dict:
    return {"id": 99, "name": "Test User", "role": "admin"}

app.dependency_overrides[get_current_user] = override_get_current_user

client = TestClient(app)

def test_me() -> None:
    response = client.get("/me")
    assert response.status_code == 200
    assert response.json()["name"] == "Test User"

# Clean up after tests
app.dependency_overrides.clear()
```

### Caching dependencies (once per request)

FastAPI caches dependency results within a single request by default when using `Depends`. Use `use_cache=False` to disable.

```python
from fastapi import FastAPI, Depends
from typing import Annotated
import uuid

call_count = 0

async def expensive_setup() -> str:
    global call_count
    call_count += 1
    return f"resource-{uuid.uuid4()}"

# Both endpoints receive the SAME resource instance per request
async def endpoint_a(resource: str = Depends(expensive_setup)) -> str:
    return resource

async def endpoint_b(resource: str = Depends(expensive_setup)) -> str:
    return resource

# Force a fresh call even within the same request
async def always_fresh(
    resource: str = Depends(expensive_setup, use_cache=False),
) -> str:
    return resource
```

### Security dependencies (OAuth2, API key, HTTP Bearer)

```python
from fastapi import FastAPI, Depends, HTTPException, Security
from fastapi.security import (
    OAuth2PasswordBearer,
    OAuth2PasswordRequestForm,
    APIKeyHeader,
    HTTPBearer,
    HTTPAuthorizationCredentials,
)
from typing import Annotated

app = FastAPI()

# OAuth2 Password Bearer
oauth2_scheme = OAuth2PasswordBearer(tokenUrl="/token")

async def get_current_user_oauth(
    token: Annotated[str, Depends(oauth2_scheme)],
) -> dict:
    # decode JWT here
    return {"id": 1, "token": token}

# API Key via header
api_key_header = APIKeyHeader(name="X-API-Key")

async def verify_api_key(
    api_key: Annotated[str, Security(api_key_header)],
) -> str:
    valid_keys = {"key-abc123", "key-xyz789"}
    if api_key not in valid_keys:
        raise HTTPException(status_code=403, detail="Invalid API key")
    return api_key

# HTTP Bearer (manual JWT validation)
bearer_scheme = HTTPBearer()

async def get_user_from_bearer(
    credentials: Annotated[HTTPAuthorizationCredentials, Security(bearer_scheme)],
) -> dict:
    token = credentials.credentials
    # validate JWT token
    return {"id": 1}

@app.post("/token")
async def login(form_data: OAuth2PasswordRequestForm = Depends()) -> dict:
    if form_data.username == "admin" and form_data.password == "secret":
        return {"access_token": "fake-jwt-token", "token_type": "bearer"}
    raise HTTPException(status_code=400, detail="Incorrect credentials")

@app.get("/protected")
async def protected(user: dict = Depends(get_current_user_oauth)) -> dict:
    return user
```

---

## 3. Pydantic v2 Model Patterns

### Field definitions with Field()

```python
from pydantic import BaseModel, Field, EmailStr
from typing import Annotated
from uuid import UUID, uuid4
from datetime import datetime

class UserCreate(BaseModel):
    id: UUID = Field(default_factory=uuid4)
    name: str = Field(min_length=1, max_length=100, examples=["Alice"])
    email: EmailStr
    age: int = Field(ge=0, le=150, description="Age in years")
    score: float = Field(default=0.0, ge=0.0, le=100.0)
    tags: list[str] = Field(default_factory=list, max_length=10)
    bio: str | None = Field(default=None, max_length=500)
    created_at: datetime = Field(default_factory=datetime.utcnow)

    # Alias for camelCase JSON input
    first_name: str = Field(alias="firstName", min_length=1)
```

### @field_validator with mode='before' and mode='after'

```python
from pydantic import BaseModel, field_validator, Field
from typing import Any

class ProductCreate(BaseModel):
    name: str
    price: float = Field(gt=0)
    sku: str
    tags: list[str] = Field(default_factory=list)

    @field_validator("name", mode="before")
    @classmethod
    def strip_name(cls, v: Any) -> str:
        """Run BEFORE type coercion — v may not be a str yet."""
        if isinstance(v, str):
            return v.strip()
        return v  # let Pydantic raise the type error

    @field_validator("sku", mode="after")
    @classmethod
    def normalize_sku(cls, v: str) -> str:
        """Run AFTER type coercion — v is guaranteed to be a str."""
        return v.upper().replace(" ", "-")

    @field_validator("tags", mode="before")
    @classmethod
    def split_tags(cls, v: Any) -> list[str]:
        """Accept comma-separated string or list."""
        if isinstance(v, str):
            return [t.strip() for t in v.split(",") if t.strip()]
        return v

    @field_validator("price", mode="after")
    @classmethod
    def round_price(cls, v: float) -> float:
        return round(v, 2)
```

### @model_validator(mode='wrap') and mode='before'/'after'

```python
from pydantic import BaseModel, model_validator
from typing import Any, Self

class DateRange(BaseModel):
    start_date: str
    end_date: str

    @model_validator(mode="after")
    def check_dates(self) -> Self:
        if self.end_date < self.start_date:
            raise ValueError("end_date must be >= start_date")
        return self

class FlexibleInput(BaseModel):
    x: int
    y: int
    label: str = ""

    @model_validator(mode="before")
    @classmethod
    def handle_legacy_format(cls, data: Any) -> Any:
        """Accept both {'x': 1, 'y': 2} and [1, 2] input formats."""
        if isinstance(data, (list, tuple)) and len(data) >= 2:
            return {"x": data[0], "y": data[1]}
        return data

    @model_validator(mode="wrap")
    @classmethod
    def wrap_validator(cls, value: Any, handler) -> "FlexibleInput":
        """Wrap mode: called with the raw value and the inner validator."""
        if isinstance(value, dict) and "coords" in value:
            coords = value.pop("coords")
            value["x"], value["y"] = coords
        return handler(value)
```

### Computed fields with @computed_field

```python
from pydantic import BaseModel, computed_field, Field
from typing import Annotated

class Rectangle(BaseModel):
    width: float = Field(gt=0)
    height: float = Field(gt=0)

    @computed_field
    @property
    def area(self) -> float:
        return self.width * self.height

    @computed_field
    @property
    def perimeter(self) -> float:
        return 2 * (self.width + self.height)

    @computed_field
    @property
    def is_square(self) -> bool:
        return abs(self.width - self.height) < 1e-9

# Computed fields appear in model_dump() and model_dump_json()
r = Rectangle(width=3.0, height=4.0)
print(r.model_dump())
# {'width': 3.0, 'height': 4.0, 'area': 12.0, 'perimeter': 14.0, 'is_square': False}
```

### ConfigDict: strict mode, json_schema_extra, from_attributes

```python
from pydantic import BaseModel, ConfigDict, Field

class StrictUser(BaseModel):
    model_config = ConfigDict(
        strict=True,            # No coercion: "1" won't become 1
        populate_by_name=True,  # Allow both alias and field name
        str_strip_whitespace=True,
        validate_default=True,
        json_schema_extra={
            "examples": [{"id": 1, "name": "Alice", "email": "alice@example.com"}]
        },
    )
    id: int
    name: str
    email: str

class OrmUser(BaseModel):
    """Maps SQLAlchemy ORM objects to Pydantic models."""
    model_config = ConfigDict(from_attributes=True)  # replaces orm_mode=True in v1

    id: int
    name: str
    email: str

# Usage with SQLAlchemy ORM object
# orm_obj = db_user  # SQLAlchemy User row
# pydantic_user = OrmUser.model_validate(orm_obj)
```

### Generic models

```python
from pydantic import BaseModel
from typing import Generic, TypeVar

T = TypeVar("T")

class Page(BaseModel, Generic[T]):
    items: list[T]
    total: int
    page: int
    page_size: int
    has_next: bool
    has_prev: bool

    @classmethod
    def create(
        cls,
        items: list[T],
        total: int,
        page: int,
        page_size: int,
    ) -> "Page[T]":
        return cls(
            items=items,
            total=total,
            page=page,
            page_size=page_size,
            has_next=(page * page_size) < total,
            has_prev=page > 1,
        )

class UserRead(BaseModel):
    id: int
    name: str

# FastAPI knows the exact schema for Page[UserRead]
@app.get("/users", response_model=Page[UserRead])
async def list_users_paged() -> Page[UserRead]:
    users = [UserRead(id=1, name="Alice"), UserRead(id=2, name="Bob")]
    return Page.create(items=users, total=2, page=1, page_size=20)
```

### Discriminated unions with Literal

```python
from pydantic import BaseModel
from typing import Literal, Annotated, Union
from pydantic import Field

class EmailNotification(BaseModel):
    type: Literal["email"]
    recipient: str
    subject: str
    body: str

class SMSNotification(BaseModel):
    type: Literal["sms"]
    phone_number: str
    message: str

class PushNotification(BaseModel):
    type: Literal["push"]
    device_token: str
    title: str
    payload: dict

Notification = Annotated[
    Union[EmailNotification, SMSNotification, PushNotification],
    Field(discriminator="type"),
]

class NotificationRequest(BaseModel):
    notification: Notification
    priority: int = 1

# Pydantic picks the right model based on "type" field — no ambiguity
req = NotificationRequest.model_validate({
    "notification": {"type": "sms", "phone_number": "+1555000", "message": "Hi"},
    "priority": 2,
})
assert isinstance(req.notification, SMSNotification)
```

### Custom types with Annotated and AfterValidator

```python
from pydantic import BaseModel, AfterValidator, BeforeValidator
from typing import Annotated
import re

def validate_phone(v: str) -> str:
    digits = re.sub(r"\D", "", v)
    if len(digits) not in (10, 11):
        raise ValueError("Phone number must have 10 or 11 digits")
    return f"+1{digits[-10:]}"

def to_lowercase_email(v: str) -> str:
    return v.strip().lower()

# Reusable annotated types
PhoneNumber = Annotated[str, AfterValidator(validate_phone)]
NormalizedEmail = Annotated[str, AfterValidator(to_lowercase_email)]
NonEmptyStr = Annotated[str, AfterValidator(lambda v: v if v.strip() else (_ for _ in ()).throw(ValueError("must not be blank")))]

class ContactForm(BaseModel):
    name: str
    email: NormalizedEmail
    phone: PhoneNumber

c = ContactForm(name="Alice", email="  ALICE@Example.COM  ", phone="(555) 867-5309")
print(c.email)   # alice@example.com
print(c.phone)   # +15558675309
```

### model_dump() and model_dump_json() options

```python
from pydantic import BaseModel, Field
from datetime import datetime
from uuid import UUID, uuid4

class ArticleRead(BaseModel):
    id: UUID = Field(default_factory=uuid4)
    title: str
    body: str
    author_id: int
    internal_notes: str = ""
    created_at: datetime = Field(default_factory=datetime.utcnow)

article = ArticleRead(title="Hello", body="World", author_id=1, internal_notes="secret")

# Exclude specific fields
print(article.model_dump(exclude={"internal_notes", "created_at"}))

# Exclude unset fields (great for PATCH endpoints)
partial = ArticleRead.model_validate({"title": "New Title", "body": "...", "author_id": 1})
print(partial.model_dump(exclude_unset=True))

# Exclude None values
print(article.model_dump(exclude_none=True))

# Include only specific fields
print(article.model_dump(include={"id", "title"}))

# Dump with aliases (camelCase)
class CamelArticle(BaseModel):
    model_config = ConfigDict(populate_by_name=True)
    article_id: UUID = Field(default_factory=uuid4, alias="articleId")
    author_name: str = Field(alias="authorName")

ca = CamelArticle(articleId=uuid4(), authorName="Alice")
print(ca.model_dump(by_alias=True))   # {"articleId": ..., "authorName": "Alice"}
print(ca.model_dump(by_alias=False))  # {"article_id": ..., "author_name": "Alice"}

# JSON bytes for sending over wire
json_bytes: bytes = article.model_dump_json(exclude={"internal_notes"}).encode()
```

### Partial update models (all Optional)

```python
from pydantic import BaseModel
from typing import Optional
from uuid import UUID

class UserBase(BaseModel):
    name: str
    email: str
    bio: Optional[str] = None
    age: Optional[int] = None

class UserCreate(UserBase):
    password: str

class UserRead(UserBase):
    id: UUID

# Pattern 1: Manual partial model
class UserUpdate(BaseModel):
    name: Optional[str] = None
    email: Optional[str] = None
    bio: Optional[str] = None
    age: Optional[int] = None

# Pattern 2: Generate partial dynamically (Pydantic v2)
def make_partial(model_cls: type[BaseModel]) -> type[BaseModel]:
    fields = {
        name: (Optional[field.annotation], None)  # type: ignore[index]
        for name, field in model_cls.model_fields.items()
    }
    return type(f"Partial{model_cls.__name__}", (BaseModel,), {"__annotations__": {k: v[0] for k, v in fields.items()}, **{k: v[1] for k, v in fields.items()}})

@app.patch("/users/{user_id}", response_model=UserRead)
async def patch_user(
    user_id: UUID,
    update: UserUpdate,
    db: AsyncSession = Depends(get_db),
) -> UserRead:
    # Only apply fields that were actually sent in the request
    update_data = update.model_dump(exclude_unset=True)
    # Apply update_data to DB row...
    ...
```

### Request/Response model separation

```python
from pydantic import BaseModel, Field
from uuid import UUID, uuid4
from datetime import datetime

# Input — what the client sends
class PostCreate(BaseModel):
    title: str = Field(min_length=1, max_length=200)
    body: str = Field(min_length=1)
    tags: list[str] = Field(default_factory=list)

# Internal — includes sensitive/computed fields
class PostInDB(PostCreate):
    id: UUID = Field(default_factory=uuid4)
    author_id: int
    created_at: datetime = Field(default_factory=datetime.utcnow)
    updated_at: datetime = Field(default_factory=datetime.utcnow)
    view_count: int = 0
    is_deleted: bool = False

# Output — what the API returns (no sensitive fields)
class PostRead(BaseModel):
    id: UUID
    title: str
    body: str
    tags: list[str]
    author_id: int
    created_at: datetime
    view_count: int

@app.post("/posts", response_model=PostRead, status_code=201)
async def create_post(payload: PostCreate, author_id: int = 1) -> PostInDB:
    post = PostInDB(**payload.model_dump(), author_id=author_id)
    # FastAPI will filter PostInDB -> PostRead automatically via response_model
    return post
```

### BaseSettings for environment config

```python
from pydantic_settings import BaseSettings, SettingsConfigDict
from pydantic import Field, PostgresDsn, RedisDsn
from functools import lru_cache

class Settings(BaseSettings):
    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        case_sensitive=False,
    )

    app_name: str = "MyApp"
    debug: bool = False
    secret_key: str = Field(min_length=32)
    database_url: PostgresDsn
    redis_url: RedisDsn = "redis://localhost:6379/0"  # type: ignore
    allowed_origins: list[str] = ["http://localhost:3000"]
    max_upload_size_mb: int = Field(default=10, ge=1, le=100)

@lru_cache
def get_settings() -> Settings:
    return Settings()

app = FastAPI()

@app.get("/config")
async def show_config(settings: Settings = Depends(get_settings)) -> dict:
    return {"app_name": settings.app_name, "debug": settings.debug}
```

---

## 4. Background Tasks & Lifespan

### BackgroundTasks parameter

FastAPI's built-in `BackgroundTasks` runs work after the response is sent. Ideal for lightweight, non-critical tasks.

```python
from fastapi import FastAPI, BackgroundTasks
import logging

app = FastAPI()
logger = logging.getLogger(__name__)

def send_welcome_email(email: str, name: str) -> None:
    """Runs in the same process, after response is returned."""
    logger.info("Sending welcome email to %s <%s>", name, email)
    # smtp or third-party API call here

def log_event(event: str, metadata: dict) -> None:
    logger.info("Event: %s | %s", event, metadata)

@app.post("/users", status_code=201)
async def create_user(
    name: str,
    email: str,
    background_tasks: BackgroundTasks,
) -> dict:
    # ... save user to DB ...
    background_tasks.add_task(send_welcome_email, email, name)
    background_tasks.add_task(log_event, "user.created", {"email": email})
    return {"id": 1, "name": name, "email": email}
```

### Starlette background tasks (manual)

```python
from starlette.background import BackgroundTask
from fastapi.responses import JSONResponse

@app.delete("/users/{user_id}")
async def delete_user(user_id: int) -> JSONResponse:
    # ... delete from DB ...
    task = BackgroundTask(log_event, "user.deleted", {"id": user_id})
    return JSONResponse(
        content={"deleted": user_id},
        background=task,
    )
```

### Lifespan context manager (startup/shutdown)

The `lifespan` parameter replaces the deprecated `@app.on_event("startup")` / `@app.on_event("shutdown")` handlers.

```python
from contextlib import asynccontextmanager
from fastapi import FastAPI
from sqlalchemy.ext.asyncio import create_async_engine, async_sessionmaker
import httpx
import logging

logger = logging.getLogger(__name__)

@asynccontextmanager
async def lifespan(app: FastAPI):
    # ---- STARTUP ----
    logger.info("Starting up...")

    app.state.db_engine = create_async_engine(
        "postgresql+asyncpg://user:pass@localhost/db",
        pool_size=20,
        max_overflow=10,
        pool_pre_ping=True,
    )
    app.state.db_session_factory = async_sessionmaker(
        app.state.db_engine, expire_on_commit=False
    )
    app.state.http_client = httpx.AsyncClient(
        timeout=httpx.Timeout(10.0),
        limits=httpx.Limits(max_connections=50),
    )
    logger.info("Database and HTTP client ready.")

    yield  # app is running here

    # ---- SHUTDOWN ----
    logger.info("Shutting down...")
    await app.state.http_client.aclose()
    await app.state.db_engine.dispose()
    logger.info("Resources released.")

app = FastAPI(lifespan=lifespan)
```

### Shared resources in lifespan

```python
from fastapi import FastAPI, Depends, Request
from sqlalchemy.ext.asyncio import AsyncSession
from collections.abc import AsyncGenerator

# Access lifespan resources through request.app.state
async def get_db(request: Request) -> AsyncGenerator[AsyncSession, None]:
    async with request.app.state.db_session_factory() as session:
        yield session

async def get_http_client(request: Request) -> httpx.AsyncClient:
    return request.app.state.http_client

@app.get("/status")
async def status(
    db: AsyncSession = Depends(get_db),
    client: httpx.AsyncClient = Depends(get_http_client),
) -> dict:
    await db.execute(select(1))  # health check
    r = await client.get("https://api.example.com/health")
    return {"db": "ok", "upstream": r.status_code}
```

### Periodic tasks with asyncio

```python
import asyncio
from contextlib import asynccontextmanager
from fastapi import FastAPI
import logging

logger = logging.getLogger(__name__)

async def cleanup_expired_sessions() -> None:
    logger.info("Cleaning expired sessions...")
    # DB delete query here

async def periodic_task(interval_seconds: int, coro_func) -> None:
    """Run coro_func every interval_seconds until cancelled."""
    while True:
        try:
            await coro_func()
        except asyncio.CancelledError:
            raise
        except Exception:
            logger.exception("Periodic task %s failed", coro_func.__name__)
        await asyncio.sleep(interval_seconds)

@asynccontextmanager
async def lifespan(app: FastAPI):
    cleanup_task = asyncio.create_task(
        periodic_task(300, cleanup_expired_sessions)  # every 5 min
    )
    yield
    cleanup_task.cancel()
    try:
        await cleanup_task
    except asyncio.CancelledError:
        pass

app = FastAPI(lifespan=lifespan)
```

### Task queues with arq

```python
# worker.py
from arq import cron
from arq.connections import RedisSettings

async def send_email_task(ctx: dict, recipient: str, subject: str, body: str) -> None:
    # heavy email sending work lives here, not in the web process
    print(f"Sending email to {recipient}")

async def startup(ctx: dict) -> None:
    import httpx
    ctx["http_client"] = httpx.AsyncClient()

async def shutdown(ctx: dict) -> None:
    await ctx["http_client"].aclose()

class WorkerSettings:
    functions = [send_email_task]
    on_startup = startup
    on_shutdown = shutdown
    redis_settings = RedisSettings(host="localhost", port=6379)

# In FastAPI app:
from arq import create_pool
from arq.connections import RedisSettings

@asynccontextmanager
async def lifespan(app: FastAPI):
    app.state.arq_pool = await create_pool(RedisSettings())
    yield
    await app.state.arq_pool.close()

@app.post("/send-email")
async def enqueue_email(request: Request, recipient: str, subject: str) -> dict:
    job = await request.app.state.arq_pool.enqueue_job(
        "send_email_task", recipient, subject, "Hello from FastAPI"
    )
    return {"job_id": job.job_id}
```

### Graceful shutdown patterns

```python
import asyncio
import signal
from contextlib import asynccontextmanager
from fastapi import FastAPI

_shutdown_event = asyncio.Event()

@asynccontextmanager
async def lifespan(app: FastAPI):
    loop = asyncio.get_running_loop()

    def handle_sigterm():
        _shutdown_event.set()

    loop.add_signal_handler(signal.SIGTERM, handle_sigterm)

    yield

    # Wait for in-flight requests to finish (uvicorn handles this)
    # But we can wait for our own long-running tasks
    if _shutdown_event.is_set():
        await asyncio.sleep(2)  # drain time

app = FastAPI(lifespan=lifespan)
```

---

## 5. Router Organization

### APIRouter with prefix, tags, dependencies

```python
from fastapi import APIRouter, Depends

users_router = APIRouter(
    prefix="/users",
    tags=["users"],
    dependencies=[Depends(get_current_user)],
    responses={404: {"description": "User not found"}},
)

@users_router.get("/")
async def list_users() -> list[dict]:
    return []

@users_router.get("/{user_id}")
async def get_user(user_id: int) -> dict:
    return {"id": user_id}

@users_router.post("/", status_code=201)
async def create_user(name: str) -> dict:
    return {"name": name}
```

### Include router patterns

```python
from fastapi import FastAPI

app = FastAPI()

# Flat include
app.include_router(users_router)

# Include with additional prefix / tag override
app.include_router(
    users_router,
    prefix="/api",        # results in /api/users/...
    tags=["users-v2"],
)

# Include with additional dependency
app.include_router(
    users_router,
    dependencies=[Depends(require_role("admin"))],
)
```

### Versioned API routes (v1, v2)

```python
from fastapi import FastAPI, APIRouter

v1 = APIRouter(prefix="/v1")
v2 = APIRouter(prefix="/v2")

@v1.get("/users")
async def list_users_v1() -> list[dict]:
    return [{"id": 1, "name": "Alice"}]

@v2.get("/users")
async def list_users_v2() -> dict:
    return {"data": [{"id": 1, "name": "Alice"}], "meta": {"total": 1}}

app = FastAPI()
app.include_router(v1)
app.include_router(v2)
```

### Nested routers

```python
from fastapi import APIRouter

# /orgs/{org_id}/repos/{repo_id}/issues
orgs_router = APIRouter(prefix="/orgs/{org_id}")
repos_router = APIRouter(prefix="/repos/{repo_id}")
issues_router = APIRouter(prefix="/issues")

@issues_router.get("/")
async def list_issues(org_id: int, repo_id: int) -> list[dict]:
    return [{"org": org_id, "repo": repo_id}]

repos_router.include_router(issues_router)
orgs_router.include_router(repos_router)

app = FastAPI()
app.include_router(orgs_router)
```

### Route naming and path operation configuration

```python
from fastapi import APIRouter
from fastapi.responses import ORJSONResponse

router = APIRouter()

@router.get(
    "/items/{item_id}",
    name="items:get",                    # for url_path_for("items:get", item_id=1)
    summary="Get a single item",
    description="Fetch an item by its integer ID.",
    response_class=ORJSONResponse,       # faster JSON serializer
    response_model=ItemRead,
    response_model_exclude_unset=True,   # hide unset fields in response
    response_model_exclude_none=True,    # hide None fields in response
    deprecated=False,
    include_in_schema=True,
    operation_id="getItemById",          # OpenAPI operationId
)
async def get_item(item_id: int) -> ItemRead:
    ...
```

### response_model_exclude_unset

```python
from fastapi import APIRouter
from pydantic import BaseModel

class ItemRead(BaseModel):
    id: int
    name: str
    description: str | None = None
    price: float | None = None

router = APIRouter()

# Useful for PATCH: only return fields that exist in the stored object
@router.get("/items/{item_id}", response_model=ItemRead, response_model_exclude_unset=True)
async def get_item_sparse(item_id: int) -> dict:
    # If the DB row has no description, it won't appear in the response
    return {"id": item_id, "name": "Widget"}
```

---

## 6. Error Handling

### HTTPException with detail and headers

```python
from fastapi import FastAPI, HTTPException

app = FastAPI()

@app.get("/items/{item_id}")
async def get_item(item_id: int) -> dict:
    if item_id <= 0:
        raise HTTPException(
            status_code=422,
            detail=f"item_id must be positive, got {item_id}",
        )
    if item_id == 999:
        raise HTTPException(
            status_code=404,
            detail={"message": "Item not found", "item_id": item_id},
        )
    if item_id == 1:
        raise HTTPException(
            status_code=401,
            detail="Authentication required",
            headers={"WWW-Authenticate": "Bearer"},
        )
    return {"id": item_id, "name": "Widget"}
```

### Custom exception classes

```python
from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse

class AppError(Exception):
    def __init__(self, status_code: int, code: str, message: str) -> None:
        self.status_code = status_code
        self.code = code
        self.message = message

class NotFoundError(AppError):
    def __init__(self, resource: str, id: int | str) -> None:
        super().__init__(404, "NOT_FOUND", f"{resource} with id={id} not found")

class ConflictError(AppError):
    def __init__(self, message: str) -> None:
        super().__init__(409, "CONFLICT", message)

app = FastAPI()

@app.exception_handler(AppError)
async def app_error_handler(request: Request, exc: AppError) -> JSONResponse:
    return JSONResponse(
        status_code=exc.status_code,
        content={"error": {"code": exc.code, "message": exc.message}},
    )

@app.get("/users/{user_id}")
async def get_user(user_id: int) -> dict:
    if user_id == 0:
        raise NotFoundError("User", user_id)
    return {"id": user_id}
```

### Exception handlers (@app.exception_handler)

```python
from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse
from fastapi.exceptions import RequestValidationError
from pydantic import ValidationError
import logging

logger = logging.getLogger(__name__)
app = FastAPI()

@app.exception_handler(RequestValidationError)
async def validation_exception_handler(
    request: Request, exc: RequestValidationError
) -> JSONResponse:
    errors = []
    for error in exc.errors():
        errors.append({
            "field": ".".join(str(loc) for loc in error["loc"]),
            "message": error["msg"],
            "type": error["type"],
        })
    return JSONResponse(
        status_code=422,
        content={"detail": "Validation failed", "errors": errors},
    )

@app.exception_handler(Exception)
async def unhandled_exception_handler(
    request: Request, exc: Exception
) -> JSONResponse:
    logger.exception("Unhandled exception for %s %s", request.method, request.url)
    return JSONResponse(
        status_code=500,
        content={"detail": "Internal server error"},
    )
```

### Error response schemas

```python
from pydantic import BaseModel

class ErrorDetail(BaseModel):
    field: str
    message: str
    type: str

class ErrorResponse(BaseModel):
    detail: str
    errors: list[ErrorDetail] | None = None
    request_id: str | None = None

class HTTPErrorResponse(BaseModel):
    detail: str

# Document error responses in route definitions
@app.get(
    "/users/{user_id}",
    responses={
        404: {"model": HTTPErrorResponse, "description": "User not found"},
        403: {"model": HTTPErrorResponse, "description": "Forbidden"},
    },
)
async def get_user_documented(user_id: int) -> dict:
    return {"id": user_id}
```

### Logging errors properly

```python
import logging
import traceback
from fastapi import Request
from fastapi.responses import JSONResponse

logger = logging.getLogger(__name__)

@app.exception_handler(Exception)
async def global_handler(request: Request, exc: Exception) -> JSONResponse:
    logger.error(
        "Unhandled error: %s %s → %s: %s",
        request.method,
        request.url.path,
        type(exc).__name__,
        exc,
        extra={
            "request_id": request.headers.get("X-Request-ID"),
            "client": request.client.host if request.client else None,
        },
        exc_info=True,  # includes traceback in log
    )
    return JSONResponse(status_code=500, content={"detail": "Internal server error"})
```

---

## 7. Middleware Patterns

### @app.middleware("http") decorator

```python
import time
import uuid
from fastapi import FastAPI, Request, Response

app = FastAPI()

@app.middleware("http")
async def add_timing_header(request: Request, call_next) -> Response:
    start = time.perf_counter()
    response = await call_next(request)
    elapsed_ms = (time.perf_counter() - start) * 1000
    response.headers["X-Process-Time-Ms"] = f"{elapsed_ms:.2f}"
    return response

@app.middleware("http")
async def inject_request_id(request: Request, call_next) -> Response:
    request_id = request.headers.get("X-Request-ID", str(uuid.uuid4()))
    request.state.request_id = request_id
    response = await call_next(request)
    response.headers["X-Request-ID"] = request_id
    return response
```

### BaseHTTPMiddleware subclass

```python
from starlette.middleware.base import BaseHTTPMiddleware
from starlette.requests import Request
from starlette.responses import Response
import logging

logger = logging.getLogger(__name__)

class RequestLoggingMiddleware(BaseHTTPMiddleware):
    async def dispatch(self, request: Request, call_next) -> Response:
        logger.info(
            "→ %s %s | client=%s",
            request.method,
            request.url.path,
            request.client.host if request.client else "unknown",
        )
        response = await call_next(request)
        logger.info(
            "← %s %s | status=%d",
            request.method,
            request.url.path,
            response.status_code,
        )
        return response

app = FastAPI()
app.add_middleware(RequestLoggingMiddleware)
```

### Pure ASGI middleware

Use pure ASGI when you need maximum performance or access to raw ASGI messages (e.g., streaming).

```python
from collections.abc import Callable, Awaitable
from starlette.types import ASGIApp, Receive, Scope, Send

class RawTimingMiddleware:
    def __init__(self, app: ASGIApp) -> None:
        self.app = app

    async def __call__(self, scope: Scope, receive: Receive, send: Send) -> None:
        if scope["type"] != "http":
            await self.app(scope, receive, send)
            return

        import time
        start = time.perf_counter()

        async def send_with_timing(message):
            if message["type"] == "http.response.start":
                elapsed = (time.perf_counter() - start) * 1000
                headers = list(message.get("headers", []))
                headers.append((b"x-process-time-ms", f"{elapsed:.2f}".encode()))
                message = {**message, "headers": headers}
            await send(message)

        await self.app(scope, receive, send_with_timing)

app = FastAPI()
app.add_middleware(RawTimingMiddleware)
```

### Custom CORS configuration

```python
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

app = FastAPI()

app.add_middleware(
    CORSMiddleware,
    allow_origins=["https://app.example.com", "https://admin.example.com"],
    allow_origin_regex=r"https://.*\.example\.com",
    allow_credentials=True,
    allow_methods=["GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"],
    allow_headers=["Authorization", "Content-Type", "X-Request-ID"],
    expose_headers=["X-Request-ID", "X-Process-Time-Ms"],
    max_age=3600,
)
```

### Trusted host and GZip middleware

```python
from starlette.middleware.trustedhost import TrustedHostMiddleware
from starlette.middleware.gzip import GZipMiddleware

app.add_middleware(GZipMiddleware, minimum_size=1000)
app.add_middleware(
    TrustedHostMiddleware,
    allowed_hosts=["example.com", "*.example.com", "localhost"],
)
```

---

## 8. WebSocket Patterns

### Basic WebSocket endpoint

```python
from fastapi import FastAPI, WebSocket, WebSocketDisconnect
import logging

app = FastAPI()
logger = logging.getLogger(__name__)

@app.websocket("/ws")
async def websocket_endpoint(websocket: WebSocket) -> None:
    await websocket.accept()
    try:
        while True:
            data = await websocket.receive_text()
            echo = f"Echo: {data}"
            await websocket.send_text(echo)
    except WebSocketDisconnect as exc:
        logger.info("Client disconnected: code=%d", exc.code)
```

### Connection manager class

```python
from fastapi import WebSocket
from typing import Any
import asyncio

class ConnectionManager:
    def __init__(self) -> None:
        self.active: list[WebSocket] = []
        self._lock = asyncio.Lock()

    async def connect(self, ws: WebSocket) -> None:
        await ws.accept()
        async with self._lock:
            self.active.append(ws)

    async def disconnect(self, ws: WebSocket) -> None:
        async with self._lock:
            self.active.remove(ws)

    async def broadcast(self, message: Any) -> None:
        dead: list[WebSocket] = []
        for ws in list(self.active):
            try:
                await ws.send_json(message)
            except Exception:
                dead.append(ws)
        for ws in dead:
            await self.disconnect(ws)

    async def send_personal(self, ws: WebSocket, message: Any) -> None:
        await ws.send_json(message)

manager = ConnectionManager()

@app.websocket("/ws/chat")
async def chat_ws(websocket: WebSocket) -> None:
    await manager.connect(websocket)
    try:
        while True:
            data = await websocket.receive_json()
            await manager.broadcast({"from": "server", "data": data})
    except WebSocketDisconnect:
        await manager.disconnect(websocket)
        await manager.broadcast({"system": "A user disconnected"})
```

### Room-based broadcasting

```python
from fastapi import WebSocket, WebSocketDisconnect
from collections import defaultdict
import asyncio

class RoomManager:
    def __init__(self) -> None:
        self.rooms: dict[str, list[WebSocket]] = defaultdict(list)
        self._lock = asyncio.Lock()

    async def join(self, room: str, ws: WebSocket) -> None:
        await ws.accept()
        async with self._lock:
            self.rooms[room].append(ws)

    async def leave(self, room: str, ws: WebSocket) -> None:
        async with self._lock:
            self.rooms[room].remove(ws)
            if not self.rooms[room]:
                del self.rooms[room]

    async def broadcast_room(self, room: str, message: dict, exclude: WebSocket | None = None) -> None:
        for ws in list(self.rooms.get(room, [])):
            if ws is exclude:
                continue
            try:
                await ws.send_json(message)
            except Exception:
                await self.leave(room, ws)

room_manager = RoomManager()

@app.websocket("/ws/rooms/{room_id}")
async def room_ws(websocket: WebSocket, room_id: str) -> None:
    await room_manager.join(room_id, websocket)
    await room_manager.broadcast_room(room_id, {"event": "join", "room": room_id}, exclude=websocket)
    try:
        while True:
            msg = await websocket.receive_json()
            await room_manager.broadcast_room(room_id, {"event": "message", "data": msg})
    except WebSocketDisconnect:
        await room_manager.leave(room_id, websocket)
        await room_manager.broadcast_room(room_id, {"event": "leave", "room": room_id})
```

### Authentication handshake

```python
from fastapi import WebSocket, WebSocketDisconnect, Query, status
from jose import JWTError, jwt

SECRET_KEY = "your-secret"
ALGORITHM = "HS256"

async def get_ws_user(token: str) -> dict | None:
    try:
        payload = jwt.decode(token, SECRET_KEY, algorithms=[ALGORITHM])
        return {"id": payload.get("sub"), "role": payload.get("role")}
    except JWTError:
        return None

@app.websocket("/ws/secure")
async def secure_ws(
    websocket: WebSocket,
    token: str = Query(...),  # ?token=<jwt>
) -> None:
    user = await get_ws_user(token)
    if user is None:
        await websocket.close(code=status.WS_1008_POLICY_VIOLATION)
        return

    await websocket.accept()
    await websocket.send_json({"event": "authenticated", "user_id": user["id"]})
    try:
        while True:
            data = await websocket.receive_json()
            await websocket.send_json({"echo": data})
    except WebSocketDisconnect:
        pass
```

### Heartbeat / ping-pong

```python
import asyncio
from fastapi import WebSocket, WebSocketDisconnect

HEARTBEAT_INTERVAL = 30  # seconds

@app.websocket("/ws/heartbeat")
async def heartbeat_ws(websocket: WebSocket) -> None:
    await websocket.accept()
    last_pong = asyncio.get_event_loop().time()

    async def pinger() -> None:
        nonlocal last_pong
        while True:
            await asyncio.sleep(HEARTBEAT_INTERVAL)
            try:
                await websocket.send_json({"type": "ping"})
                if asyncio.get_event_loop().time() - last_pong > HEARTBEAT_INTERVAL * 2:
                    await websocket.close()
                    return
            except Exception:
                return

    ping_task = asyncio.create_task(pinger())
    try:
        while True:
            msg = await websocket.receive_json()
            if msg.get("type") == "pong":
                last_pong = asyncio.get_event_loop().time()
            else:
                await websocket.send_json({"echo": msg})
    except WebSocketDisconnect:
        pass
    finally:
        ping_task.cancel()
```

### Error handling and graceful disconnect

```python
from fastapi import WebSocket, WebSocketDisconnect
from fastapi.websockets import WebSocketState
import logging

logger = logging.getLogger(__name__)

@app.websocket("/ws/robust")
async def robust_ws(websocket: WebSocket) -> None:
    await websocket.accept()
    try:
        while True:
            try:
                data = await asyncio.wait_for(
                    websocket.receive_json(),
                    timeout=60.0,  # close idle connections after 60s
                )
            except asyncio.TimeoutError:
                logger.info("WebSocket idle timeout, closing gracefully.")
                break
            except Exception as exc:
                logger.warning("WS receive error: %s", exc)
                break

            try:
                result = process_message(data)  # your business logic
                await websocket.send_json({"ok": True, "result": result})
            except ValueError as exc:
                await websocket.send_json({"ok": False, "error": str(exc)})
            except Exception:
                logger.exception("Error processing WS message")
                await websocket.send_json({"ok": False, "error": "Internal error"})

    except WebSocketDisconnect as exc:
        logger.info("WS disconnected: code=%d reason=%s", exc.code, exc.reason)
    finally:
        if websocket.client_state != WebSocketState.DISCONNECTED:
            await websocket.close(code=1000)

def process_message(data: dict) -> dict:
    return {"processed": data}
```

---

## Quick-Reference Cheatsheet

| Pattern | Import |
|---|---|
| `AsyncSession` | `from sqlalchemy.ext.asyncio import AsyncSession` |
| `async_sessionmaker` | `from sqlalchemy.ext.asyncio import async_sessionmaker` |
| `AsyncClient` (httpx) | `import httpx; httpx.AsyncClient()` |
| `BackgroundTasks` | `from fastapi import BackgroundTasks` |
| `lifespan` | `from contextlib import asynccontextmanager` |
| `APIRouter` | `from fastapi import APIRouter` |
| `HTTPException` | `from fastapi import HTTPException` |
| `BaseHTTPMiddleware` | `from starlette.middleware.base import BaseHTTPMiddleware` |
| `WebSocketDisconnect` | `from fastapi import WebSocketDisconnect` |
| `OAuth2PasswordBearer` | `from fastapi.security import OAuth2PasswordBearer` |
| `Field` | `from pydantic import Field` |
| `field_validator` | `from pydantic import field_validator` |
| `model_validator` | `from pydantic import model_validator` |
| `computed_field` | `from pydantic import computed_field` |
| `ConfigDict` | `from pydantic import ConfigDict` |
| `BaseSettings` | `from pydantic_settings import BaseSettings` |

### Key Rules

- Always use `async def` when you `await` anything inside the handler.
- Never call blocking I/O (`requests`, `time.sleep`, sync SQLAlchemy) inside `async def`.
- Prefer a single shared `httpx.AsyncClient` (created in lifespan) over per-request clients.
- Use `response_model_exclude_unset=True` on PATCH endpoints so unchanged fields are not serialized.
- Prefer `lifespan` over deprecated `@app.on_event` for startup/shutdown logic.
- Use `Annotated[..., Depends(...)]` style for dependency injection — it is cleaner than default-value `Depends`.
- Use Pydantic v2 `model_validate()` instead of v1 `from_orm()` or `.parse_obj()`.
- Use `model_dump(exclude_unset=True)` when applying partial updates to avoid overwriting fields with defaults.
- Discriminated unions with `Literal` are faster and produce better OpenAPI schemas than untagged unions.
- Add `pool_pre_ping=True` to SQLAlchemy engines to recover from dropped DB connections.
