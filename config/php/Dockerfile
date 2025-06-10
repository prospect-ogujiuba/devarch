# =================================================================
# DEVARCH MULTI-LANGUAGE RUNTIME CONTAINER
# =================================================================
# This Dockerfile creates a comprehensive development environment
# supporting PHP, Node.js, Python, and common development tools.

# Use PHP 8.3 FPM as the base image
ARG PHP_VERSION=8.3
FROM php:${PHP_VERSION}-fpm-bullseye

# Build arguments
ARG NODE_VERSION=20
ARG PYTHON_VERSION=3.11
ARG COMPOSER_VERSION=2.7

# Set environment variables
ENV DEBIAN_FRONTEND=noninteractive
ENV TZ=America/Toronto
ENV COMPOSER_ALLOW_SUPERUSER=1
ENV NODE_ENV=development
ENV PYTHONUNBUFFERED=1

# =================================================================
# SYSTEM DEPENDENCIES AND TOOLS
# =================================================================
RUN apt-get update && apt-get install -y \
    # System utilities
    curl \
    wget \
    git \
    unzip \
    zip \
    vim \
    nano \
    htop \
    tree \
    jq \
    # Build tools
    build-essential \
    software-properties-common \
    pkg-config \
    # Image processing
    libfreetype6-dev \
    libjpeg62-turbo-dev \
    libpng-dev \
    libwebp-dev \
    libavif-dev \
    # Database clients
    default-mysql-client \
    postgresql-client \
    # Development libraries
    libzip-dev \
    libicu-dev \
    libonig-dev \
    libxml2-dev \
    libcurl4-openssl-dev \
    libssl-dev \
    # System libraries
    zlib1g-dev \
    libpq-dev \
    libbz2-dev \
    libxslt1-dev \
    libreadline-dev \
    libsqlite3-dev \
    # Process management
    supervisor \
    # SSL certificates
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# =================================================================
# PHP EXTENSIONS AND CONFIGURATION
# =================================================================
# Configure and install PHP extensions
RUN docker-php-ext-configure gd \
        --with-freetype \
        --with-jpeg \
        --with-webp \
    && docker-php-ext-configure intl \
    && docker-php-ext-install -j$(nproc) \
        gd \
        zip \
        pdo \
        pdo_mysql \
        pdo_pgsql \
        mysqli \
        bcmath \
        intl \
        mbstring \
        xml \
        curl \
        exif \
        fileinfo \
        tokenizer \
        opcache \
        bz2 \
        calendar \
        ctype \
        dom \
        filter \
        hash \
        iconv \
        json \
        libxml \
        pcre \
        session \
        simplexml \
        spl \
        xmlreader \
        xmlwriter \
        xsl

# Install PECL extensions
RUN pecl install \
        redis-6.0.2 \
        mongodb-1.17.1 \
        xdebug-3.3.1 \
    && docker-php-ext-enable \
        redis \
        mongodb \
        xdebug

# =================================================================
# COMPOSER INSTALLATION
# =================================================================
COPY --from=composer:${COMPOSER_VERSION} /usr/bin/composer /usr/bin/composer

