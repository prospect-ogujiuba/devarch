<?php
/**
 * Service Management Library
 * Advanced service operations with dependency ordering and health checks
 */

require_once __DIR__ . '/shell.php';
require_once __DIR__ . '/containers.php';

// Service startup order (must match config.sh)
const SERVICE_STARTUP_ORDER = [
    'database', 'storage', 'dbms', 'erp', 'security', 'registry',
    'gateway', 'proxy', 'management', 'backend', 'ci', 'project',
    'mail', 'exporters', 'analytics', 'messaging', 'search',
    'workflow', 'docs', 'testing', 'collaboration', 'ai', 'support'
];

// Services that should have health checks
const HEALTH_CHECK_SERVICES = [
    'postgres', 'mariadb', 'mysql', 'mongodb', 'redis',
    'nginx-proxy-manager', 'krakend', 'traefik'
];

/**
 * Get services in dependency order
 */
function getServicesInDependencyOrder(array $categories, bool $reverse = false): array {
    $ordered = array_filter(SERVICE_STARTUP_ORDER, fn($cat) => in_array($cat, $categories));

    if ($reverse) {
        $ordered = array_reverse($ordered);
    }

    return $ordered;
}

/**
 * Get all service names from a category
 */
function getServicesInCategory(string $category): array {
    $composeFiles = glob(COMPOSE_BASE_PATH . "/{$category}/*.yml");

    if ($composeFiles === false) {
        return [];
    }

    $services = [];
    foreach ($composeFiles as $file) {
        $content = file_get_contents($file);
        if (preg_match('/container_name:\s*([^\s\n]+)/', $content, $match)) {
            $services[] = trim($match[1]);
        }
    }

    return $services;
}

/**
 * Start service with options
 */
function startService(string $service, array $options = []): array {
    if (!validateServiceName($service)) {
        return [
            'success' => false,
            'service' => $service,
            'error' => 'Invalid service name'
        ];
    }

    $composeFile = findComposeFile($service);
    if (!$composeFile) {
        return [
            'success' => false,
            'service' => $service,
            'error' => 'Compose file not found'
        ];
    }

    $startTime = microtime(true);

    // Build command options
    $cmdOptions = ['d' => true];

    if ($options['forceRecreate'] ?? false) {
        $cmdOptions['force-recreate'] = true;
    }

    if ($options['rebuild'] ?? false) {
        // Build first
        $buildCmd = buildComposeCommand($composeFile, 'build', [
            'no-cache' => $options['noCache'] ?? false
        ]);

        $buildResult = execSecure($buildCmd);
        if (!$buildResult['success']) {
            return [
                'success' => false,
                'service' => $service,
                'error' => 'Build failed',
                'output' => $buildResult['output']
            ];
        }
    }

    // Start service
    $cmd = buildComposeCommand($composeFile, 'up', $cmdOptions);
    $result = execSecure($cmd);

    if (!$result['success']) {
        return [
            'success' => false,
            'service' => $service,
            'error' => 'Start failed',
            'output' => $result['output']
        ];
    }

    // Wait for health if requested
    $healthStatus = 'unknown';
    if ($options['waitHealthy'] ?? false) {
        $healthTimeout = $options['healthTimeout'] ?? 120;
        $healthStatus = waitForServiceHealth($service, $healthTimeout);
    }

    $timeTaken = round(microtime(true) - $startTime, 2);

    return [
        'success' => true,
        'service' => $service,
        'status' => getPodmanContainerStatus($service) ?? 'running',
        'healthStatus' => $healthStatus,
        'timeTaken' => $timeTaken
    ];
}

/**
 * Stop service with options
 */
function stopService(string $service, array $options = []): array {
    if (!validateServiceName($service)) {
        return [
            'success' => false,
            'service' => $service,
            'error' => 'Invalid service name'
        ];
    }

    $composeFile = findComposeFile($service);
    if (!$composeFile) {
        return [
            'success' => false,
            'service' => $service,
            'error' => 'Compose file not found'
        ];
    }

    // Build command options
    $cmdOptions = [
        'timeout' => $options['timeout'] ?? 30
    ];

    if ($options['removeVolumes'] ?? false) {
        $cmdOptions['volumes'] = true;
    }

    // Stop service
    $cmd = buildComposeCommand($composeFile, 'down', $cmdOptions);
    $result = execSecure($cmd);

    return [
        'success' => $result['success'],
        'service' => $service,
        'error' => $result['success'] ? null : 'Stop failed',
        'output' => $result['output']
    ];
}

/**
 * Start multiple services with advanced options
 */
