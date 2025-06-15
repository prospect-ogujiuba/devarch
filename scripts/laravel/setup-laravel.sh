#!/bin/zsh

# Simple Laravel setup script
# Usage: ./setup-laravel.sh <git-url> [project-name]

if [[ -z "$1" ]]; then
    echo "Usage: $0 <git-url> [project-name]"
    exit 1
fi

GIT_URL="$1"
PROJECT_NAME="${2:-$(basename "$1" .git)}"

echo "🚀 Setting up Laravel project: $PROJECT_NAME"

# Clone the project
echo "📥 Cloning project..."
cd /var/www/html
git clone "$GIT_URL" "$PROJECT_NAME"
cd "$PROJECT_NAME"

# Copy .env file
echo "📝 Setting up .env file..."
if [[ -f ".env.example" ]]; then
    cp .env.example .env
elif [[ -f ".env.sample" ]]; then
    cp .env.sample .env
else
    echo "⚠️  No .env template found"
fi

# Fix permissions
echo "🔐 Setting permissions..."
chmod -R 775 storage bootstrap/cache
chown -R www-data:www-data .

# Install dependencies
echo "📦 Installing Composer dependencies..."
composer install

echo "📦 Installing NPM dependencies..."
npm install

# Generate key
echo "🔑 Generating app key..."
php artisan key:generate

# Build assets
echo "🏗️  Building assets..."
npm run build

# Run migrations
echo "🗄️  Running migrations..."
php artisan migrate

echo "✅ Done! Project available at: https://${PROJECT_NAME}.test"