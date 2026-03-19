---
name: python-web-setup
description: >
  Python Web Framework Suite — complete development toolkit for building production-grade
  Django 5.x, Flask 3.x, and FastAPI applications. Django with ORM, admin, DRF, Celery, and Channels.
  FastAPI with async endpoints, Pydantic v2, dependency injection, and WebSocket. Flask with
  blueprints, SQLAlchemy 2.0, Flask-Login, and Jinja2 templating. Security coverage for CSRF, XSS,
  SQL injection, auth, CORS, rate limiting, and secrets management.
  Triggers: "django project", "django app", "django rest framework", "drf", "django admin",
  "django celery", "django channels", "django orm", "django model",
  "fastapi project", "fastapi endpoint", "fastapi async", "pydantic model", "pydantic v2",
  "fastapi dependency injection", "fastapi websocket",
  "flask project", "flask app", "flask blueprint", "flask sqlalchemy", "flask login",
  "flask migrate", "jinja2 template",
  "python web security", "python csrf", "python auth", "python deployment",
  "gunicorn", "uvicorn", "python docker".
  Dispatches the appropriate specialist agent: django-architect, fastapi-expert,
  flask-developer, or python-web-security.
  NOT for: Data science, machine learning, desktop apps, mobile apps, or non-Python frameworks.
version: 1.0.0
argument-hint: "<django|fastapi|flask|security> [target]"
user-invocable: true
allowed-tools: Read, Grep, Glob, Bash
model: sonnet
---

# Python Web Framework Suite

Four specialist agents covering every major Python web framework and the security layer that protects them. Whether you're scaffolding a new Django monolith, building a high-throughput FastAPI service, assembling a Flask app with blueprints, or hardening an existing Python backend against OWASP threats — dispatch the right agent and get production-ready code, not tutorials.

## Available Agents

### Django Architect (`django-architect`)

Expert in Django 5.x full-stack and API development. Handles the full lifecycle from project scaffolding to production deployment: settings management, ORM modeling with migrations, the admin interface, Django REST Framework serializers and viewsets, Celery background tasks, and real-time features via Django Channels and WebSockets.

**Invoke**: Dispatch via Task tool with `subagent_type: "django-architect"`.

**Example prompts**:
- "Scaffold a new Django 5 project with split settings (base/dev/prod), PostgreSQL, and a custom user model."
- "Create a DRF viewset for a Product model with nested serializers, filtering, pagination, and JWT auth."
- "Set up Celery with Redis as the broker and beat scheduler for sending daily digest emails."
- "Add Django Channels to an existing project and implement a WebSocket notification endpoint."

---

### FastAPI Expert (`fastapi-expert`)

Expert in high-performance async Python APIs. Handles async endpoint design, Pydantic v2 models with validators and computed fields, dependency injection patterns, background tasks, OAuth2 and JWT auth flows, Alembic migrations with SQLAlchemy 2.0 async, and WebSocket connections. Generates OpenAPI-complete, production-wired code.

**Invoke**: Dispatch via Task tool with `subagent_type: "fastapi-expert"`.

**Example prompts**:
- "Set up a FastAPI project with async PostgreSQL via asyncpg, SQLAlchemy 2.0, Alembic, and Pydantic v2 schemas."
- "Create a dependency injection chain for database sessions, current user, and role-based permission checks."
- "Build a WebSocket endpoint that broadcasts real-time order status updates to subscribed clients."
- "Add OAuth2 password flow with JWT access tokens, refresh tokens, and token blacklisting via Redis."

---

### Flask Developer (`flask-developer`)

Expert in Flask 3.x application architecture. Handles application factory pattern, blueprint registration, SQLAlchemy 2.0 models with Flask-Migrate, Flask-Login session management, WTForms with CSRF protection, Jinja2 template inheritance, and REST API layering. Produces clean, testable code with a logical directory structure.

**Invoke**: Dispatch via Task tool with `subagent_type: "flask-developer"`.

**Example prompts**:
- "Create a Flask app using the application factory pattern with blueprints for auth, dashboard, and API routes."
- "Set up SQLAlchemy 2.0 with Flask-Migrate, define User and Post models, and generate the initial migration."
- "Implement Flask-Login with remember-me, email-based password reset, and role-based access decorators."
- "Build a Jinja2 base template with block inheritance, macros for form rendering, and Bootstrap 5 integration."

---

### Python Web Security (`python-web-security`)

Expert in securing Python web applications against the OWASP Top 10 and beyond. Performs read-only audits, produces severity-rated findings, and writes the exact remediation code — CSRF middleware, parameterized queries, XSS escaping, CORS configuration, rate limiting, secrets management, and auth hardening. Framework-agnostic: works with Django, FastAPI, or Flask.

**Invoke**: Dispatch via Task tool with `subagent_type: "python-web-security"`.

