<?php
/**
 * DevArch Dashboard - Container Detection Library
 *
 * Detects and monitors containers via Podman CLI, parses nginx configs
 * for .test domain mappings, and compose files for port mappings.
 *
 * @version 1.0.0
 * @author DevArch
 */

// =============================================================================
// CONFIGURATION
// =============================================================================

// Detect if running inside container or on host
$isContainer = file_exists('/.dockerenv') || file_exists('/run/.containerenv');
$basePathPrefix = $isContainer ? '/var/www' : '/home/fhcadmin/projects/devarch';

define('PODMAN_BINARY', 'podman');
define('NGINX_CONFIG_PATH', $basePathPrefix . ($isContainer ? '/../config/nginx/custom/http.conf' : '/config/nginx/custom/http.conf'));
define('COMPOSE_BASE_PATH', $basePathPrefix . ($isContainer ? '/../compose' : '/compose'));
const CACHE_DURATION = 300; // 5 minutes

// Category colors (matching service-manager.sh categories)
const CATEGORY_COLORS = [
    'database' => '#3b82f6',
    'dbms' => '#8b5cf6',
    'proxy' => '#10b981',
    'management' => '#f59e0b',
    'backend' => '#ef4444',
    'project' => '#ec4899',
    'mail' => '#6366f1',
    'exporters' => '#14b8a6',
    'analytics' => '#06b6d4',
    'messaging' => '#f97316',
    'search' => '#84cc16',
    'ai' => '#a855f7',
];

// =============================================================================
// CORE FUNCTIONS
// =============================================================================

/**
 * Get all containers with stats and metadata
 */
function getAllContainers(): array
{
    $containers = getContainerList();
    $stats = getContainerStats();
    $domains = parseNginxDomains();
    $ports = parseComposePorts();

    // Merge stats and inspect data into containers
    foreach ($containers as &$container) {
        $name = $container['name'];

        // Add resource stats if running
        if (isset($stats[$name])) {
            $container['cpu'] = $stats[$name]['cpu'];
            $container['memory'] = $stats[$name]['memory'];
            $container['network'] = $stats[$name]['network'];

            // Parse CPU for percentage
            $cpuParsed = parseCpuUsage($stats[$name]['cpu']);
            $container['cpuPercentage'] = $cpuParsed['percentage'];

            // Parse memory for MB values and percentage
            $memParsed = parseMemoryUsage($stats[$name]['memory']);
            $container['memoryUsedMb'] = $memParsed['usedMb'];
            $container['memoryLimitMb'] = $memParsed['limitMb'];
            $container['memoryPercentage'] = $memParsed['percentage'];
        } else {
            $container['cpu'] = 'N/A';
            $container['memory'] = 'N/A';
            $container['network'] = 'N/A';
            $container['cpuPercentage'] = 0;
            $container['memoryUsedMb'] = 0;
            $container['memoryLimitMb'] = 0;
            $container['memoryPercentage'] = 0;
        }

        // Get detailed inspect data (health, restarts, uptime)
        $inspectData = getContainerInspectData($name);
        $container['restartCount'] = $inspectData['restartCount'];
        $container['healthStatus'] = $inspectData['healthStatus'];
        $container['uptimeSeconds'] = $inspectData['uptimeSeconds'];
        $container['startedAt'] = $inspectData['startedAt'];

        // Add .test domains
        $container['testDomains'] = $domains[$name] ?? [];

        // Add localhost URLs
        $container['localhostUrls'] = $ports[$name] ?? [];
    }

    // Calculate summary stats
    $summaryStats = calculateStats($containers);

    return [
        'containers' => $containers,
        'stats' => $summaryStats,
    ];
}

/**
 * Get container list via podman ps
 */