# Install global Composer packages
RUN composer global require \
        laravel/installer \
        symfony/console \
        phpunit/phpunit \
        squizlabs/php_codesniffer \
        friendsofphp/php-cs-fixer \
        phpstan/phpstan \
    && ln -s /root/.composer/vendor/bin/* /usr/local/bin/

# =================================================================
# NODE.JS AND NPM INSTALLATION
# =================================================================
# Install Node.js using NodeSource repository
RUN curl -fsSL https://deb.nodesource.com/setup_${NODE_VERSION}.x | bash - \
    && apt-get install -y nodejs

# Install global npm packages
RUN npm install -g \
        npm@latest \
        yarn \
        pnpm \
        @vue/cli \
        @angular/cli \
        create-react-app \
        next \
        nuxt \
        vite \
        webpack \
        webpack-cli \
        nodemon \
        pm2 \
        typescript \
        ts-node \
        eslint \
        prettier \
        @nestjs/cli

# =================================================================
# PYTHON INSTALLATION
# =================================================================
# Install Python and pip
RUN apt-get update && apt-get install -y \
        python${PYTHON_VERSION} \
        python${PYTHON_VERSION}-dev \
        python3-pip \
        python3-venv \
    && ln -sf /usr/bin/python${PYTHON_VERSION} /usr/bin/python \
    && ln -sf /usr/bin/python${PYTHON_VERSION} /usr/bin/python3 \
    && rm -rf /var/lib/apt/lists/*

# Install Python packages
RUN pip3 install --upgrade pip setuptools wheel \
    && pip3 install \
        django \
        djangorestframework \
        flask \
        fastapi \
        uvicorn \
        gunicorn \
        celery \
        redis \
        psycopg2-binary \
        mysqlclient \
        sqlalchemy \
        alembic \
        pytest \
        black \
        flake8 \
        poetry

# =================================================================
# ADDITIONAL DEVELOPMENT TOOLS
# =================================================================
# Install WordPress CLI
RUN curl -O https://raw.githubusercontent.com/wp-cli/builds/gh-pages/phar/wp-cli.phar \
    && chmod +x wp-cli.phar \
    && mv wp-cli.phar /usr/local/bin/wp

# Install Drupal Console
RUN curl https://drupalconsole.com/installer -L -o drupal.phar \
    && chmod +x drupal.phar \
    && mv drupal.phar /usr/local/bin/drupal

# Install Go (for Go applications)
RUN wget https://go.dev/dl/go1.21.6.linux-amd64.tar.gz \
    && tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz \
    && rm go1.21.6.linux-amd64.tar.gz
ENV PATH=$PATH:/usr/local/go/bin

# Install Rust (for Rust applications)
RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
ENV PATH="/root/.cargo/bin:${PATH}"

# =================================================================
# USER AND PERMISSIONS SETUP
# =================================================================
# Create a development user (optional, for better security)
RUN groupadd -g 1000 devarch \
    && useradd -u 1000 -g devarch -m -s /bin/bash devarch \
    && usermod -aG sudo devarch

# Set up working directory
WORKDIR /var/www/html

# Create necessary directories
RUN mkdir -p \
        /var/log/php \
        /var/log/supervisor \
        /run/php \
    && chown -R www-data:www-data /var/www/html \
    && chown -R www-data:www-data /var/log/php \
    && chmod -R 755 /var/www/html

# =================================================================
# CONFIGURATION FILES
# =================================================================
# Copy configuration files (will be provided separately)
COPY php.ini /usr/local/etc/php/php.ini
COPY php-fpm.conf /usr/local/etc/php-fpm.conf
COPY supervisord.conf /etc/supervisor/conf.d/supervisord.conf

# =================================================================
# HEALTH CHECK AND MONITORING
# =================================================================
# Add health check script
RUN echo '#!/bin/bash\n\
set -e\n\
\n\
# Check PHP-FPM\n\
if ! pgrep -f "php-fpm: master process" > /dev/null; then\n\
    echo "PHP-FPM is not running"\n\
    exit 1\n\
fi\n\
\n\
# Check if PHP-FPM is responding\n\
if ! cgi-fcgi -bind -connect 127.0.0.1:9000 > /dev/null 2>&1; then\n\
    echo "PHP-FPM is not responding"\n\
    exit 1\n\
fi\n\
\n\
echo "All services are healthy"\n\
exit 0\n\
' > /usr/local/bin/health-check.sh \
    && chmod +x /usr/local/bin/health-check.sh

# Install fcgi for health checks
RUN apt-get update && apt-get install -y libfcgi0ldbl \
    && rm -rf /var/lib/apt/lists/*

# =================================================================
# STARTUP SCRIPT
# =================================================================
# Create startup script for multiple services
RUN echo '#!/bin/bash\n\
set -e\n\
\n\
echo "Starting DevArch Multi-Language Runtime..."\n\
\n\
# Start PHP-FPM\n\
echo "Starting PHP-FPM..."\n\
php-fpm --daemonize\n\
\n\
# Function to start development servers based on detected projects\n\
start_dev_servers() {\n\
    cd /var/www/html\n\
    \n\
    # Look for different project types and start appropriate dev servers\n\
    for dir in */; do\n\
        if [ -d "$dir" ]; then\n\
            cd "$dir"\n\
            \n\
            # Next.js project\n\
            if [ -f "next.config.js" ] || [ -f "next.config.mjs" ]; then\n\
                echo "Starting Next.js dev server for $dir..."\n\
                npm run dev > /var/log/php/${dir%/}-nextjs.log 2>&1 &\n\
            # Django project\n\
            elif [ -f "manage.py" ]; then\n\
                echo "Starting Django dev server for $dir..."\n\
                python manage.py runserver 0.0.0.0:8000 > /var/log/php/${dir%/}-django.log 2>&1 &\n\
            # Node.js project (generic)\n\
            elif [ -f "package.json" ] && [ ! -f "next.config.js" ]; then\n\
                if grep -q "\"dev\"" package.json; then\n\
                    echo "Starting Node.js dev server for $dir..."\n\
                    npm run dev > /var/log/php/${dir%/}-nodejs.log 2>&1 &\n\
                fi\n\
            fi\n\
            \n\
            cd ..\n\
        fi\n\
    done\n\
}\n\
\n\
# Start development servers if in development mode\n\
if [ "$DEV_MODE" = "true" ]; then\n\
    echo "Development mode enabled - starting dev servers..."\n\
    start_dev_servers\n\
fi\n\
\n\
echo "DevArch Runtime started successfully!"\n\
\n\
# Keep container running and handle signals\n\
trap "echo \"Shutting down...\"; pkill -TERM php-fpm; exit 0" SIGTERM SIGINT\n\
\n\
# Monitor PHP-FPM\n\
while kill -0 $(pgrep -f "php-fpm: master process") 2>/dev/null; do\n\
    sleep 30\n\
done\n\
\n\
echo "PHP-FPM stopped. Exiting..."\n\
exit 1\n\
' > /usr/local/bin/start-runtime.sh \
    && chmod +x /usr/local/bin/start-runtime.sh

# =================================================================
# EXPOSE PORTS
# =================================================================
# PHP-FPM
EXPOSE 9000
# Node.js applications
EXPOSE 3000
# Vite dev server
EXPOSE 5173
# Django dev server
EXPOSE 8000
# Additional dev ports
EXPOSE 3001 8001 8080

# =================================================================
# FINAL SETUP
# =================================================================
# Set proper ownership
RUN chown -R www-data:www-data /var/www/html /var/log/php

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=30s --retries=3 \
    CMD /usr/local/bin/health-check.sh

# Default command
CMD ["/usr/local/bin/start-runtime.sh"]