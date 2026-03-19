# Django Patterns Reference

Quick-reference guide for Django 5.x patterns including models, views, forms, Django REST Framework serializers, signals, and middleware. Consult this when building or reviewing Django applications.

---

## Model Patterns

### Abstract Base Models

```python
# models/base.py
import uuid
from django.db import models
from django.utils import timezone


class TimeStampedModel(models.Model):
    """Abstract base model providing created_at and updated_at fields."""

    created_at = models.DateTimeField(auto_now_add=True, db_index=True)
    updated_at = models.DateTimeField(auto_now=True)

    class Meta:
        abstract = True


class UUIDModel(models.Model):
    """Abstract base model using UUID as the primary key."""

    id = models.UUIDField(primary_key=True, default=uuid.uuid4, editable=False)

    class Meta:
        abstract = True


class SoftDeleteManager(models.Manager):
    def get_queryset(self) -> models.QuerySet:
        return super().get_queryset().filter(deleted_at__isnull=True)


class AllObjectsManager(models.Manager):
    def get_queryset(self) -> models.QuerySet:
        return super().get_queryset()


class SoftDeleteModel(models.Model):
    """Abstract base model providing soft delete functionality."""

    deleted_at = models.DateTimeField(null=True, blank=True, db_index=True)

    objects = SoftDeleteManager()
    all_objects = AllObjectsManager()

    class Meta:
        abstract = True

    def delete(self, using=None, keep_parents=False):
        self.deleted_at = timezone.now()
        self.save(update_fields=["deleted_at"])

    def hard_delete(self, using=None, keep_parents=False):
        return super().delete(using=using, keep_parents=keep_parents)

    def restore(self) -> None:
        self.deleted_at = None
        self.save(update_fields=["deleted_at"])

    @property
    def is_deleted(self) -> bool:
        return self.deleted_at is not None


class BaseModel(UUIDModel, TimeStampedModel, SoftDeleteModel):
    """Composite abstract base combining UUID, timestamps, and soft delete."""

    class Meta:
        abstract = True
```

### Custom Managers and QuerySets

```python
# models/managers.py
from django.db import models
from django.db.models import QuerySet
from django.utils import timezone


class ArticleQuerySet(models.QuerySet):
    def published(self) -> QuerySet:
        return self.filter(status=Article.Status.PUBLISHED, published_at__lte=timezone.now())

    def drafts(self) -> QuerySet:
        return self.filter(status=Article.Status.DRAFT)

    def by_author(self, author_id: int) -> QuerySet:
        return self.filter(author_id=author_id)

    def with_comment_count(self) -> QuerySet:
        from django.db.models import Count
        return self.annotate(comment_count=Count("comments"))

    def recent(self, days: int = 30) -> QuerySet:
        cutoff = timezone.now() - timezone.timedelta(days=days)
        return self.filter(created_at__gte=cutoff)


class ArticleManager(models.Manager):
    def get_queryset(self) -> ArticleQuerySet:
        return ArticleQuerySet(self.model, using=self._db)

    def published(self) -> ArticleQuerySet:
        return self.get_queryset().published()

    def drafts(self) -> ArticleQuerySet:
        return self.get_queryset().drafts()
```

### Model Field Choices with TextChoices/IntegerChoices

```python
# models/choices.py
from django.db import models


class Article(models.Model):
    class Status(models.TextChoices):
        DRAFT = "draft", "Draft"
        REVIEW = "review", "In Review"
        PUBLISHED = "published", "Published"
        ARCHIVED = "archived", "Archived"

    class Priority(models.IntegerChoices):
        LOW = 1, "Low"
        MEDIUM = 2, "Medium"
        HIGH = 3, "High"
        CRITICAL = 4, "Critical"

    title = models.CharField(max_length=255)
    status = models.CharField(
        max_length=20,
        choices=Status.choices,
        default=Status.DRAFT,
        db_index=True,
    )
    priority = models.IntegerField(
        choices=Priority.choices,
        default=Priority.MEDIUM,
    )

    objects = ArticleManager()

    def is_published(self) -> bool:
        return self.status == self.Status.PUBLISHED
```

### Property Decorators and cached_property

```python
# models/properties.py
from functools import cached_property
from django.db import models


class UserProfile(models.Model):
    user = models.OneToOneField("auth.User", on_delete=models.CASCADE, related_name="profile")
    birth_date = models.DateField(null=True, blank=True)
    bio = models.TextField(blank=True)

    @property
    def full_name(self) -> str:
        return f"{self.user.first_name} {self.user.last_name}".strip()

    @cached_property
    def post_count(self) -> int:
        # cached_property avoids repeated DB queries within the same request
        return self.user.posts.count()

    @cached_property
    def follower_count(self) -> int:
        return self.followers.count()
```

### Model Meta Options

```python
# models/meta_example.py
from django.db import models


class Product(models.Model):
    sku = models.CharField(max_length=50)
    name = models.CharField(max_length=255)
    category = models.ForeignKey("Category", on_delete=models.PROTECT, related_name="products")
    price = models.DecimalField(max_digits=10, decimal_places=2)
    stock = models.IntegerField(default=0)

    class Meta:
        ordering = ["category", "name"]
        verbose_name = "Product"
        verbose_name_plural = "Products"
        # Unique constraint across multiple fields
        unique_together = [("sku", "category")]
        # Django 4+ constraints
        constraints = [
            models.UniqueConstraint(fields=["sku", "category"], name="unique_sku_per_category"),
            models.CheckConstraint(check=models.Q(price__gte=0), name="price_non_negative"),
            models.CheckConstraint(check=models.Q(stock__gte=0), name="stock_non_negative"),
        ]
        indexes = [
            models.Index(fields=["category", "price"], name="product_category_price_idx"),
            models.Index(fields=["-created_at"], name="product_recent_idx"),
        ]
```

### Custom save() and clean() Methods

```python
# models/custom_save.py
from django.core.exceptions import ValidationError
from django.db import models
from django.utils.text import slugify


class BlogPost(models.Model):
    title = models.CharField(max_length=255)
    slug = models.SlugField(max_length=255, unique=True, blank=True)
    body = models.TextField()
    published_at = models.DateTimeField(null=True, blank=True)
    status = models.CharField(max_length=20, choices=[("draft", "Draft"), ("published", "Published")])

    def clean(self) -> None:
        super().clean()
        if self.status == "published" and not self.published_at:
            raise ValidationError({"published_at": "Published posts must have a published_at date."})
        if self.published_at and self.status == "draft":
            raise ValidationError({"status": "Cannot set published_at on a draft post."})

    def save(self, *args, **kwargs) -> None:
        if not self.slug:
            base_slug = slugify(self.title)
            slug = base_slug
            counter = 1
            while BlogPost.objects.filter(slug=slug).exclude(pk=self.pk).exists():
                slug = f"{base_slug}-{counter}"
                counter += 1
            self.slug = slug
        self.full_clean()
        super().save(*args, **kwargs)
```

### F Expressions and Q Objects

