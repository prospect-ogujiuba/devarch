# Use the official PHP 8.3 FPM base image with Debian bookworm
FROM php:8.3-fpm

# Arguments for versions (makes it easier to update)
ARG NODE_VERSION=lts
ARG COMPOSER_VERSION=latest

# Install dependencies and PHP extensions in a single layer to reduce image size
RUN apt-get update && apt-get install -y \
    sudo \
    default-mysql-client \
    vim \
    nano \
    git \
    curl \
    wget \
    zsh \
    msmtp \
    msmtp-mta \
    podman \
    libpng-dev \
    libjpeg-dev \
    libwebp-dev \
    libfreetype6-dev \
    libonig-dev \
    libzip-dev \
    libpq-dev \
    libmagickwand-dev \
    librabbitmq-dev \
    libicu-dev \
    libxml2-dev \
    libxslt1-dev \
    unzip \
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
    # Install Node.js in the same layer
    && curl -fsSL https://deb.nodesource.com/setup_${NODE_VERSION}.x | bash - \
    && apt-get install -y nodejs \
    && npm install -g npm@latest \
    # Cleanup
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* \
    && rm -rf /tmp/pear

    
# Install and configure Composer
COPY --from=composer:${COMPOSER_VERSION} /usr/bin/composer /usr/bin/composer
ENV COMPOSER_ALLOW_SUPERUSER=1
ENV COMPOSER_MEMORY_LIMIT=-1
ENV PATH="/root/.composer/vendor/bin:${PATH}"

# Install development tools
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

# Configure msmtp for mail relay
RUN echo "account default" > /etc/msmtprc \
    && echo "host mailpit" >> /etc/msmtprc \
    && echo "port 1025" >> /etc/msmtprc \
    && echo "from noreply@devarch.local" >> /etc/msmtprc \
    && chmod 644 /etc/msmtprc

# Install Oh My Zsh and set Zsh as default shell
RUN sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" \
    && chsh -s $(which zsh)

# Set permissions for the container user
RUN chown -R www-data:www-data /var/www/html

# Add www-data to sudoers for podman socket access (read-only socket is safe)
RUN echo "www-data ALL=(ALL) NOPASSWD: /usr/bin/podman" >> /etc/sudoers.d/www-data \
    && chmod 0440 /etc/sudoers.d/www-data

# Set the working directory
WORKDIR /var/www/html

# Expose ports for PHP-FPM and development servers
EXPOSE 8000 5173 9000

# Default command
CMD ["php-fpm"]