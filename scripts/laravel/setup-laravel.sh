#!/bin/zsh

# Laravel setup script
# Usage: ./setup-laravel.sh <project-name> [-r <git-url>]
# Usage: ./setup-laravel.sh <project-name>  # Creates new Laravel project

# Parse arguments
PROJECT_NAME=""
GIT_URL=""

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -r|--repo)
            GIT_URL="$2"
            shift 2
            ;;
        *)
            if [[ -z "$PROJECT_NAME" ]]; then
                PROJECT_NAME="$1"
            else
                echo "❌ Unknown argument: $1"
                exit 1
            fi
            shift
            ;;
    esac
done

# Validate project name
if [[ -z "$PROJECT_NAME" ]]; then
    echo "Usage: $0 <project-name> [-r <git-url>]"
    echo ""
    echo "Examples:"
    echo "  $0 my-blog -r https://github.com/user/laravel-blog.git    # Clone existing repo"
    echo "  $0 my-new-app                                             # Create fresh Laravel project"
    exit 1
fi

echo "🚀 Setting up Laravel project: $PROJECT_NAME"

# Navigate to apps directory
cd ./apps || { echo "❌ Apps directory not found"; exit 1; }

# Check if project already exists
if [[ -d "$PROJECT_NAME" ]]; then
    echo "❌ Project $PROJECT_NAME already exists"
    exit 1
fi

# Setup project based on whether repo is provided
if [[ -n "$GIT_URL" ]]; then
    echo "📥 Cloning project from repository..."
    git clone "$GIT_URL" "$PROJECT_NAME" || { echo "❌ Failed to clone repository"; exit 1; }
else
    echo "🆕 Creating fresh Laravel project..."
    # Execute Laravel installer inside PHP container
    sudo podman exec -w /var/www/html php zsh -c "
        laravel new $PROJECT_NAME || { echo '❌ Failed to create Laravel project'; exit 1; }
    "
fi

# Navigate to project directory
cd "$PROJECT_NAME" || { echo "❌ Failed to enter project directory"; exit 1; }

# Copy .env file
echo "📝 Setting up .env file..."
if [[ -f ".env.example" ]]; then
    cp .env.example .env
elif [[ -f ".env.sample" ]]; then
    cp .env.sample .env
else
    echo "⚠️  No .env template found"
fi

# Execute remaining commands inside the PHP container
echo "🔧 Running setup commands in PHP container..."
sudo podman exec -w /var/www/html/$PROJECT_NAME php zsh -c "
# Fix permissions
echo '🔐 Setting permissions...'
chmod -R 775 /var/www/html/$PROJECT_NAME
chown -R www-data:www-data .

# Install dependencies
echo '📦 Installing Composer dependencies...'
composer install || { echo '❌ Composer install failed'; exit 1; }

echo '📦 Installing NPM dependencies...'
npm install || { echo '❌ NPM install failed'; exit 1; }

# Generate key (only if .env exists and APP_KEY is empty)
if [[ -f '.env' ]] && ! grep -q 'APP_KEY=.*[^=]' .env; then
    echo '🔑 Generating app key...'
    php artisan key:generate || { echo '❌ Key generation failed'; exit 1; }
fi

# Build assets
echo '🏗️  Building assets...'
npm run build || { echo '❌ Asset build failed'; exit 1; }

# Run migrations (only if database is configured)
if grep -q 'DB_DATABASE=.*[^=]' .env 2>/dev/null; then
    echo '🗄️  Running migrations...'
    php artisan migrate --force || echo '⚠️  Migration failed (database might not be configured)'
else
    echo 'ℹ️  Skipping migrations (no database configured)'
fi
"

echo "✅ Done! Project available at: https://${PROJECT_NAME}.test"
echo "📁 Project location: ./apps/$PROJECT_NAME"