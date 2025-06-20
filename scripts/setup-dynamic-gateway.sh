#!/bin/zsh

# =============================================================================
# DYNAMIC GATEWAY INTEGRATION SCRIPT
# =============================================================================
# Integrates the dynamic gateway into your existing microservices setup
# Handles file creation, configuration updates, and conflict resolution

# Source the central configuration
SCRIPT_DIR="$(dirname "$0")"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../" && pwd)"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

print_status() {
    local level="$1"
    local message="$2"
    
    case "$level" in
        "info")
            echo -e "${BLUE}ℹ️  $message${NC}"
            ;;
        "success")
            echo -e "${GREEN}✅ $message${NC}"
            ;;
        "warning")
            echo -e "${YELLOW}⚠️  $message${NC}"
            ;;
        "error")
            echo -e "${RED}❌ $message${NC}"
            ;;
        "step")
            echo -e "${BLUE}🔄 $message${NC}"
            ;;
    esac
}

# =============================================================================
# BACKUP FUNCTIONS
# =============================================================================

create_backup() {
    local backup_dir="$PROJECT_ROOT/backup-$(date +%Y%m%d-%H%M%S)"
    
    print_status "step" "Creating backup of current configuration..."
    mkdir -p "$backup_dir"
    
    # Backup existing files that will be modified
    [[ -f "$PROJECT_ROOT/scripts/config.sh" ]] && cp "$PROJECT_ROOT/scripts/config.sh" "$backup_dir/"
    [[ -f "$PROJECT_ROOT/config/traefik/traefik.yml" ]] && cp "$PROJECT_ROOT/config/traefik/traefik.yml" "$backup_dir/"
    [[ -d "$PROJECT_ROOT/config/node" ]] && cp -r "$PROJECT_ROOT/config/node" "$backup_dir/"
    [[ -d "$PROJECT_ROOT/config/python" ]] && cp -r "$PROJECT_ROOT/config/python" "$backup_dir/"
    [[ -d "$PROJECT_ROOT/config/go" ]] && cp -r "$PROJECT_ROOT/config/go" "$backup_dir/"
    [[ -d "$PROJECT_ROOT/config/dotnet" ]] && cp -r "$PROJECT_ROOT/config/dotnet" "$backup_dir/"
    
    echo "$backup_dir" > "$PROJECT_ROOT/.gateway-backup-location"
    print_status "success" "Backup created: $backup_dir"
}

# =============================================================================
# DIRECTORY STRUCTURE CREATION
# =============================================================================

create_directory_structure() {
    print_status "step" "Creating directory structure..."
    
    mkdir -p "$PROJECT_ROOT/compose/gateway"
    mkdir -p "$PROJECT_ROOT/config/gateway"
    
    print_status "success" "Directory structure created"
}

# =============================================================================
# FILE CREATION FUNCTIONS
# =============================================================================

create_gateway_compose() {
    print_status "step" "Creating gateway compose file..."
    
    cat > "$PROJECT_ROOT/compose/gateway/dynamic-gateway.yml" << 'EOF'
networks:
  microservices-net:
    external: true

services:
  dynamic-gateway:
    container_name: dynamic-gateway
    build:
      context: ../../config/gateway
      dockerfile: Dockerfile
    ports:
      - "8888:8888"
    volumes:
      - ../../apps:/var/www/html
      - ../../config/gateway/gateway.php:/var/www/gateway.php:ro
    networks:
      - microservices-net
    environment:
      - NODE_CONTAINER=node
      - PYTHON_CONTAINER=python
      - GO_CONTAINER=go
      - DOTNET_CONTAINER=dotnet
      - PHP_CONTAINER=php
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.dynamic-gateway.rule=HostRegexp(`^[a-zA-Z0-9-]+\\.test$$`)"
      - "traefik.http.routers.dynamic-gateway.entrypoints=websecure"
      - "traefik.http.routers.dynamic-gateway.tls=true"
      - "traefik.http.routers.dynamic-gateway.priority=10"
      - "traefik.http.services.dynamic-gateway.loadbalancer.server.port=8888"
EOF

    print_status "success" "Gateway compose file created"
}

create_gateway_dockerfile() {
    print_status "step" "Creating gateway Dockerfile..."
    
    cat > "$PROJECT_ROOT/config/gateway/Dockerfile" << 'EOF'
FROM php:8.3-cli

# Install required packages
RUN apt-get update && apt-get install -y \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Install Composer
COPY --from=composer:latest /usr/bin/composer /usr/bin/composer

WORKDIR /var/www

# Copy gateway script
COPY gateway.php /var/www/gateway.php

EXPOSE 8888

CMD ["php", "-S", "0.0.0.0:8888", "gateway.php"]
EOF

    print_status "success" "Gateway Dockerfile created"
}

