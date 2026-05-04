FROM php:8.4-fpm

ARG COMPOSER_VERSION=latest

RUN apt-get update && apt-get install -y \
    default-mysql-client \
    postgresql-client \
    git \
    vim \
    nano \
    curl \
    wget \
    unzip \
    zsh \
    nodejs \
    npm \
    msmtp \
    msmtp-mta \
    ffmpeg \
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
        pcntl \
    && pecl install redis mongodb imagick xdebug amqp \
    && docker-php-ext-enable redis opcache pdo_mysql pdo_pgsql mongodb imagick xdebug amqp exif intl bcmath soap xsl pcntl \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* \
    && rm -rf /tmp/pear

COPY --from=composer:${COMPOSER_VERSION} /usr/bin/composer /usr/bin/composer
ENV COMPOSER_ALLOW_SUPERUSER=1
ENV COMPOSER_MEMORY_LIMIT=-1
ENV PATH="/root/.composer/vendor/bin:${PATH}"

RUN composer global require laravel/installer && \
    curl -O https://raw.githubusercontent.com/wp-cli/builds/gh-pages/phar/wp-cli.phar && \
    /bin/bash -c "$(curl -fsSL https://php.new/install/linux/8.4)" \
    chmod +x wp-cli.phar && \
    mv wp-cli.phar /usr/local/bin/wp

RUN curl -1sLf 'https://dl.cloudsmith.io/public/symfony/stable/setup.deb.sh' | bash \
    && apt-get update \
    && apt-get install -y symfony-cli \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

RUN npm install -g @anthropic-ai/claude-code @openai/codex

RUN echo "account default" > /etc/msmtprc \
    && echo "host mailpit" >> /etc/msmtprc \
    && echo "port 1025" >> /etc/msmtprc \
    && echo "from noreply@devarch.local" >> /etc/msmtprc \
    && chmod 644 /etc/msmtprc

RUN sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" "" --unattended \
    && chsh -s $(which zsh)

RUN chown -R www-data:www-data /var/www/html

WORKDIR /var/www/html

EXPOSE 8000 9000

CMD ["php-fpm"]
