#!/bin/zsh
# Dynamic Project Creator - Updated for .dev domains

PROJECT_NAME="$1"
PROJECT_TYPE="$2"

if [[ -z "$PROJECT_NAME" ]]; then
    echo "Usage: $0 <project-name> [type]"
    echo ""
    echo "Types:"
    echo "  node       - Node.js/Express project"
    echo "  react      - React SPA (static build)"
    echo "  nextjs     - Next.js project"
    echo "  python     - Python/Flask project"
    echo "  django     - Django project"
    echo "  go         - Go web service"
    echo "  dotnet     - .NET Core project"
    echo "  php        - PHP project"
    echo "  laravel    - Laravel project"
    echo ""
    echo "Examples:"
    echo "  $0 my-api node"
    echo "  $0 my-site react"
    echo "  $0 my-blog laravel"
    exit 1
fi

PROJECT_DIR="./apps/$PROJECT_NAME"

if [[ -d "$PROJECT_DIR" ]]; then
    echo "❌ Project $PROJECT_NAME already exists"
    exit 1
fi

echo "🚀 Creating $PROJECT_TYPE project: $PROJECT_NAME"
mkdir -p "$PROJECT_DIR"
cd "$PROJECT_DIR"

case "$PROJECT_TYPE" in
    "node"|"express")
        npm init -y
        npm install express
        cat > index.js << 'JS'
const express = require('express');
const app = express();
const PORT = process.env.PORT || 3000;

app.get('/', (req, res) => {
    res.send(`
        <h1>🟢 Node.js Project: ${process.env.npm_package_name || 'Unknown'}</h1>
        <p>Port: ${PORT}</p>
        <p>Environment: ${process.env.NODE_ENV || 'development'}</p>
        <p>Time: ${new Date().toISOString()}</p>
        <p>Domain: ${req.get('host')}</p>
    `);
});