```python
# queries/f_and_q.py
from django.db import models
from django.db.models import F, Q


# F expressions — reference field values in DB operations (avoids race conditions)
Product.objects.filter(stock__gt=0).update(stock=F("stock") - 1)
Product.objects.update(price=F("price") * 1.1)  # 10% price increase

# Annotate with F expression arithmetic
from django.db.models import ExpressionWrapper, DecimalField
Product.objects.annotate(
    discounted_price=ExpressionWrapper(
        F("price") * 0.9,
        output_field=DecimalField(max_digits=10, decimal_places=2),
    )
)

# Q objects — complex boolean filter logic
active_discounted = Product.objects.filter(
    Q(stock__gt=0) & (Q(price__lt=50) | Q(category__name="Sale"))
)

# Negation with ~Q
not_out_of_stock = Product.objects.filter(~Q(stock=0))

# OR queries
results = Article.objects.filter(
    Q(title__icontains="django") | Q(body__icontains="django")
)
```

### Aggregation and Annotation Patterns

```python
# queries/aggregations.py
from django.db.models import Avg, Count, Max, Min, Sum, F, Value
from django.db.models.functions import Coalesce, TruncMonth


# Basic aggregation
stats = Order.objects.aggregate(
    total_revenue=Sum("total"),
    avg_order=Avg("total"),
    order_count=Count("id"),
    max_order=Max("total"),
)

# Annotation (per-row computation)
Category.objects.annotate(
    product_count=Count("products"),
    avg_price=Avg("products__price"),
    total_stock=Sum("products__stock"),
).filter(product_count__gt=5).order_by("-product_count")

# Group by month
Order.objects.annotate(
    month=TruncMonth("created_at")
).values("month").annotate(
    revenue=Sum("total"),
    count=Count("id"),
).order_by("month")

# Coalesce — fallback if null
UserProfile.objects.annotate(
    display_name=Coalesce(F("nickname"), F("user__first_name"), Value("Anonymous"))
)
```

### Subquery and OuterRef

```python
# queries/subquery.py
from django.db.models import OuterRef, Subquery, Exists


# Latest order total for each user
latest_order = Order.objects.filter(
    user=OuterRef("pk")
).order_by("-created_at").values("total")[:1]

users_with_latest_order = User.objects.annotate(
    latest_order_total=Subquery(latest_order)
)

# Exists subquery — more efficient than annotate + count
has_recent_order = Order.objects.filter(
    user=OuterRef("pk"),
    created_at__gte=thirty_days_ago,
)
active_users = User.objects.filter(Exists(has_recent_order))
```

### Signal Receivers

```python
# signals/receivers.py
from django.db import models
from django.db.models.signals import post_save, pre_delete, m2m_changed
from django.dispatch import receiver


@receiver(post_save, sender=User)
def create_user_profile(sender, instance: User, created: bool, **kwargs) -> None:
    if created:
        UserProfile.objects.create(user=instance)


@receiver(pre_delete, sender=Order)
def restore_inventory_on_order_delete(sender, instance: Order, **kwargs) -> None:
    for item in instance.items.all():
        Product.objects.filter(pk=item.product_id).update(
            stock=F("stock") + item.quantity
        )


@receiver(m2m_changed, sender=Article.tags.through)
def update_tag_counts(sender, instance: Article, action: str, pk_set, **kwargs) -> None:
    if action in ("post_add", "post_remove", "post_clear"):
        Tag.objects.filter(pk__in=(pk_set or [])).update(
            article_count=Count("articles")
        )
```

---

## View Patterns

### Generic Class-Based Views

```python
# views/cbv.py
from django.contrib.auth.mixins import LoginRequiredMixin, PermissionRequiredMixin
from django.urls import reverse_lazy
from django.views.generic import (
    CreateView, DeleteView, DetailView, ListView, UpdateView,
)

from .models import Article
from .forms import ArticleForm


class ArticleListView(LoginRequiredMixin, ListView):
    model = Article
    template_name = "articles/list.html"
    context_object_name = "articles"
    paginate_by = 20

    def get_queryset(self):
        qs = Article.objects.published().select_related("author")
        search = self.request.GET.get("q")
        if search:
            qs = qs.filter(title__icontains=search)
        return qs

    def get_context_data(self, **kwargs):
        context = super().get_context_data(**kwargs)
        context["search_query"] = self.request.GET.get("q", "")
        return context


class ArticleDetailView(DetailView):
    model = Article
    template_name = "articles/detail.html"
    context_object_name = "article"
    slug_url_kwarg = "slug"

    def get_queryset(self):
        return Article.objects.select_related("author").prefetch_related("tags", "comments__author")


class ArticleCreateView(LoginRequiredMixin, PermissionRequiredMixin, CreateView):
    model = Article
    form_class = ArticleForm
    template_name = "articles/form.html"
    permission_required = "articles.add_article"

    def form_valid(self, form):
        form.instance.author = self.request.user
        return super().form_valid(form)

    def get_success_url(self):
        return reverse_lazy("articles:detail", kwargs={"slug": self.object.slug})


class ArticleUpdateView(LoginRequiredMixin, UpdateView):
    model = Article
    form_class = ArticleForm
    template_name = "articles/form.html"

    def get_queryset(self):
        # Users can only edit their own articles
        return Article.objects.filter(author=self.request.user)

    def get_success_url(self):
        return reverse_lazy("articles:detail", kwargs={"slug": self.object.slug})


class ArticleDeleteView(LoginRequiredMixin, DeleteView):
    model = Article
    template_name = "articles/confirm_delete.html"
    success_url = reverse_lazy("articles:list")

    def get_queryset(self):
        return Article.objects.filter(author=self.request.user)
```

### Custom Mixins: Pagination, Filtering, Sorting

```python
# views/mixins.py
from typing import Any
from django.core.paginator import Paginator
from django.db.models import QuerySet
from django.views.generic import ListView


class FilterMixin:
    """Mixin that applies GET parameter filtering to a queryset."""

    filter_fields: list[str] = []

    def get_queryset(self) -> QuerySet:
        qs = super().get_queryset()
        for field in self.filter_fields:
            value = self.request.GET.get(field)
            if value:
                qs = qs.filter(**{field: value})
        return qs


class SortMixin:
    """Mixin that allows queryset sorting via ?sort= GET parameter."""

    default_sort: str = "-created_at"
    allowed_sort_fields: list[str] = []

    def get_queryset(self) -> QuerySet:
        qs = super().get_queryset()
        sort = self.request.GET.get("sort", self.default_sort)
        field = sort.lstrip("-")
        if field in self.allowed_sort_fields:
            qs = qs.order_by(sort)
        else:
            qs = qs.order_by(self.default_sort)
        return qs

    def get_context_data(self, **kwargs) -> dict[str, Any]:
        context = super().get_context_data(**kwargs)
        context["current_sort"] = self.request.GET.get("sort", self.default_sort)
        return context


class OwnerFilterMixin:
    """Restrict queryset to objects owned by the current user."""

    owner_field: str = "user"

    def get_queryset(self) -> QuerySet:
        qs = super().get_queryset()
        return qs.filter(**{self.owner_field: self.request.user})
```

### Function-Based Views with Decorators

