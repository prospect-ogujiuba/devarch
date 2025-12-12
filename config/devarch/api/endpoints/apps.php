<?php
/**
 * DevArch API - Apps Endpoint
 * Returns application data with filtering, sorting, and searching
 */

require_once __DIR__ . '/../lib/common.php';
require_once __DIR__ . '/../lib/apps.php';

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
    successResponse([
        'apps' => $apps,
        'stats' => $stats,
        'count' => count($apps),
        'total' => $stats['total'],
    ], [
        'timestamp' => time(),
        'base_path' => APPS_BASE_PATH,
        'filter' => $filter,
        'search' => $search,
        'sort' => $sort,
        'order' => $order,
    ]);

} catch (Exception $e) {
    errorResponse('Internal server error', $e->getMessage(), 500);
}
