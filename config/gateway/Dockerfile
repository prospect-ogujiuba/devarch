FROM php:8.3-cli

# Install required packages
RUN apt-get update && apt-get install -y \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Install Composer
COPY --from=composer:latest /usr/bin/composer /usr/bin/composer

WORKDIR /var/www

# Copy gateway script
COPY gateway.php /var/www/gateway.php

EXPOSE 8888

CMD ["php", "-S", "0.0.0.0:8888", "gateway.php"]