create_gateway_php() {
    print_status "step" "Creating gateway router script..."
    
    cat > "$PROJECT_ROOT/config/gateway/gateway.php" << 'EOF'
<?php
/**
 * Dynamic Multi-Runtime Gateway
 * Routes requests to appropriate runtime containers or serves static files
 */

class DynamicGateway {
    private $host;
    private $projectName;
    private $requestUri;
    private $projectPath;
    
    // Container mappings
    private $containers = [
        'node' => ['host' => 'node', 'port' => 3000],
        'python' => ['host' => 'python', 'port' => 8000],
        'go' => ['host' => 'go', 'port' => 8080],
        'dotnet' => ['host' => 'dotnet', 'port' => 80],
        'php' => ['host' => 'php', 'port' => 8000]
    ];
    
    public function __construct() {
        $this->host = $_SERVER['HTTP_HOST'] ?? 'localhost';
        $this->requestUri = $_SERVER['REQUEST_URI'] ?? '/';
        $this->extractProjectName();
        $this->projectPath = "/var/www/html/{$this->projectName}";
    }
    
    private function extractProjectName() {
        if (preg_match('/^([a-zA-Z0-9-]+)\.test$/', $this->host, $matches)) {
            $this->projectName = $matches[1];
        } else {
            $this->projectName = 'default';
        }
    }
    
    public function route() {
        // Check if project exists
        if (!is_dir($this->projectPath)) {
            $this->showProjectListing();
            return;
        }
        
        // Detect project type and route accordingly
        $projectType = $this->detectProjectType();
        
        switch ($projectType) {
            case 'node_server':
                $this->proxyToContainer('node');
                break;
                
            case 'python_server':
                $this->proxyToContainer('python');
                break;
                
            case 'go_server':
                $this->proxyToContainer('go');
                break;
                
            case 'dotnet_server':
                $this->proxyToContainer('dotnet');
                break;
                
            case 'php_server':
                $this->proxyToContainer('php');
                break;
                
            case 'static':
                $this->serveStaticFiles();
                break;
                
            default:
                $this->showProjectInfo($projectType);
                break;
        }
    }
    
    private function detectProjectType() {
        $detections = [];
        
        // Node.js detection
        if (file_exists("{$this->projectPath}/package.json")) {
            $package = json_decode(file_get_contents("{$this->projectPath}/package.json"), true);
            
            // Check for server-side frameworks
            if (isset($package['scripts']['start']) || isset($package['scripts']['dev'])) {
                $dependencies = array_merge(
                    $package['dependencies'] ?? [],
                    $package['devDependencies'] ?? []
                );
                
                // Server-side frameworks
                if (isset($dependencies['next']) || 
                    isset($dependencies['express']) || 
                    isset($dependencies['fastify']) ||
                    isset($dependencies['nestjs']) ||
                    isset($dependencies['nuxt'])) {
                    $detections[] = 'node_server';
                }
            }
            
            // Check for static build output
            if (is_dir("{$this->projectPath}/build") || 
                is_dir("{$this->projectPath}/dist") || 
                is_dir("{$this->projectPath}/out")) {
                $detections[] = 'static';
            }
        }
        
        // Python detection
        if (file_exists("{$this->projectPath}/requirements.txt") || 
            file_exists("{$this->projectPath}/main.py") ||
            file_exists("{$this->projectPath}/app.py") ||
            file_exists("{$this->projectPath}/manage.py")) {
            $detections[] = 'python_server';
        }
        
        // Go detection
        if (file_exists("{$this->projectPath}/go.mod") || 
            file_exists("{$this->projectPath}/main.go")) {
            $detections[] = 'go_server';
        }
        
        // .NET detection
        if (file_exists("{$this->projectPath}/Program.cs") ||
            glob("{$this->projectPath}/*.csproj") ||
            file_exists("{$this->projectPath}/appsettings.json")) {
            $detections[] = 'dotnet_server';
        }
        
        // PHP detection
        if (file_exists("{$this->projectPath}/index.php") ||
            file_exists("{$this->projectPath}/public/index.php") ||
            file_exists("{$this->projectPath}/composer.json")) {
            $detections[] = 'php_server';
        }
        
        // Return the most likely type (prefer server over static)
        $serverTypes = ['node_server', 'python_server', 'go_server', 'dotnet_server', 'php_server'];
        foreach ($serverTypes as $type) {
            if (in_array($type, $detections)) {
                return $type;
            }
        }
        
        if (in_array('static', $detections)) {
            return 'static';
        }
        
        return 'unknown';
    }
    
    private function proxyToContainer($containerType) {
        $container = $this->containers[$containerType];
        $targetUrl = "https://{$container['host']}:{$container['port']}{$this->requestUri}";
        
        // Set up headers for the proxied request
        $headers = [
            "Host: {$this->host}",
            "X-Forwarded-For: " . ($_SERVER['REMOTE_ADDR'] ?? ''),
            "X-Forwarded-Proto: " . (isset($_SERVER['HTTPS']) ? 'https' : 'http'),
            "X-Project-Name: {$this->projectName}",
            "X-Project-Path: {$this->projectPath}"
        ];
        
        // Forward original headers
        foreach ($_SERVER as $key => $value) {
            if (strpos($key, 'HTTP_') === 0 && $key !== 'HTTP_HOST') {
                $headerName = str_replace('_', '-', substr($key, 5));
                $headers[] = "{$headerName}: {$value}";
            }
        }
        
        // Initialize cURL
        $ch = curl_init();
        curl_setopt_array($ch, [
            CURLOPT_URL => $targetUrl,
            CURLOPT_RETURNTRANSFER => false,
            CURLOPT_FOLLOWLOCATION => true,
            CURLOPT_HTTPHEADER => $headers,
            CURLOPT_WRITEFUNCTION => function($ch, $data) {
                echo $data;
                return strlen($data);
            },
            CURLOPT_HEADERFUNCTION => function($ch, $header) {
                // Forward response headers (except some proxy-specific ones)
                $headerLower = strtolower($header);
                if (!preg_match('/^(transfer-encoding|connection|keep-alive):/i', $headerLower)) {
                    header(trim($header), false);
                }
                return strlen($header);
            },
            CURLOPT_TIMEOUT => 30,
            CURLOPT_CONNECTTIMEOUT => 10
        ]);
        
        // Handle different HTTP methods
        switch ($_SERVER['REQUEST_METHOD']) {
            case 'POST':
                curl_setopt($ch, CURLOPT_POST, true);
                curl_setopt($ch, CURLOPT_POSTFIELDS, file_get_contents('php://input'));
                break;
            case 'PUT':
            case 'DELETE':
            case 'PATCH':
                curl_setopt($ch, CURLOPT_CUSTOMREQUEST, $_SERVER['REQUEST_METHOD']);
                curl_setopt($ch, CURLOPT_POSTFIELDS, file_get_contents('php://input'));
                break;
        }
        
        // Execute request
        $result = curl_exec($ch);
        $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        
        if (curl_error($ch)) {
            http_response_code(502);
            echo "<h1>Service Unavailable</h1>";
            echo "<p>Could not connect to {$containerType} container for project: {$this->projectName}</p>";
            echo "<p>Error: " . curl_error($ch) . "</p>";
            echo "<p>Make sure the {$containerType} container is running and the project is started.</p>";
            echo "<h2>Debug Info:</h2>";
            echo "<p>Target URL: {$targetUrl}</p>";
            echo "<p>Container: {$container['host']}:{$container['port']}</p>";
        } else {
            http_response_code($httpCode);
        }
        
        curl_close($ch);
    }
    
    private function serveStaticFiles() {
        // Try common static build directories
        $staticDirs = ['build', 'dist', 'out', 'public'];
        $documentRoot = null;
        
        foreach ($staticDirs as $dir) {
            if (is_dir("{$this->projectPath}/{$dir}")) {
                $documentRoot = "{$this->projectPath}/{$dir}";
                break;
            }
        }
        
        if (!$documentRoot) {
            $documentRoot = $this->projectPath;
        }
        
        $requestedFile = $documentRoot . $this->requestUri;
        
        // Handle directory requests
        if (is_dir($requestedFile)) {
            $indexFiles = ['index.html', 'index.htm', 'index.php'];
            foreach ($indexFiles as $indexFile) {
                if (file_exists($requestedFile . '/' . $indexFile)) {
                    $requestedFile = $requestedFile . '/' . $indexFile;
                    break;
                }
            }
        }
        
        // Serve the file
        if (file_exists($requestedFile) && is_file($requestedFile)) {
            $mimeType = $this->getMimeType($requestedFile);
            header("Content-Type: {$mimeType}");
            
            // Handle PHP files
            if (pathinfo($requestedFile, PATHINFO_EXTENSION) === 'php') {
                chdir(dirname($requestedFile));
                include $requestedFile;
            } else {
                readfile($requestedFile);
            }
        } else {
            http_response_code(404);
            echo "<h1>404 - File Not Found</h1>";
            echo "<p>File not found in static project: {$this->projectName}</p>";
            echo "<p>Looking for: {$this->requestUri}</p>";
            echo "<p>Document root: {$documentRoot}</p>";
        }
    }
    
    private function showProjectListing() {
        $projects = glob('/var/www/html/*', GLOB_ONLYDIR);
        
        echo "<h1>🚀 Dynamic Gateway - Available Projects</h1>";
        echo "<p>No project found for: <code>{$this->projectName}.test</code></p>";
        echo "<h2>Available Projects:</h2>";
        echo "<div style='display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; margin: 20px 0;'>";
        
        foreach ($projects as $project) {
            $projectName = basename($project);
            $projectType = $this->detectProjectTypeForPath($project);
            $icon = $this->getProjectIcon($projectType);
            
            echo "<div style='border: 1px solid #ddd; padding: 15px; border-radius: 5px;'>";
            echo "<h3>{$icon} <a href='https://{$projectName}.test'>{$projectName}</a></h3>";
            echo "<p><strong>Type:</strong> {$projectType}</p>";
            echo "<p><strong>Path:</strong> <code>{$project}</code></p>";
            echo "</div>";
        }
        
        echo "</div>";
        echo "<hr>";
        echo "<h2>🛠️ Quick Setup</h2>";
        echo "<pre># Node.js project
mkdir -p /var/www/html/my-node-app
cd /var/www/html/my-node-app
npm init -y
npm install express
echo 'const express = require(\"express\"); const app = express(); app.get(\"/\", (req,res) => res.send(\"Hello from Node!\")); app.listen(3000);' > index.js
# Visit: https://my-node-app.test

# Python project  
mkdir -p /var/www/html/my-python-app
cd /var/www/html/my-python-app
echo 'from flask import Flask; app = Flask(__name__); @app.route(\"/\"); def hello(): return \"Hello from Python!\"; app.run(host=\"0.0.0.0\", port=8000)' > app.py
echo 'flask' > requirements.txt
# Visit: https://my-python-app.test

# Static React project
mkdir -p /var/www/html/my-react-app
cd /var/www/html/my-react-app
npx create-react-app .
npm run build
# Visit: https://my-react-app.test</pre>";
    }
    
    private function detectProjectTypeForPath($path) {
        $tempPath = $this->projectPath;
        $this->projectPath = $path;
        $type = $this->detectProjectType();
        $this->projectPath = $tempPath;
        return $type;
    }
    
    private function getProjectIcon($type) {
        $icons = [
            'node_server' => '🟢',
            'python_server' => '🐍',
            'go_server' => '🐹',
            'dotnet_server' => '🔵',
            'php_server' => '🐘',
            'static' => '📄',
            'unknown' => '❓'
        ];
        return $icons[$type] ?? '❓';
    }
    
    private function showProjectInfo($projectType) {
        echo "<h1>Project: {$this->projectName}</h1>";
        echo "<p><strong>Detected Type:</strong> {$projectType}</p>";
        echo "<p><strong>Path:</strong> {$this->projectPath}</p>";
        
        if ($projectType === 'unknown') {
            echo "<h2>⚠️ Unknown Project Type</h2>";
            echo "<p>This project doesn't match any known patterns. Supported types:</p>";
            echo "<ul>";
            echo "<li><strong>Node.js:</strong> package.json with start/dev scripts</li>";
            echo "<li><strong>Python:</strong> requirements.txt, main.py, app.py, or manage.py</li>";
            echo "<li><strong>Go:</strong> go.mod or main.go</li>";
            echo "<li><strong>.NET:</strong> *.csproj, Program.cs, or appsettings.json</li>";
            echo "<li><strong>PHP:</strong> index.php or composer.json</li>";
            echo "<li><strong>Static:</strong> build/, dist/, or out/ directories</li>";
            echo "</ul>";
        }
        
        echo "<h2>📁 Project Contents:</h2>";
        $this->showDirectoryListing($this->projectPath);
    }
    
    private function showDirectoryListing($path, $level = 0) {
        if ($level > 2) return; // Limit depth
        
        $items = scandir($path);
        echo "<ul>";
        foreach ($items as $item) {
            if ($item === '.' || $item === '..') continue;
            $fullPath = $path . '/' . $item;
            $indent = str_repeat('  ', $level);
            
            if (is_dir($fullPath)) {
                echo "<li>📁 {$item}/</li>";
                if ($level < 2) {
                    $this->showDirectoryListing($fullPath, $level + 1);
                }
            } else {
                echo "<li>📄 {$item}</li>";
            }
        }
        echo "</ul>";
    }
    
    private function getMimeType($file) {
        $mimeTypes = [
            'html' => 'text/html',
            'htm' => 'text/html',
            'css' => 'text/css',
            'js' => 'application/javascript',
            'json' => 'application/json',
            'png' => 'image/png',
            'jpg' => 'image/jpeg',
            'jpeg' => 'image/jpeg',
            'gif' => 'image/gif',
            'svg' => 'image/svg+xml',
            'ico' => 'image/x-icon',
            'txt' => 'text/plain',
            'pdf' => 'application/pdf',
            'woff' => 'font/woff',
            'woff2' => 'font/woff2',
            'ttf' => 'font/ttf',
            'eot' => 'application/vnd.ms-fontobject'
        ];
        
        $extension = strtolower(pathinfo($file, PATHINFO_EXTENSION));
        return $mimeTypes[$extension] ?? 'application/octet-stream';
    }
}

// Initialize and route the request
$gateway = new DynamicGateway();
$gateway->route();
?>
EOF

    print_status "success" "Gateway router script created"
}

