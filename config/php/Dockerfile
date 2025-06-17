# Use the official PHP 8.3 FPM base image with Debian bookworm
FROM php:8.3-fpm

# Arguments for versions (makes it easier to update)
ARG NODE_VERSION=lts
ARG MAILPIT_VERSION=v1.22.3
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
    libpng-dev \
    libjpeg-dev \
    libwebp-dev \
    libfreetype6-dev \
    libonig-dev \
    libzip-dev \
    libpq-dev \
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
    && pecl install redis mongodb \
    && docker-php-ext-enable redis opcache pdo_mysql pdo_pgsql mongodb \
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
ENV PATH="/root/.composer/vendor/bin:${PATH}"
    
# Install development tools
RUN composer global require laravel/installer && \
    curl -O https://raw.githubusercontent.com/wp-cli/builds/gh-pages/phar/wp-cli.phar && \
    chmod +x wp-cli.phar && \
    mv wp-cli.phar /usr/local/bin/wp

# Install Mailpit
RUN wget https://github.com/axllent/mailpit/releases/download/${MAILPIT_VERSION}/mailpit-linux-amd64.tar.gz \
    && tar xzf mailpit-linux-amd64.tar.gz \
    && mv mailpit /usr/local/bin/ \
    && chmod +x /usr/local/bin/mailpit \
    && rm mailpit-linux-amd64.tar.gz \
    && mailpit version

# Install Oh My Zsh and set Zsh as default shell
RUN sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" \
    && chsh -s $(which zsh)

# Copy smart entrypoint script
COPY smart-entrypoint.sh /usr/local/bin/smart-entrypoint.sh
RUN chmod +x /usr/local/bin/smart-entrypoint.sh

# Set permissions for the container user
RUN chown -R www-data:www-data /var/www/html

# Set the working directory
WORKDIR /var/www/html

# Expose ports for PHP-FPM and Mailpit
EXPOSE 9000 8025 1025

# Default command
CMD ["php-fpm"]