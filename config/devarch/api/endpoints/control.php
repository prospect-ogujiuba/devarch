<?php
/**
 * DevArch API - Container Control Endpoint
 * Controls containers (start/stop/restart/rebuild)
 */

require_once __DIR__ . '/../lib/common.php';
require_once __DIR__ . '/../lib/containers.php';

// Parse JSON body
$input = getJsonBody();

$action = $input['action'] ?? $_POST['action'] ?? '';
$container = $input['container'] ?? $_POST['container'] ?? '';

// Validate inputs
$validActions = ['start', 'stop', 'restart', 'rebuild', 'rebuild-no-cache'];
if (!in_array($action, $validActions)) {
    errorResponse('Invalid action', 'Action must be one of: ' . implode(', ', $validActions));
}

if (empty($container)) {
    errorResponse('Missing container', 'Container name or ID is required');
}

// Sanitize container name (alphanumeric, dash, underscore only)
if (!preg_match('/^[a-zA-Z0-9_-]+$/', $container)) {
    errorResponse('Invalid container name', 'Container name contains invalid characters');
}

// Log action for security audit
error_log("Container control: {$action} on {$container}");

// Direct Podman access in devarch container
putenv('CONTAINER_HOST=unix:///run/podman/podman.sock');
$podmanBinary = PODMAN_BINARY;
$podmanComposeBinary = 'podman-compose';

// Handle rebuild operations
if ($action === 'rebuild' || $action === 'rebuild-no-cache') {
    $composeFile = findComposeFile($container);

    if (!$composeFile) {
        errorResponse('Compose file not found', "Could not find compose file for container: {$container}", 404);
    }

    // Stop container if running
    exec("{$podmanBinary} stop {$container} 2>&1", $stopOutput);

    // Build with or without cache
    $buildCmd = "{$podmanComposeBinary} -f " . escapeshellarg($composeFile) . " build";
    if ($action === 'rebuild-no-cache') {
        $buildCmd .= " --no-cache";
    }
    $buildCmd .= " 2>&1";

    exec($buildCmd, $buildOutput, $buildCode);

    if ($buildCode !== 0) {
        errorResponse('Build failed', 'Failed to rebuild container', 500);
    }

    // Start container with force-recreate
    $upCmd = "{$podmanComposeBinary} -f " . escapeshellarg($composeFile) . " up -d --force-recreate 2>&1";
    exec($upCmd, $output, $returnCode);

} elseif ($action === 'start') {
    // Smart start: check if container exists
    $inspectCmd = "{$podmanBinary} inspect " . escapeshellarg($container) . " 2>&1";
    exec($inspectCmd, $inspectOut, $inspectCode);

    if ($inspectCode !== 0) {
        // Container doesn't exist - use compose
        $composeFile = findComposeFile($container);

        if (!$composeFile) {
            errorResponse('Compose file not found', "Container not found and no compose file exists for: {$container}", 404);
        }

        // Use podman-compose up for new containers
        $cmd = "{$podmanComposeBinary} -f " . escapeshellarg($composeFile) . " up -d 2>&1";
    } else {
        // Container exists - use regular podman start
        $cmd = "{$podmanBinary} start " . escapeshellarg($container) . " 2>&1";
    }

    exec($cmd, $output, $returnCode);

} else {
    // Regular stop/restart commands
    $cmd = "{$podmanBinary} {$action} " . escapeshellarg($container) . " 2>&1";
    exec($cmd, $output, $returnCode);
}

if ($returnCode === 0) {
    // Get updated container status
    $statusCmd = "{$podmanBinary} inspect " . escapeshellarg($container) . " --format '{{.State.Status}}' 2>&1";
    exec($statusCmd, $statusOutput, $statusReturn);
    $status = $statusReturn === 0 ? trim($statusOutput[0]) : 'unknown';

    successResponse([
        'message' => ucfirst($action) . ' successful',
        'container' => $container,
        'status' => $status,
        'output' => implode("\n", $output)
    ]);
} else {
    errorResponse('Command failed', 'Failed to ' . $action . ' container: ' . implode("\n", $output), 500);
}
