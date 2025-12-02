# PHP Container Optimization Summary

**Date:** 2025-12-02
**Container:** PHP 8.3-FPM
**Optimization Focus:** Laravel, WordPress, and JavaScript-heavy application development

---

## Overview

This optimization addresses missing critical PHP extensions, adds modern development tools, and removes redundant services to create a lean, efficient PHP development container tailored for Laravel and WordPress workflows.

---

## Changes Summary

### Removed Items

#### 1. Mailpit Binary Installation (REMOVED)
**Lines Removed:** 61-67 in Dockerfile
```dockerfile
# REMOVED: Mailpit binary installation
ARG MAILPIT_VERSION=v1.22.3
RUN wget https://github.com/axllent/mailpit/releases/download/${MAILPIT_VERSION}/mailpit-linux-amd64.tar.gz \
    && tar xzf mailpit-linux-amd64.tar.gz \
    && mv mailpit /usr/local/bin/ \
    && chmod +x /usr/local/bin/mailpit \
    && rm mailpit-linux-amd64.tar.gz \
    && mailpit version
```

**Rationale:**
- Redundant installation - dedicated mailpit service exists at `compose/mail/mailpit.yml`
- PHP container should only SEND email, not BE a mail server
- msmtp SMTP relay is sufficient for PHP to send to the mailpit service
- Reduces container size and complexity

**Replacement:** msmtp SMTP client configured to relay to mailpit service

#### 2. Exposed Ports Cleanup
**Changed:** `EXPOSE 8000 5173 9000 8025 1025` → `EXPOSE 8000 5173 9000`

**Rationale:**
- Port 8025 (Mailpit web UI) and 1025 (Mailpit SMTP) are exposed by the mailpit service
- No need to expose these ports in PHP container

---

## Added Extensions

### Core PHP Extensions (via docker-php-ext-install)

| Extension | Purpose | Use Case |
|-----------|---------|----------|
| **exif** | Read EXIF data from images | WordPress media library metadata |
| **intl** | Internationalization support | Laravel localization, multi-language apps |
| **bcmath** | Arbitrary precision mathematics | E-commerce calculations, financial apps |
| **soap** | SOAP web services client | Legacy API integrations |
| **xsl** | XSL transformations | XML processing, data transformation |

### PECL Extensions

| Extension | Purpose | Use Case |
|-----------|---------|----------|
| **imagick** | ImageMagick integration | WordPress image manipulation, advanced image processing |
| **xdebug** | Debugging and profiling | Development debugging (was configured in php.ini but not installed) |
| **amqp** | RabbitMQ support | Message queue processing, async jobs |

### System Dependencies Added

```dockerfile
msmtp                  # SMTP client for mail relay
msmtp-mta             # MTA replacement
libmagickwand-dev     # Required for imagick
librabbitmq-dev       # Required for amqp
libicu-dev            # Required for intl
libxml2-dev           # Required for soap/xsl
libxslt1-dev          # Required for xsl
```

---

## Added Development Tools

### Modern PHP Analysis & Quality Tools

Installed via `composer global require`:

| Tool | Package | Purpose |
|------|---------|---------|
| **PHPStan** | `phpstan/phpstan` | Static analysis, find bugs before runtime |
| **PHP-CS-Fixer** | `friendsofphp/php-cs-fixer` | Code formatting, PSR compliance |
| **PHP_CodeSniffer** | `squizlabs/php_codesniffer` | Coding standards (phpcs/phpcbf commands) |
| **Rector** | `rector/rector` | Automated refactoring, PHP version upgrades |
| **Pest** | `pestphp/pest` | Modern testing framework for PHP |

### Symfony CLI

**Installation:**
```bash
curl -1sLf 'https://dl.cloudsmith.io/public/symfony/stable/setup.deb.sh' | bash
apt-get install symfony-cli
```

**Benefits:**
- Enhanced Laravel development experience
- Local web server with TLS support
- Environment variable management
- Better console output formatting

### Mail Configuration (msmtp)

**Configuration:** `/etc/msmtprc`
```
account default
host mailpit
port 1025
from noreply@devarch.local
```

**php.ini Update:**
- **Before:** `sendmail_path = "/usr/local/bin/mailpit sendmail --smtp-addr=mailpit:1025"`
- **After:** `sendmail_path = "/usr/bin/msmtp -t"`

---

