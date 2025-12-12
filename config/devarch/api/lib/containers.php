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

// DevArch API runs inside devarch container with /workspace mount
define('PODMAN_BINARY', 'podman');
define('NGINX_CONFIG_PATH', '/workspace/config/nginx/custom/http.conf');
define('COMPOSE_BASE_PATH', '/workspace/compose');
const CACHE_DURATION = 300; // 5 minutes

// Category colors (matching service-manager.sh categories)
// Updated 2025-12-12: Added storage, ci, security, registry, workflow, docs, testing, collaboration
const CATEGORY_COLORS = [
    'database' => '#3b82f6',
    'dbms' => '#8b5cf6',
    'gateway' => '#ffcc33',
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
    'storage' => '#9333ea',
    'ci' => '#22c55e',
    'security' => '#dc2626',
    'registry' => '#0ea5e9',
    'workflow' => '#7c3aed',
    'docs' => '#fbbf24',
    'testing' => '#059669',
    'collaboration' => '#ec4899',
];

// Remote Podman API helpers removed - direct Podman access in devarch container

// =============================================================================
// CORE FUNCTIONS
// =============================================================================

/**
 * Get all containers with stats and metadata
 * Merges compose file definitions with actual running containers
 */
function getAllContainers(): array
{
    // Get all service definitions from compose files
    $allServices = parseComposeServices();

    // Get actual running/stopped containers
    $runningContainers = getContainerList();
    $stats = getContainerStats();
    $domains = parseNginxDomains();
    $ports = parseComposePorts();

    // Create lookup map of running containers by name
    $runningMap = [];
    foreach ($runningContainers as $container) {
        $runningMap[$container['name']] = $container;
    }

    // Merge: start with all defined services, enrich with runtime data
    $containers = [];
    foreach ($allServices as $serviceName => $serviceData) {
        if (isset($runningMap[$serviceName])) {
            // Container exists - use runtime data but enrich with compose metadata
            $container = $runningMap[$serviceName];
            $container['category'] = $serviceData['category']; // Use category from compose file
            $container['color'] = CATEGORY_COLORS[$serviceData['category']] ?? '#94a3b8';
        } else {
            // Container not created yet - use compose definition
            $container = [
                'id' => '',
                'name' => $serviceName,
                'status' => 'not-created',
                'image' => $serviceData['image'],
                'version' => $serviceData['version'],
                'uptime' => 'Not created',
                'ports' => $serviceData['ports'] ?? [],
                'category' => $serviceData['category'],
                'color' => CATEGORY_COLORS[$serviceData['category']] ?? '#94a3b8',
                'statusColor' => '#cbd5e1',
            ];
        }

        // Add resource stats if running
        if (isset($stats[$serviceName])) {
            $container['cpu'] = $stats[$serviceName]['cpu'];
            $container['memory'] = $stats[$serviceName]['memory'];
            $container['network'] = $stats[$serviceName]['network'];

            $cpuParsed = parseCpuUsage($stats[$serviceName]['cpu']);
            $container['cpuPercentage'] = $cpuParsed['percentage'];

            $memParsed = parseMemoryUsage($stats[$serviceName]['memory']);
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

        // Get detailed inspect data if container exists
        if ($container['status'] !== 'not-created') {
            $inspectData = getContainerInspectData($serviceName);
            $container['restartCount'] = $inspectData['restartCount'];
            $container['healthStatus'] = $inspectData['healthStatus'];
            $container['uptimeSeconds'] = $inspectData['uptimeSeconds'];
            $container['startedAt'] = $inspectData['startedAt'];
        } else {
            $container['restartCount'] = 0;
            $container['healthStatus'] = null;
            $container['uptimeSeconds'] = 0;
            $container['startedAt'] = null;
        }

        // Add .test domains
        $container['testDomains'] = $domains[$serviceName] ?? [];

        // Add localhost URLs
        $container['localhostUrls'] = $ports[$serviceName] ?? [];

        $containers[] = $container;
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
    // Direct Podman access in devarch container
    putenv('CONTAINER_HOST=unix:///run/podman/podman.sock');

    $whichCmd = 'which ' . PODMAN_BINARY . ' 2>&1';
    exec($whichCmd, $whichOutput, $whichReturn);

    if ($whichReturn !== 0) {
        error_log("Podman binary not found - container monitoring unavailable");
        return [];
    }

    $cmd = PODMAN_BINARY . ' ps -a --format json 2>&1';
    exec($cmd, $output, $returnCode);

    if ($returnCode !== 0) {
        error_log("Failed to execute podman ps: " . implode("\n", $output));
        return [];
    }

    $jsonString = implode('', $output);
    $rawContainers = json_decode($jsonString, true);

    if (!is_array($rawContainers)) {
        error_log("Failed to parse podman ps JSON output");
        return [];
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

        // Determine category - will be enriched later from compose file data
        $category = 'other';

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
    // Direct Podman access in devarch container
    putenv('CONTAINER_HOST=unix:///run/podman/podman.sock');

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
        $name = $stat['name'] ?? '';
        if (empty($name)) continue;

        $stats[$name] = [
            'cpu' => $stat['cpu_percent'] ?? '0%',
            'memory' => ($stat['mem_usage'] ?? 'N/A'),
            'network' => ($stat['net_io'] ?? 'N/A'),
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

    // Direct Podman access in devarch container
    putenv('CONTAINER_HOST=unix:///run/podman/podman.sock');

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

    // Parse line by line to extract server_name and container mappings
    $lines = explode("\n", $content);
    $currentServerNames = [];
    $inServerBlock = false;

    foreach ($lines as $line) {
        $line = trim($line);

        // Detect start of server block
        if (strpos($line, 'server {') !== false) {
            $inServerBlock = true;
            $currentServerNames = [];
            continue;
        }

        // Reset on next server block (don't exit on nested braces like location {})
        // Just stay in server block until we find the next one

        if (!$inServerBlock) {
            continue;
        }

        // Extract server_name
        if (preg_match('/^server_name\s+([^;]+);/i', $line, $nameMatch)) {
            $serverNames = $nameMatch[1];
            $names = preg_split('/[\s,]+/', trim($serverNames));
            $currentServerNames = array_filter($names, fn($n) => str_ends_with($n, '.test'));
        }

        // Extract container from proxy_pass
        if (preg_match('/proxy_pass\s+http:\/\/([^:]+):/i', $line, $containerMatch)) {
            $containerName = $containerMatch[1];

            if (!empty($containerName) && !empty($currentServerNames)) {
                if (!isset($domains[$containerName])) {
                    $domains[$containerName] = [];
                }
                $domains[$containerName] = array_merge($domains[$containerName], $currentServerNames);
            }
        }

        // Extract container from set $upstream_*
        if (preg_match('/set\s+\$upstream_\w+\s+([^:;]+):/i', $line, $upstreamMatch)) {
            $containerName = trim($upstreamMatch[1]);

            if (!empty($containerName) && !empty($currentServerNames)) {
                if (!isset($domains[$containerName])) {
                    $domains[$containerName] = [];
                }
                $domains[$containerName] = array_merge($domains[$containerName], $currentServerNames);
            }
        }
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
 * Parse all compose files to get complete service definitions
 */
function parseComposeServices(): array
{
    static $cache = null;
    static $cacheTime = 0;

    if ($cache !== null && (time() - $cacheTime) < CACHE_DURATION) {
        return $cache;
    }

    $services = [];

    if (!is_dir(COMPOSE_BASE_PATH)) {
        error_log("Compose directory not found: " . COMPOSE_BASE_PATH);
        $cache = $services;
        $cacheTime = time();
        return $services;
    }

    // Find all compose files
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

        // Extract service category from path (e.g., /compose/database/mariadb.yml -> database)
        $pathParts = explode('/', $file);
        $category = 'other';
        foreach ($pathParts as $i => $part) {
            if ($part === 'compose' && isset($pathParts[$i + 1])) {
                $category = $pathParts[$i + 1];
                break;
            }
        }

        // Parse container_name and image
        if (preg_match('/container_name:\s*([^\s\n]+)/i', $content, $nameMatch)) {
            $containerName = trim($nameMatch[1]);

            // Extract image name
            $imageName = 'unknown';
            $imageVersion = 'latest';
            if (preg_match('/image:\s*([^\s\n]+)/i', $content, $imageMatch)) {
                $imageFullName = trim($imageMatch[1]);
                $imageParts = explode(':', $imageFullName);
                $imageName = $imageParts[0] ?? $imageFullName;
                $imageVersion = $imageParts[1] ?? 'latest';

                // Clean up image name (remove docker.io/library/ prefix)
                $imageName = preg_replace('#^(docker\.io/library/|docker\.io/)#', '', $imageName);
            }

            // Extract port mappings
            $ports = [];
            if (preg_match('/ports:\s*((?:\s*-\s*["\']?[\d:\.]+["\']?\s*\n?)+)/s', $content, $portsMatch)) {
                $portsBlock = $portsMatch[1];
                preg_match_all('/["\']?((?:[\d.]+:)?(\d+):\d+)["\']?/', $portsBlock, $portMatches);
                $ports = array_map(function($p) {
                    // Convert "127.0.0.1:8501:3306" to "8501:3306"
                    return preg_replace('/^[\d.]+:/', '', $p);
                }, $portMatches[1]);
            }

            $services[$containerName] = [
                'name' => $containerName,
                'image' => $imageName,
                'version' => $imageVersion,
                'category' => $category, // Use path-based category from filesystem
                'ports' => $ports,
            ];
        }
    }

    $cache = $services;
    $cacheTime = time();

    return $services;
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
    // Exporters - CHECK FIRST before database patterns
    if (str_contains($name, 'exporter') || $name === 'cadvisor') {
        return 'exporters';
    }

    // Gateway services
    if (preg_match('/^(krakend|kong|traefik|envoy)/', $name)) {
        return 'gateway';
    }
    // DBMS tools - CHECK BEFORE database (redis-commander would match redis pattern)
    if (preg_match('/^(adminer|phpmyadmin|pgadmin|mongo-express|redis-commander|cloudbeaver|metabase|nocodb|drawdb|memcached-admin)/', $name)) {
        return 'dbms';
    }

    // Database services
    if (preg_match('/^(mariadb|mysql|postgres|mongodb|redis|memcached)/', $name)) {
        return 'database';
    }

    // Proxy
    if (preg_match('/^(nginx|traefik|proxy)/', $name)) {
        return 'proxy';
    }

    // Backend runtimes
    if (preg_match('/^(php|node|python|go|dotnet)$/', $name)) {
        return 'backend';
    }

    // Analytics
    if (preg_match('/^(prometheus|grafana|elasticsearch|kibana|logstash|otel|matomo|jaeger|loki)/', $name)) {
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

    // AI/Workflow - CHECK BEFORE workflow (n8n is both AI and workflow)
    if (preg_match('/^(langflow|n8n)/', $name)) {
        return 'ai';
    }

    // Management
    if (preg_match('/^(portainer|devarch)/', $name)) {
        return 'management';
    }

    // Storage services
    if (preg_match('/^(localstack|seaweedfs|azurite|minio)/', $name)) {
        return 'storage';
    }

    // CI/CD services
    if (preg_match('/^(concourse|jenkins|drone|woodpecker|gitlab-runner)/', $name)) {
        return 'ci';
    }

    // Security services
    if (preg_match('/^(keycloak|authelia|vault|trivy|authentik)/', $name)) {
        return 'security';
    }

    // Registry services
    if (preg_match('/^(docker-registry|harbor|nexus|verdaccio)/', $name)) {
        return 'registry';
    }

    // Workflow automation
    if (preg_match('/^(prefect|airflow|temporal)/', $name)) {
        return 'workflow';
    }

    // Documentation services
    if (preg_match('/^(docusaurus|bookstack|wikijs|outline)/', $name)) {
        return 'docs';
    }

    // Testing services
    if (preg_match('/^(playwright|selenium|gatling|k6)/', $name)) {
        return 'testing';
    }

    // Collaboration services
    if (preg_match('/^(matrix|element|zulip|mattermost|nextcloud|rocketchat)/', $name)) {
        return 'collaboration';
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
    // Initialize stats with basic counters
    $stats = [
        'total' => count($containers),
        'running' => 0,
        'stopped' => 0,
        'notCreated' => 0,
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

    // Dynamically initialize category counters from containers
    // This replaces the hardcoded category list
    foreach ($containers as $container) {
        $category = $container['category'] ?? 'other';
        if (!isset($stats[$category])) {
            $stats[$category] = 0;
        }
    }

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
        } elseif ($container['status'] === 'not-created') {
            $stats['notCreated']++;
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
 * Find the compose file that defines a specific container
 * Used for smart start and rebuild operations
 */
function findComposeFile(string $containerName): ?string
{
    if (!is_dir(COMPOSE_BASE_PATH)) {
        error_log("Compose directory not found: " . COMPOSE_BASE_PATH);
        return null;
    }

    // Scan all category directories dynamically
    $categoryDirs = scandir(COMPOSE_BASE_PATH);

    foreach ($categoryDirs as $category) {
        if ($category === '.' || $category === '..') continue;

        $categoryPath = COMPOSE_BASE_PATH . '/' . $category;
        if (!is_dir($categoryPath)) continue;

        $pattern = $categoryPath . '/*.yml';
        $files = glob($pattern);

        if ($files === false) {
            continue;
        }

        foreach ($files as $file) {
            $content = file_get_contents($file);
            if ($content === false) {
                continue;
            }

            // Match container_name with or without quotes
            if (preg_match("/container_name:\s*['\"]?{$containerName}['\"]?/", $content)) {
                return $file;
            }
        }
    }

    error_log("Compose file not found for container: {$containerName}");
    return null;
}

/**
 * Detect base path depending on if running in container or on host
 * Exported for use in other files
 */
function detectBasePath(): string
{
    global $basePathPrefix;
    return $basePathPrefix;
}