```python
# views/fbv.py
from django.contrib.auth.decorators import login_required, permission_required
from django.http import HttpRequest, HttpResponse, JsonResponse
from django.shortcuts import get_object_or_404, redirect, render
from django.views.decorators.http import require_http_methods, require_POST


@login_required
@require_http_methods(["GET", "POST"])
def article_publish(request: HttpRequest, pk: int) -> HttpResponse:
    article = get_object_or_404(Article, pk=pk, author=request.user)

    if request.method == "POST":
        article.status = Article.Status.PUBLISHED
        article.published_at = timezone.now()
        article.save(update_fields=["status", "published_at"])
        messages.success(request, f"'{article.title}' has been published.")
        return redirect("articles:detail", slug=article.slug)

    return render(request, "articles/publish_confirm.html", {"article": article})


@login_required
@require_POST
def toggle_like(request: HttpRequest, pk: int) -> JsonResponse:
    article = get_object_or_404(Article, pk=pk)
    like, created = Like.objects.get_or_create(article=article, user=request.user)
    if not created:
        like.delete()
        liked = False
    else:
        liked = True
    return JsonResponse({"liked": liked, "count": article.likes.count()})


@permission_required("articles.change_article", raise_exception=True)
def bulk_archive(request: HttpRequest) -> HttpResponse:
    if request.method == "POST":
        ids = request.POST.getlist("article_ids")
        Article.objects.filter(pk__in=ids).update(status=Article.Status.ARCHIVED)
        return redirect("articles:list")
    return render(request, "articles/bulk_archive.html")
```

### Redirect Patterns

```python
# views/redirects.py
from django.http import HttpRequest, HttpResponse
from django.shortcuts import redirect
from django.urls import reverse
from django.views.generic.base import RedirectView


# Simple redirect in FBV
def old_article_url(request: HttpRequest, pk: int) -> HttpResponse:
    article = get_object_or_404(Article, pk=pk)
    return redirect("articles:detail", slug=article.slug, permanent=True)


# Redirect view for URL migrations
class LegacyArticleRedirectView(RedirectView):
    permanent = True

    def get_redirect_url(self, *args, **kwargs) -> str:
        article = get_object_or_404(Article, pk=kwargs["pk"])
        return reverse("articles:detail", kwargs={"slug": article.slug})


# Conditional redirect with next parameter
@login_required
def profile_required_view(request: HttpRequest) -> HttpResponse:
    if not hasattr(request.user, "profile"):
        return redirect(f"{reverse('profile:create')}?next={request.path}")
    return render(request, "dashboard.html")
```

### get_queryset, get_context_data, form_valid Overrides

```python
# views/overrides.py
from django.contrib import messages
from django.views.generic import CreateView, ListView
from django.utils import timezone


class DashboardListView(ListView):
    model = Article
    template_name = "dashboard/articles.html"
    paginate_by = 25

    def get_queryset(self):
        return (
            Article.objects.filter(author=self.request.user)
            .select_related("category")
            .prefetch_related("tags")
            .annotate(comment_count=Count("comments"))
            .order_by("-updated_at")
        )

    def get_context_data(self, **kwargs):
        context = super().get_context_data(**kwargs)
        context["draft_count"] = Article.objects.filter(
            author=self.request.user, status=Article.Status.DRAFT
        ).count()
        context["published_count"] = Article.objects.filter(
            author=self.request.user, status=Article.Status.PUBLISHED
        ).count()
        context["now"] = timezone.now()
        return context


class CommentCreateView(LoginRequiredMixin, CreateView):
    model = Comment
    fields = ["body"]

    def form_valid(self, form):
        form.instance.author = self.request.user
        form.instance.article = get_object_or_404(Article, slug=self.kwargs["slug"])
        response = super().form_valid(form)
        messages.success(self.request, "Comment posted successfully.")
        return response

    def get_success_url(self):
        return self.object.article.get_absolute_url()
```

---

## Form Patterns

### ModelForm Customization

```python
# forms/model_forms.py
from django import forms
from django.core.validators import FileExtensionValidator

from .models import Article, UserProfile


class ArticleForm(forms.ModelForm):
    # Override widget or add extra attributes
    body = forms.CharField(
        widget=forms.Textarea(attrs={"rows": 20, "class": "prose-editor"}),
        help_text="Supports Markdown formatting.",
    )
    tags = forms.CharField(
        required=False,
        help_text="Comma-separated list of tags.",
        widget=forms.TextInput(attrs={"placeholder": "django, python, web"}),
    )

    class Meta:
        model = Article
        fields = ["title", "body", "category", "status", "tags"]
        widgets = {
            "title": forms.TextInput(attrs={"autofocus": True}),
            "status": forms.RadioSelect(),
        }
        labels = {
            "body": "Article Content",
        }

    def __init__(self, *args, user=None, **kwargs):
        super().__init__(*args, **kwargs)
        self.user = user
        # Filter category choices based on user's organization
        if user and hasattr(user, "organization"):
            self.fields["category"].queryset = Category.objects.filter(
                organization=user.organization
            )
```

### Form Validation

```python
# forms/validation.py
from django import forms
from django.core.exceptions import ValidationError


class RegistrationForm(forms.Form):
    username = forms.CharField(max_length=150)
    email = forms.EmailField()
    password = forms.CharField(widget=forms.PasswordInput)
    password_confirm = forms.CharField(widget=forms.PasswordInput)
    age = forms.IntegerField()

    def clean_username(self) -> str:
        username = self.cleaned_data["username"]
        if User.objects.filter(username=username).exists():
            raise ValidationError("This username is already taken.")
        if len(username) < 3:
            raise ValidationError("Username must be at least 3 characters.")
        return username.lower()

    def clean_age(self) -> int:
        age = self.cleaned_data["age"]
        if age < 13:
            raise ValidationError("You must be at least 13 years old to register.")
        return age

    def clean(self) -> dict:
        cleaned_data = super().clean()
        password = cleaned_data.get("password")
        password_confirm = cleaned_data.get("password_confirm")
        if password and password_confirm and password != password_confirm:
            self.add_error("password_confirm", "Passwords do not match.")
        return cleaned_data
```

### Formsets and Inline Formsets

```python
# forms/formsets.py
from django.forms import inlineformset_factory, modelformset_factory
from .models import Order, OrderItem, Product


# Inline formset — OrderItems inline within an Order
OrderItemFormSet = inlineformset_factory(
    parent_model=Order,
    model=OrderItem,
    fields=["product", "quantity", "unit_price"],
    extra=3,
    can_delete=True,
    min_num=1,
    validate_min=True,
)


# Usage in a view
class OrderCreateView(LoginRequiredMixin, CreateView):
    model = Order
    fields = ["customer", "notes"]
    template_name = "orders/create.html"

    def get_context_data(self, **kwargs):
        context = super().get_context_data(**kwargs)
        if self.request.POST:
            context["item_formset"] = OrderItemFormSet(self.request.POST)
        else:
            context["item_formset"] = OrderItemFormSet()
        return context

    def form_valid(self, form):
        context = self.get_context_data()
        item_formset = context["item_formset"]
        if item_formset.is_valid():
            self.object = form.save()
            item_formset.instance = self.object
            item_formset.save()
            return redirect(self.get_success_url())
        return self.render_to_response(self.get_context_data(form=form))
```

