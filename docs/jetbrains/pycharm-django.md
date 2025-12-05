# PyCharm - Django Quick Start

Create Django applications in DevArch using PyCharm with container-based Python interpreter.

## Prerequisites

- PyCharm Professional (required for Django support)
- DevArch containers running:
  ```bash
  ./scripts/service-manager.sh start database proxy backend
  ```

## 1. Create Project

**Using django-admin in container:**

```bash
podman exec -it python bash
cd /app
django-admin startproject my_django_app
cd my_django_app
python manage.py startapp core
exit
```

**Open in PyCharm:**
1. File → Open → `/home/fhcadmin/projects/devarch/apps/my_django_app`

## 2. Configure `public/` Structure

Django serves static/media files. Configure to output to `public/`:

**Update `my_django_app/settings.py`:**

```python
import os
from pathlib import Path

BASE_DIR = Path(__file__).resolve().parent.parent

# Static files (CSS, JavaScript, Images)
STATIC_URL = '/static/'
STATIC_ROOT = BASE_DIR / 'public' / 'static'

# Media files (User uploads)
MEDIA_URL = '/media/'
MEDIA_ROOT = BASE_DIR / 'public' / 'media'

# For serving via whitenoise in production
STATICFILES_STORAGE = 'whitenoise.storage.CompressedManifestStaticFilesStorage'
```

**Directory structure:**
```
apps/my_django_app/
├── public/              # Build output for static/media
│   ├── static/          # Collected static files
│   └── media/           # User uploads
├── my_django_app/
│   ├── settings.py
│   ├── urls.py
│   └── wsgi.py
├── core/                # Django app
│   ├── views.py
│   ├── models.py
│   └── urls.py
└── manage.py
```

## 3. Configure Python Interpreter (Container)

1. Settings → Project → Python Interpreter
2. Click gear icon → "Add..."
3. Select "Docker Compose"
4. Configuration:
   - Configuration file: `/home/fhcadmin/projects/devarch/compose/backend/python.yml`
   - Service: `python`
   - Python interpreter path: `/usr/local/bin/python`
5. Click "OK"

**Wait for indexing to complete** - PyCharm will detect Django, installed packages.

## 4. Enable Django Support

1. Settings → Languages & Frameworks → Django
2. Check "Enable Django Support"
3. Configuration:
   - Django project root: `/home/fhcadmin/projects/devarch/apps/my_django_app`
   - Settings: `my_django_app/settings.py`
   - Manage script: `manage.py`
4. Click "OK"

## 5. Database Configuration

**Create PostgreSQL database:**

```bash
podman exec -it postgres bash
psql -U postgres -c "CREATE DATABASE my_django_db;"
exit
```

**Update `settings.py`:**

```python
DATABASES = {
    'default': {
        'ENGINE': 'django.db.backends.postgresql',
        'NAME': 'my_django_db',
        'USER': 'postgres',
        'PASSWORD': 'admin1234567',
        'HOST': 'postgres',
        'PORT': '5432',
    }
}
```

**Install psycopg2 (already in container):**

```bash
podman exec -it python bash
cd /app/my_django_app
pip install psycopg2-binary
# Or add to requirements.txt and pip install -r requirements.txt
```

**Run migrations:**

```bash
python manage.py makemigrations
python manage.py migrate
```

## 6. Create Superuser

```bash
podman exec -it python bash
cd /app/my_django_app
python manage.py createsuperuser
# Username: admin
# Email: admin@devarch.test
# Password: admin1234567
```

## 7. Configure Run Configurations

**Django server:**

1. Run → Edit Configurations → "+"
2. Select "Django Server"
3. Configuration:
   - Name: `runserver (container)`
   - Host: `0.0.0.0`
   - Port: `8000`
   - Python interpreter: Container interpreter
   - Environment variables:
     ```
     DJANGO_SETTINGS_MODULE=my_django_app.settings
     PYTHONUNBUFFERED=1
     ```
4. Click "OK"

**PyCharm creates this automatically if Django support enabled.**

## 8. Development Workflow

**Start Django dev server:**

```bash
podman exec -it python bash
cd /app/my_django_app
python manage.py runserver 0.0.0.0:8000
```

Or use PyCharm run configuration: Click "Run" (green play).

**Server runs on:** http://localhost:8300

**Admin panel:** http://localhost:8300/admin

## 9. Create Views and URLs

**Create view:** `core/views.py`

```python
from django.http import JsonResponse
from django.shortcuts import render

def index(request):
    return render(request, 'core/index.html')

def api_hello(request):
    return JsonResponse({'message': 'Hello from Django'})
```

**Create URLs:** `core/urls.py`

```python
from django.urls import path
from . import views

urlpatterns = [
    path('', views.index, name='index'),
    path('api/hello', views.api_hello, name='api_hello'),
]
```

**Include in project URLs:** `my_django_app/urls.py`

```python
from django.contrib import admin
from django.urls import path, include

urlpatterns = [
    path('admin/', admin.site.urls),
    path('', include('core.urls')),
]
```

**Create template:** `core/templates/core/index.html`

```html
<!DOCTYPE html>
<html>
<head>
    <title>Django App</title>
</head>
<body>
    <h1>Welcome to Django</h1>
    <p>Running in DevArch container</p>
</body>
</html>
```

## 10. Database Models

**Create model:** `core/models.py`

```python
from django.db import models

class Article(models.Model):
    title = models.CharField(max_length=200)
    content = models.TextField()
    created_at = models.DateTimeField(auto_now_add=True)
    updated_at = models.DateTimeField(auto_now=True)

    def __str__(self):
        return self.title

    class Meta:
        ordering = ['-created_at']
```

