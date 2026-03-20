---
name: django-patterns
description: >
  Production Django patterns — models, views, serializers, authentication,
  migrations, queryset optimization, testing, and deployment patterns.
  Triggers: "django", "django models", "django views", "django rest framework",
  "drf", "django authentication", "django testing", "django queryset".
  NOT for: FastAPI projects (use fastapi-patterns).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Django Production Patterns

## Project Structure

```
project/
  manage.py
  config/
    __init__.py
    settings/
      __init__.py
      base.py         # Shared settings
      development.py
      production.py
    urls.py
    wsgi.py
    asgi.py
  apps/
    users/
      __init__.py
      models.py
      serializers.py
      views.py
      urls.py
      admin.py
      tests/
        test_models.py
        test_views.py
        factories.py
    orders/
      ...
```

## Models

```python
# apps/users/models.py
from django.db import models
from django.contrib.auth.models import AbstractUser
import uuid

class TimeStampedModel(models.Model):
    """Abstract base for all models."""
    id = models.UUIDField(primary_key=True, default=uuid.uuid4, editable=False)
    created_at = models.DateTimeField(auto_now_add=True, db_index=True)
    updated_at = models.DateTimeField(auto_now=True)

    class Meta:
        abstract = True
        ordering = ["-created_at"]

class User(AbstractUser, TimeStampedModel):
    email = models.EmailField(unique=True)
    avatar = models.ImageField(upload_to="avatars/", blank=True)
    bio = models.TextField(max_length=500, blank=True)

    USERNAME_FIELD = "email"
    REQUIRED_FIELDS = ["username"]

    class Meta:
        db_table = "users"
        indexes = [
            models.Index(fields=["email"]),
            models.Index(fields=["-created_at"]),
        ]

    def __str__(self):
        return self.email

class Profile(TimeStampedModel):
    user = models.OneToOneField(User, on_delete=models.CASCADE, related_name="profile")
    company = models.CharField(max_length=200, blank=True)
    website = models.URLField(blank=True)

    def __str__(self):
        return f"Profile of {self.user.email}"
```

```python
# apps/orders/models.py
class Order(TimeStampedModel):
    class Status(models.TextChoices):
        PENDING = "pending", "Pending"
        PAID = "paid", "Paid"
        SHIPPED = "shipped", "Shipped"
        DELIVERED = "delivered", "Delivered"
        CANCELLED = "cancelled", "Cancelled"

    user = models.ForeignKey(User, on_delete=models.CASCADE, related_name="orders")
    status = models.CharField(max_length=20, choices=Status.choices, default=Status.PENDING, db_index=True)
    total = models.DecimalField(max_digits=10, decimal_places=2)
    notes = models.TextField(blank=True)

    class Meta:
        db_table = "orders"
        indexes = [
            models.Index(fields=["user", "-created_at"]),
            models.Index(fields=["status"]),
        ]

class OrderItem(TimeStampedModel):
    order = models.ForeignKey(Order, on_delete=models.CASCADE, related_name="items")
    product_name = models.CharField(max_length=200)
    quantity = models.PositiveIntegerField()
    unit_price = models.DecimalField(max_digits=10, decimal_places=2)

    @property
    def subtotal(self):
        return self.quantity * self.unit_price
```

## Django REST Framework Views

```python
# apps/users/serializers.py
from rest_framework import serializers
from .models import User, Profile

class ProfileSerializer(serializers.ModelSerializer):
    class Meta:
        model = Profile
        fields = ["company", "website"]

class UserSerializer(serializers.ModelSerializer):
    profile = ProfileSerializer(read_only=True)

    class Meta:
        model = User
        fields = ["id", "email", "username", "bio", "profile", "created_at"]
        read_only_fields = ["id", "created_at"]

class UserCreateSerializer(serializers.ModelSerializer):
    password = serializers.CharField(write_only=True, min_length=8)

    class Meta:
        model = User
        fields = ["email", "username", "password"]

    def validate_email(self, value):
        if User.objects.filter(email=value).exists():
            raise serializers.ValidationError("Email already registered")
        return value

    def create(self, validated_data):
        return User.objects.create_user(**validated_data)
```