### Dynamic Form Fields

```python
# forms/dynamic.py
from django import forms


class DynamicCategoryForm(forms.Form):
    """Form that adds subcategory field only when a parent category is selected."""

    category = forms.ModelChoiceField(queryset=Category.objects.filter(parent__isnull=True))

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        # When bound with data, check if we should add subcategory field
        if "category" in self.data:
            try:
                category_id = int(self.data.get("category"))
                self.fields["subcategory"] = forms.ModelChoiceField(
                    queryset=Category.objects.filter(parent_id=category_id),
                    required=False,
                )
            except (ValueError, TypeError):
                pass
        elif self.initial.get("category"):
            self.fields["subcategory"] = forms.ModelChoiceField(
                queryset=Category.objects.filter(parent=self.initial["category"]),
                required=False,
            )
```

### File Upload Handling

```python
# forms/file_upload.py
from django import forms
from django.core.validators import FileExtensionValidator


class DocumentUploadForm(forms.Form):
    title = forms.CharField(max_length=255)
    file = forms.FileField(
        validators=[FileExtensionValidator(allowed_extensions=["pdf", "docx", "xlsx"])],
        help_text="Allowed formats: PDF, DOCX, XLSX. Max size: 10MB.",
    )
    thumbnail = forms.ImageField(required=False)

    def clean_file(self):
        file = self.cleaned_data.get("file")
        if file:
            if file.size > 10 * 1024 * 1024:  # 10MB
                raise forms.ValidationError("File size must not exceed 10MB.")
        return file


# In the view
def upload_document(request: HttpRequest) -> HttpResponse:
    if request.method == "POST":
        form = DocumentUploadForm(request.POST, request.FILES)
        if form.is_valid():
            document = Document(
                title=form.cleaned_data["title"],
                file=form.cleaned_data["file"],
                uploaded_by=request.user,
            )
            document.save()
            return redirect("documents:detail", pk=document.pk)
    else:
        form = DocumentUploadForm()
    return render(request, "documents/upload.html", {"form": form})
```

### Custom Form Widgets

```python
# forms/widgets.py
from django import forms


class StarRatingWidget(forms.RadioSelect):
    """Custom widget rendering a 1-5 star rating selector."""

    template_name = "widgets/star_rating.html"

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.choices = [(i, f"{i} star{'s' if i != 1 else ''}") for i in range(1, 6)]


class TagInputWidget(forms.TextInput):
    """Widget that renders a tag input with autocomplete."""

    class Media:
        css = {"all": ("css/tagify.min.css",)}
        js = ("js/tagify.min.js", "js/tag-input-init.js")

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.attrs.setdefault("class", "tag-input")
```

---

## DRF Serializer Patterns

### ModelSerializer with Extra Fields

```python
# serializers/model_serializer.py
from rest_framework import serializers
from .models import Article, UserProfile


class ArticleSerializer(serializers.ModelSerializer):
    author_name = serializers.SerializerMethodField()
    tag_names = serializers.SerializerMethodField()
    is_liked = serializers.SerializerMethodField()
    comment_count = serializers.IntegerField(read_only=True)  # populated via annotation

    class Meta:
        model = Article
        fields = [
            "id", "title", "slug", "body", "status",
            "author", "author_name", "tag_names",
            "is_liked", "comment_count",
            "created_at", "updated_at", "published_at",
        ]
        read_only_fields = ["id", "slug", "author", "created_at", "updated_at"]

    def get_author_name(self, obj: Article) -> str:
        return obj.author.get_full_name() or obj.author.username

    def get_tag_names(self, obj: Article) -> list[str]:
        return list(obj.tags.values_list("name", flat=True))

    def get_is_liked(self, obj: Article) -> bool:
        request = self.context.get("request")
        if request and request.user.is_authenticated:
            return obj.likes.filter(user=request.user).exists()
        return False
```

### Nested Serializers (Read/Write)

```python
# serializers/nested.py
from rest_framework import serializers
from .models import Order, OrderItem, Product


class ProductMinimalSerializer(serializers.ModelSerializer):
    class Meta:
        model = Product
        fields = ["id", "name", "sku", "price"]


class OrderItemReadSerializer(serializers.ModelSerializer):
    product = ProductMinimalSerializer(read_only=True)

    class Meta:
        model = OrderItem
        fields = ["id", "product", "quantity", "unit_price", "line_total"]


class OrderReadSerializer(serializers.ModelSerializer):
    items = OrderItemReadSerializer(many=True, read_only=True)

    class Meta:
        model = Order
        fields = ["id", "customer", "status", "total", "items", "created_at"]
```

### Writable Nested Serializers with create/update

```python
# serializers/writable_nested.py
from django.db import transaction
from rest_framework import serializers
from .models import Order, OrderItem


class OrderItemWriteSerializer(serializers.ModelSerializer):
    class Meta:
        model = OrderItem
        fields = ["product", "quantity", "unit_price"]


class OrderWriteSerializer(serializers.ModelSerializer):
    items = OrderItemWriteSerializer(many=True)

    class Meta:
        model = Order
        fields = ["customer", "notes", "items"]

    def validate_items(self, value: list) -> list:
        if not value:
            raise serializers.ValidationError("An order must have at least one item.")
        return value

    @transaction.atomic
    def create(self, validated_data: dict) -> Order:
        items_data = validated_data.pop("items")
        order = Order.objects.create(**validated_data)
        for item_data in items_data:
            OrderItem.objects.create(order=order, **item_data)
        return order

    @transaction.atomic
    def update(self, instance: Order, validated_data: dict) -> Order:
        items_data = validated_data.pop("items", None)
        for attr, value in validated_data.items():
            setattr(instance, attr, value)
        instance.save()
        if items_data is not None:
            instance.items.all().delete()
            for item_data in items_data:
                OrderItem.objects.create(order=instance, **item_data)
        return instance
```

### SerializerMethodField and Custom Validation

```python
# serializers/custom_validation.py
from rest_framework import serializers
from django.utils import timezone


class EventSerializer(serializers.ModelSerializer):
    duration_minutes = serializers.SerializerMethodField()
    is_upcoming = serializers.SerializerMethodField()

    class Meta:
        model = Event
        fields = [
            "id", "title", "start_time", "end_time",
            "duration_minutes", "is_upcoming", "capacity", "registered_count",
        ]

    def get_duration_minutes(self, obj: Event) -> int | None:
        if obj.start_time and obj.end_time:
            delta = obj.end_time - obj.start_time
            return int(delta.total_seconds() / 60)
        return None

    def get_is_upcoming(self, obj: Event) -> bool:
        return obj.start_time > timezone.now()

    def validate_end_time(self, value):
        start_time = self.initial_data.get("start_time") or getattr(self.instance, "start_time", None)
        if start_time and value and value <= start_time:
            raise serializers.ValidationError("End time must be after start time.")
        return value

    def validate(self, attrs: dict) -> dict:
        if attrs.get("capacity") and attrs.get("registered_count", 0) > attrs["capacity"]:
            raise serializers.ValidationError("Registered count cannot exceed capacity.")
        return attrs
```

### SlugRelatedField and PrimaryKeyRelatedField

