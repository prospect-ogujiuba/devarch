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

    // Test podman connectivity (works from inside container)
    $podmanInfoCmd = 'podman info --format json 2>&1';
    $podmanInfoResult = execSecure($podmanInfoCmd);

    $active = 'none';
    $rootlessActive = false;
    $rootfulActive = false;

    if ($podmanInfoResult['success']) {
        $info = json_decode($podmanInfoResult['output'], true);
        if ($info && isset($info['host']['security']['rootless'])) {
            $isRootless = $info['host']['security']['rootless'];
            $active = $isRootless ? 'rootless' : 'rootful';
            $rootlessActive = $isRootless;
            $rootfulActive = !$isRootless;
        }
    }

    // Socket paths (for reference, may not be accessible from container)
    $rootlessSocket = "/run/user/{$uid}/podman/podman.sock";
    $rootfulSocket = '/run/podman/podman.sock';

    // Get systemd status
    $rootlessSystemdStatus = getSystemdServiceStatus('podman.socket', true);
    $rootfulSystemdStatus = getSystemdServiceStatus('podman.socket', false);

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
                'active' => $rootlessActive,
                'socketPath' => $rootlessSocket,
                'exists' => $rootlessActive,  // If podman works, socket exists
                'connectivity' => $rootlessActive ? 'working' : 'unavailable',
                'systemdStatus' => $rootlessSystemdStatus
            ],
            'rootful' => [
                'active' => $rootfulActive,
                'socketPath' => $rootfulSocket,
                'exists' => $rootfulActive,  // If podman works, socket exists
                'connectivity' => $rootfulActive ? 'working' : 'unavailable',
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