```python
# apps/users/views.py
from rest_framework import viewsets, permissions, status, filters
from rest_framework.decorators import action
from rest_framework.response import Response
from django_filters.rest_framework import DjangoFilterBackend
from .models import User
from .serializers import UserSerializer, UserCreateSerializer

class UserViewSet(viewsets.ModelViewSet):
    queryset = User.objects.select_related("profile").all()
    permission_classes = [permissions.IsAuthenticated]
    filter_backends = [DjangoFilterBackend, filters.SearchFilter, filters.OrderingFilter]
    search_fields = ["email", "username"]
    ordering_fields = ["created_at", "email"]
    ordering = ["-created_at"]

    def get_serializer_class(self):
        if self.action == "create":
            return UserCreateSerializer
        return UserSerializer

    def get_permissions(self):
        if self.action == "create":
            return [permissions.AllowAny()]
        return super().get_permissions()

    @action(detail=False, methods=["get"])
    def me(self, request):
        serializer = self.get_serializer(request.user)
        return Response(serializer.data)

    @action(detail=True, methods=["post"], permission_classes=[permissions.IsAdminUser])
    def deactivate(self, request, pk=None):
        user = self.get_object()
        user.is_active = False
        user.save(update_fields=["is_active"])
        return Response(status=status.HTTP_204_NO_CONTENT)
```

## QuerySet Optimization

```python
# N+1 prevention
# BAD: N+1 — each order triggers a user query
orders = Order.objects.all()
for order in orders:
    print(order.user.email)  # SQL query per iteration!

# GOOD: select_related for ForeignKey/OneToOne (SQL JOIN)
orders = Order.objects.select_related("user").all()

# GOOD: prefetch_related for ManyToMany/reverse FK (separate query)
users = User.objects.prefetch_related("orders").all()

# Optimized queryset example
orders = (
    Order.objects
    .select_related("user")                          # JOIN user table
    .prefetch_related("items")                       # Separate query for items
    .filter(status=Order.Status.PENDING)
    .only("id", "total", "status", "user__email")    # Only needed columns
    .order_by("-created_at")[:20]
)

# Aggregation without loading objects
from django.db.models import Count, Sum, Avg

stats = Order.objects.aggregate(
    total_orders=Count("id"),
    revenue=Sum("total"),
    avg_order=Avg("total"),
)

# Exists check — don't load the object
if Order.objects.filter(user=user, status="pending").exists():
    raise ValidationError("Existing pending order")

# Bulk operations
Order.objects.filter(status="pending", created_at__lt=threshold).update(status="cancelled")
OrderItem.objects.bulk_create([OrderItem(order=order, **item) for item in items])
```

## Custom Managers

```python
class OrderQuerySet(models.QuerySet):
    def pending(self):
        return self.filter(status=Order.Status.PENDING)

    def for_user(self, user):
        return self.filter(user=user)

    def with_items(self):
        return self.prefetch_related("items")

    def total_revenue(self):
        return self.aggregate(total=Sum("total"))["total"] or 0

class OrderManager(models.Manager):
    def get_queryset(self):
        return OrderQuerySet(self.model, using=self._db)

    def pending(self):
        return self.get_queryset().pending()

# Model
class Order(TimeStampedModel):
    objects = OrderManager()
    # ...

# Usage
pending = Order.objects.pending().for_user(user).with_items()
revenue = Order.objects.filter(created_at__year=2026).total_revenue()
```

## Settings

