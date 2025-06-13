#!/bin/bash

echo "🔍 PHP Smart Entrypoint: Detecting project structure..."

# Function to detect and setup PHP projects
detect_php_projects() {
    for app_dir in /var/www/html/*/; do
        if [ -d "$app_dir" ]; then
            app_name=$(basename "$app_dir")
            echo "📁 Found app: $app_name"
            
            cd "$app_dir"
            
            # Laravel Detection
            if [ -f "artisan" ] && [ -f "composer.json" ]; then
                echo "🚀 Laravel project detected in $app_name"
                composer install --no-dev --optimize-autoloader
                php artisan config:cache 2>/dev/null || true
                php artisan route:cache 2>/dev/null || true
                php artisan view:cache 2>/dev/null || true
                
            # WordPress Detection
            elif [ -f "wp-config.php" ] || [ -f "wp-config-sample.php" ]; then
                echo "📝 WordPress project detected in $app_name"
                # WordPress doesn't need build steps, just ensure permissions
                
            # Symfony Detection
            elif [ -f "bin/console" ] && [ -f "composer.json" ]; then
                echo "🎵 Symfony project detected in $app_name"
                composer install --no-dev --optimize-autoloader
                php bin/console cache:clear --env=prod 2>/dev/null || true
                
            # CodeIgniter Detection
            elif [ -f "spark" ] && [ -f "composer.json" ]; then
                echo "🔥 CodeIgniter project detected in $app_name"
                composer install --no-dev --optimize-autoloader
                
            # Generic Composer Project
            elif [ -f "composer.json" ]; then
                echo "📦 Composer project detected in $app_name"
                composer install --no-dev --optimize-autoloader
                
            # Static PHP/HTML Project
            elif [ -f "index.php" ] || [ -f "index.html" ]; then
                echo "📄 Static PHP/HTML project detected in $app_name"
                # No build steps needed
                
            else
                echo "❓ Unknown PHP project structure in $app_name"
            fi
            
            # Set proper permissions
            chown -R www-data:www-data "$app_dir"
            find "$app_dir" -type d -exec chmod 755 {} \;
            find "$app_dir" -type f -exec chmod 644 {} \;
            
            # Make special files executable
            [ -f "artisan" ] && chmod +x artisan
            [ -f "bin/console" ] && chmod +x bin/console
            [ -f "spark" ] && chmod +x spark
        fi
    done
}

# Run detection
detect_php_projects

echo "✅ PHP Smart Entrypoint: Setup complete!"

# Execute the original command
exec "$@"