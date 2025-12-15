<?php
/**
 * GET /api/runtime/status
 * Get current container runtime status (Docker vs Podman)
 */

require_once __DIR__ . '/../lib/common.php';
require_once __DIR__ . '/../lib/shell.php';

try {
    // Check Podman
    $podmanInstalled = false;
    $podmanVersion = null;
    $podmanRunning = false;
    $podmanResponsive = false;

    $podmanVersionCmd = 'podman version --format "{{.Client.Version}}" 2>&1';
    $podmanVersionResult = execSecure($podmanVersionCmd);

    if ($podmanVersionResult['success']) {
        $podmanInstalled = true;
        $podmanVersion = trim($podmanVersionResult['output']);

        // Check if socket is responsive
        $activePodmanSocket = getActivePodmanSocket();
        if ($activePodmanSocket) {
            $podmanRunning = true;
            $podmanResponsive = testPodmanSocketConnectivity($activePodmanSocket);
        }
    }

    // Check Docker
    $dockerInstalled = false;
    $dockerVersion = null;
    $dockerRunning = false;
    $dockerResponsive = false;

    $dockerVersionCmd = 'docker version --format "{{.Client.Version}}" 2>&1';
    $dockerVersionResult = execSecure($dockerVersionCmd);

    if ($dockerVersionResult['success']) {
        $dockerInstalled = true;
        $dockerVersion = trim($dockerVersionResult['output']);

        // Check if Docker daemon is running
        $dockerPsCmd = 'docker ps 2>&1';
        $dockerPsResult = execSecure($dockerPsCmd);

        if ($dockerPsResult['success']) {
            $dockerRunning = true;
            $dockerResponsive = true;
        }
    }

    // Determine current runtime
    $current = 'none';
    if ($podmanRunning && $podmanResponsive) {
        $current = 'podman';
    } elseif ($dockerRunning && $dockerResponsive) {
        $current = 'docker';
    }

    // Count containers for each runtime
    $podmanContainers = 0;
    $dockerContainers = 0;

    if ($podmanResponsive) {
        $podmanPsCmd = 'podman ps -q 2>&1';
        $podmanPsResult = execSecure($podmanPsCmd);
        if ($podmanPsResult['success']) {
            $podmanContainers = count(array_filter(explode("\n", trim($podmanPsResult['output']))));
        }
    }

    if ($dockerResponsive) {
        $dockerPsCmd = 'docker ps -q 2>&1';
        $dockerPsResult = execSecure($dockerPsCmd);
        if ($dockerPsResult['success']) {
            $dockerContainers = count(array_filter(explode("\n", trim($dockerPsResult['output']))));
        }
    }

    // Check microservices network status
    $microservicesRunning = 0;
    $networkExists = false;

    if ($current === 'podman') {
        $networkCmd = 'podman network exists microservices-net 2>&1';
        $networkResult = execSecure($networkCmd);
        $networkExists = $networkResult['success'];

        if ($networkExists) {
            $psCmd = 'podman ps --filter network=microservices-net --format "{{.Names}}" 2>&1';
            $psResult = execSecure($psCmd);
            if ($psResult['success']) {
                $microservicesRunning = count(array_filter(explode("\n", trim($psResult['output']))));
            }
        }
    } elseif ($current === 'docker') {
        $networkCmd = 'docker network ls --filter name=microservices-net --format "{{.Name}}" 2>&1';
        $networkResult = execSecure($networkCmd);
        $networkExists = $networkResult['success'] && !empty(trim($networkResult['output']));

        if ($networkExists) {
            $psCmd = 'docker ps --filter network=microservices-net --format "{{.Names}}" 2>&1';
            $psResult = execSecure($psCmd);
            if ($psResult['success']) {
                $microservicesRunning = count(array_filter(explode("\n", trim($psResult['output']))));
            }
        }
    }

    $statusData = [
        'current' => $current,
        'available' => [
            'podman' => [
                'installed' => $podmanInstalled,
                'version' => $podmanVersion,
                'running' => $podmanRunning,
                'responsive' => $podmanResponsive
            ],
            'docker' => [
                'installed' => $dockerInstalled,
                'version' => $dockerVersion,
                'running' => $dockerRunning,
                'responsive' => $dockerResponsive
            ]
        ],
        'containers' => [
            'podman' => $podmanContainers,
            'docker' => $dockerContainers
        ],
        'microservices' => [
            'running' => $microservicesRunning,
            'network' => 'microservices-net',
            'networkExists' => $networkExists
        ]
    ];

    successResponse($statusData);
} catch (Exception $e) {
    errorResponse('Failed to get runtime status', $e->getMessage(), 500);
}