**Register in admin:** `core/admin.py`

```python
from django.contrib import admin
from .models import Article

@admin.register(Article)
class ArticleAdmin(admin.ModelAdmin):
    list_display = ('title', 'created_at', 'updated_at')
    search_fields = ('title', 'content')
```

**Run migrations:**

```bash
python manage.py makemigrations
python manage.py migrate
```

**PyCharm shortcuts:**
- Tools → Run manage.py Task → `makemigrations`
- Tools → Run manage.py Task → `migrate`

## 11. Django REST Framework (Optional)

**Install:**

```bash
pip install djangorestframework
```

**Add to `settings.py`:**

```python
INSTALLED_APPS = [
    # ... existing apps
    'rest_framework',
    'core',
]

REST_FRAMEWORK = {
    'DEFAULT_PAGINATION_CLASS': 'rest_framework.pagination.PageNumberPagination',
    'PAGE_SIZE': 10
}
```

**Create serializer:** `core/serializers.py`

```python
from rest_framework import serializers
from .models import Article

class ArticleSerializer(serializers.ModelSerializer):
    class Meta:
        model = Article
        fields = ['id', 'title', 'content', 'created_at', 'updated_at']
```

**Create API view:** `core/views.py`

```python
from rest_framework import viewsets
from .models import Article
from .serializers import ArticleSerializer

class ArticleViewSet(viewsets.ModelViewSet):
    queryset = Article.objects.all()
    serializer_class = ArticleSerializer
```

**Add API URLs:** `my_django_app/urls.py`

```python
from django.urls import path, include
from rest_framework.routers import DefaultRouter
from core.views import ArticleViewSet

router = DefaultRouter()
router.register(r'articles', ArticleViewSet)

urlpatterns = [
    path('admin/', admin.site.urls),
    path('api/', include(router.urls)),
    path('', include('core.urls')),
]
```

**Test API:** http://localhost:8300/api/articles/

## 12. Debugging

**Set breakpoints in PyCharm:**
1. Click line number gutter to set breakpoint
2. Click "Debug" (green bug icon)
3. Access view in browser
4. Debugger pauses at breakpoint

**Inspect variables, evaluate expressions, step through code.**

**Django template debugging:**
- Breakpoints work in views
- Template errors shown in browser with Django debug page

## 13. Static Files

**Collect static files:**

```bash
python manage.py collectstatic --noinput
```

**Output:** `apps/my_django_app/public/static/`

**Install whitenoise for serving static files:**

```bash
pip install whitenoise
```

**Update `settings.py`:**

```python
MIDDLEWARE = [
    'django.middleware.security.SecurityMiddleware',
    'whitenoise.middleware.WhiteNoiseMiddleware',  # Add this
    # ... other middleware
]
```

## 14. Configure nginx-proxy-manager

1. Open http://localhost:81
2. Proxy Hosts → Add Proxy Host
   - Domain: `my-django-app.test`
   - Forward to: `python:8000`
   - WebSockets: ✓
3. Custom Nginx Configuration:
   ```nginx
   location /static/ {
       alias /app/my_django_app/public/static/;
   }
   location /media/ {
       alias /app/my_django_app/public/media/;
   }
   ```
4. SSL: Request new certificate
5. Click "Save"

**Update /etc/hosts:**
```bash
sudo sh -c 'echo "127.0.0.1 my-django-app.test" >> /etc/hosts'
```

**Update `settings.py`:**

```python
ALLOWED_HOSTS = ['my-django-app.test', 'localhost', '127.0.0.1']
```

**Access:** https://my-django-app.test

## 15. Testing

**Create tests:** `core/tests.py`

```python
from django.test import TestCase, Client
from django.urls import reverse
from .models import Article

class ArticleModelTest(TestCase):
    def test_create_article(self):
        article = Article.objects.create(
            title='Test Article',
            content='Test content'
        )
        self.assertEqual(article.title, 'Test Article')
        self.assertTrue(article.created_at)

class ArticleViewTest(TestCase):
    def setUp(self):
        self.client = Client()

    def test_index_view(self):
        response = self.client.get(reverse('index'))
        self.assertEqual(response.status_code, 200)
```

**Run tests:**

```bash
python manage.py test
```

**PyCharm test runner:**
- Right-click `tests.py` → Run 'Unittests in tests.py'
- View results in Test Runner panel

## 16. Database Management in PyCharm

1. View → Tool Windows → Database
2. "+" → Data Source → PostgreSQL
3. Connection:
   - Host: `localhost`
   - Port: `5432`
   - User: `postgres`
   - Password: `admin1234567`
   - Database: `my_django_db`
4. Test Connection → OK

**Query Django tables, browse data, edit records directly.**

## Port Allocation

- **8300**: Django dev server
- **8301**: Flask (if running multiple Python apps)

## Troubleshooting

**Issue:** PyCharm can't find Django modules
- Verify Python interpreter set to container
- Rebuild project index: File → Invalidate Caches / Restart

**Issue:** Database connection refused
- Use container name (`postgres`, not `localhost`) in settings.py
- Verify database container running
- Check credentials match .env

**Issue:** Static files not loading
- Run `python manage.py collectstatic`
- Verify `STATIC_ROOT = BASE_DIR / 'public' / 'static'`
- Check whitenoise installed and in middleware

**Issue:** PyCharm Django console not working
- Settings → Build, Execution, Deployment → Console → Django Console
- Verify interpreter and settings module configured

## Next Steps

- Setup Celery for background tasks (with Redis)
- Add Django Channels for WebSockets
- Configure Django Debug Toolbar
- Setup pytest for testing
- Add Django Allauth for authentication
- Configure CORS headers for API
