# PHPStorm - Laravel Quick Start

Create Laravel projects in DevArch using PHPStorm with container-based PHP interpreter.

## Prerequisites

- PHPStorm installed
- DevArch backend containers running:
  ```bash
  ./scripts/service-manager.sh start database proxy backend
  ```

## 1. Create Project

**Option A: New Laravel Project via Composer**

1. Open PHPStorm → File → New Project
2. Select "PHP" from left sidebar
3. **Location:** `/home/fhcadmin/projects/devarch/apps/my-laravel-app`
4. Click "..." next to "PHP Interpreter"
   - Click "+" → "From Docker, Vagrant, VM, WSL, Remote..."
   - Select "Docker Compose"
   - Configuration file: `/home/fhcadmin/projects/devarch/compose/backend/php.yml`
   - Service: `php`
   - Click "OK"
5. Under "Composer", select "Install dependencies"
6. Click "Create"

**Option B: Use Container's Laravel Installer**

1. Execute in PHP container:
   ```bash
   podman exec -it php bash
   cd /var/www/html
   laravel new my-laravel-app
   exit
   ```
2. PHPStorm → File → Open → `/home/fhcadmin/projects/devarch/apps/my-laravel-app`

## 2. Ensure `public/` Structure

Laravel already uses `public/` by default. Verify:

```
apps/my-laravel-app/
├── public/           # ✓ Already exists - web server root
│   ├── index.php     # Entry point
│   └── .htaccess
├── app/
├── config/
├── routes/
└── ...
```

**No changes needed** - Laravel follows DevArch standard.

## 3. Configure PHP Interpreter (Container)

1. PHPStorm → Settings → PHP
2. CLI Interpreter: Click "..."
3. If not already set:
   - Click "+" → "From Docker, Vagrant, VM, WSL, Remote..."
   - Docker Compose
   - Configuration file: `/home/fhcadmin/projects/devarch/compose/backend/php.yml`
   - Service: `php`
   - Lifecycle: Connect to existing container
   - Click "OK"
4. Verify PHP 8.3 detected
5. Click "OK"

## 4. Configure Xdebug

1. Settings → PHP → Debug
2. Xdebug port: `9003`
3. Enable "Break at first line in PHP scripts" (test only)
4. Settings → PHP → Servers
   - Click "+"
   - Name: `my-laravel-app`
   - Host: `my-laravel-app.test`
   - Port: `443`
   - Debugger: Xdebug
   - Use path mappings: ✓
   - Map project root → `/var/www/html/my-laravel-app`
   - Click "OK"

## 5. Database Configuration

1. Copy `.env.example` → `.env`
2. Update database credentials:
   ```env
   DB_CONNECTION=mysql
   DB_HOST=mariadb
   DB_PORT=3306
   DB_DATABASE=my_laravel_app
   DB_USERNAME=root
   DB_PASSWORD=admin1234567
   ```
3. Create database:
   ```bash
   podman exec -it php bash
   cd /var/www/html/my-laravel-app
   php artisan migrate
   ```

**PHPStorm Database Tool:**
1. View → Tool Windows → Database
2. Click "+" → Data Source → MySQL
3. Host: `localhost`, Port: `3306`
4. User: `root`, Password: `admin1234567`
5. Database: `my_laravel_app`
6. Test Connection → OK

## 6. Configure nginx-proxy-manager

1. Open http://localhost:81
2. Login: `admin@devarch.test` / `admin1234567`
3. Proxy Hosts → Add Proxy Host
   - Domain Names: `my-laravel-app.test`
   - Scheme: `http`
   - Forward Hostname/IP: `php`
   - Forward Port: `8000`
   - Block Common Exploits: ✓
   - Websockets Support: ✓
4. SSL tab:
   - SSL Certificate: Request New SSL Certificate
   - Force SSL: ✓
   - HTTP/2 Support: ✓
   - Click "Save"

## 7. Update /etc/hosts

```bash
sudo sh -c 'echo "127.0.0.1 my-laravel-app.test" >> /etc/hosts'
```

Or run:
```bash
sudo ./scripts/update-hosts.sh
```

## 8. Development Workflow

**Start Laravel Dev Server (in container):**

1. PHPStorm → Tools → Start SSH Session → php container
2. In terminal:
   ```bash
   cd /var/www/html/my-laravel-app
   php artisan serve --host=0.0.0.0 --port=8000
   ```

**Or create Run Configuration:**

1. Run → Edit Configurations → "+" → PHP Script
2. File: `/var/www/html/my-laravel-app/artisan`
3. Arguments: `serve --host=0.0.0.0 --port=8000`
4. Interpreter: Select container interpreter
5. Click "OK"

**Access:** https://my-laravel-app.test

## 9. Debugging

1. Set breakpoint in `routes/web.php` or controller
2. PHPStorm → Run → Start Listening for PHP Debug Connections
3. Browser: Install Xdebug browser extension
4. Enable Xdebug session
5. Refresh page in browser
6. Debugger should pause at breakpoint

**Container Xdebug Configuration (already in PHP container):**
```ini
; /usr/local/etc/php/conf.d/docker-php-ext-xdebug.ini
xdebug.mode=debug
xdebug.client_host=host.docker.internal
xdebug.client_port=9003
xdebug.start_with_request=yes
```

## 10. Artisan Commands

Use PHPStorm's built-in Artisan support:

1. Tools → Laravel → Run Artisan Command
2. Or terminal: `php artisan {command}`

Common commands:
```bash
php artisan migrate
php artisan make:controller MyController
php artisan make:model MyModel -m
php artisan queue:work
php artisan cache:clear
```

## Port Allocation

- **8100**: PHP applications (Laravel, custom PHP)
- **8102**: Vite dev server (for Laravel + Vite)

## Troubleshooting

**Issue:** Composer dependencies not installing
- Solution: Ensure PHP container has network access, run `composer install` manually in container

**Issue:** Xdebug not connecting
- Verify Xdebug extension loaded: `php -m | grep xdebug`
- Check IDE listening: Run → Start Listening for PHP Debug Connections
- Verify path mappings in Server configuration

**Issue:** Database connection refused
- Ensure database container running: `./scripts/service-manager.sh status database`
- Use container name (`mariadb`) not `localhost` in `.env`

## Next Steps

- Setup Laravel queues with Redis
- Configure Laravel Horizon for queue monitoring
- Add Telescope for debugging
- Setup Laravel Mix/Vite for frontend assets