create_node_project_handler() {
    print_status "step" "Creating Node.js project handler..."
    
    cat > "$PROJECT_ROOT/config/node/project-handler.js" << 'EOF'
const express = require('express');
const path = require('path');
const fs = require('fs');
const { spawn } = require('child_process');

const app = express();
const PORT = 3000;

// Store running projects
const runningProjects = new Map();

// Middleware to handle project routing
app.use((req, res, next) => {
    const projectName = req.headers['x-project-name'];
    const projectPath = req.headers['x-project-path'];
    
    if (projectName && projectPath && fs.existsSync(projectPath)) {
        const packageJsonPath = path.join(projectPath, 'package.json');
        
        if (fs.existsSync(packageJsonPath)) {
            try {
                const packageJson = JSON.parse(fs.readFileSync(packageJsonPath, 'utf8'));
                
                // Check if it's a server project
                if (packageJson.scripts && (packageJson.scripts.start || packageJson.scripts.dev)) {
                    // Try to start project if not already running
                    if (!runningProjects.has(projectName)) {
                        startProject(projectName, projectPath, packageJson);
                    }
                    
                    // Proxy to the project
                    const projectPort = 3000 + Math.abs(hashCode(projectName) % 1000);
                    proxyToProject(req, res, projectPort);
                    return;
                }
            } catch (error) {
                console.error('Error processing project:', error);
            }
        }
    }
    
    next();
});

function hashCode(str) {
    let hash = 0;
    for (let i = 0; i < str.length; i++) {
        const char = str.charCodeAt(i);
        hash = ((hash << 5) - hash) + char;
        hash = hash & hash; // Convert to 32-bit integer
    }
    return hash;
}

function startProject(projectName, projectPath, packageJson) {
    const projectPort = 3000 + Math.abs(hashCode(projectName) % 1000);
    
    // Determine start command
    let command, args;
    if (packageJson.scripts.dev) {
        command = 'npm';
        args = ['run', 'dev'];
    } else if (packageJson.scripts.start) {
        command = 'npm';
        args = ['start'];
    } else {
        return;
    }
    
    // Set environment variables
    const env = {
        ...process.env,
        PORT: projectPort.toString(),
        NODE_ENV: 'development'
    };
    
    // Start the project
    const child = spawn(command, args, {
        cwd: projectPath,
        env: env,
        stdio: 'pipe'
    });
    
    child.stdout.on('data', (data) => {
        console.log(`[${projectName}] ${data}`);
    });
    
    child.stderr.on('data', (data) => {
        console.error(`[${projectName}] ${data}`);
    });
    
    child.on('close', (code) => {
        console.log(`[${projectName}] Process exited with code ${code}`);
        runningProjects.delete(projectName);
    });
    
    runningProjects.set(projectName, { child, port: projectPort });
    console.log(`Started project ${projectName} on port ${projectPort}`);
}

function proxyToProject(req, res, port) {
    // Simple proxy implementation
    const http = require('http');
    const options = {
        hostname: 'localhost',
        port: port,
        path: req.url,
        method: req.method,
        headers: req.headers
    };
    
    const proxyReq = http.request(options, (proxyRes) => {
        res.writeHead(proxyRes.statusCode, proxyRes.headers);
        proxyRes.pipe(res);
    });
    
    proxyReq.on('error', (err) => {
        res.writeHead(502);
        res.end(`<h1>Project Starting...</h1><p>Please wait while the project starts up, then refresh.</p><p>Error: ${err.message}</p>`);
    });
    
    req.pipe(proxyReq);
}

// Fallback handler
app.use('*', (req, res) => {
    res.send(`
        <h1>Node.js Project Handler</h1>
        <p>Project: ${req.headers['x-project-name'] || 'Unknown'}</p>
        <p>Path: ${req.headers['x-project-path'] || 'Unknown'}</p>
        <p>Running Projects: ${Array.from(runningProjects.keys()).join(', ') || 'None'}</p>
        <p>This project either hasn't started yet or doesn't have a proper entry point.</p>
        <p>Make sure your package.json has a "start" or "dev" script.</p>
    `);
});

app.listen(PORT, '0.0.0.0', () => {
    console.log(`Node.js project handler listening on port ${PORT}`);
});
EOF

    print_status "success" "Node.js project handler created"
}