```python
# serializers/related_fields.py
from rest_framework import serializers
from .models import Article, Tag, Category


class ArticleRelatedFieldSerializer(serializers.ModelSerializer):
    # Represent tags by slug in read and write
    tags = serializers.SlugRelatedField(
        many=True,
        slug_field="slug",
        queryset=Tag.objects.all(),
    )
    # Represent category by PK in write, nested in read
    category = serializers.PrimaryKeyRelatedField(queryset=Category.objects.all())

    class Meta:
        model = Article
        fields = ["id", "title", "tags", "category"]
```

### Dynamic Field Serializers

```python
# serializers/dynamic_fields.py
from rest_framework import serializers


class DynamicFieldsModelSerializer(serializers.ModelSerializer):
    """Serializer that accepts a `fields` parameter to restrict output."""

    def __init__(self, *args, **kwargs):
        fields = kwargs.pop("fields", None)
        exclude = kwargs.pop("exclude", None)
        super().__init__(*args, **kwargs)

        if fields is not None:
            allowed = set(fields)
            existing = set(self.fields)
            for field_name in existing - allowed:
                self.fields.pop(field_name)

        if exclude is not None:
            for field_name in exclude:
                self.fields.pop(field_name, None)


class ArticleDynamicSerializer(DynamicFieldsModelSerializer):
    class Meta:
        model = Article
        fields = ["id", "title", "slug", "body", "status", "author", "created_at"]


# Usage:
# ArticleDynamicSerializer(article, fields=["id", "title", "slug"])
# ArticleDynamicSerializer(articles, many=True, exclude=["body"])
```

### Serializer Inheritance

```python
# serializers/inheritance.py
from rest_framework import serializers
from .models import Article


class ArticleListSerializer(serializers.ModelSerializer):
    """Lightweight serializer for list endpoints."""
    class Meta:
        model = Article
        fields = ["id", "title", "slug", "status", "created_at"]


class ArticleDetailSerializer(ArticleListSerializer):
    """Full serializer for detail endpoints — extends list serializer."""
    author_name = serializers.SerializerMethodField()
    tags = serializers.SlugRelatedField(many=True, read_only=True, slug_field="name")

    class Meta(ArticleListSerializer.Meta):
        fields = ArticleListSerializer.Meta.fields + [
            "body", "author_name", "tags", "updated_at", "published_at"
        ]

    def get_author_name(self, obj: Article) -> str:
        return obj.author.get_full_name()
```

### Pagination Classes

```python
# pagination.py
from rest_framework.pagination import CursorPagination, PageNumberPagination
from rest_framework.response import Response


class StandardPageNumberPagination(PageNumberPagination):
    page_size = 20
    page_size_query_param = "page_size"
    max_page_size = 100

    def get_paginated_response(self, data):
        return Response({
            "count": self.page.paginator.count,
            "next": self.get_next_link(),
            "previous": self.get_previous_link(),
            "total_pages": self.page.paginator.num_pages,
            "current_page": self.page.number,
            "results": data,
        })


class ArticleCursorPagination(CursorPagination):
    """Cursor-based pagination for large, frequently-updated datasets."""
    page_size = 25
    ordering = "-created_at"
    cursor_query_param = "cursor"
```

### ViewSet Patterns with Custom Actions

```python
# views/viewsets.py
from rest_framework import status, viewsets
from rest_framework.decorators import action
from rest_framework.permissions import IsAuthenticated
from rest_framework.response import Response
from django_filters.rest_framework import DjangoFilterBackend
from rest_framework.filters import OrderingFilter, SearchFilter

from .models import Article
from .serializers import ArticleListSerializer, ArticleDetailSerializer
from .pagination import StandardPageNumberPagination
from .permissions import IsAuthorOrReadOnly


class ArticleViewSet(viewsets.ModelViewSet):
    queryset = Article.objects.select_related("author").prefetch_related("tags")
    permission_classes = [IsAuthenticated, IsAuthorOrReadOnly]
    pagination_class = StandardPageNumberPagination
    filter_backends = [DjangoFilterBackend, SearchFilter, OrderingFilter]
    filterset_fields = ["status", "category"]
    search_fields = ["title", "body"]
    ordering_fields = ["created_at", "published_at", "title"]
    ordering = ["-created_at"]

    def get_serializer_class(self):
        if self.action == "list":
            return ArticleListSerializer
        return ArticleDetailSerializer

    def perform_create(self, serializer):
        serializer.save(author=self.request.user)

    @action(detail=True, methods=["post"], url_path="publish")
    def publish(self, request, pk=None):
        article = self.get_object()
        if article.author != request.user:
            return Response({"detail": "Not authorized."}, status=status.HTTP_403_FORBIDDEN)
        article.status = Article.Status.PUBLISHED
        article.published_at = timezone.now()
        article.save(update_fields=["status", "published_at"])
        return Response(ArticleDetailSerializer(article).data)

    @action(detail=False, methods=["get"], url_path="my-articles")
    def my_articles(self, request):
        qs = self.get_queryset().filter(author=request.user)
        page = self.paginate_queryset(qs)
        if page is not None:
            serializer = self.get_serializer(page, many=True)
            return self.get_paginated_response(serializer.data)
        serializer = self.get_serializer(qs, many=True)
        return Response(serializer.data)

    @action(detail=True, methods=["post"], url_path="toggle-like")
    def toggle_like(self, request, pk=None):
        article = self.get_object()
        like, created = Like.objects.get_or_create(article=article, user=request.user)
        if not created:
            like.delete()
        return Response({"liked": created, "count": article.likes.count()})
```

### Filter Backends

```python
# filters.py
import django_filters
from .models import Article


class ArticleFilter(django_filters.FilterSet):
    title = django_filters.CharFilter(lookup_expr="icontains")
    status = django_filters.MultipleChoiceFilter(choices=Article.Status.choices)
    created_after = django_filters.DateTimeFilter(field_name="created_at", lookup_expr="gte")
    created_before = django_filters.DateTimeFilter(field_name="created_at", lookup_expr="lte")
    author_username = django_filters.CharFilter(field_name="author__username", lookup_expr="iexact")
    has_tags = django_filters.BooleanFilter(method="filter_has_tags")

    class Meta:
        model = Article
        fields = ["status", "category", "author"]

    def filter_has_tags(self, queryset, name, value):
        if value:
            return queryset.filter(tags__isnull=False).distinct()
        return queryset.filter(tags__isnull=True)
```

---

## Signal Patterns

### Defining Custom Signals

```python
# signals/custom.py
from django.dispatch import Signal, receiver

# Define custom signals
article_published = Signal()  # provides: article, publisher
order_completed = Signal()    # provides: order, total

# Send a custom signal
def publish_article(article, publisher):
    article.status = Article.Status.PUBLISHED
    article.published_at = timezone.now()
    article.save()
    article_published.send(sender=Article, article=article, publisher=publisher)


# Receive a custom signal
@receiver(article_published, sender=Article)
def notify_subscribers_on_publish(sender, article, publisher, **kwargs) -> None:
    subscribers = article.author.followers.select_related("user")
    for follower in subscribers:
        Notification.objects.create(
            recipient=follower.user,
            message=f"{publisher.username} published '{article.title}'",
            link=article.get_absolute_url(),
        )
```

