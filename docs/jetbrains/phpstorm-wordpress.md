# PHPStorm - WordPress Quick Start

Create WordPress projects in DevArch using PHPStorm with container-based PHP interpreter and WP-CLI.

## Prerequisites

- PHPStorm installed
- DevArch containers running:
  ```bash
  ./scripts/service-manager.sh start database proxy backend
  ```

## 1. Create Project Structure

**Create directory and initialize WordPress:**

```bash
# Create app directory
mkdir -p /home/fhcadmin/projects/devarch/apps/my-wordpress-site

# Install WordPress via WP-CLI in container
podman exec -it php bash -c "cd /var/www/html/my-wordpress-site && wp core download --allow-root"
```

**Open in PHPStorm:**
1. File → Open → `/home/fhcadmin/projects/devarch/apps/my-wordpress-site`

## 2. Verify `public/` Structure

WordPress core files are typically in root. For DevArch, move to `public/`:

```bash
podman exec -it php bash
cd /var/www/html/my-wordpress-site
mkdir -p public
mv wp-* public/
mv index.php public/
mv xmlrpc.php public/
mv wp-config-sample.php public/
```

**Or use symlink approach (cleaner):**
```bash
# Keep WordPress in root, symlink public/ to root
cd /home/fhcadmin/projects/devarch/apps/my-wordpress-site
ln -s . public
```

Final structure:
```
apps/my-wordpress-site/
├── public/           # Symlink to . or actual directory
│   ├── wp-admin/
│   ├── wp-content/
│   ├── wp-includes/
│   ├── index.php
│   └── wp-config.php
└── ...
```

## 3. Configure PHP Interpreter

1. Settings → PHP → CLI Interpreter → "..."
2. Click "+" → "From Docker, Vagrant, VM, WSL, Remote..."
3. Docker Compose:
   - Configuration file: `/home/fhcadmin/projects/devarch/compose/backend/php.yml`
   - Service: `php`
   - Lifecycle: Connect to existing container
4. Click "OK"
5. Verify PHP 8.3, Xdebug detected

## 4. Database Setup

**Create WordPress database:**

```bash
podman exec -it php bash
mysql -h mariadb -u root -padmin1234567 -e "CREATE DATABASE my_wordpress_site CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
```

**Configure WordPress:**

1. Create `wp-config.php`:
   ```bash
   podman exec -it php bash
   cd /var/www/html/my-wordpress-site/public
   wp config create \
     --dbname=my_wordpress_site \
     --dbuser=root \
     --dbpass=admin1234567 \
     --dbhost=mariadb \
     --allow-root
   ```

2. Install WordPress:
   ```bash
   wp core install \
     --url=https://my-wordpress-site.test \
     --title="My WordPress Site" \
     --admin_user=admin \
     --admin_password=admin1234567 \
     --admin_email=admin@devarch.test \
     --allow-root
   ```

## 5. Configure nginx-proxy-manager

1. Open http://localhost:81
2. Login: `admin@devarch.test` / `admin1234567`
3. Proxy Hosts → Add Proxy Host
   - Domain Names: `my-wordpress-site.test`
   - Scheme: `http`
   - Forward Hostname/IP: `php`
   - Forward Port: `8000`
   - Custom Nginx Configuration:
     ```nginx
     location / {
         try_files $uri $uri/ /index.php?$args;
     }
     location ~ \.php$ {
         fastcgi_pass php:9000;
         fastcgi_index index.php;
         fastcgi_param SCRIPT_FILENAME /var/www/html/my-wordpress-site/public$fastcgi_script_name;
         include fastcgi_params;
     }
     ```
4. SSL tab:
   - Request New SSL Certificate
   - Force SSL: ✓
5. Click "Save"

## 6. Update /etc/hosts

```bash
sudo sh -c 'echo "127.0.0.1 my-wordpress-site.test" >> /etc/hosts'
```

## 7. Development Workflow

**Access WordPress:**
- Frontend: https://my-wordpress-site.test
- Admin: https://my-wordpress-site.test/wp-admin

