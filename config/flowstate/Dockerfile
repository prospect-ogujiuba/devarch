# FlowState Production Container
# Purpose: Production-ready PHP container with Nginx, PHP-FPM, and FFmpeg
# Serves: React SPA + Laravel API + Filament Admin

FROM php:8.3-fpm-alpine

# Install system dependencies and build tools
RUN apk add --no-cache \
    nginx \
    supervisor \
    curl \
    ffmpeg \
    postgresql-client \
    libpng \
    libjpeg-turbo \
    freetype \
    libzip \
    libpq \
    icu-libs \
    oniguruma \
    netcat-openbsd \
    # Temp build dependencies
    && apk add --no-cache --virtual .build-deps \
        $PHPIZE_DEPS \
        libpng-dev \
        libjpeg-turbo-dev \
        freetype-dev \
        libzip-dev \
        libpq-dev \
        icu-dev \
        oniguruma-dev \
    && docker-php-ext-configure gd --with-freetype --with-jpeg \
    && docker-php-ext-install -j$(nproc) \
        gd \
        mbstring \
        zip \
        opcache \
        pdo \
        pdo_pgsql \
        exif \
        intl \
        bcmath \
        pcntl \
    && pecl install redis \
    && docker-php-ext-enable redis opcache \
    # Cleanup build dependencies
    && apk del .build-deps \
    && rm -rf /tmp/* /var/cache/apk/*

# Configure PHP for production
RUN mv "$PHP_INI_DIR/php.ini-production" "$PHP_INI_DIR/php.ini"

# PHP production settings
COPY php.ini /usr/local/etc/php/conf.d/99-flowstate.ini

# Nginx configuration
COPY nginx.conf /etc/nginx/nginx.conf

# Supervisor configuration
COPY supervisord.conf /etc/supervisor/conf.d/supervisord.conf

# Create required directories
RUN mkdir -p /var/run/php /var/log/supervisor /var/log/php /var/www/html

# Set permissions
RUN chown -R www-data:www-data /var/www/html /var/run/php /var/log/php

WORKDIR /var/www/html

EXPOSE 80 8080

CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/conf.d/supervisord.conf"]
