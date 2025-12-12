<?php
/**
 * DevArch API - Containers Endpoint
 * Returns container data with filtering, sorting, and searching
 */

require_once __DIR__ . '/../lib/common.php';
require_once __DIR__ . '/../lib/containers.php';

try {
    // Get all containers with metadata
    $data = getAllContainers();
    $containers = $data['containers'];
    $stats = $data['stats'];

    // Apply filters if provided
    $filter = $_GET['filter'] ?? null;
    if ($filter && $filter !== 'all') {
        if (in_array($filter, ['running', 'stopped', 'not-created'])) {
            // Filter by status
            $containers = array_filter($containers, function($container) use ($filter) {
                if ($filter === 'running') {
                    return $container['status'] === 'running';
                } elseif ($filter === 'not-created') {
                    return $container['status'] === 'not-created';
                } else {
                    // stopped filter includes exited, stopped, and other non-running states
                    return $container['status'] !== 'running' && $container['status'] !== 'not-created';
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
    successResponse([
        'containers' => $containers,
        'stats' => $stats,
        'count' => count($containers),
        'total' => $stats['total'],
    ], [
        'timestamp' => time(),
        'filter' => $filter,
        'search' => $search,
        'sort' => $sort,
        'order' => $order,
    ]);

} catch (Exception $e) {
    errorResponse('Internal server error', $e->getMessage(), 500);
}
