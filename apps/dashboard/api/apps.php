<?php
/**
 * DevArch Dashboard - Apps API Endpoint
 *
 * REST API endpoint that returns application data as JSON.
 * Supports filtering, sorting, and pagination.
 *
 * @version 2.0.0
 * @author DevArch
 */

// Enable CORS for development
header('Access-Control-Allow-Origin: *');
header('Access-Control-Allow-Methods: GET, OPTIONS');
header('Access-Control-Allow-Headers: Content-Type');
header('Content-Type: application/json');

// Handle preflight requests
if ($_SERVER['REQUEST_METHOD'] === 'OPTIONS') {
    http_response_code(204);
    exit;
}

// Only allow GET requests
if ($_SERVER['REQUEST_METHOD'] !== 'GET') {
    http_response_code(405);
    echo json_encode([
        'error' => 'Method not allowed',
        'message' => 'Only GET requests are supported'
    ]);
    exit;
}

// Load detection library
require_once __DIR__ . '/lib/detection.php';

try {
    // Get all apps with metadata
    $data = getAllApps();
    $apps = $data['apps'];
    $stats = $data['stats'];

    // Apply filters if provided
    $filter = $_GET['filter'] ?? null;
    if ($filter && $filter !== 'all') {
        $apps = array_filter($apps, fn($app) => $app['runtime'] === $filter);
        $apps = array_values($apps); // Re-index array
    }

    // Apply search if provided
    $search = $_GET['search'] ?? null;
    if ($search) {
        $searchLower = strtolower($search);
        $apps = array_filter($apps, function($app) use ($searchLower) {
            return str_contains(strtolower($app['name']), $searchLower) ||
                   str_contains(strtolower($app['framework']), $searchLower) ||
                   str_contains(strtolower($app['runtime']), $searchLower);
        });
        $apps = array_values($apps); // Re-index array
    }

    // Apply sorting if provided
    $sort = $_GET['sort'] ?? 'name';
    $order = $_GET['order'] ?? 'asc';

    usort($apps, function($a, $b) use ($sort, $order) {
        $aVal = $a[$sort] ?? '';
        $bVal = $b[$sort] ?? '';

        $result = match(gettype($aVal)) {
            'string' => strcasecmp($aVal, $bVal),
            'integer', 'double' => $aVal <=> $bVal,
            default => 0
        };

        return $order === 'desc' ? -$result : $result;
    });

    // Return successful response
    http_response_code(200);
    echo json_encode([
        'success' => true,
        'data' => [
            'apps' => $apps,
            'stats' => $stats,
            'count' => count($apps),
            'total' => $stats['total'],
        ],
        'meta' => [
            'timestamp' => time(),
            'base_path' => APPS_BASE_PATH,
            'filter' => $filter,
            'search' => $search,
            'sort' => $sort,
            'order' => $order,
        ]
    ], JSON_PRETTY_PRINT);

} catch (Exception $e) {
    // Return error response
    http_response_code(500);
    echo json_encode([
        'success' => false,
        'error' => 'Internal server error',
        'message' => $e->getMessage(),
    ], JSON_PRETTY_PRINT);
}