## Before/After Comparison

### Extensions Count

| Category | Before | After | Added |
|----------|--------|-------|-------|
| Core Extensions | 8 | 13 | +5 (exif, intl, bcmath, soap, xsl) |
| PECL Extensions | 2 | 5 | +3 (imagick, xdebug, amqp) |
| **Total** | **10** | **18** | **+8** |

### Development Tools

| Category | Before | After |
|----------|--------|-------|
| Laravel Tools | Laravel Installer, WP-CLI | Same + 5 modern tools + Symfony CLI |
| Code Quality | None | PHPStan, PHP-CS-Fixer, PHPCS, Rector |
| Testing | None | Pest |
| CLI Tools | None | Symfony CLI |

### Container Size Impact

| Item | Size Impact | Notes |
|------|-------------|-------|
| ImageMagick libs | +~50MB | Essential for WordPress image processing |
| PHP Extensions | +~20MB | Critical functionality |
| Symfony CLI | +~15MB | Significant DX improvement |
| Development Tools | +~30MB | Installed in vendor, not runtime |
| Removed Mailpit | -~10MB | Redundant binary removed |
| **Net Impact** | **+~105MB** | Acceptable for complete dev environment |

---

## Verification Commands

### Test All Extensions Load

```bash
docker exec php php -m
```

**Expected output should include:**
```
[PHP Modules]
amqp
bcmath
exif
imagick
intl
mongodb
redis
soap
xdebug
xsl
... (other modules)
```

### Verify XDebug Configuration

```bash
docker exec php php -i | grep xdebug
```

**Should show:**
```
xdebug.mode => develop,debug
xdebug.client_host => host.docker.internal
xdebug.start_with_request => yes
```

### Check Development Tools

```bash
# PHPStan
docker exec php phpstan --version

# PHP-CS-Fixer
docker exec php php-cs-fixer --version

# PHPCS
docker exec php phpcs --version

# Rector
docker exec php rector --version

# Pest
docker exec php pest --version

# Symfony CLI
docker exec php symfony -V
```

### Verify Mail Configuration

```bash
# Check msmtp config
docker exec php cat /etc/msmtprc

# Test sending email
docker exec php php -r "mail('test@example.com', 'Test', 'Test message');"

# Check mailpit UI at http://localhost:9200
```

### Check Mailpit Service

```bash
# Verify mailpit service is running
docker ps | grep mailpit

# Should show:
# mailpit   axllent/mailpit   Up   127.0.0.1:9200->8025/tcp, 127.0.0.1:9201->1025/tcp
```

---

## Build Instructions

### Full Rebuild (Recommended)

```bash
cd /home/fhcadmin/projects/devarch
docker-compose -f compose/backend/php.yml build --no-cache
docker-compose -f compose/backend/php.yml up -d
```

### Verify Build Success

```bash
# Check container status
docker ps | grep php

# Enter container
docker exec -it php zsh

# Test extensions
php -m | grep -E '(imagick|xdebug|amqp|intl|exif|bcmath|soap|xsl)'

# Test tools
phpstan --version
symfony -V
wp --version
```

---

## Extension Use Cases

### WordPress Development

| Extension | WordPress Feature |
|-----------|-------------------|
| **exif** | Media library metadata extraction |
| **imagick** | Advanced image manipulation, thumbnails, filters |
| **intl** | Multilingual sites, locale support |
| **bcmath** | WooCommerce price calculations |
| **soap** | Payment gateway integrations |

### Laravel Development

| Extension | Laravel Feature |
|-----------|-----------------|
| **intl** | Localization, string manipulation |
| **bcmath** | Cashier, financial calculations |
| **soap** | Legacy API integrations |
| **amqp** | Queue workers (RabbitMQ driver) |
| **xdebug** | Debugging, code coverage |
| **xsl** | XML processing (RSS feeds, sitemaps) |

### Modern Development Workflow

| Tool | Workflow Stage |
|------|----------------|
| **PHPStan** | Pre-commit hooks, CI/CD static analysis |
| **PHP-CS-Fixer** | Code formatting automation |
| **PHPCS** | Enforce coding standards |
| **Rector** | PHP version upgrades, automated refactoring |
| **Pest** | Test-driven development |
| **Symfony CLI** | Local development server, debugging |

---

## Architecture Notes

### Mail Flow