### post_save with created Flag

```python
# signals/post_save.py
from django.db.models.signals import post_save
from django.dispatch import receiver
from django.contrib.auth.models import User


@receiver(post_save, sender=User)
def on_user_created(sender, instance: User, created: bool, **kwargs) -> None:
    if created:
        # Create related objects only on first save
        UserProfile.objects.create(user=instance)
        UserPreferences.objects.create(user=instance)
        send_welcome_email.delay(instance.pk)  # Celery task


@receiver(post_save, sender=Order)
def on_order_status_change(sender, instance: Order, created: bool, **kwargs) -> None:
    if not created and "status" in (kwargs.get("update_fields") or []):
        if instance.status == Order.Status.COMPLETED:
            order_completed.send(sender=Order, order=instance, total=instance.total)
```

### pre_save for Auto-Computed Fields

```python
# signals/pre_save.py
from django.db.models.signals import pre_save
from django.dispatch import receiver
from django.utils.text import slugify


@receiver(pre_save, sender=Article)
def auto_generate_slug(sender, instance: Article, **kwargs) -> None:
    if not instance.slug:
        base = slugify(instance.title)
        slug = base
        n = 1
        while Article.objects.filter(slug=slug).exclude(pk=instance.pk).exists():
            slug = f"{base}-{n}"
            n += 1
        instance.slug = slug


@receiver(pre_save, sender=OrderItem)
def compute_line_total(sender, instance: OrderItem, **kwargs) -> None:
    instance.line_total = instance.quantity * instance.unit_price
```

### m2m_changed for Many-to-Many Updates

```python
# signals/m2m.py
from django.db.models.signals import m2m_changed
from django.dispatch import receiver


@receiver(m2m_changed, sender=Article.tags.through)
def sync_tag_article_counts(sender, instance, action: str, pk_set, model, **kwargs) -> None:
    if action in ("post_add", "post_remove", "post_clear"):
        tag_ids = pk_set if pk_set else model.objects.filter(articles=instance).values_list("pk", flat=True)
        for tag in Tag.objects.filter(pk__in=tag_ids):
            tag.article_count = tag.articles.count()
            tag.save(update_fields=["article_count"])
```

### Avoiding Signal Recursion

```python
# signals/safe.py
from django.db.models.signals import post_save
from django.dispatch import receiver


_updating_profile = set()  # thread-local would be safer in production


@receiver(post_save, sender=UserProfile)
def on_profile_save(sender, instance: UserProfile, **kwargs) -> None:
    if instance.pk in _updating_profile:
        return  # Prevent recursion

    _updating_profile.add(instance.pk)
    try:
        # Do something that might trigger another save
        instance.compute_score()
        instance.save(update_fields=["score"])
    finally:
        _updating_profile.discard(instance.pk)
```

### Signal vs Overriding save()

```python
# Prefer overriding save() when:
# - Logic is intrinsic to the model's behavior
# - You need access to the old state (fetch before save)
# - The action must always happen (not optional)

class Article(models.Model):
    def save(self, *args, **kwargs):
        self._auto_slug()         # Intrinsic to the model
        self.updated_at = timezone.now()
        super().save(*args, **kwargs)

# Prefer signals when:
# - Decoupling app-level side effects (email, cache, notifications)
# - Cross-app dependencies (avoid circular imports)
# - The receiver lives in a different app

@receiver(post_save, sender=Article)
def invalidate_article_cache(sender, instance: Article, **kwargs) -> None:
    cache.delete(f"article:{instance.slug}")
    cache.delete("article:list")
```

---

## Middleware Patterns

### Custom Middleware Class Structure

```python
# middleware/base.py
from collections.abc import Callable
from django.http import HttpRequest, HttpResponse


class ExampleMiddleware:
    """
    Django middleware using the new-style callable pattern.
    Compatible with both WSGI and ASGI.
    """

    def __init__(self, get_response: Callable) -> None:
        self.get_response = get_response
        # One-time configuration and initialization here

    def __call__(self, request: HttpRequest) -> HttpResponse:
        # Code executed before the view (and later middleware)
        self.process_request(request)

        response = self.get_response(request)

        # Code executed after the view
        self.process_response(request, response)

        return response

    def process_request(self, request: HttpRequest) -> None:
        pass  # Override in subclasses

    def process_response(self, request: HttpRequest, response: HttpResponse) -> None:
        pass  # Override in subclasses

    def process_exception(self, request: HttpRequest, exception: Exception) -> HttpResponse | None:
        return None  # Return None to let Django handle it, or return a response
```

### Request Timing Middleware

```python
# middleware/timing.py
import time
import logging
from collections.abc import Callable
from django.http import HttpRequest, HttpResponse

logger = logging.getLogger("request.timing")


class RequestTimingMiddleware:
    SLOW_REQUEST_THRESHOLD_MS = 500

    def __init__(self, get_response: Callable) -> None:
        self.get_response = get_response

    def __call__(self, request: HttpRequest) -> HttpResponse:
        start = time.monotonic()
        response = self.get_response(request)
        duration_ms = (time.monotonic() - start) * 1000

        response["X-Request-Duration-Ms"] = f"{duration_ms:.1f}"

        if duration_ms > self.SLOW_REQUEST_THRESHOLD_MS:
            logger.warning(
                "Slow request: %s %s took %.1fms",
                request.method,
                request.path,
                duration_ms,
            )
        return response
```

### User Activity Tracking Middleware

```python
# middleware/activity.py
import logging
from collections.abc import Callable
from django.http import HttpRequest, HttpResponse
from django.utils import timezone

logger = logging.getLogger("user.activity")

SKIP_PATHS = frozenset(["/health/", "/metrics/", "/favicon.ico"])


class UserActivityMiddleware:
    """Track last-seen time for authenticated users."""

    def __init__(self, get_response: Callable) -> None:
        self.get_response = get_response

    def __call__(self, request: HttpRequest) -> HttpResponse:
        response = self.get_response(request)

        if request.path in SKIP_PATHS:
            return response

        if request.user.is_authenticated:
            # Use cache to avoid DB write on every request
            from django.core.cache import cache
            cache_key = f"user_last_seen:{request.user.pk}"
            if not cache.get(cache_key):
                UserProfile.objects.filter(user=request.user).update(
                    last_seen_at=timezone.now()
                )
                cache.set(cache_key, True, timeout=60)  # Update at most once per minute

        return response
```

### IP-Based Access Control Middleware

