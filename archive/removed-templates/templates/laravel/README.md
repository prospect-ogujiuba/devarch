# DevArch Laravel Template

Laravel framework template with standard `public/` directory structure.

## Quick Start

```bash
# Create new Laravel project
composer create-project laravel/laravel .

# Install dependencies
composer install

# Configure environment
cp .env.example .env
php artisan key:generate

# Run migrations
php artisan migrate

# Development server (inside PHP container)
php artisan serve --host=0.0.0.0 --port=8100
```

## Structure

Laravel already follows DevArch's public/ standard:

```
laravel-app/
├── public/              # WEB ROOT - Already configured by Laravel
│   ├── index.php        # Entry point
│   ├── .htaccess
│   └── assets/
├── app/
├── config/
├── database/
├── resources/
└── routes/
```

## DevArch Integration

### Port: 8100-8199 (PHP range)
### Domain: Configure via Nginx Proxy Manager
### Container: php service

## Documentation

See: https://laravel.com/docs