**Example prompts**:
- "Audit the views/ and serializers/ directories for injection vulnerabilities, IDOR, and missing auth checks."
- "Implement rate limiting on the /api/auth/ endpoints using slowapi (FastAPI) or Django Ratelimit."
- "Review our secrets and environment variable handling — check for hardcoded credentials and insecure config."
- "Add CORS, HSTS, Content-Security-Policy, and X-Frame-Options headers to the Flask application."

---

## Quick Start

Invoke this skill directly or dispatch agents via Task:

```
/python-web-setup django Scaffold a new e-commerce API project with DRF and Celery
```

```
/python-web-setup fastapi Set up async PostgreSQL with SQLAlchemy 2.0 and Alembic
```

```
/python-web-setup flask Add authentication to the existing Flask app in ./src/
```

```
/python-web-setup security Audit the authentication layer for OWASP vulnerabilities
```

---

## Agent Selection Guide

Use this table to pick the right agent before dispatching.

| Task | Agent |
|------|-------|
| New Django project with ORM, admin, custom user model | `django-architect` |
| Django REST Framework — serializers, viewsets, routers | `django-architect` |
| Celery tasks, beat schedules, task queues | `django-architect` |
| Django Channels — WebSockets, consumers, channel layers | `django-architect` |
| Django admin customizations, inlines, list filters | `django-architect` |
| High-performance async API endpoints | `fastapi-expert` |
| Pydantic v2 models, validators, computed fields | `fastapi-expert` |
| FastAPI dependency injection, lifespan events | `fastapi-expert` |
| Alembic migrations with async SQLAlchemy 2.0 | `fastapi-expert` |
| FastAPI WebSocket connections, connection managers | `fastapi-expert` |
| Flask application factory and blueprint structure | `flask-developer` |
| SQLAlchemy 2.0 models and Flask-Migrate | `flask-developer` |
| Flask-Login, session management, password reset | `flask-developer` |
| Jinja2 templates, macros, template inheritance | `flask-developer` |
| WTForms with validation and CSRF protection | `flask-developer` |
| Security audit — OWASP, injection, auth flaws | `python-web-security` |
| CSRF, XSS, SQL injection hardening | `python-web-security` |
| CORS, security headers, HSTS, CSP | `python-web-security` |
| Rate limiting, brute-force protection | `python-web-security` |
| Secrets management, environment variable hygiene | `python-web-security` |

---

## Project Templates

### Django 5.x Project

**Directory structure**:
```
myproject/
├── config/
│   ├── settings/
│   │   ├── base.py
│   │   ├── development.py
│   │   └── production.py
│   ├── urls.py
│   └── wsgi.py
├── apps/
│   ├── users/
│   │   ├── models.py
│   │   ├── serializers.py
│   │   ├── views.py
│   │   └── urls.py
│   └── core/
├── requirements/
│   ├── base.txt
│   ├── development.txt
│   └── production.txt
├── Dockerfile
├── docker-compose.yml
└── manage.py
```

**Setup commands**:
```bash
# Create project and virtualenv
python -m venv .venv && source .venv/bin/activate
pip install django==5.2 djangorestframework psycopg[binary] celery redis

# Scaffold project with split settings
django-admin startproject config .
mkdir -p config/settings apps

# Run migrations with custom settings module
DJANGO_SETTINGS_MODULE=config.settings.development python manage.py migrate

# Create superuser and start dev server
python manage.py createsuperuser
python manage.py runserver
```

**Key requirements** (`requirements/base.txt`):
```
django==5.2
djangorestframework==3.15
django-filter==24.3
djangorestframework-simplejwt==5.3
celery==5.4
redis==5.0
psycopg[binary]==3.2
django-environ==0.11
gunicorn==23.0
```

**Docker** (`Dockerfile`):
```dockerfile
FROM python:3.12-slim
WORKDIR /app
COPY requirements/production.txt .
RUN pip install --no-cache-dir -r production.txt
COPY . .
RUN python manage.py collectstatic --noinput
CMD ["gunicorn", "config.wsgi:application", "--bind", "0.0.0.0:8000", "--workers", "4"]
```

---

### FastAPI Project

**Directory structure**:
```
myapi/
├── app/
│   ├── api/
│   │   ├── v1/
│   │   │   ├── endpoints/
│   │   │   └── router.py
│   │   └── deps.py
│   ├── core/
│   │   ├── config.py
│   │   └── security.py
│   ├── db/
│   │   ├── base.py
│   │   └── session.py
│   ├── models/
│   ├── schemas/
│   └── main.py
├── alembic/
│   └── versions/
├── alembic.ini
├── Dockerfile
├── docker-compose.yml
└── requirements.txt
```

**Setup commands**:
```bash
# Create virtualenv and install deps
python -m venv .venv && source .venv/bin/activate
pip install fastapi==0.115 uvicorn[standard] sqlalchemy==2.0 asyncpg alembic pydantic-settings

# Initialize Alembic
alembic init alembic

# Generate and run first migration
alembic revision --autogenerate -m "initial"
alembic upgrade head

# Start dev server with reload
uvicorn app.main:app --reload --host 0.0.0.0 --port 8000
```