function getContainerList(): array
{
    // Check if podman is available
    $whichCmd = 'which ' . PODMAN_BINARY . ' 2>&1';
    exec($whichCmd, $whichOutput, $whichReturn);

    if ($whichReturn !== 0) {
        error_log("Podman binary not found - container monitoring unavailable");
        return getMockContainers(); // Return mock data for development
    }

    $cmd = PODMAN_BINARY . ' ps -a --format json 2>&1';
    exec($cmd, $output, $returnCode);

    if ($returnCode !== 0) {
        error_log("Failed to execute podman ps: " . implode("\n", $output));
        return getMockContainers(); // Return mock data on error
    }

    $jsonString = implode('', $output);
    $rawContainers = json_decode($jsonString, true);

    if (!is_array($rawContainers)) {
        error_log("Failed to parse podman ps JSON output");
        return getMockContainers(); // Return mock data on parse error
    }

    $containers = [];
    foreach ($rawContainers as $raw) {
        $name = $raw['Names'][0] ?? 'unknown';
        $imageFullName = $raw['Image'] ?? '';

        // Split image into name and version
        $imageParts = explode(':', $imageFullName);
        $imageName = $imageParts[0] ?? $imageFullName;
        $imageVersion = $imageParts[1] ?? 'latest';

        // Parse status
        $status = strtolower($raw['State'] ?? 'unknown');

        // Parse uptime
        $uptime = $raw['Status'] ?? '';

        // Parse ports
        $portMappings = [];
        if (isset($raw['Ports']) && is_array($raw['Ports'])) {
            foreach ($raw['Ports'] as $portInfo) {
                if (isset($portInfo['host_port']) && isset($portInfo['container_port'])) {
                    $portMappings[] = $portInfo['host_port'] . ':' . $portInfo['container_port'];
                }
            }
        }

        // Determine category
        $category = getCategoryFromName($name);

        $containers[] = [
            'id' => $raw['Id'] ?? '',
            'name' => $name,
            'status' => $status,
            'image' => $imageName,
            'version' => $imageVersion,
            'uptime' => $uptime,
            'ports' => $portMappings,
            'category' => $category,
            'color' => CATEGORY_COLORS[$category] ?? '#94a3b8',
            'statusColor' => getContainerStatusColor($status),
        ];
    }

    return $containers;
}

/**
 * Get container resource stats via podman stats
 */
function getContainerStats(): array
{
    $cmd = PODMAN_BINARY . ' stats --no-stream --format json 2>&1';
    exec($cmd, $output, $returnCode);

    if ($returnCode !== 0) {
        error_log("Failed to execute podman stats: " . implode("\n", $output));
        return [];
    }

    $jsonString = implode('', $output);
    $rawStats = json_decode($jsonString, true);

    if (!is_array($rawStats)) {
        return [];
    }

    $stats = [];
    foreach ($rawStats as $stat) {
        $name = $stat['Name'] ?? '';
        if (empty($name)) continue;

        $stats[$name] = [
            'cpu' => $stat['CPU'] ?? '0%',
            'memory' => ($stat['MemUsage'] ?? 'N/A'),
            'network' => ($stat['NetIO'] ?? 'N/A'),
        ];
    }

    return $stats;
}

/**
 * Parse memory usage string to extract numeric values and percentage
 * Example: "256MiB / 2GiB" → ['usedMb' => 256, 'limitMb' => 2048, 'percentage' => 12.5]
 */
function parseMemoryUsage(string $memoryStr): array
{
    $default = ['usedMb' => 0, 'limitMb' => 0, 'percentage' => 0];

    if ($memoryStr === 'N/A' || empty($memoryStr)) {
        return $default;
    }

    // Match pattern: "256MiB / 2GiB" or "1.2GiB / 4GiB"
    if (!preg_match('/([0-9.]+)\s*([KMGT]i?B)\s*\/\s*([0-9.]+)\s*([KMGT]i?B)/', $memoryStr, $matches)) {
        return $default;
    }

    $usedValue = (float)$matches[1];
    $usedUnit = strtoupper($matches[2]);
    $limitValue = (float)$matches[3];
    $limitUnit = strtoupper($matches[4]);

    // Convert to MB
    $usedMb = convertToMB($usedValue, $usedUnit);
    $limitMb = convertToMB($limitValue, $limitUnit);

    $percentage = $limitMb > 0 ? round(($usedMb / $limitMb) * 100, 1) : 0;

    return [
        'usedMb' => round($usedMb, 1),
        'limitMb' => round($limitMb, 1),
        'percentage' => $percentage,
    ];
}

/**
 * Convert memory value to MB based on unit
 */
function convertToMB(float $value, string $unit): float
{
    return match (strtoupper($unit)) {
        'KIB', 'KB' => $value / 1024,
        'MIB', 'MB' => $value,
        'GIB', 'GB' => $value * 1024,
        'TIB', 'TB' => $value * 1024 * 1024,
        default => $value
    };
}

/**
 * Parse CPU usage string to extract percentage value
 * Example: "0.50%" → ['raw' => "0.50%", 'percentage' => 0.50]
 */
