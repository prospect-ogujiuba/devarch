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
        $targetUrl = "http://{$container['host']}:{$container['port']}{$this->requestUri}";
        
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
            echo "<h3>{$icon} <a href='http://{$projectName}.test'>{$projectName}</a></h3>";
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
# Visit: http://my-node-app.test

# Python project  
mkdir -p /var/www/html/my-python-app
cd /var/www/html/my-python-app
echo 'from flask import Flask; app = Flask(__name__); @app.route(\"/\"); def hello(): return \"Hello from Python!\"; app.run(host=\"0.0.0.0\", port=8000)' > app.py
echo 'flask' > requirements.txt
# Visit: http://my-python-app.test

# Static React project
mkdir -p /var/www/html/my-react-app
cd /var/www/html/my-react-app
npx create-react-app .
npm run build
# Visit: http://my-react-app.test</pre>";
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

<?php phpinfo(); ?>