**Key requirements** (`requirements.txt`):
```
fastapi==0.115
uvicorn[standard]==0.32
sqlalchemy==2.0
asyncpg==0.30
alembic==1.14
pydantic==2.10
pydantic-settings==2.7
python-jose[cryptography]==3.3
passlib[bcrypt]==1.7
python-multipart==0.0.20
```

**Docker** (`Dockerfile`):
```dockerfile
FROM python:3.12-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY . .
CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8000", "--workers", "4"]
```

---

### Flask 3.x Project

**Directory structure**:
```
myflaskapp/
├── app/
│   ├── auth/
│   │   ├── __init__.py
│   │   ├── forms.py
│   │   ├── routes.py
│   │   └── templates/auth/
│   ├── main/
│   │   ├── __init__.py
│   │   ├── routes.py
│   │   └── templates/main/
│   ├── models.py
│   ├── extensions.py
│   └── __init__.py          # application factory
├── migrations/
├── tests/
├── config.py
├── Dockerfile
├── docker-compose.yml
└── requirements.txt
```

**Setup commands**:
```bash
# Create virtualenv and install deps
python -m venv .venv && source .venv/bin/activate
pip install flask==3.1 flask-sqlalchemy flask-migrate flask-login flask-wtf

# Initialize migrations
flask db init
flask db migrate -m "initial schema"
flask db upgrade

# Start dev server
flask --app app run --debug
```

**Key requirements** (`requirements.txt`):
```
flask==3.1
flask-sqlalchemy==3.1
flask-migrate==4.0
flask-login==0.6
flask-wtf==1.2
sqlalchemy==2.0
psycopg2-binary==2.9
email-validator==2.2
gunicorn==23.0
python-dotenv==1.0
```

**Docker** (`Dockerfile`):
```dockerfile
FROM python:3.12-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY . .
CMD ["gunicorn", "app:create_app()", "--bind", "0.0.0.0:8000", "--workers", "4"]
```

---

## Common Workflows

### Add a new Django REST API endpoint

Use `django-architect` to scaffold the full endpoint stack: model, serializer, viewset, and URL registration.

```
Task tool:
  subagent_type: "django-architect"
  description: "Add Product API endpoint"
  prompt: |
    Add a new REST API endpoint for a Product model in the apps/products/ directory.
    Fields: name (str), sku (str, unique), price (decimal), stock (int), is_active (bool).
    Needs: ModelSerializer with validation, ModelViewSet with filtering by is_active and
    price range, pagination (20 per page), and JWT authentication. Register under /api/v1/products/.
  mode: "bypassPermissions"
```

---

### Set up FastAPI with async PostgreSQL

Use `fastapi-expert` to wire up the full async database stack from scratch.

```
Task tool:
  subagent_type: "fastapi-expert"
  description: "Set up async PostgreSQL stack"
  prompt: |
    Set up async PostgreSQL in this FastAPI project. Use SQLAlchemy 2.0 with asyncpg,
    async session factory with proper lifespan context, Alembic for migrations, and
    Pydantic v2 schemas that are separate from ORM models. Create a User model as the
    first example with id, email (unique), hashed_password, created_at, and is_active.
    Include the dependency injection function for getting a database session.
  mode: "bypassPermissions"
```

---

### Add authentication to Flask app

Use `flask-developer` to implement a complete auth system with login, logout, registration, and protected routes.

```
Task tool:
  subagent_type: "flask-developer"
  description: "Implement Flask auth system"
  prompt: |
    Add a full authentication system to the Flask app at ./app/. Use Flask-Login for
    session management and Flask-WTF for form validation with CSRF. Needs: User model
    with email/password (hashed with bcrypt), registration with email validation, login
    with remember-me, logout, and a password reset flow using time-limited tokens sent
    by email. Protect non-auth routes with @login_required. Add auth/ blueprint with
    templates for all forms.
  mode: "bypassPermissions"
```

---

### Security audit a Python web app

Use `python-web-security` to get a structured, severity-rated audit with exact remediation code.

```
Task tool:
  subagent_type: "python-web-security"
  description: "Full security audit"
  prompt: |
    Perform a comprehensive security audit of this Python web application. Read all files
    in app/ and check for: (1) SQL injection via raw queries or ORM misuse, (2) missing
    CSRF protection on state-changing endpoints, (3) XSS in template rendering, (4) weak
    or missing authentication on protected routes, (5) hardcoded secrets or credentials,
    (6) overly permissive CORS configuration, (7) missing rate limiting on auth endpoints,
    (8) insecure direct object references. Produce a report with severity (Critical/High/
    Medium/Low) and the exact code fix for each finding.
  mode: "bypassPermissions"
```