```python
# middleware/ip_whitelist.py
import ipaddress
from collections.abc import Callable
from django.conf import settings
from django.http import HttpRequest, HttpResponse, HttpResponseForbidden


class IPWhitelistMiddleware:
    """Allow access only from whitelisted IP ranges for admin paths."""

    def __init__(self, get_response: Callable) -> None:
        self.get_response = get_response
        self.allowed_networks = [
            ipaddress.ip_network(cidr)
            for cidr in getattr(settings, "ADMIN_ALLOWED_CIDR_RANGES", [])
        ]
        self.protected_paths = getattr(settings, "IP_PROTECTED_PATHS", ["/admin/"])

    def __call__(self, request: HttpRequest) -> HttpResponse:
        if any(request.path.startswith(path) for path in self.protected_paths):
            client_ip = self._get_client_ip(request)
            if not self._is_allowed(client_ip):
                return HttpResponseForbidden("Access denied from your IP address.")
        return self.get_response(request)

    def _get_client_ip(self, request: HttpRequest) -> str:
        x_forwarded_for = request.META.get("HTTP_X_FORWARDED_FOR")
        if x_forwarded_for:
            return x_forwarded_for.split(",")[0].strip()
        return request.META.get("REMOTE_ADDR", "")

    def _is_allowed(self, ip_str: str) -> bool:
        if not self.allowed_networks:
            return True  # No restrictions configured
        try:
            ip = ipaddress.ip_address(ip_str)
            return any(ip in network for network in self.allowed_networks)
        except ValueError:
            return False
```

### ASGI Middleware

```python
# middleware/asgi.py
from collections.abc import Callable


class ASGITimingMiddleware:
    """ASGI-compatible middleware for timing async requests."""

    def __init__(self, app: Callable) -> None:
        self.app = app

    async def __call__(self, scope, receive, send) -> None:
        import time
        if scope["type"] not in ("http", "websocket"):
            await self.app(scope, receive, send)
            return

        start = time.monotonic()

        async def send_with_timing(message):
            if message["type"] == "http.response.start":
                duration_ms = (time.monotonic() - start) * 1000
                headers = list(message.get("headers", []))
                headers.append((b"x-duration-ms", f"{duration_ms:.1f}".encode()))
                message = {**message, "headers": headers}
            await send(message)

        await self.app(scope, receive, send_with_timing)
```

### Middleware Ordering in settings.py

```python
# settings/middleware_order.py
MIDDLEWARE = [
    # Security first
    "django.middleware.security.SecurityMiddleware",
    "whitenoise.middleware.WhiteNoiseMiddleware",  # Serve static files early

    # Session and auth before custom middleware that needs request.user
    "django.contrib.sessions.middleware.SessionMiddleware",
    "django.middleware.common.CommonMiddleware",
    "django.middleware.csrf.CsrfViewMiddleware",
    "django.contrib.auth.middleware.AuthenticationMiddleware",
    "django.contrib.messages.middleware.MessageMiddleware",
    "django.middleware.clickjacking.XFrameOptionsMiddleware",

    # Custom middleware — goes after auth so request.user is available
    "myapp.middleware.ip_whitelist.IPWhitelistMiddleware",
    "myapp.middleware.timing.RequestTimingMiddleware",
    "myapp.middleware.activity.UserActivityMiddleware",
]
```

---

## URL & Routing Patterns

### Path Converters

```python
# urls/converters.py
from django.urls import path, register_converter


class FourDigitYearConverter:
    regex = r"\d{4}"

    def to_python(self, value: str) -> int:
        return int(value)

    def to_url(self, value: int) -> str:
        return f"{value:04d}"


register_converter(FourDigitYearConverter, "yyyy")


# urls/articles.py
from django.urls import path, include
import uuid

urlpatterns = [
    # int: — matches positive integers
    path("articles/<int:pk>/", views.ArticleDetailView.as_view(), name="detail"),

    # slug: — matches [a-zA-Z0-9_-]+
    path("articles/<slug:slug>/", views.ArticleBySlugView.as_view(), name="detail-slug"),

    # uuid: — matches UUID formatted strings
    path("documents/<uuid:pk>/", views.DocumentDetailView.as_view(), name="doc-detail"),

    # str: — matches any non-empty string without a slash (default)
    path("pages/<str:page_name>/", views.StaticPageView.as_view(), name="static-page"),

    # Custom converter
    path("archive/<yyyy:year>/", views.ArchiveYearView.as_view(), name="archive-year"),
]
```

### Include with Namespaces

```python
# urls/main.py
from django.urls import include, path

urlpatterns = [
    path("articles/", include("articles.urls", namespace="articles")),
    path("api/v1/", include("api.urls", namespace="api-v1")),
    path("accounts/", include("accounts.urls", namespace="accounts")),
]

# articles/urls.py
app_name = "articles"

urlpatterns = [
    path("", views.ArticleListView.as_view(), name="list"),
    path("create/", views.ArticleCreateView.as_view(), name="create"),
    path("<slug:slug>/", views.ArticleDetailView.as_view(), name="detail"),
    path("<slug:slug>/edit/", views.ArticleUpdateView.as_view(), name="edit"),
    path("<slug:slug>/delete/", views.ArticleDeleteView.as_view(), name="delete"),
]
```

### Reverse URL Resolution

```python
# Reverse in Python code
from django.urls import reverse, reverse_lazy

# With namespace
url = reverse("articles:detail", kwargs={"slug": article.slug})

# reverse_lazy — use when URL conf may not be loaded yet (class attributes, module level)
success_url = reverse_lazy("articles:list")

# In models — get_absolute_url convention
class Article(models.Model):
    slug = models.SlugField(unique=True)

    def get_absolute_url(self) -> str:
        return reverse("articles:detail", kwargs={"slug": self.slug})

# In templates
# {% url 'articles:detail' slug=article.slug %}
# {% url 'articles:list' %}
```

---

## Template Patterns

### Template Inheritance Chain

```html
{# templates/base.html — root layout #}
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>{% block title %}My Site{% endblock %}</title>
    {% block extra_css %}{% endblock %}
</head>
<body>
    {% block header %}
        {% include "partials/navbar.html" %}
    {% endblock %}

    <main>{% block content %}{% endblock %}</main>

    {% block footer %}{% include "partials/footer.html" %}{% endblock %}
    {% block extra_js %}{% endblock %}
</body>
</html>

{# templates/articles/base.html — section layout #}
{% extends "base.html" %}
{% block content %}
<div class="article-layout">
    <aside>{% block sidebar %}{% endblock %}</aside>
    <section>{% block article_content %}{% endblock %}</section>
</div>
{% endblock %}

{# templates/articles/detail.html — page template #}
{% extends "articles/base.html" %}
{% block title %}{{ article.title }} | My Site{% endblock %}
{% block article_content %}
    <h1>{{ article.title }}</h1>
    <p>By {{ article.author.get_full_name }} on {{ article.published_at|date:"N j, Y" }}</p>
    {{ article.body|linebreaks }}
{% endblock %}
```

### Custom Template Tags

```python
# templatetags/article_tags.py
from django import template
from django.utils.safestring import mark_safe

register = template.Library()


@register.simple_tag
def active_link(request, url_name: str, css_class: str = "active") -> str:
    """Return css_class if the current URL matches url_name."""
    from django.urls import reverse, NoReverseMatch
    try:
        url = reverse(url_name)
        if request.path == url:
            return css_class
    except NoReverseMatch:
        pass
    return ""


@register.simple_tag(takes_context=True)
def query_transform(context, **kwargs) -> str:
    """Update URL query parameters while preserving existing ones."""
    request = context["request"]
    updated = request.GET.copy()
    for key, value in kwargs.items():
        updated[key] = value
    return updated.urlencode()


@register.assignment_tag  # or simple_tag with 'as' in template
@register.simple_tag
def get_related_articles(article, count: int = 5):
    return Article.objects.filter(
        category=article.category
    ).exclude(pk=article.pk).order_by("-published_at")[:count]
```