function startServices(array $targets, array $options = []): array {
    $results = [];
    $failed = [];
    $totalTime = 0;

    // Determine if targets are categories or individual services
    $categories = [];
    $individualServices = [];

    foreach ($targets as $target) {
        if (validateCategoryName($target)) {
            $categories[] = $target;
        } else {
            $individualServices[] = $target;
        }
    }

    // Get services in dependency order
    if (!empty($categories)) {
        $orderedCategories = getServicesInDependencyOrder($categories);

        foreach ($orderedCategories as $category) {
            $services = getServicesInCategory($category);

            foreach ($services as $service) {
                if (in_array($service, $options['exceptServices'] ?? [])) {
                    continue;
                }

                $result = startService($service, $options);
                $results[] = $result;
                $totalTime += $result['timeTaken'] ?? 0;

                if (!$result['success']) {
                    $failed[] = $service;
                }
            }
        }
    }

    // Start individual services
    foreach ($individualServices as $service) {
        $result = startService($service, $options);
        $results[] = $result;
        $totalTime += $result['timeTaken'] ?? 0;

        if (!$result['success']) {
            $failed[] = $service;
        }
    }

    return [
        'success' => empty($failed),
        'started' => array_filter($results, fn($r) => $r['success']),
        'failed' => array_filter($results, fn($r) => !$r['success']),
        'summary' => [
            'total' => count($results),
            'successful' => count($results) - count($failed),
            'failed' => count($failed),
            'totalTime' => round($totalTime, 2)
        ]
    ];
}

/**
 * Stop multiple services
 */
function stopServices(array $targets, array $options = []): array {
    $results = [];
    $failed = [];

    // Determine if targets are categories or individual services
    $categories = [];
    $individualServices = [];

    foreach ($targets as $target) {
        if (validateCategoryName($target)) {
            $categories[] = $target;
        } else {
            $individualServices[] = $target;
        }
    }

    // Stop in reverse dependency order
    if (!empty($categories)) {
        $orderedCategories = getServicesInDependencyOrder($categories, true);

        foreach ($orderedCategories as $category) {
            $services = getServicesInCategory($category);

            foreach ($services as $service) {
                $result = stopService($service, $options);
                $results[] = $result;

                if (!$result['success']) {
                    $failed[] = $service;
                }
            }
        }
    }

    // Stop individual services
    foreach ($individualServices as $service) {
        $result = stopService($service, $options);
        $results[] = $result;

        if (!$result['success']) {
            $failed[] = $service;
        }
    }

    return [
        'success' => empty($failed),
        'stopped' => array_map(fn($r) => $r['service'], array_filter($results, fn($r) => $r['success'])),
        'failed' => array_map(fn($r) => $r['service'], array_filter($results, fn($r) => !$r['success'])),
        'summary' => [
            'total' => count($results),
            'successful' => count($results) - count($failed),
            'failed' => count($failed)
        ]
    ];
}

/**
 * Wait for service health check
 */
function waitForServiceHealth(string $service, int $timeout = 120): string {
    $start = time();

    while ((time() - $start) < $timeout) {
        // Check Podman health status
        $cmd = 'podman inspect ' . escapeArg($service) . ' --format "{{.State.Health.Status}}" 2>&1';
        $result = execSecure($cmd);

        if ($result['success']) {
            $health = trim($result['output']);

            if ($health === 'healthy') {
                return 'healthy';
            } elseif ($health === 'unhealthy') {
                return 'unhealthy';
            }
        }

        // Fallback: check if container is running
        $status = getPodmanContainerStatus($service);
        if ($status !== 'running') {
            return 'not-running';
        }

        sleep(2);
    }

    return 'timeout';
}

/**
 * Get enhanced service status with health and uptime
 */
function getEnhancedServiceStatus(): array {
    $allContainers = getAllContainers();
    $containers = $allContainers['containers'];

    // Group by category
    $categories = [];
    foreach ($containers as $container) {
        $cat = $container['category'];

        if (!isset($categories[$cat])) {
            $categories[$cat] = [
                'name' => $cat,
                'services' => [],
                'stats' => [
                    'total' => 0,
                    'running' => 0,
                    'stopped' => 0,
                    'unhealthy' => 0
                ]
            ];
        }

        $categories[$cat]['services'][] = [
            'name' => $container['name'],
            'status' => $container['status'],
            'health' => $container['healthStatus'] ?? null,
            'uptime' => $container['uptimeSeconds'] ?? 0,
            'restartCount' => $container['restartCount'] ?? 0
        ];

        $categories[$cat]['stats']['total']++;

        if ($container['status'] === 'running') {
            $categories[$cat]['stats']['running']++;
        } else {
            $categories[$cat]['stats']['stopped']++;
        }

        if (($container['healthStatus'] ?? null) === 'unhealthy') {
            $categories[$cat]['stats']['unhealthy']++;
        }
    }

    return [
        'categories' => array_values($categories),
        'summary' => $allContainers['stats']
    ];
}