# =============================================================================
# CONFIGURATION UPDATE FUNCTIONS
# =============================================================================

update_config_sh() {
    print_status "step" "Updating scripts/config.sh..."
    
    local config_file="$PROJECT_ROOT/scripts/config.sh"
    
    if [[ ! -f "$config_file" ]]; then
        print_status "error" "config.sh not found at $config_file"
        return 1
    fi
    
    # Check if gateway is already added
    if grep -q "gateway.*dynamic-gateway.yml" "$config_file"; then
        print_status "warning" "Gateway already exists in config.sh"
        return 0
    fi
    
    # Create temporary file with updated config
    local temp_file=$(mktemp)
    
    # Add gateway to SERVICE_CATEGORIES
    sed '/SERVICE_CATEGORIES=(/,/)/{
        /\[proxy\]/i\
    [gateway]="dynamic-gateway.yml"
    }' "$config_file" > "$temp_file"
    
    # Add gateway to SERVICE_STARTUP_ORDER
    sed -i '/SERVICE_STARTUP_ORDER=(/,/)/{
        /"management"/a\
    "gateway"
    }' "$temp_file"
    
    # Replace original file
    mv "$temp_file" "$config_file"
    
    print_status "success" "config.sh updated with gateway service"
}

update_backend_containers() {
    print_status "step" "Updating backend container configurations..."
    
    # Update Node.js Dockerfile
    if [[ -f "$PROJECT_ROOT/config/node/Dockerfile" ]]; then
        if ! grep -q "project-handler.js" "$PROJECT_ROOT/config/node/Dockerfile"; then
            cat >> "$PROJECT_ROOT/config/node/Dockerfile" << 'EOF'

# Add project handler for dynamic routing
COPY project-handler.js /var/www/project-handler.js

# Use the project handler as entry point  
CMD ["node", "/var/www/project-handler.js"]
EOF
            print_status "success" "Node.js Dockerfile updated"
        fi
    fi
    
    # Update Python container to handle project headers
    if [[ -f "$PROJECT_ROOT/config/python/Dockerfile" ]]; then
        if ! grep -q "X-Project-Name" "$PROJECT_ROOT/config/python/Dockerfile"; then
            cat >> "$PROJECT_ROOT/config/python/Dockerfile" << 'EOF'

# Add support for project detection headers
ENV PYTHONPATH=/var/www/html
EOF
            print_status "success" "Python Dockerfile updated"
        fi
    fi
}

