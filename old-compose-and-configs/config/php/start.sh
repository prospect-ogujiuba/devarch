#!/bin/bash
# Start PHP-FPM in background
php-fpm --daemonize

# Start PHP built-in server on port 8000 for dashboard API
exec php -S 0.0.0.0:8000 -t /var/www/html/dashboard/api
