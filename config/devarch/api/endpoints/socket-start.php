<?php
/**
 * POST /api/socket/start
 * Start Podman socket (rootless or rootful)
 */

require_once __DIR__ . '/../lib/common.php';
require_once __DIR__ . '/../lib/shell.php';

// Parse JSON body
$input = getJsonBody();

// Validate socket type
$type = $input['type'] ?? 'rootless';
if (!in_array($type, ['rootless', 'rootful'])) {
    errorResponse('Invalid type', 'Type must be "rootless" or "rootful"', 400);
}

// Parse options
$options = $input['options'] ?? [];
$enableLingering = $options['enableLingering'] ?? true;
$stopConflicting = $options['stopConflicting'] ?? true;

try {
    $user = getCurrentUser();
    $uid = getCurrentUid();
    $isUser = $type === 'rootless';

    // Stop conflicting socket if requested
    if ($stopConflicting) {
        $conflictingType = $type === 'rootless' ? 'rootful' : 'rootless';
        $conflictingIsUser = $conflictingType === 'rootless';

        if (isSystemdServiceActive('podman.socket', $conflictingIsUser)) {
            $stopped = stopSystemdService('podman.socket', $conflictingIsUser);
            if (!$stopped) {
                errorResponse(
                    'Failed to stop conflicting socket',
                    "Could not stop {$conflictingType} socket",
                    500
                );
            }
        }
    }

    // Enable lingering for rootless
    if ($type === 'rootless' && $enableLingering) {
        $lingerEnabled = enableSystemdLinger($user);
        if (!$lingerEnabled) {
            // Log warning but continue
            error_log("Warning: Failed to enable lingering for user {$user}");
        }
    }

    // Start socket
    $started = startSystemdService('podman.socket', $isUser);
    if (!$started) {
        errorResponse(
            'Failed to start socket',
            "Could not start {$type} Podman socket",
            500
        );
    }

    // Wait a moment for socket to become available
    sleep(2);

    // Verify connectivity
    $socketPath = $type === 'rootless'
        ? "/run/user/{$uid}/podman/podman.sock"
        : '/run/podman/podman.sock';

    $connectivity = testPodmanSocketConnectivity($socketPath);
    if (!$connectivity) {
        errorResponse(
            'Socket unresponsive',
            'Socket started but is not responding to API calls',
            503
        );
    }

    // Get systemd status
    $systemdStatus = getSystemdServiceStatus('podman.socket', $isUser);

    successResponse([
        'type' => $type,
        'socketPath' => $socketPath,
        'status' => $systemdStatus,
        'connectivity' => 'working',
        'message' => ucfirst($type) . ' socket started successfully',
        'environmentSetup' => "export DOCKER_HOST=\"unix://{$socketPath}\""
    ]);
} catch (Exception $e) {
    errorResponse('Operation failed', $e->getMessage(), 500);
}
