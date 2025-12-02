# PHP Container Optimization - Verification Checklist

**Date:** 2025-12-02
**Purpose:** Pre-build verification checklist for PHP container optimization

---

## Pre-Build Verification

### 1. Dockerfile Changes

- [x] **System dependencies added:**
  - msmtp, msmtp-mta (mail relay)
  - libmagickwand-dev (for imagick)
  - librabbitmq-dev (for amqp)
  - libicu-dev (for intl)
  - libxml2-dev (for soap)
  - libxslt1-dev (for xsl)

- [x] **Core PHP extensions added:**
  - exif (line 44)
  - intl (line 45)
  - bcmath (line 46)
  - soap (line 47)
  - xsl (line 48)

- [x] **PECL extensions added:**
  - imagick (line 49)
  - xdebug (line 49)
  - amqp (line 49)

- [x] **All extensions enabled:**
  - Line 50 includes all new extensions in docker-php-ext-enable

- [x] **Modern PHP tools added:**
  - phpstan/phpstan (line 74)
  - friendsofphp/php-cs-fixer (line 75)
  - squizlabs/php_codesniffer (line 76)
  - rector/rector (line 77)
  - pestphp/pest (line 78)

- [x] **Symfony CLI added:**
  - Installation commands (lines 80-85)

- [x] **msmtp configuration added:**
  - Configuration block (lines 87-92)

- [x] **Mailpit binary removed:**
  - ARG MAILPIT_VERSION removed (line 6)
  - wget/tar/install commands removed (formerly lines 61-67)
  - mailpit version check removed

- [x] **Exposed ports updated:**
  - Removed ports 8025 and 1025 (line 105)
  - Kept ports 8000, 5173, 9000

### 2. php.ini Changes

- [x] **Mail configuration updated:**
  - Changed from: `/usr/local/bin/mailpit sendmail --smtp-addr=mailpit:1025`
  - Changed to: `/usr/bin/msmtp -t`

- [x] **XDebug configuration present:**
  - xdebug.mode = develop,debug
  - xdebug.client_host = host.docker.internal
  - xdebug.start_with_request = yes
  - xdebug.log_level = 0

### 3. File Integrity

```bash
# Verify Dockerfile exists
ls -lh /home/fhcadmin/projects/devarch/config/php/Dockerfile

# Verify php.ini exists
ls -lh /home/fhcadmin/projects/devarch/config/php/php.ini

# Verify mailpit service exists
ls -lh /home/fhcadmin/projects/devarch/compose/mail/mailpit.yml
```

---

## Build Commands

### Full Rebuild (Recommended)

```bash
cd /home/fhcadmin/projects/devarch

# Stop existing container
docker-compose -f compose/backend/php.yml down

# Build with no cache (ensures fresh build)
docker-compose -f compose/backend/php.yml build --no-cache

# Start container
docker-compose -f compose/backend/php.yml up -d

# Check logs
docker-compose -f compose/backend/php.yml logs -f
```

### Quick Rebuild (if dependencies haven't changed)

```bash
cd /home/fhcadmin/projects/devarch

docker-compose -f compose/backend/php.yml down
docker-compose -f compose/backend/php.yml build
docker-compose -f compose/backend/php.yml up -d
```

---

## Post-Build Verification

### 1. Container Status

```bash
# Verify container is running
docker ps | grep php

# Expected output:
# php   Up   8000/tcp, 5173/tcp, 9000/tcp
```

### 2. PHP Extensions

```bash
# List all loaded extensions
docker exec php php -m

# Should include:
# - amqp
# - bcmath
# - exif
# - imagick
# - intl
# - mongodb
# - opcache
# - pdo_mysql
# - pdo_pgsql
# - redis
# - soap
# - xdebug
# - xsl
```

#### Specific Extension Tests

```bash
# Test each new extension individually
docker exec php php -r "echo extension_loaded('exif') ? 'exif: OK' : 'exif: MISSING';" && echo
docker exec php php -r "echo extension_loaded('intl') ? 'intl: OK' : 'intl: MISSING';" && echo
docker exec php php -r "echo extension_loaded('bcmath') ? 'bcmath: OK' : 'bcmath: MISSING';" && echo
docker exec php php -r "echo extension_loaded('soap') ? 'soap: OK' : 'soap: MISSING';" && echo
docker exec php php -r "echo extension_loaded('xsl') ? 'xsl: OK' : 'xsl: MISSING';" && echo
docker exec php php -r "echo extension_loaded('imagick') ? 'imagick: OK' : 'imagick: MISSING';" && echo
docker exec php php -r "echo extension_loaded('xdebug') ? 'xdebug: OK' : 'xdebug: MISSING';" && echo
docker exec php php -r "echo extension_loaded('amqp') ? 'amqp: OK' : 'amqp: MISSING';" && echo
```

