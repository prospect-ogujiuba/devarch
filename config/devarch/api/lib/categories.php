<?php
/**
 * DevArch API - Category Auto-Discovery
 * Dynamically discovers categories from compose directory structure
 */

require_once __DIR__ . '/containers.php';

// Category cache
class CategoryCache {
    private static $cache = null;
    private static $cacheTime = 0;
    private const CACHE_TTL = 300; // 5 minutes

    public static function get(): ?array {
        if (self::$cache !== null && (time() - self::$cacheTime) < self::CACHE_TTL) {
            return self::$cache;
        }
        return null;
    }

    public static function set(array $data): void {
        self::$cache = $data;
        self::$cacheTime = time();
    }

    public static function invalidate(): void {
        self::$cache = null;
        self::$cacheTime = 0;
    }

    public static function getCacheAge(): int {
        return time() - self::$cacheTime;
    }

    public static function isCached(): bool {
        return self::$cache !== null && (time() - self::$cacheTime) < self::CACHE_TTL;
    }
}

/**
 * Auto-discover categories from compose directory structure
 *
 * Scans /workspace/compose/ and returns category metadata
 * Each subdirectory becomes a category with its associated containers
 *
 * @return array Category metadata indexed by category name
 */
function discoverCategories(): array {
    // Check cache first
    $cached = CategoryCache::get();
    if ($cached !== null) {
        return $cached;
    }

    $composePath = COMPOSE_BASE_PATH;
    $categories = [];

    if (!is_dir($composePath)) {
        error_log("Compose directory not found: {$composePath}");
        return [];
    }

    // Scan compose directory for subdirectories
    $items = scandir($composePath);

    foreach ($items as $item) {
        if ($item === '.' || $item === '..') continue;

        $categoryPath = $composePath . '/' . $item;

        if (!is_dir($categoryPath)) continue;

        // Category name = directory name
        $categoryName = $item;

        // Find all .yml files in category
        $composeFiles = glob($categoryPath . '/*.yml');
        if ($composeFiles === false) $composeFiles = [];

        // Parse containers from compose files
        $containers = [];
        $containerCount = 0;

        foreach ($composeFiles as $file) {
            $content = @file_get_contents($file);
            if ($content === false) continue;

            // Extract container_name from compose file
            if (preg_match_all('/container_name:\s*([^\s\n]+)/i', $content, $matches)) {
                foreach ($matches[1] as $containerName) {
                    $containerName = trim($containerName);
                    $containers[] = [
                        'name' => $containerName,
                        'composeFile' => $file,
                        'category' => $categoryName
                    ];
                    $containerCount++;
                }
            }
        }

        $categories[$categoryName] = [
            'name' => $categoryName,
            'displayName' => ucfirst(str_replace('-', ' ', $categoryName)),
            'containerCount' => $containerCount,
            'composeFileCount' => count($composeFiles),
            'containers' => $containers,
            'color' => getCategoryColor($categoryName),
        ];
    }

    // Sort categories by name
    ksort($categories);

    // Cache the result
    CategoryCache::set($categories);

    return $categories;
}

/**
 * Get category color (with sensible defaults)
 */
function getCategoryColor(string $category): string {
    // Keep existing colors for consistency
    $colors = CATEGORY_COLORS;

    return $colors[$category] ?? '#94a3b8'; // Default gray
}

/**
 * Get metadata for a specific category
 *
 * @param string $categoryName Category name to look up
 * @return array|null Category metadata or null if not found
 */
function getCategoryMetadata(string $categoryName): ?array {
    $categories = discoverCategories();
    return $categories[$categoryName] ?? null;
}

/**
 * Determine category from compose file path
 *
 * Instead of regex pattern matching, derive category from file structure:
 * /workspace/compose/{category}/file.yml → category name
 *
 * @param string $composePath Full path to compose file
 * @return string Category name or 'other' if not found
 */
function getCategoryFromComposePath(string $composePath): string {
    // Extract category from path: /workspace/compose/{category}/file.yml
    $parts = explode('/', $composePath);

    foreach ($parts as $i => $part) {
        if ($part === 'compose' && isset($parts[$i + 1])) {
            return $parts[$i + 1]; // Return directory name after 'compose'
        }
    }

    return 'other';
}

/**
 * Get all categories with enriched stats (running/stopped counts)
 *
 * @return array Categories with runtime statistics
 */
function getCategoriesWithStats(): array {
    $categories = discoverCategories();

    // Get all containers to determine status counts
    $containerData = getAllContainers();
    $allContainers = $containerData['containers'];

    // Enrich categories with runtime statistics
    foreach ($categories as $name => $category) {
        $categoryContainers = array_filter($allContainers, fn($c) => $c['category'] === $name);

        $running = count(array_filter($categoryContainers, fn($c) => $c['status'] === 'running'));
        $stopped = count(array_filter($categoryContainers, fn($c) => $c['status'] !== 'running' && $c['status'] !== 'not-created'));
        $notCreated = count(array_filter($categoryContainers, fn($c) => $c['status'] === 'not-created'));

        $categories[$name]['runningCount'] = $running;
        $categories[$name]['stoppedCount'] = $stopped;
        $categories[$name]['notCreatedCount'] = $notCreated;
    }

    return $categories;
}
