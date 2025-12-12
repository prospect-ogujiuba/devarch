<?php
/**
 * DevArch API Router
 * Entry point for all API requests
 */

// CORS headers
header('Access-Control-Allow-Origin: *');
header('Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS');
header('Access-Control-Allow-Headers: Content-Type, Authorization');
header('Content-Type: application/json');

// Handle OPTIONS preflight
if ($_SERVER['REQUEST_METHOD'] === 'OPTIONS') {
    http_response_code(204);
    exit;
}

// Get request method and path
$method = $_SERVER['REQUEST_METHOD'];
$path = parse_url($_SERVER['REQUEST_URI'], PHP_URL_PATH);

// Route to appropriate endpoint
try {
    switch (true) {
        // Categories endpoints
        case $path === '/api/categories' && $method === 'GET':
            require __DIR__ . '/../endpoints/categories.php';
            break;

        case preg_match('#^/api/categories/([^/]+)/containers$#', $path, $matches) && $method === 'GET':
            $_GET['category'] = $matches[1];
            require __DIR__ . '/../endpoints/category-containers.php';
            break;

        case $path === '/api/categories/refresh' && $method === 'POST':
            require __DIR__ . '/../endpoints/category-refresh.php';
            break;

        // Container endpoints
        case $path === '/api/containers' && $method === 'GET':
            require __DIR__ . '/../endpoints/containers.php';
            break;

        // Container control - both new and simple paths
        case preg_match('#^/api/containers/([^/]+)/control$#', $path, $matches) && $method === 'POST':
            $_POST['container'] = $matches[1];
            require __DIR__ . '/../endpoints/control.php';
            break;

        case $path === '/api/control' && $method === 'POST':
            require __DIR__ . '/../endpoints/control.php';
            break;

        // Bulk actions
        case $path === '/api/containers/bulk' && $method === 'POST':
            require __DIR__ . '/../endpoints/bulk.php';
            break;

        case $path === '/api/bulk' && $method === 'POST':
            require __DIR__ . '/../endpoints/bulk.php';
            break;

        // Container logs - both new and simple paths
        case preg_match('#^/api/containers/([^/]+)/logs$#', $path, $matches) && $method === 'GET':
            $_GET['container'] = $matches[1];
            require __DIR__ . '/../endpoints/logs.php';
            break;

        case $path === '/api/logs' && $method === 'GET':
            require __DIR__ . '/../endpoints/logs.php';
            break;

        // Apps endpoint
        case $path === '/api/apps' && $method === 'GET':
        case $path === '/api/apps.php' && $method === 'GET': // Backwards compatibility
            require __DIR__ . '/../endpoints/apps.php';
            break;

        // Domains endpoint
        case $path === '/api/domains' && $method === 'GET':
        case $path === '/api/domains.php' && $method === 'GET': // Backwards compatibility
            require __DIR__ . '/../endpoints/domains.php';
            break;

        // Backwards compatibility for old container control endpoint
        case $path === '/api/container-control.php' && $method === 'POST':
            require __DIR__ . '/../endpoints/control.php';
            break;

        // Backwards compatibility for old container logs endpoint
        case $path === '/api/container-logs.php' && $method === 'GET':
            require __DIR__ . '/../endpoints/logs.php';
            break;

        // Backwards compatibility for old container bulk endpoint
        case $path === '/api/container-bulk.php' && $method === 'POST':
            require __DIR__ . '/../endpoints/bulk.php';
            break;

        // Health check
        case $path === '/health' && $method === 'GET':
            http_response_code(200);
            echo json_encode([
                'status' => 'healthy',
                'service' => 'devarch-api',
                'timestamp' => time()
            ]);
            break;

        // Not found
        default:
            http_response_code(404);
            echo json_encode([
                'success' => false,
                'error' => 'Endpoint not found',
                'path' => $path,
                'method' => $method
            ]);
            break;
    }
} catch (Exception $e) {
    http_response_code(500);
    echo json_encode([
        'success' => false,
        'error' => 'Internal server error',
        'message' => $e->getMessage()
    ]);
}