### 3. XDebug Configuration

```bash
# Check XDebug is loaded and configured
docker exec php php -i | grep xdebug

# Should show:
# xdebug.mode => develop,debug
# xdebug.client_host => host.docker.internal
# xdebug.start_with_request => yes
```

### 4. Development Tools

```bash
# PHPStan
docker exec php phpstan --version
# Expected: PHPStan - PHP Static Analysis Tool x.x.x

# PHP-CS-Fixer
docker exec php php-cs-fixer --version
# Expected: PHP CS Fixer x.x.x

# PHPCS
docker exec php phpcs --version
# Expected: PHP_CodeSniffer version x.x.x

# PHPCBF
docker exec php phpcbf --version
# Expected: PHP_CodeSniffer version x.x.x

# Rector
docker exec php rector --version
# Expected: Rector x.x.x

# Pest
docker exec php pest --version
# Expected: Pest x.x.x

# Symfony CLI
docker exec php symfony -V
# Expected: Symfony CLI version x.x.x
```

### 5. Laravel & WordPress Tools

```bash
# Laravel Installer
docker exec php laravel --version
# Expected: Laravel Installer x.x.x

# WP-CLI
docker exec php wp --version
# Expected: WP-CLI x.x.x
```

### 6. Mail Configuration

```bash
# Verify msmtp installed
docker exec php which msmtp
# Expected: /usr/bin/msmtp

# Check msmtp configuration
docker exec php cat /etc/msmtprc
# Expected output:
# account default
# host mailpit
# port 1025
# from noreply@devarch.local

# Verify mailpit binary NOT installed
docker exec php which mailpit 2>&1
# Expected: (no output or command not found)

# Test mail sending
docker exec php php -r "mail('test@example.com', 'Test Subject', 'Test message from PHP container');"

# Check mailpit web UI
# Open: http://localhost:9200
# Should see the test email
```

### 7. Mailpit Service

```bash
# Verify mailpit service is running
docker ps | grep mailpit

# Expected output:
# mailpit   axllent/mailpit   Up   127.0.0.1:9200->8025/tcp, 127.0.0.1:9201->1025/tcp

# Test mailpit service directly
docker exec mailpit mailpit version
```

### 8. ImageMagick (for imagick extension)

```bash
# Verify ImageMagick installed
docker exec php convert --version
# Expected: Version: ImageMagick 6.x.x or 7.x.x

# Test imagick extension
docker exec php php -r "echo 'ImageMagick: ' . phpversion('imagick');" && echo

# Test imagick functionality
docker exec php php -r "
\$img = new Imagick();
\$img->newImage(100, 100, new ImagickPixel('red'));
echo 'Imagick test: OK';
"
```

### 9. Composer Global Packages

```bash
# List global packages
docker exec php composer global show

# Should include:
# friendsofphp/php-cs-fixer
# laravel/installer
# pestphp/pest
# phpstan/phpstan
# rector/rector
# squizlabs/php_codesniffer
```

### 10. Container Access

```bash
# Enter container with zsh
docker exec -it php zsh

# Once inside, test commands:
php -v
composer --version
node --version
npm --version
phpstan --version
symfony -V
wp --version
```

---

## Troubleshooting

### Extensions Not Loading

**Issue:** Extension shows in `php -m` but doesn't work

```bash
# Check extension configuration
docker exec php php --ini

# Check for errors
docker exec php php -i | grep -A 5 "extension_name"

# Restart PHP-FPM
docker exec php pkill -USR2 php-fpm
```

### Imagick Fails to Build

**Issue:** Build fails during imagick installation

```bash
# Check ImageMagick libraries are installed
docker exec php dpkg -l | grep imagemagick

# If missing, rebuild with:
docker-compose -f compose/backend/php.yml build --no-cache --build-arg BUILDKIT_INLINE_CACHE=0
```

### XDebug Not Working

**Issue:** XDebug not connecting to IDE

```bash
# Verify XDebug configuration
docker exec php php -i | grep xdebug

# Check client_host setting
# Should be: host.docker.internal

# Test connection
docker exec php php -r "xdebug_info();"
```

### Composer Tools Not Found

**Issue:** Commands like `phpstan` return "command not found"

