<?php
/**
 * DevArch API - Categories Endpoint
 * Returns all categories with metadata
 */

require_once __DIR__ . '/../lib/common.php';
require_once __DIR__ . '/../lib/categories.php';

try {
    $categories = getCategoriesWithStats();

    // Get total container count
    $totalContainers = array_sum(array_column($categories, 'containerCount'));

    successResponse([
        'categories' => array_values($categories), // Re-index for JSON array
        'totalCategories' => count($categories),
        'totalContainers' => $totalContainers
    ], [
        'timestamp' => time(),
        'cached' => \CategoryCache::isCached(),
        'cacheAge' => \CategoryCache::getCacheAge()
    ]);

} catch (Exception $e) {
    errorResponse('Internal server error', $e->getMessage(), 500);
}