```python
# config/settings/production.py
from .base import *

DEBUG = False
ALLOWED_HOSTS = env.list("ALLOWED_HOSTS")
SECRET_KEY = env("SECRET_KEY")

DATABASES = {
    "default": {
        "ENGINE": "django.db.backends.postgresql",
        "NAME": env("DB_NAME"),
        "USER": env("DB_USER"),
        "PASSWORD": env("DB_PASSWORD"),
        "HOST": env("DB_HOST"),
        "PORT": env.int("DB_PORT", 5432),
        "CONN_MAX_AGE": 600,            # Connection pooling
        "CONN_HEALTH_CHECKS": True,
        "OPTIONS": {"sslmode": "require"},
    }
}

# Cache
CACHES = {
    "default": {
        "BACKEND": "django.core.cache.backends.redis.RedisCache",
        "LOCATION": env("REDIS_URL"),
        "TIMEOUT": 300,
    }
}

# Security
SECURE_SSL_REDIRECT = True
SECURE_HSTS_SECONDS = 31536000
SESSION_COOKIE_SECURE = True
CSRF_COOKIE_SECURE = True
SECURE_BROWSER_XSS_FILTER = True
```

## Testing

```python
# apps/users/tests/factories.py
import factory
from apps.users.models import User

class UserFactory(factory.django.DjangoModelFactory):
    class Meta:
        model = User
    username = factory.Sequence(lambda n: f"user{n}")
    email = factory.LazyAttribute(lambda o: f"{o.username}@test.com")
    password = factory.PostGenerationMethodCall("set_password", "testpass123")

# apps/users/tests/test_views.py
from rest_framework.test import APITestCase
from rest_framework import status

class UserViewSetTest(APITestCase):
    def setUp(self):
        self.user = UserFactory()
        self.client.force_authenticate(user=self.user)

    def test_list_users(self):
        UserFactory.create_batch(5)
        response = self.client.get("/api/users/")
        self.assertEqual(response.status_code, status.HTTP_200_OK)
        self.assertEqual(len(response.data["results"]), 6)  # 5 + setUp user

    def test_create_user_unauthenticated(self):
        self.client.force_authenticate(user=None)
        response = self.client.post("/api/users/", {
            "email": "new@test.com", "username": "newuser", "password": "securepass123",
        })
        self.assertEqual(response.status_code, status.HTTP_201_CREATED)
        self.assertNotIn("password", response.data)

    def test_create_duplicate_email(self):
        response = self.client.post("/api/users/", {
            "email": self.user.email, "username": "other", "password": "securepass123",
        })
        self.assertEqual(response.status_code, status.HTTP_400_BAD_REQUEST)

    def test_me_endpoint(self):
        response = self.client.get("/api/users/me/")
        self.assertEqual(response.data["email"], self.user.email)
```

## Gotchas

1. **`select_related` vs `prefetch_related`** — Use `select_related` for ForeignKey/OneToOne (generates a SQL JOIN). Use `prefetch_related` for ManyToMany and reverse ForeignKey (separate query). Using the wrong one either generates extra queries or cartesian product issues.

2. **`queryset.count()` vs `len(queryset)`** — `count()` runs `SELECT COUNT(*)` in SQL. `len()` loads ALL objects into memory then counts. For checking existence, use `exists()` which is even cheaper than `count()`.

3. **`auto_now_add` prevents manual setting** — Fields with `auto_now_add=True` ignore any value you pass during creation. If you need to set timestamps manually (e.g., data import), use `default=timezone.now` instead.

4. **Signals are implicit and hard to test** — `post_save` signals create hidden side effects. Prefer explicit method calls in views/services. If you must use signals, document them and test the signal handler independently.

5. **`CONN_MAX_AGE` without `CONN_HEALTH_CHECKS`** — Setting `CONN_MAX_AGE` reuses database connections across requests. Without `CONN_HEALTH_CHECKS=True` (Django 4.1+), a closed connection causes 500 errors until the stale connection is recycled.

6. **Migrations in production** — Never use `RunPython` with `apps.get_model` and `objects.all()` for large tables — it loads everything into memory. Use raw SQL for data migrations on large tables: `migrations.RunSQL("UPDATE ...")`.