**WP-CLI Usage (in container):**

```bash
# Access container
podman exec -it php bash
cd /var/www/html/my-wordpress-site/public

# Plugin management
wp plugin list --allow-root
wp plugin install wordpress-seo --activate --allow-root
wp plugin update --all --allow-root

# Theme management
wp theme list --allow-root
wp theme install twentytwentyfour --activate --allow-root

# Database operations
wp db export backup.sql --allow-root
wp db query "SELECT * FROM wp_posts LIMIT 5" --allow-root
```

**PHPStorm Terminal Integration:**

1. Tools → Start SSH Session → php container
2. Set working directory: `cd /var/www/html/my-wordpress-site/public`
3. Run WP-CLI commands directly

## 8. Custom Theme/Plugin Development

**Create custom theme:**

```bash
podman exec -it php bash
cd /var/www/html/my-wordpress-site/public/wp-content/themes
wp scaffold _s mytheme --theme_name="My Theme" --author="Your Name" --allow-root
```

**Structure:**
```
apps/my-wordpress-site/
├── public/
│   └── wp-content/
│       ├── themes/
│       │   └── mytheme/    # Custom theme
│       ├── plugins/
│       │   └── myplugin/   # Custom plugin
│       └── mu-plugins/     # Must-use plugins
```

**Enable WordPress Coding Standards in PHPStorm:**

1. Settings → PHP → Quality Tools → PHP_CodeSniffer
2. Configuration: Set path to `phpcs` in container
3. Coding Standard: WordPress

## 9. Debugging with Xdebug

1. Settings → PHP → Servers
   - Name: `my-wordpress-site`
   - Host: `my-wordpress-site.test`
   - Port: `443`
   - Debugger: Xdebug
   - Path mappings:
     - `/home/fhcadmin/projects/devarch/apps/my-wordpress-site/public` → `/var/www/html/my-wordpress-site/public`

2. Set breakpoint in theme/plugin PHP file
3. Run → Start Listening for PHP Debug Connections
4. Enable Xdebug in browser
5. Refresh WordPress page
6. Debugger should pause at breakpoint

**Debug specific hooks:**
```php
// In theme functions.php or plugin
add_action('init', function() {
    xdebug_break(); // Force breakpoint
});
```

## 10. Database Management

**PHPStorm Database Tool:**

1. View → Tool Windows → Database
2. "+" → Data Source → MySQL
3. Connection:
   - Host: `localhost`
   - Port: `3306`
   - User: `root`
   - Password: `admin1234567`
   - Database: `my_wordpress_site`
4. Test Connection → OK

**Query WordPress database:**
```sql
SELECT * FROM wp_posts WHERE post_status = 'publish';
SELECT * FROM wp_options WHERE option_name LIKE '%siteurl%';
```

## 11. TypeRocket Integration (Optional)

If using TypeRocket Pro framework:

```bash
# Install TypeRocket as mu-plugin
cd /home/fhcadmin/projects/devarch/apps/my-wordpress-site/public/wp-content/mu-plugins
# Extract TypeRocket Pro zip here
# Configure via wp-config.php
```

## Port Allocation

- **8100**: WordPress/PHP applications
- **8102**: Vite dev server (for theme development with Vite)

## Troubleshooting

**Issue:** WordPress installation fails
- Verify database exists and credentials correct
- Check WP-CLI: `wp --info --allow-root`

**Issue:** Permalinks not working
- Verify nginx configuration includes `try_files` directive
- Run: `wp rewrite flush --allow-root`

**Issue:** File permissions
- Container runs as root, files owned by root
- For local editing: `sudo chown -R $USER:$USER apps/my-wordpress-site/`

**Issue:** Xdebug not connecting
- Verify path mappings include `/public` suffix
- Check Xdebug installed: `php -m | grep xdebug`

## Next Steps

- Install Makermaker plugin for custom post types
- Setup Makerblocks for Gutenberg block development
- Configure TypeRocket for advanced custom fields
- Setup WordPress multisite
- Integrate with Mailpit for email testing (already configured)