### Custom Template Filters

```python
# templatetags/article_tags.py (continued)
import re
from django import template
from django.utils.html import strip_tags
from django.utils.safestring import mark_safe

register = template.Library()


@register.filter(name="truncate_words_html")
def truncate_words_html(value: str, arg: int) -> str:
    """Truncate HTML content to N words, stripping tags."""
    text = strip_tags(value)
    words = text.split()
    if len(words) > arg:
        return " ".join(words[:arg]) + "..."
    return text


@register.filter
def multiply(value, arg):
    """Multiply a value by arg. Usage: {{ price|multiply:quantity }}"""
    try:
        return float(value) * float(arg)
    except (ValueError, TypeError):
        return ""


@register.filter(name="highlight")
def highlight_search(text: str, search: str) -> str:
    """Wrap occurrences of search term in <mark> tags."""
    if not search:
        return text
    pattern = re.compile(re.escape(search), re.IGNORECASE)
    highlighted = pattern.sub(lambda m: f"<mark>{m.group()}</mark>", str(text))
    return mark_safe(highlighted)
```

### Inclusion Tags

```python
# templatetags/article_tags.py (continued)
from django import template
register = template.Library()


@register.inclusion_tag("partials/article_card.html")
def article_card(article, show_author: bool = True) -> dict:
    """Render an article preview card."""
    return {
        "article": article,
        "show_author": show_author,
    }


@register.inclusion_tag("partials/pagination.html", takes_context=True)
def pagination_controls(context, page_obj) -> dict:
    """Render pagination controls preserving existing query parameters."""
    request = context["request"]
    params = request.GET.copy()
    params.pop("page", None)
    return {
        "page_obj": page_obj,
        "request": request,
        "base_params": params.urlencode(),
    }
```

### Context Processors

```python
# context_processors.py
from django.conf import settings
from django.http import HttpRequest


def site_settings(request: HttpRequest) -> dict:
    """Inject global site configuration into all templates."""
    return {
        "SITE_NAME": settings.SITE_NAME,
        "SUPPORT_EMAIL": settings.SUPPORT_EMAIL,
        "ENVIRONMENT": settings.ENVIRONMENT,
        "DEBUG": settings.DEBUG,
    }


def notifications(request: HttpRequest) -> dict:
    """Inject unread notification count for authenticated users."""
    if not request.user.is_authenticated:
        return {"unread_notification_count": 0}

    from django.core.cache import cache
    cache_key = f"unread_notifications:{request.user.pk}"
    count = cache.get(cache_key)
    if count is None:
        count = Notification.objects.filter(
            recipient=request.user, read_at__isnull=True
        ).count()
        cache.set(cache_key, count, timeout=30)

    return {"unread_notification_count": count}


# Register in settings.py:
# TEMPLATES[0]["OPTIONS"]["context_processors"] += [
#     "myapp.context_processors.site_settings",
#     "myapp.context_processors.notifications",
# ]
```

---

## Quick Reference: Common Patterns at a Glance

### ORM Cheat Sheet

```python
# Select related (FK/OneToOne) — single JOIN
Article.objects.select_related("author", "category")

# Prefetch related (M2M / reverse FK) — separate query
Article.objects.prefetch_related("tags", "comments__author")

# Combine both
Article.objects.select_related("author").prefetch_related("tags")

# only() — fetch subset of columns
Article.objects.only("id", "title", "slug")

# defer() — exclude heavy columns (e.g. large text fields)
Article.objects.defer("body")

# values() and values_list()
Article.objects.values("id", "title")
Article.objects.values_list("id", flat=True)

# exists() — cheaper than count() for boolean check
Article.objects.filter(author=user).exists()

# update_or_create
obj, created = UserProfile.objects.update_or_create(
    user=user,
    defaults={"bio": "Updated bio"},
)

# get_or_create
obj, created = Tag.objects.get_or_create(
    slug=slugify(name),
    defaults={"name": name},
)

# bulk_create — insert many rows in one query
Article.objects.bulk_create([
    Article(title="A", author=user),
    Article(title="B", author=user),
], batch_size=500, ignore_conflicts=True)

# bulk_update — update many rows in one query
articles = list(Article.objects.filter(status="draft"))
for a in articles:
    a.status = "published"
Article.objects.bulk_update(articles, ["status"], batch_size=500)
```

### DRF Permission Classes

```python
# permissions.py
from rest_framework.permissions import BasePermission, SAFE_METHODS


class IsAuthorOrReadOnly(BasePermission):
    """Object-level permission: only the author can write."""

    def has_permission(self, request, view) -> bool:
        return request.method in SAFE_METHODS or request.user.is_authenticated

    def has_object_permission(self, request, view, obj) -> bool:
        if request.method in SAFE_METHODS:
            return True
        return obj.author == request.user


class IsOrganizationMember(BasePermission):
    """Allow access only to users who belong to the object's organization."""

    def has_object_permission(self, request, view, obj) -> bool:
        return request.user.organization == obj.organization
```

### settings.py Snippets

```python
# Common production-ready settings snippets

# Database with connection pooling (Django 5.x)
DATABASES = {
    "default": {
        "ENGINE": "django.db.backends.postgresql",
        "NAME": env("DB_NAME"),
        "USER": env("DB_USER"),
        "PASSWORD": env("DB_PASSWORD"),
        "HOST": env("DB_HOST", default="localhost"),
        "PORT": env("DB_PORT", default="5432"),
        "CONN_MAX_AGE": 60,
        "CONN_HEALTH_CHECKS": True,
        "OPTIONS": {"sslmode": "require"},
    }
}

# Cache with Redis
CACHES = {
    "default": {
        "BACKEND": "django.core.cache.backends.redis.RedisCache",
        "LOCATION": env("REDIS_URL"),
    }
}

# Celery
CELERY_BROKER_URL = env("REDIS_URL")
CELERY_RESULT_BACKEND = env("REDIS_URL")
CELERY_TASK_SERIALIZER = "json"
CELERY_ACCEPT_CONTENT = ["json"]

# DRF defaults
REST_FRAMEWORK = {
    "DEFAULT_AUTHENTICATION_CLASSES": [
        "rest_framework_simplejwt.authentication.JWTAuthentication",
    ],
    "DEFAULT_PERMISSION_CLASSES": [
        "rest_framework.permissions.IsAuthenticated",
    ],
    "DEFAULT_PAGINATION_CLASS": "myapp.pagination.StandardPageNumberPagination",
    "PAGE_SIZE": 20,
    "DEFAULT_FILTER_BACKENDS": [
        "django_filters.rest_framework.DjangoFilterBackend",
        "rest_framework.filters.SearchFilter",
        "rest_framework.filters.OrderingFilter",
    ],
    "DEFAULT_THROTTLE_CLASSES": [
        "rest_framework.throttling.AnonRateThrottle",
        "rest_framework.throttling.UserRateThrottle",
    ],
    "DEFAULT_THROTTLE_RATES": {
        "anon": "100/hour",
        "user": "1000/hour",
    },
}
```
