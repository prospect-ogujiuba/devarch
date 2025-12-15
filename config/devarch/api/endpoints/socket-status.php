<?php
/**
 * GET /api/socket/status
 * Get Podman socket status (rootless vs rootful)
 */

require_once __DIR__ . '/../lib/common.php';
require_once __DIR__ . '/../lib/shell.php';

try {
    $uid = getCurrentUid();
    $user = getCurrentUser();

    // Check rootless socket
    $rootlessSocket = "/run/user/{$uid}/podman/podman.sock";
    $rootlessExists = fileOrSocketExists($rootlessSocket);
    $rootlessConnectivity = $rootlessExists ? testPodmanSocketConnectivity($rootlessSocket) : false;
    $rootlessSystemdStatus = getSystemdServiceStatus('podman.socket', true);

    // Check rootful socket
    $rootfulSocket = '/run/podman/podman.sock';
    $rootfulExists = fileOrSocketExists($rootfulSocket);
    $rootfulConnectivity = $rootfulExists ? testPodmanSocketConnectivity($rootfulSocket) : false;
    $rootfulSystemdStatus = getSystemdServiceStatus('podman.socket', false);

    // Determine active socket
    $active = 'none';
    if ($rootlessConnectivity) {
        $active = 'rootless';
    } elseif ($rootfulConnectivity) {
        $active = 'rootful';
    }

    // Get environment info
    $dockerHost = getenv('DOCKER_HOST') ?: getenv('CONTAINER_HOST') ?: 'not set';

    // Check network integration
    $networkCmd = 'podman network exists microservices-net 2>&1';
    $networkResult = execSecure($networkCmd);
    $networkExists = $networkResult['success'];

    // Count running services
    $psCmd = 'podman ps --filter network=microservices-net --format "{{.Names}}" 2>&1';
    $psResult = execSecure($psCmd);
    $runningServices = $psResult['success'] ? count(array_filter(explode("\n", trim($psResult['output'])))) : 0;

    $statusData = [
        'active' => $active,
        'sockets' => [
            'rootless' => [
                'active' => $active === 'rootless',
                'socketPath' => $rootlessSocket,
                'exists' => $rootlessExists,
                'connectivity' => $rootlessConnectivity ? 'working' : 'unavailable',
                'systemdStatus' => $rootlessSystemdStatus
            ],
            'rootful' => [
                'active' => $active === 'rootful',
                'socketPath' => $rootfulSocket,
                'exists' => $rootfulExists,
                'connectivity' => $rootfulConnectivity ? 'working' : 'unavailable',
                'systemdStatus' => $rootfulSystemdStatus
            ]
        ],
        'environment' => [
            'dockerHost' => $dockerHost,
            'user' => $user,
            'uid' => $uid
        ],
        'integration' => [
            'projectNetwork' => 'microservices-net',
            'networkExists' => $networkExists,
            'runningServices' => $runningServices
        ]
    ];

    successResponse($statusData);
} catch (Exception $e) {
    errorResponse('Failed to get socket status', $e->getMessage(), 500);
}
