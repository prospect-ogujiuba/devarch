<?php
/**
 * DevArch API - Bulk Container Operations Endpoint
 * Performs bulk operations on multiple containers
 */

require_once __DIR__ . '/../lib/common.php';
require_once __DIR__ . '/../lib/containers.php';

// Parse JSON body
$input = getJsonBody();

$action = $input['action'] ?? '';
$containers = $input['containers'] ?? [];

// Validate action
$validActions = ['start', 'stop', 'restart'];
if (!in_array($action, $validActions)) {
    errorResponse('Invalid action', 'Action must be one of: ' . implode(', ', $validActions));
}

// Validate containers array
if (!is_array($containers) || empty($containers)) {
    errorResponse('Invalid containers', 'Containers must be a non-empty array');
}

// Log action for security audit
error_log("Container bulk: {$action} on " . count($containers) . " container(s)");

// Direct Podman access in devarch container
putenv('CONTAINER_HOST=unix:///run/podman/podman.sock');
$podmanBinary = PODMAN_BINARY;

// Execute operations on all containers
$results = [];
$successful = 0;
$failed = 0;

foreach ($containers as $container) {
    // Sanitize container name
    if (!preg_match('/^[a-zA-Z0-9_-]+$/', $container)) {
        $results[] = [
            'container' => $container,
            'success' => false,
            'error' => 'Invalid container name'
        ];
        $failed++;
        continue;
    }

    // Execute podman command
    $cmd = "{$podmanBinary} {$action} " . escapeshellarg($container) . " 2>&1";
    exec($cmd, $output, $code);

    if ($code === 0) {
        $results[] = [
            'container' => $container,
            'success' => true,
            'message' => ucfirst($action) . ' successful',
            'output' => implode("\n", $output)
        ];
        $successful++;
    } else {
        $results[] = [
            'container' => $container,
            'success' => false,
            'error' => implode("\n", $output),
            'message' => 'Failed to ' . $action . ' container'
        ];
        $failed++;
    }

    // Clear output array for next iteration
    $output = [];
}

// Return results
successResponse([
    'results' => $results,
    'summary' => [
        'total' => count($containers),
        'successful' => $successful,
        'failed' => $failed
    ]
]);
