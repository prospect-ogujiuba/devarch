#!/bin/bash

echo "🔍 Python Smart Entrypoint: Detecting project structure..."

# Function to detect and setup Python projects
detect_python_projects() {
    for app_dir in /app/*/; do
        if [ -d "$app_dir" ]; then
            app_name=$(basename "$app_dir")
            echo "📁 Found app: $app_name"
            
            cd "$app_dir"
            
            # Poetry Detection
            if [ -f "pyproject.toml" ] && grep -q "tool.poetry" pyproject.toml 2>/dev/null; then
                echo "📖 Poetry project detected in $app_name"
                poetry install --no-dev
                
            # Pipenv Detection
            elif [ -f "Pipfile" ]; then
                echo "🚰 Pipenv project detected in $app_name"
                pipenv install --system --deploy
                
            # Requirements.txt Detection
            elif [ -f "requirements.txt" ]; then
                echo "📋 Requirements.txt project detected in $app_name"
                pip install --no-cache-dir -r requirements.txt
                
            # Django Detection
            elif [ -f "manage.py" ]; then
                echo "🎸 Django project detected in $app_name"
                # Install requirements if they exist
                [ -f "requirements.txt" ] && pip install --no-cache-dir -r requirements.txt
                # Run Django setup commands
                python manage.py collectstatic --noinput 2>/dev/null || true
                python manage.py migrate --noinput 2>/dev/null || true
                
            # FastAPI Detection (main.py with FastAPI imports)
            elif [ -f "main.py" ] && grep -q "from fastapi" main.py 2>/dev/null; then
                echo "🚀 FastAPI project detected in $app_name"
                [ -f "requirements.txt" ] && pip install --no-cache-dir -r requirements.txt
                
            # Flask Detection (app.py or main.py with Flask imports)
            elif ([ -f "app.py" ] && grep -q "from flask" app.py 2>/dev/null) || \
                 ([ -f "main.py" ] && grep -q "from flask" main.py 2>/dev/null); then
                echo "🌶️  Flask project detected in $app_name"
                [ -f "requirements.txt" ] && pip install --no-cache-dir -r requirements.txt
                
            # Streamlit Detection
            elif find . -name "*.py" -exec grep -l "import streamlit" {} \; | head -1 >/dev/null 2>&1; then
                echo "📊 Streamlit project detected in $app_name"
                [ -f "requirements.txt" ] && pip install --no-cache-dir -r requirements.txt
                
            # Generic Python project with setup.py
            elif [ -f "setup.py" ]; then
                echo "⚙️  Setup.py project detected in $app_name"
                pip install -e .
                
            # Standalone Python files
            elif ls *.py >/dev/null 2>&1; then
                echo "🐍 Python files detected in $app_name"
                [ -f "requirements.txt" ] && pip install --no-cache-dir -r requirements.txt
                
            else
                echo "❓ No recognizable Python project structure in $app_name"
            fi
            
            # Set proper permissions
            chown -R app:app "$app_dir"
        fi
    done
}

# Run detection
detect_python_projects

echo "✅ Python Smart Entrypoint: Setup complete!"

# If we're in a specific app directory, use it
if [ -f "/app/main.py" ] || [ -f "/app/app.py" ] || [ -f "/app/manage.py" ]; then
    cd /app
elif [ -d "/app" ]; then
    # Find the first app with Python files and cd into it
    for app_dir in /app/*/; do
        if [ -f "$app_dir/main.py" ] || [ -f "$app_dir/app.py" ] || [ -f "$app_dir/manage.py" ]; then
            cd "$app_dir"
            echo "🎯 Switching to $app_dir for execution"
            break
        fi
    done
fi

# Switch to app user for security
exec su-exec app "$@" 2>/dev/null || exec "$@"