update_traefik_priority() {
    print_status "step" "Updating Traefik configuration for gateway priority..."
    
    local traefik_config="$PROJECT_ROOT/config/traefik/traefik.yml"
    
    if [[ -f "$traefik_config" ]]; then
        # The gateway will have priority=10, so existing services should have lower priority
        # Most existing services should work fine with default priority (0)
        print_status "success" "Traefik priority handled by gateway configuration"
    fi
}

create_project_creation_script() {
    print_status "step" "Creating project creation helper script..."
    
    cat > "$PROJECT_ROOT/scripts/create-project.sh" << 'EOF'
#!/bin/zsh
# Dynamic Project Creator

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
        npx create-react-app . --template typescript
        npm run build
        ;;
        
    "nextjs")
        npx create-next-app@latest . --typescript --tailwind --eslint --app --src-dir --import-alias "@/*"
        # Update next.config.js for dynamic routing
        cat > next.config.js << 'JS'
/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone',
  experimental: {
    appDir: true,
  },
}

module.exports = nextConfig
JS
        ;;
        
    "python"|"flask")
        cat > app.py << 'PY'
from flask import Flask, jsonify
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
    """

@app.route('/api/health')
def health():
    return jsonify({
        'status': 'healthy',
        'project': os.path.basename(os.getcwd()),
        'timestamp': datetime.now().isoformat()
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
    `, projectName, port, time.Now().Format(time.RFC3339), r.Method, r.URL.Path)
    
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

app.MapGet("/", () => 
{
    var projectName = System.IO.Path.GetFileName(Directory.GetCurrentDirectory());
    var port = Environment.GetEnvironmentVariable("PORT") ?? "80";
    
    return Results.Content($@"
        <h1>🔵 .NET Project: {projectName}</h1>
        <p>Port: {port}</p>
        <p>Environment: {app.Environment.EnvironmentName}</p>
        <p>Time: {DateTime.Now:yyyy-MM-dd HH:mm:ss}</p>
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
</body>
</html>
BLADE
        ;;
        
    *)
        echo "❓ Unknown project type: $PROJECT_TYPE"
        echo "Creating basic project structure..."
        echo "<h1>Project: $PROJECT_NAME</h1>" > index.html
        ;;
esac

echo "✅ Project created: $PROJECT_NAME"
echo "🌐 Access at: https://$PROJECT_NAME.test"
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
EOF

    chmod +x "$PROJECT_ROOT/scripts/create-project.sh"
    print_status "success" "Project creation script created"
}

# =============================================================================
# TESTING FUNCTIONS
# =============================================================================

create_test_projects() {
    print_status "step" "Creating test projects for validation..."
    
    # Create a simple test project for each type
    local test_dir="$PROJECT_ROOT/apps"
    
    # Node.js test
    if [[ ! -d "$test_dir/test-node" ]]; then
        mkdir -p "$test_dir/test-node"
        cat > "$test_dir/test-node/package.json" << 'EOF'
{
  "name": "test-node",
  "version": "1.0.0",
  "scripts": {
    "start": "node index.js",
    "dev": "node index.js"
  },
  "dependencies": {
    "express": "^4.18.0"
  }
}
EOF
        cat > "$test_dir/test-node/index.js" << 'EOF'
const express = require('express');
const app = express();
const PORT = process.env.PORT || 3000;

app.get('/', (req, res) => {
    res.send('<h1>🟢 Test Node.js Project</h1><p>Gateway routing working!</p>');
});

app.listen(PORT, '0.0.0.0', () => {
    console.log(`Test Node.js app running on port ${PORT}`);
});
EOF
        print_status "success" "Created test Node.js project"
    fi
    
    # Python test
    if [[ ! -d "$test_dir/test-python" ]]; then
        mkdir -p "$test_dir/test-python"
        cat > "$test_dir/test-python/app.py" << 'EOF'
from flask import Flask
import os

app = Flask(__name__)

@app.route('/')
def hello():
    return '<h1>🐍 Test Python Project</h1><p>Gateway routing working!</p>'

if __name__ == '__main__':
    port = int(os.environ.get('PORT', 8000))
    app.run(host='0.0.0.0', port=port)
EOF
        echo "flask" > "$test_dir/test-python/requirements.txt"
        print_status "success" "Created test Python project"
    fi
    
    # Static test
    if [[ ! -d "$test_dir/test-static" ]]; then
        mkdir -p "$test_dir/test-static/build"
        cat > "$test_dir/test-static/build/index.html" << 'EOF'
<!DOCTYPE html>
<html>
<head><title>Test Static Project</title></head>
<body>
    <h1>📄 Test Static Project</h1>
    <p>Gateway static serving working!</p>
</body>
</html>
EOF
        print_status "success" "Created test static project"
    fi
}

# =============================================================================
# VALIDATION FUNCTIONS
# =============================================================================

validate_installation() {
    print_status "step" "Validating installation..."
    
    # Check if all required files were created
    local errors=0
    
    local required_files=(
        "$PROJECT_ROOT/compose/gateway/dynamic-gateway.yml"
        "$PROJECT_ROOT/config/gateway/Dockerfile"
        "$PROJECT_ROOT/config/gateway/gateway.php"
        "$PROJECT_ROOT/config/node/project-handler.js"
        "$PROJECT_ROOT/scripts/create-project.sh"
    )
    
    for file in "${required_files[@]}"; do
        if [[ ! -f "$file" ]]; then
            print_status "error" "Missing file: $file"
            ((errors++))
        fi
    done
    
    # Check if config.sh was updated
    if ! grep -q "gateway.*dynamic-gateway.yml" "$PROJECT_ROOT/scripts/config.sh"; then
        print_status "error" "config.sh was not updated properly"
        ((errors++))
    fi
    
    if [[ $errors -eq 0 ]]; then
        print_status "success" "Installation validation passed"
        return 0
    else
        print_status "error" "Installation validation failed with $errors errors"
        return 1
    fi
}

# =============================================================================
# ROLLBACK FUNCTIONS
# =============================================================================

create_rollback_script() {
    print_status "step" "Creating rollback script..."
    
    local backup_location
    if [[ -f "$PROJECT_ROOT/.gateway-backup-location" ]]; then
        backup_location=$(cat "$PROJECT_ROOT/.gateway-backup-location")
    else
        print_status "warning" "No backup location found"
        return 1
    fi
    
    cat > "$PROJECT_ROOT/scripts/rollback-gateway.sh" << EOF
#!/bin/zsh
# Rollback Dynamic Gateway Installation

echo "🔄 Rolling back dynamic gateway installation..."

# Stop gateway container
echo "Stopping gateway container..."
$PROJECT_ROOT/scripts/service-manager.sh down dynamic-gateway 2>/dev/null || true

# Remove gateway files
echo "Removing gateway files..."
rm -rf "$PROJECT_ROOT/compose/gateway"
rm -rf "$PROJECT_ROOT/config/gateway"
rm -f "$PROJECT_ROOT/config/node/project-handler.js"
rm -f "$PROJECT_ROOT/scripts/create-project.sh"
rm -f "$PROJECT_ROOT/scripts/rollback-gateway.sh"

# Restore original files
if [[ -d "$backup_location" ]]; then
    echo "Restoring original files from $backup_location..."
    [[ -f "$backup_location/config.sh" ]] && cp "$backup_location/config.sh" "$PROJECT_ROOT/scripts/"
    [[ -f "$backup_location/traefik.yml" ]] && cp "$backup_location/traefik.yml" "$PROJECT_ROOT/config/traefik/"
    [[ -d "$backup_location/node" ]] && cp -r "$backup_location/node" "$PROJECT_ROOT/config/"
    [[ -d "$backup_location/python" ]] && cp -r "$backup_location/python" "$PROJECT_ROOT/config/"
    [[ -d "$backup_location/go" ]] && cp -r "$backup_location/go" "$PROJECT_ROOT/config/"
    [[ -d "$backup_location/dotnet" ]] && cp -r "$backup_location/dotnet" "$PROJECT_ROOT/config/"
fi

# Remove test projects
rm -rf "$PROJECT_ROOT/apps/test-node"
rm -rf "$PROJECT_ROOT/apps/test-python"
rm -rf "$PROJECT_ROOT/apps/test-static"

# Clean up
rm -f "$PROJECT_ROOT/.gateway-backup-location"

echo "✅ Rollback completed!"
echo "ℹ️  You may need to restart your services: ./scripts/start-services.sh"
EOF

    chmod +x "$PROJECT_ROOT/scripts/rollback-gateway.sh"
    print_status "success" "Rollback script created"
}

# =============================================================================
# MAIN EXECUTION FUNCTIONS
# =============================================================================

show_completion_message() {
    print_status "success" "Dynamic Gateway Integration Complete!"
    echo ""
    echo "========================================================"
    echo "🚀 DYNAMIC GATEWAY SUCCESSFULLY INSTALLED"
    echo "========================================================"
    echo ""
    echo "✅ What was installed:"
    echo "   • Dynamic gateway service (routes all *.test domains)"
    echo "   • Smart project detection (Node.js, Python, Go, .NET, PHP, Static)"
    echo "   • Enhanced backend containers with project handling"
    echo "   • Project creation helper script"
    echo "   • Test projects for validation"
    echo ""
    echo "🛠️  Next Steps:"
    echo ""
    echo "1. Start the gateway:"
    echo "   ./scripts/service-manager.sh up dynamic-gateway"
    echo ""
    echo "2. Test with existing projects:"
    echo "   https://test-node.test      (Node.js test)"
    echo "   https://test-python.test    (Python test)"
    echo "   https://test-static.test    (Static test)"
    echo ""
    echo "3. Create new projects:"
    echo "   ./scripts/create-project.sh my-api node"
    echo "   ./scripts/create-project.sh my-site react"
    echo "   ./scripts/create-project.sh my-blog laravel"
    echo ""
    echo "🔧 Management Commands:"
    echo "   • Start gateway:    ./scripts/service-manager.sh up dynamic-gateway"
    echo "   • Stop gateway:     ./scripts/service-manager.sh down dynamic-gateway"
    echo "   • View logs:        ./scripts/service-manager.sh logs dynamic-gateway"
    echo "   • Create project:   ./scripts/create-project.sh <name> <type>"
    echo "   • Rollback:         ./scripts/rollback-gateway.sh"
    echo ""
    echo "📁 Project Structure:"
    echo "   Drop any project in: ./apps/project-name/"
    echo "   Access at: https://project-name.test"
    echo ""
    echo "🆘 If something goes wrong:"
    echo "   ./scripts/rollback-gateway.sh"
    echo ""
    echo "========================================================"
}

# =============================================================================
# MAIN EXECUTION
# =============================================================================

main() {
    echo "========================================================"
    echo "🚀 DYNAMIC GATEWAY INTEGRATION SCRIPT"
    echo "========================================================"
    echo ""
    print_status "info" "This script will integrate the dynamic gateway into your microservices setup"
    echo ""
    
    # Check if we're in the right directory
    if [[ ! -f "$PROJECT_ROOT/scripts/config.sh" ]]; then
        print_status "error" "This script must be run from the project root or scripts directory"
        print_status "info" "Looking for: $PROJECT_ROOT/scripts/config.sh"
        exit 1
    fi
    
    # Ask for confirmation
    read -q "REPLY?Do you want to proceed with the installation? (y/N): "
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_status "info" "Installation cancelled"
        exit 0
    fi
    
    echo ""
    print_status "step" "Starting dynamic gateway integration..."
    
    # Run installation steps
    create_backup || exit 1
    create_directory_structure || exit 1
    create_gateway_compose || exit 1
    create_gateway_dockerfile || exit 1
    create_gateway_php || exit 1
    create_node_project_handler || exit 1
    update_config_sh || exit 1
    update_backend_containers || exit 1
    update_traefik_priority || exit 1
    create_project_creation_script || exit 1
    create_test_projects || exit 1
    create_rollback_script || exit 1
    
    # Validate installation
    if validate_installation; then
        show_completion_message
        exit 0
    else
        print_status "error" "Installation validation failed"
        print_status "info" "You can rollback using: ./scripts/rollback-gateway.sh"
        exit 1
    fi
}

# Run main function
main "$@"