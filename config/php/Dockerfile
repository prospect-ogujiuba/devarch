# PHP Development Container
# Purpose: Lean container focused on PHP application development
# Provides: PHP 8.3-FPM, Composer, framework CLIs, database clients, development tools
# For container management: Use devarch-manager container instead

FROM php:8.3-fpm

# Arguments for versions
ARG COMPOSER_VERSION=latest

# Install system dependencies and PHP extensions
# Order: System packages first, then PHP extensions (for better layer caching)
RUN apt-get update && apt-get install -y \
    # Database clients (for migrations, debugging, queries)
    default-mysql-client \
    postgresql-client \
    # Development tools
    git \
    vim \
    nano \
    curl \
    wget \
    unzip \
    zsh \
    nodejs \
    npm \
    # Mail relay for local development
    msmtp \
    msmtp-mta \
    # Media processing
    ffmpeg \
    # PHP extension dependencies
    libpng-dev \
    libwebp-dev \
    libjpeg-dev \
    libfreetype6-dev \
    libonig-dev \
    libzip-dev \
    libpq-dev \
    libmagickwand-dev \
    librabbitmq-dev \
    libicu-dev \
    libxml2-dev \
    libxslt1-dev \
    ca-certificates \
    && docker-php-ext-configure gd --with-freetype --with-jpeg --with-webp \
    && docker-php-ext-install -j$(nproc) \
        gd \
        mbstring \
        zip \
        opcache \
        pdo \
        pdo_mysql \
        pdo_pgsql \
        mysqli \
        exif \
        intl \
        bcmath \
        soap \
        xsl \
    && pecl install redis mongodb imagick xdebug amqp \
    && docker-php-ext-enable redis opcache pdo_mysql pdo_pgsql mongodb imagick xdebug amqp exif intl bcmath soap xsl \
    # Cleanup
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* \
    && rm -rf /tmp/pear

# Install and configure Composer
COPY --from=composer:${COMPOSER_VERSION} /usr/bin/composer /usr/bin/composer
ENV COMPOSER_ALLOW_SUPERUSER=1
ENV COMPOSER_MEMORY_LIMIT=-1
ENV PATH="/root/.composer/vendor/bin:${PATH}"

# Install PHP framework CLIs (Laravel, WordPress, Symfony)
RUN composer global require laravel/installer && \
    curl -O https://raw.githubusercontent.com/wp-cli/builds/gh-pages/phar/wp-cli.phar && \
    chmod +x wp-cli.phar && \
    mv wp-cli.phar /usr/local/bin/wp

# Install Symfony CLI
RUN curl -1sLf 'https://dl.cloudsmith.io/public/symfony/stable/setup.deb.sh' | bash \
    && apt-get update \
    && apt-get install -y symfony-cli \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Install Claude Code and OpenAI Codex CLIs
RUN npm install -g @anthropic-ai/claude-code @openai/codex

# Configure msmtp for mail relay (routes to mailpit in compose)
RUN echo "account default" > /etc/msmtprc \
    && echo "host mailpit" >> /etc/msmtprc \
    && echo "port 1025" >> /etc/msmtprc \
    && echo "from noreply@devarch.local" >> /etc/msmtprc \
    && chmod 644 /etc/msmtprc

# Install Oh My Zsh and set as default shell
RUN sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" "" --unattended \
    && chsh -s $(which zsh)

# Set permissions for PHP-FPM user
RUN chown -R www-data:www-data /var/www/html

# Set working directory
WORKDIR /var/www/html

# Expose PHP-FPM and development server ports
EXPOSE 8000 9000

# Default command
CMD ["php-fpm"]
