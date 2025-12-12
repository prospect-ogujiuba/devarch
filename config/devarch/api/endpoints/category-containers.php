<?php
/**
 * DevArch API - Category Containers Endpoint
 * Returns containers for a specific category
 */

require_once __DIR__ . '/../lib/common.php';
require_once __DIR__ . '/../lib/categories.php';
require_once __DIR__ . '/../lib/containers.php';

try {
    $categoryName = $_GET['category'] ?? '';

    if (empty($categoryName)) {
        errorResponse('Missing category', 'Category name is required');
    }

    $category = getCategoryMetadata($categoryName);

    if (!$category) {
        errorResponse('Category not found', "Category '{$categoryName}' does not exist", 404);
    }

    // Get all containers, filter by category
    $containerData = getAllContainers();
    $categoryContainers = array_filter(
        $containerData['containers'],
        fn($c) => $c['category'] === $categoryName
    );

    successResponse([
        'category' => [
            'name' => $category['name'],
            'displayName' => $category['displayName'],
            'color' => $category['color'],
        ],
        'containers' => array_values($categoryContainers), // Re-index for JSON array
        'count' => count($categoryContainers),
    ], [
        'timestamp' => time()
    ]);

} catch (Exception $e) {
    errorResponse('Internal server error', $e->getMessage(), 500);
}