```
PHP Application
    ↓ (uses mail() function)
msmtp (SMTP client in PHP container)
    ↓ (relays to mailpit:1025)
Mailpit Service (separate container)
    ↓ (stores and displays)
Mailpit Web UI (http://localhost:9200)
```

### Why msmtp Instead of Mailpit Binary?

1. **Separation of Concerns**: PHP container sends, mailpit service receives
2. **Single Source of Truth**: One mailpit service for all containers
3. **Reduced Complexity**: PHP doesn't need mail server functionality
4. **Container Size**: Smaller PHP image
5. **Better Architecture**: Microservices principle - one service, one job

---

## Performance Considerations

### Build Time

- **First build:** ~10-15 minutes (imagick compilation)
- **Subsequent builds:** ~3-5 minutes (Docker layer caching)

### Runtime Performance

- **opcache:** Enabled and optimized for development
- **xdebug:** Configured but can be disabled in production
- **Extensions:** All extensions are loaded but have minimal overhead when not used

### Production Optimization

For production deployments, consider:
```dockerfile
# Disable xdebug in production
RUN sed -i 's/xdebug.mode = develop,debug/xdebug.mode = off/' /usr/local/etc/php/php.ini

# Or remove xdebug entirely
RUN docker-php-ext-disable xdebug
```

---

## Compatibility Matrix

| PHP Version | Extensions | Status |
|-------------|------------|--------|
| 8.3.x | All extensions | Fully supported |
| 8.4.x | All extensions | Should work (test required) |
| 8.2.x | All extensions | Downgrade supported |

---

## Troubleshooting

### Extension Won't Load

```bash
# Check extension files
docker exec php ls -la /usr/local/lib/php/extensions/

# Check PHP error log
docker exec php tail -f /var/log/php-fpm.log
```

### Imagick Issues

```bash
# Verify ImageMagick installation
docker exec php convert --version

# Check imagick PHP extension
docker exec php php -i | grep imagick
```

### Mail Not Sending

```bash
# Test msmtp directly
docker exec php msmtp --version

# Check msmtp configuration
docker exec php cat /etc/msmtprc

# Verify mailpit service connection
docker exec php nc -zv mailpit 1025
```

### Composer Tools Not Found

```bash
# Check PATH
docker exec php echo $PATH

# Verify tools installed
docker exec php ls -la /root/.composer/vendor/bin/

# Reinstall if needed
docker exec php composer global require phpstan/phpstan
```

---

## Future Optimization Opportunities

1. **Multi-stage build**: Separate build and runtime stages
2. **Alpine variant**: Consider php:8.3-fpm-alpine for smaller base (~50MB savings)
3. **Extension toggling**: Environment-based extension enabling/disabling
4. **Tool versioning**: Pin specific versions of development tools
5. **Cache optimization**: Use BuildKit cache mounts for composer

---

## Related Services

| Service | Compose File | Purpose |
|---------|--------------|---------|
| Mailpit | `compose/mail/mailpit.yml` | Email testing and viewing |
| PHP | `compose/backend/php.yml` | Application runtime |
| MySQL | (assumed) | Database |
| Redis | (assumed) | Cache/sessions |

---

## Success Criteria Met

- [x] All 8 missing extensions added (exif, intl, bcmath, soap, xsl, imagick, xdebug, amqp)
- [x] All 5 modern PHP tools added (PHPStan, PHP-CS-Fixer, PHPCS, Rector, Pest)
- [x] Symfony CLI installed and accessible
- [x] Mailpit binary installation removed (msmtp configured)
- [x] No redundant or duplicate installations
- [x] Build order is correct (system deps → php extensions → pecl → tools)
- [x] php.ini updated to use msmtp
- [x] Summary document created with clear before/after comparison

---

## References

- [PHP Docker Official Images](https://hub.docker.com/_/php)
- [ImageMagick PHP Extension](https://www.php.net/manual/en/book.imagick.php)
- [Xdebug Documentation](https://xdebug.org/docs/)
- [Symfony CLI](https://symfony.com/download)
- [PHPStan Documentation](https://phpstan.org/)
- [msmtp SMTP Client](https://marlam.de/msmtp/)
- [Mailpit](https://github.com/axllent/mailpit)

---

**Optimization completed successfully.**
The PHP container now provides complete coverage for Laravel, WordPress, and JavaScript-heavy application development with modern tooling and a clean, efficient configuration.