app.listen(PORT, '0.0.0.0', () => {
    console.log(\`Server running on port \${PORT}\`);
});
JS
        # Update package.json scripts
        node -e "
            const pkg = require('./package.json');
            pkg.scripts = { ...pkg.scripts, start: 'node index.js', dev: 'node index.js' };
            require('fs').writeFileSync('package.json', JSON.stringify(pkg, null, 2));
        "
        ;;
        
    "react")
        npm create vite@latest . -- --template react-ts
        npm install
        npm run build
        ;;
        
    "nextjs")
        npx create-next-app@latest . --typescript --tailwind --eslint --app --src-dir --import-alias "@/*"
        ;;
        
    "python"|"flask")
        cat > app.py << 'PY'
from flask import Flask, jsonify, request
import os
from datetime import datetime

app = Flask(__name__)

@app.route('/')
def hello():
    return f"""
    <h1>🐍 Python/Flask Project</h1>
    <p>Project: {os.path.basename(os.getcwd())}</p>
    <p>Port: {os.environ.get('PORT', 8000)}</p>
    <p>Time: {datetime.now().isoformat()}</p>
    <p>Host: {request.host}</p>
    """

@app.route('/api/health')
def health():
    return jsonify({
        'status': 'healthy',
        'project': os.path.basename(os.getcwd()),
        'timestamp': datetime.now().isoformat(),
        'host': request.host
    })

if __name__ == '__main__':
    port = int(os.environ.get('PORT', 8000))
    app.run(host='0.0.0.0', port=port, debug=True)
PY
        echo "flask" > requirements.txt
        ;;
        
    "django")
        pip install django
        django-admin startproject . .
        cat >> requirements.txt << 'TXT'
django>=4.2
TXT
        ;;
        
    "go")
        go mod init "$PROJECT_NAME"
        cat > main.go << 'GO'
package main

import (
    "fmt"
    "net/http"
    "os"
    "time"
)

func handler(w http.ResponseWriter, r *http.Request) {
    projectName := os.Getenv("PROJECT_NAME")
    if projectName == "" {
        projectName = "go-project"
    }
    
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    
    html := fmt.Sprintf(`
        <h1>🐹 Go Project: %s</h1>
        <p>Port: %s</p>
        <p>Time: %s</p>
        <p>Method: %s</p>
        <p>Path: %s</p>
        <p>Host: %s</p>
    `, projectName, port, time.Now().Format(time.RFC3339), r.Method, r.URL.Path, r.Host)
    
    w.Header().Set("Content-Type", "text/html")
    fmt.Fprint(w, html)
}

func main() {
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    
    http.HandleFunc("/", handler)
    fmt.Printf("Go server starting on port %s\n", port)
    http.ListenAndServe(":"+port, nil)
}
GO
        go mod tidy
        ;;
        
    "dotnet")
        dotnet new web
        # Update Program.cs for custom response
        cat > Program.cs << 'CS'
var builder = WebApplication.CreateBuilder(args);
var app = builder.Build();

app.MapGet("/", (HttpContext context) => 
{
    var projectName = System.IO.Path.GetFileName(Directory.GetCurrentDirectory());
    var port = Environment.GetEnvironmentVariable("PORT") ?? "80";
    var host = context.Request.Host.ToString();
    
    return Results.Content($@"
        <h1>🔵 .NET Project: {projectName}</h1>
        <p>Port: {port}</p>
        <p>Environment: {app.Environment.EnvironmentName}</p>
        <p>Time: {DateTime.Now:yyyy-MM-dd HH:mm:ss}</p>
        <p>Host: {host}</p>
    ", "text/html");
});

app.Run();
CS
        ;;
        
    "php")
        mkdir -p public
        cat > public/index.php << 'PHP'
<?php
$projectName = basename(dirname(__DIR__));
$port = $_ENV['PORT'] ?? '8000';
$host = $_SERVER['HTTP_HOST'] ?? 'localhost';
?>
<!DOCTYPE html>
<html>
<head>
    <title><?= $projectName ?></title>
</head>
<body>
    <h1>🐘 PHP Project: <?= $projectName ?></h1>
    <p>Port: <?= $port ?></p>
    <p>PHP Version: <?= phpversion() ?></p>
    <p>Time: <?= date('Y-m-d H:i:s') ?></p>
    <p>Host: <?= $host ?></p>
    <p>Server: <?= $_SERVER['SERVER_SOFTWARE'] ?? 'Unknown' ?></p>
</body>
</html>
PHP
        ;;
        
    "laravel")
        composer create-project laravel/laravel . --prefer-dist
        # Update welcome blade
        cat > resources/views/welcome.blade.php << 'BLADE'
<!DOCTYPE html>
<html>
<head>
    <title>{{ config('app.name') }}</title>
</head>
<body>
    <h1>🐘 Laravel Project: {{ config('app.name') }}</h1>
    <p>Environment: {{ app()->environment() }}</p>
    <p>Time: {{ now()->format('Y-m-d H:i:s') }}</p>
    <p>Laravel Version: {{ app()->version() }}</p>
    <p>Host: {{ request()->getHost() }}</p>
</body>
</html>
BLADE
        ;;
        
    *)
        echo "❓ Unknown project type: $PROJECT_TYPE"
        echo "Creating basic project structure..."
        mkdir -p public
        echo "<h1>Project: $PROJECT_NAME</h1><p>Host: \${_SERVER['HTTP_HOST'] ?? 'localhost'}</p>" > public/index.html
        ;;
esac

echo "✅ Project created: $PROJECT_NAME"
echo "🌐 Access at: https://$PROJECT_NAME.dev"  # Updated to .dev
echo "📁 Files in: $PROJECT_DIR"

# Show next steps based on project type
case "$PROJECT_TYPE" in
    "node"|"express"|"nextjs")
        echo ""
        echo "Next steps:"
        echo "  cd $PROJECT_DIR"
        echo "  npm install"
        echo "  npm run dev"
        ;;
    "python"|"flask")
        echo ""
        echo "Next steps:"
        echo "  cd $PROJECT_DIR"
        echo "  pip install -r requirements.txt"
        echo "  python app.py"
        ;;
    "django")
        echo ""
        echo "Next steps:"
        echo "  cd $PROJECT_DIR"
        echo "  pip install -r requirements.txt"
        echo "  python manage.py runserver 0.0.0.0:8000"
        ;;
    "go")
        echo ""
        echo "Next steps:"
        echo "  cd $PROJECT_DIR"
        echo "  go run main.go"
        ;;
    "laravel")
        echo ""
        echo "Next steps:"
        echo "  cd $PROJECT_DIR"
        echo "  cp .env.example .env"
        echo "  php artisan key:generate"
        echo "  php artisan serve --host=0.0.0.0 --port=8000"
        ;;
esac