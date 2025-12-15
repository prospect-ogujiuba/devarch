<?php
/**
 * POST /api/runtime/switch
 * Switch between Docker and Podman runtimes
 */

require_once __DIR__ . '/../lib/common.php';
require_once __DIR__ . '/../lib/shell.php';

// Parse JSON body
$input = getJsonBody();

// Validate runtime
$runtime = $input['runtime'] ?? '';
if (!in_array($runtime, ['docker', 'podman'])) {
    errorResponse('Invalid runtime', 'Runtime must be "docker" or "podman"', 400);
}

// Parse options
$options = $input['options'] ?? [];
$stopServices = $options['stopServices'] ?? true;
$preserveData = $options['preserveData'] ?? true;
$updateConfig = $options['updateConfig'] ?? true;

try {
    // Get current runtime status
    $statusCmd = 'podman ps -q 2>&1';
    $podmanStatus = execSecure($statusCmd);
    $currentRuntime = $podmanStatus['success'] ? 'podman' : 'docker';

    // Check if already on target runtime
    if ($currentRuntime === $runtime) {
        successResponse([
            'message' => "Already running on {$runtime}",
            'current' => $runtime,
            'previous' => $runtime,
            'noChangeRequired' => true
        ]);
    }

    // Count running services
    $psCmd = "{$currentRuntime} ps --filter network=microservices-net -q 2>&1";
    $psResult = execSecure($psCmd);
    $runningCount = 0;

    if ($psResult['success']) {
        $containers = array_filter(explode("\n", trim($psResult['output'])));
        $runningCount = count($containers);
    }

    // Check for running services
    if ($runningCount > 0 && !$stopServices) {
        errorResponse(
            'Services still running',
            "{$runningCount} service(s) running on {$currentRuntime}. Set stopServices to true to auto-stop.",
            409
        );
    }

    // Stop services if requested
    $servicesStopped = 0;
    if ($runningCount > 0 && $stopServices) {
        // Use service-manager to stop all services
        $stopCmd = '/workspace/scripts/service-manager.sh stop-all';
        if (!$preserveData) {
            $stopCmd .= ' --remove-volumes';
        }

        $stopResult = execSecure($stopCmd);
        if (!$stopResult['success']) {
            errorResponse(
                'Failed to stop services',
                'Could not stop running services: ' . $stopResult['output'],
                500
            );
        }

        $servicesStopped = $runningCount;
    }

    // Update config.sh if requested
    $configUpdated = false;
    if ($updateConfig) {
        $configPath = '/workspace/scripts/config.sh';

        if (file_exists($configPath)) {
            $configContent = file_get_contents($configPath);

            // Update CONTAINER_RUNTIME variable
            $newRuntime = $runtime === 'docker' ? 'docker' : 'podman';
            $configContent = preg_replace(
                '/^export CONTAINER_RUNTIME=.*$/m',
                "export CONTAINER_RUNTIME=\"{$newRuntime}\"",
                $configContent
            );

            file_put_contents($configPath, $configContent);
            $configUpdated = true;
        }
    }

    // Check if target runtime is available
    $targetCheck = "{$runtime} version 2>&1";
    $targetResult = execSecure($targetCheck);

    if (!$targetResult['success']) {
        errorResponse(
            'Target runtime not available',
            "{$runtime} is not installed or not accessible",
            503
        );
    }

    successResponse([
        'success' => true,
        'previous' => $currentRuntime,
        'current' => $runtime,
        'servicesStopped' => $servicesStopped,
        'configUpdated' => $configUpdated,
        'message' => "Successfully switched to {$runtime}",
        'nextSteps' => [
            'Start services with: POST /api/services/start'
        ]
    ]);
} catch (Exception $e) {
    errorResponse('Switch failed', $e->getMessage(), 500);
}
