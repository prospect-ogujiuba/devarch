<?php
/**
 * DevArch Dashboard - Containers API Endpoint
 *
 * REST API endpoint that returns container data as JSON.
 * Supports filtering, sorting, and searching.
 *
 * @version 1.0.0
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

// Load container detection library
require_once __DIR__ . '/lib/containers.php';

try {
    // Get all containers with metadata
    $data = getAllContainers();
    $containers = $data['containers'];
    $stats = $data['stats'];

    // Apply filters if provided
    $filter = $_GET['filter'] ?? null;
    if ($filter && $filter !== 'all') {
        if (in_array($filter, ['running', 'stopped'])) {
            // Filter by status
            $containers = array_filter($containers, function($container) use ($filter) {
                if ($filter === 'running') {
                    return $container['status'] === 'running';
                } else {
                    return $container['status'] !== 'running';
                }
            });
        } else {
            // Filter by category
            $containers = array_filter($containers, fn($container) => $container['category'] === $filter);
        }
        $containers = array_values($containers); // Re-index array
    }

    // Apply search if provided
    $search = $_GET['search'] ?? null;
    if ($search) {
        $searchLower = strtolower($search);
        $containers = array_filter($containers, function($container) use ($searchLower) {
            return str_contains(strtolower($container['name']), $searchLower) ||
                   str_contains(strtolower($container['image']), $searchLower) ||
                   str_contains(strtolower($container['category']), $searchLower);
        });
        $containers = array_values($containers); // Re-index array
    }

    // Apply sorting if provided
    $sort = $_GET['sort'] ?? 'name';
    $order = $_GET['order'] ?? 'asc';

    usort($containers, function($a, $b) use ($sort, $order) {
        $aVal = $a[$sort] ?? '';
        $bVal = $b[$sort] ?? '';

        // Special handling for CPU and memory (remove % and parse)
        if ($sort === 'cpu') {
            $aVal = floatval(str_replace('%', '', $aVal));
            $bVal = floatval(str_replace('%', '', $bVal));
        }

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
            'containers' => $containers,
            'stats' => $stats,
            'count' => count($containers),
            'total' => $stats['total'],
        ],
        'meta' => [
            'timestamp' => time(),
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