```bash
# Check PATH
docker exec php echo $PATH
# Should include: /root/.composer/vendor/bin

# Verify tools installed
docker exec php ls -la /root/.composer/vendor/bin/

# If missing, reinstall
docker exec php composer global require phpstan/phpstan
```

### Mail Not Sending to Mailpit

**Issue:** PHP mail() doesn't appear in Mailpit

```bash
# Test msmtp directly
docker exec php echo "Subject: Test\n\nTest message" | msmtp -d test@example.com

# Check msmtp logs
docker exec php tail -f /var/log/msmtp.log

# Verify mailpit service is accessible
docker exec php nc -zv mailpit 1025

# Check mailpit logs
docker logs mailpit
```

---

## Performance Benchmarks

### Extension Overhead

```bash
# Without extensions (baseline)
docker exec php php -r "echo memory_get_usage();"

# Load test script
docker exec php php -r "
\$start = microtime(true);
for (\$i = 0; \$i < 10000; \$i++) {
    bcadd('1.1', '2.2', 10);
}
\$end = microtime(true);
echo 'bcmath: ' . (\$end - \$start) . 's';
"
```

### Build Time Tracking

Expected build times:
- **First build (no cache):** 10-15 minutes
- **Rebuilds (with cache):** 3-5 minutes
- **Extension compilation (imagick):** 5-8 minutes alone

---

## Success Criteria

All items below should be verified:

- [ ] Container builds without errors
- [ ] Container starts successfully
- [ ] All 8 new extensions load (exif, intl, bcmath, soap, xsl, imagick, xdebug, amqp)
- [ ] All 5 dev tools work (phpstan, php-cs-fixer, phpcs, rector, pest)
- [ ] Symfony CLI accessible
- [ ] Laravel installer works
- [ ] WP-CLI works
- [ ] Mail sending works (appears in Mailpit UI)
- [ ] Mailpit binary NOT present in PHP container
- [ ] XDebug configuration correct
- [ ] ImageMagick functional
- [ ] No build warnings or errors

---

## Quick Test Script

Save this as `test-php-container.sh` and run after build:

```bash
#!/bin/bash
set -e

echo "=== PHP Container Verification ==="
echo

echo "1. Testing extensions..."
docker exec php php -r "
\$extensions = ['exif', 'intl', 'bcmath', 'soap', 'xsl', 'imagick', 'xdebug', 'amqp'];
foreach (\$extensions as \$ext) {
    echo \$ext . ': ' . (extension_loaded(\$ext) ? 'OK' : 'MISSING') . PHP_EOL;
}
"

echo
echo "2. Testing dev tools..."
docker exec php phpstan --version | grep -q "PHPStan" && echo "phpstan: OK" || echo "phpstan: MISSING"
docker exec php php-cs-fixer --version | grep -q "PHP CS Fixer" && echo "php-cs-fixer: OK" || echo "php-cs-fixer: MISSING"
docker exec php phpcs --version | grep -q "CodeSniffer" && echo "phpcs: OK" || echo "phpcs: MISSING"
docker exec php rector --version | grep -q "Rector" && echo "rector: OK" || echo "rector: MISSING"
docker exec php pest --version | grep -q "Pest" && echo "pest: OK" || echo "pest: MISSING"
docker exec php symfony -V | grep -q "Symfony CLI" && echo "symfony: OK" || echo "symfony: MISSING"

echo
echo "3. Testing mail..."
docker exec php which msmtp | grep -q "msmtp" && echo "msmtp: OK" || echo "msmtp: MISSING"
docker exec php which mailpit 2>&1 | grep -q "no mailpit" && echo "mailpit binary: REMOVED (correct)" || echo "mailpit binary: STILL PRESENT (wrong)"

echo
echo "4. Testing Laravel/WordPress tools..."
docker exec php laravel --version | grep -q "Laravel Installer" && echo "laravel: OK" || echo "laravel: MISSING"
docker exec php wp --version | grep -q "WP-CLI" && echo "wp-cli: OK" || echo "wp-cli: MISSING"

echo
echo "=== All tests completed ==="
```

Make executable and run:
```bash
chmod +x test-php-container.sh
./test-php-container.sh
```

---

## Files Modified

1. `/home/fhcadmin/projects/devarch/config/php/Dockerfile` - Main optimization changes
2. `/home/fhcadmin/projects/devarch/config/php/php.ini` - Mail configuration update
3. `/home/fhcadmin/projects/devarch/context/php-optimization-summary.md` - Detailed summary
4. `/home/fhcadmin/projects/devarch/context/php-verification-checklist.md` - This checklist

---

**Ready to build!**

Run the build commands above and use this checklist to verify everything works correctly.