function parseCpuUsage(string $cpuStr): array
{
    $default = ['raw' => '0%', 'percentage' => 0];

    if ($cpuStr === 'N/A' || empty($cpuStr)) {
        return $default;
    }

    // Extract numeric value from percentage string
    if (preg_match('/([0-9.]+)/', $cpuStr, $matches)) {
        return [
            'raw' => $cpuStr,
            'percentage' => (float)$matches[1],
        ];
    }

    return $default;
}

/**
 * Get detailed container inspect data (health, restarts, uptime)
 */
function getContainerInspectData(string $containerName): array
{
    $default = [
        'restartCount' => 0,
        'startedAt' => null,
        'healthStatus' => null,
        'uptimeSeconds' => 0,
    ];

    $cmd = PODMAN_BINARY . ' inspect ' . escapeshellarg($containerName) . ' --format \'{{json .}}\' 2>&1';
    exec($cmd, $output, $returnCode);

    if ($returnCode !== 0) {
        return $default;
    }

    $jsonString = implode('', $output);
    $data = json_decode($jsonString, true);

    if (!is_array($data)) {
        return $default;
    }

    $result = $default;

    // Extract restart count
    if (isset($data['RestartCount'])) {
        $result['restartCount'] = (int)$data['RestartCount'];
    }

    // Extract started time and calculate uptime
    if (isset($data['State']['StartedAt'])) {
        $result['startedAt'] = $data['State']['StartedAt'];

        try {
            $startTime = new DateTime($data['State']['StartedAt']);
            $now = new DateTime();
            $result['uptimeSeconds'] = $now->getTimestamp() - $startTime->getTimestamp();
        } catch (Exception $e) {
            error_log("Failed to parse StartedAt time: " . $e->getMessage());
        }
    }

    // Extract health status if healthcheck is configured
    if (isset($data['State']['Health']['Status'])) {
        $result['healthStatus'] = strtolower($data['State']['Health']['Status']);
    } elseif (isset($data['State']['Healthcheck']['Status'])) {
        $result['healthStatus'] = strtolower($data['State']['Healthcheck']['Status']);
    }

    return $result;
}

/**
 * Parse nginx config for .test domain to container mappings
 */
function parseNginxDomains(): array
{
    static $cache = null;
    static $cacheTime = 0;

    // Return cached result if still valid
    if ($cache !== null && (time() - $cacheTime) < CACHE_DURATION) {
        return $cache;
    }

    $domains = [];

    if (!file_exists(NGINX_CONFIG_PATH)) {
        error_log("Nginx config not found: " . NGINX_CONFIG_PATH);
        $cache = $domains;
        $cacheTime = time();
        return $domains;
    }

    $content = file_get_contents(NGINX_CONFIG_PATH);

    // Parse server blocks to extract server_name and upstream mappings
    // Pattern: server_name xyz.test; ... proxy_pass http://container-name:port;
    preg_match_all(
        '/server\s*\{[^}]*server_name\s+([^;]+);[^}]*(?:proxy_pass\s+http:\/\/([^:]+):|set\s+\$upstream_\w+\s+([^:]+):)/s',
        $content,
        $matches,
        PREG_SET_ORDER
    );

    foreach ($matches as $match) {
        $serverNames = $match[1] ?? '';
        $containerName = $match[2] ?? $match[3] ?? '';

        if (empty($serverNames) || empty($containerName)) {
            continue;
        }

        // Parse multiple server names
        $names = preg_split('/\s+/', trim($serverNames));
        $testDomains = array_filter($names, fn($n) => str_ends_with($n, '.test'));

        if (!isset($domains[$containerName])) {
            $domains[$containerName] = [];
        }

        $domains[$containerName] = array_merge($domains[$containerName], $testDomains);
    }

    // Deduplicate
    foreach ($domains as $container => $domainList) {
        $domains[$container] = array_values(array_unique($domainList));
    }

    $cache = $domains;
    $cacheTime = time();

    return $domains;
}

/**
 * Parse compose files for port mappings to generate localhost URLs
 */
