# DevArch Django Template

Django framework template configured to serve static files from `public/` directory.

## Quick Start

```bash
# Create Django project
pip install django
django-admin startproject myproject .

# Configure static files
# Edit settings.py (see below)

# Collect static files
python manage.py collectstatic --noinput

# Run server
python manage.py runserver 0.0.0.0:8300
```

## Configure settings.py

```python
import os
from pathlib import Path

BASE_DIR = Path(__file__).resolve().parent.parent

# Static files configuration
STATIC_URL = '/static/'
STATIC_ROOT = os.path.join(BASE_DIR, 'public/static')
STATICFILES_DIRS = [
    os.path.join(BASE_DIR, 'static'),
]

# Media files
MEDIA_URL = '/media/'
MEDIA_ROOT = os.path.join(BASE_DIR, 'public/media')
```

## Structure

```
django-app/
├── public/              # WEB ROOT
│   ├── static/         # Collected static files
│   └── media/          # Uploaded media
├── myproject/
│   ├── settings.py
│   ├── urls.py
│   └── wsgi.py
├── manage.py
└── requirements.txt
```

## DevArch Integration

### Port: 8300-8399 (Python range)
### Domain: Configure via Nginx Proxy Manager
### Container: python service

## Documentation

See: https://docs.djangoproject.com/
