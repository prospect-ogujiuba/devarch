<?php
/**
 * DevArch API - Category Refresh Endpoint
 * Forces cache invalidation and category rescan
 */

require_once __DIR__ . '/../lib/common.php';
require_once __DIR__ . '/../lib/categories.php';

try {
    // Invalidate cache
    CategoryCache::invalidate();

    // Force re-scan
    $categories = discoverCategories();

    $totalContainers = array_sum(array_column($categories, 'containerCount'));

    successResponse([
        'message' => 'Categories rescanned',
        'categoriesFound' => count($categories),
        'containersFound' => $totalContainers
    ], [
        'timestamp' => time()
    ]);

} catch (Exception $e) {
    errorResponse('Internal server error', $e->getMessage(), 500);
}