function parseComposePorts(): array
{
    static $cache = null;
    static $cacheTime = 0;

    // Return cached result if still valid
    if ($cache !== null && (time() - $cacheTime) < CACHE_DURATION) {
        return $cache;
    }

    $ports = [];

    if (!is_dir(COMPOSE_BASE_PATH)) {
        error_log("Compose directory not found: " . COMPOSE_BASE_PATH);
        $cache = $ports;
        $cacheTime = time();
        return $ports;
    }

    // Recursively find all .yml files
    $composeFiles = [];
    $iterator = new RecursiveIteratorIterator(
        new RecursiveDirectoryIterator(COMPOSE_BASE_PATH, RecursiveDirectoryIterator::SKIP_DOTS)
    );

    foreach ($iterator as $file) {
        if ($file->isFile() && preg_match('/\.ya?ml$/i', $file->getFilename())) {
            $composeFiles[] = $file->getPathname();
        }
    }

    foreach ($composeFiles as $file) {
        $content = file_get_contents($file);

        // Very basic YAML parsing for container_name and ports
        // Pattern: container_name: xyz ... ports: - "8080:80"
        preg_match_all(
            '/(?:container_name|services):\s*([^\s\n]+).*?ports:\s*((?:\s*-\s*["\']?[\d:]+["\']?\s*\n?)+)/s',
            $content,
            $matches,
            PREG_SET_ORDER
        );

        foreach ($matches as $match) {
            $containerName = trim($match[1]);
            $portsBlock = $match[2];

            // Extract port mappings like "8080:80" or "127.0.0.1:8080:80"
            preg_match_all('/["\']?((?:[\d.]+:)?(\d+):\d+)["\']?/', $portsBlock, $portMatches);

            foreach ($portMatches[2] as $hostPort) {
                if (!isset($ports[$containerName])) {
                    $ports[$containerName] = [];
                }
                $ports[$containerName][] = "http://127.0.0.1:$hostPort";
            }
        }
    }

    // Deduplicate
    foreach ($ports as $container => $urlList) {
        $ports[$container] = array_values(array_unique($urlList));
    }

    $cache = $ports;
    $cacheTime = time();

    return $ports;
}

/**
 * Determine container category from name
 */
function getCategoryFromName(string $name): string
{
    // Database services
    if (preg_match('/^(mariadb|mysql|postgres|mongodb|redis|memcached)/', $name)) {
        return 'database';
    }

    // DBMS tools
    if (preg_match('/^(adminer|phpmyadmin|pgadmin|mongo-express|redis-commander|cloudbeaver|metabase|nocodb|drawdb)/', $name)) {
        return 'dbms';
    }

    // Proxy
    if (preg_match('/^(nginx|traefik|proxy)/', $name)) {
        return 'proxy';
    }

    // Backend runtimes
    if (preg_match('/^(php|node|python|go|dotnet)$/', $name)) {
        return 'backend';
    }

    // Exporters
    if (str_contains($name, 'exporter') || $name === 'cadvisor') {
        return 'exporters';
    }

    // Analytics
    if (preg_match('/^(prometheus|grafana|elasticsearch|kibana|logstash|otel|matomo)/', $name)) {
        return 'analytics';
    }

    // Messaging
    if (preg_match('/^(kafka|rabbitmq|zookeeper)/', $name)) {
        return 'messaging';
    }

    // Search
    if (preg_match('/^(meilisearch|typesense)/', $name)) {
        return 'search';
    }

    // Mail
    if (preg_match('/^(mailpit|mail)/', $name)) {
        return 'mail';
    }

    // Project management
    if (preg_match('/^(gitea|openproject)/', $name)) {
        return 'project';
    }

    // AI services
    if (preg_match('/^(langflow|n8n)/', $name)) {
        return 'ai';
    }

    // Management
    if (preg_match('/^(portainer)/', $name)) {
        return 'management';
    }

    return 'other';
}

/**
 * Get status color for container
 */
function getContainerStatusColor(string $status): string
{
    return match ($status) {
        'running' => '#22c55e',
        'exited' => '#ef4444',
        'stopped' => '#6b7280',
        'paused' => '#eab308',
        default => '#cbd5e1'
    };
}

/**
 * Calculate summary statistics
 */
