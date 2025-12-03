# DevArch WordPress Template

WordPress template with standard `public/` directory structure.

## Quick Start

```bash
# Use WordPress installation script
cd /home/fhcadmin/projects/devarch
./scripts/wordpress/install-wordpress.sh

# Or manual installation:
# Download WordPress to public/ directory
cd apps/your-app
wget https://wordpress.org/latest.tar.gz
tar -xzf latest.tar.gz --strip-components=1 -C public/
rm latest.tar.gz
```

## Structure

```
wordpress-app/
├── public/              # WEB ROOT - WordPress installation
│   ├── index.php
│   ├── wp-config.php
│   ├── wp-content/
│   │   ├── plugins/
│   │   ├── themes/
│   │   └── mu-plugins/
│   ├── wp-admin/
│   └── wp-includes/
└── README.md
```

## Database Configuration

WordPress requires MySQL/MariaDB database. Configure in `public/wp-config.php`:

```php
define('DB_NAME', 'your_database');
define('DB_USER', 'your_user');
define('DB_PASSWORD', 'your_password');
define('DB_HOST', 'mariadb');
```

## DevArch Integration

### Port: 8100-8199 (PHP range)
### Domain: Configure via Nginx Proxy Manager
### Container: php service
### Database: mariadb or mysql

## Documentation

See: https://wordpress.org/documentation/