function calculateStats(array $containers): array
{
    $stats = [
        'total' => count($containers),
        'running' => 0,
        'stopped' => 0,
        'database' => 0,
        'backend' => 0,
        'proxy' => 0,
        'analytics' => 0,
        'dbms' => 0,
        'exporters' => 0,
        'messaging' => 0,
        'search' => 0,
        'mail' => 0,
        'project' => 0,
        'management' => 0,
        'ai' => 0,
        'other' => 0,
        // Health stats
        'healthy' => 0,
        'unhealthy' => 0,
        'starting' => 0,
        // Resource aggregates
        'totalCpuPercentage' => 0,
        'avgCpuPercentage' => 0,
        'totalMemoryMb' => 0,
        'avgMemoryMb' => 0,
        'totalRestarts' => 0,
        'maxRestarts' => 0,
    ];

    $cpuSum = 0;
    $memSum = 0;
    $runningCount = 0;

    foreach ($containers as $container) {
        // Count by status
        if ($container['status'] === 'running') {
            $stats['running']++;
            $runningCount++;

            // Aggregate resource stats for running containers
            if (isset($container['cpuPercentage'])) {
                $cpuSum += $container['cpuPercentage'];
            }
            if (isset($container['memoryUsedMb'])) {
                $memSum += $container['memoryUsedMb'];
            }
        } else {
            $stats['stopped']++;
        }

        // Count by category
        $category = $container['category'];
        if (isset($stats[$category])) {
            $stats[$category]++;
        }

        // Count health status
        if (isset($container['healthStatus'])) {
            switch ($container['healthStatus']) {
                case 'healthy':
                    $stats['healthy']++;
                    break;
                case 'unhealthy':
                    $stats['unhealthy']++;
                    break;
                case 'starting':
                    $stats['starting']++;
                    break;
            }
        }

        // Track restart counts
        if (isset($container['restartCount'])) {
            $stats['totalRestarts'] += $container['restartCount'];
            $stats['maxRestarts'] = max($stats['maxRestarts'], $container['restartCount']);
        }
    }

    // Calculate averages
    $stats['totalCpuPercentage'] = round($cpuSum, 1);
    $stats['avgCpuPercentage'] = $runningCount > 0 ? round($cpuSum / $runningCount, 1) : 0;
    $stats['totalMemoryMb'] = round($memSum);
    $stats['avgMemoryMb'] = $runningCount > 0 ? round($memSum / $runningCount) : 0;

    return $stats;
}

/**
 * Get mock container data for development/demo
 */
function getMockContainers(): array
{
    return [
        [
            'id' => 'abc123',
            'name' => 'nginx-proxy-manager',
            'status' => 'running',
            'image' => 'jc21/nginx-proxy-manager',
            'version' => 'latest',
            'uptime' => 'Up 2 days',
            'ports' => ['80:80', '443:443', '81:81'],
            'category' => 'proxy',
            'color' => CATEGORY_COLORS['proxy'],
            'statusColor' => '#22c55e',
        ],
        [
            'id' => 'def456',
            'name' => 'mariadb',
            'status' => 'running',
            'image' => 'mariadb',
            'version' => '10.11',
            'uptime' => 'Up 3 hours',
            'ports' => ['3306:3306'],
            'category' => 'database',
            'color' => CATEGORY_COLORS['database'],
            'statusColor' => '#22c55e',
        ],
        [
            'id' => 'ghi789',
            'name' => 'php',
            'status' => 'running',
            'image' => 'custom-php',
            'version' => '8.3-fpm',
            'uptime' => 'Up 1 hour',
            'ports' => ['8100:8000', '8102:5173'],
            'category' => 'backend',
            'color' => CATEGORY_COLORS['backend'],
            'statusColor' => '#22c55e',
        ],
        [
            'id' => 'jkl012',
            'name' => 'postgres',
            'status' => 'running',
            'image' => 'postgres',
            'version' => '15',
            'uptime' => 'Up 2 days',
            'ports' => ['5432:5432'],
            'category' => 'database',
            'color' => CATEGORY_COLORS['database'],
            'statusColor' => '#22c55e',
        ],
        [
            'id' => 'mno345',
            'name' => 'redis',
            'status' => 'running',
            'image' => 'redis',
            'version' => '7-alpine',
            'uptime' => 'Up 5 hours',
            'ports' => ['6379:6379'],
            'category' => 'database',
            'color' => CATEGORY_COLORS['database'],
            'statusColor' => '#22c55e',
        ],
        [
            'id' => 'pqr678',
            'name' => 'phpmyadmin',
            'status' => 'exited',
            'image' => 'phpmyadmin',
            'version' => 'latest',
            'uptime' => 'Exited (0) 2 hours ago',
            'ports' => [],
            'category' => 'dbms',
            'color' => CATEGORY_COLORS['dbms'],
            'statusColor' => '#ef4444',
        ],
    ];
}
