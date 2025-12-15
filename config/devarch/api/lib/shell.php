<?php
/**
 * Shell Command Execution Helpers
 * Secure shell command execution with validation and error handling
 */

/**
 * Execute shell command with security validation
 *
 * @param string $command Command to execute
 * @param bool $background Run in background
 * @return array ['success' => bool, 'output' => string, 'returnCode' => int]
 */
function execSecure(string $command, bool $background = false): array {
    // Prepend CONTAINER_HOST if set (needed for podman commands via PHP-FPM)
    $containerHost = getenv('CONTAINER_HOST');
    if ($containerHost && (str_contains($command, 'podman') || str_contains($command, 'docker'))) {
        $command = 'CONTAINER_HOST=' . escapeshellarg($containerHost) . ' ' . $command;
    }

    if ($background) {
        $command .= ' > /dev/null 2>&1 &';
    } else {
        $command .= ' 2>&1';
    }

    $output = [];
    $returnCode = 0;

    exec($command, $output, $returnCode);

    return [
        'success' => $returnCode === 0,
        'output' => implode("\n", $output),
        'returnCode' => $returnCode
    ];
}

/**
 * Validate service name (alphanumeric, dash, underscore only)
 */
function validateServiceName(string $name): bool {
    return preg_match('/^[a-zA-Z0-9_-]+$/', $name) === 1;
}

/**
 * Validate category name
 */
function validateCategoryName(string $name): bool {
    $validCategories = [
        'database', 'storage', 'dbms', 'erp', 'security', 'registry',
        'gateway', 'proxy', 'management', 'backend', 'ci', 'project',
        'mail', 'exporters', 'analytics', 'messaging', 'search',
        'workflow', 'docs', 'testing', 'collaboration', 'ai', 'support'
    ];

    return in_array($name, $validCategories, true);
}

/**
 * Escape shell argument safely
 */
function escapeArg(string $arg): string {
    return escapeshellarg($arg);
}

/**
 * Build podman-compose command
 */
function buildComposeCommand(string $composeFile, string $action, array $options = []): string {
    $cmd = 'podman-compose -f ' . escapeArg($composeFile) . ' ' . $action;

    foreach ($options as $key => $value) {
        if ($value === true) {
            $cmd .= ' --' . $key;
        } elseif ($value !== false && $value !== null) {
            $cmd .= ' --' . $key . ' ' . escapeArg((string)$value);
        }
    }

    return $cmd;
}

/**
 * Get podman container status
 */
function getPodmanContainerStatus(string $container): ?string {
    $cmd = 'podman inspect ' . escapeArg($container) . ' --format "{{.State.Status}}" 2>&1';
    $result = execSecure($cmd);

    if (!$result['success']) {
        return null;
    }

    return trim($result['output']);
}

/**
 * Check if podman container exists
 */
function podmanContainerExists(string $container): bool {
    $cmd = 'podman inspect ' . escapeArg($container) . ' 2>/dev/null';
    $result = execSecure($cmd);

    return $result['success'];
}

/**
 * Get systemd service status
 */
function getSystemdServiceStatus(string $service, bool $user = true): ?string {
    $cmd = $user
        ? 'systemctl --user is-active ' . escapeArg($service) . ' 2>&1'
        : 'sudo systemctl is-active ' . escapeArg($service) . ' 2>&1';

    $result = execSecure($cmd);

    return trim($result['output']);
}

/**
 * Check if systemd service is active
 */
function isSystemdServiceActive(string $service, bool $user = true): bool {
    $status = getSystemdServiceStatus($service, $user);
    return $status === 'active';
}

/**
 * Start systemd service
 */
function startSystemdService(string $service, bool $user = true): bool {
    $cmd = $user
        ? 'systemctl --user start ' . escapeArg($service) . ' 2>&1'
        : 'sudo systemctl start ' . escapeArg($service) . ' 2>&1';

    $result = execSecure($cmd);

    return $result['success'];
}

/**
 * Stop systemd service
 */
function stopSystemdService(string $service, bool $user = true): bool {
    $cmd = $user
        ? 'systemctl --user stop ' . escapeArg($service) . ' 2>&1'
        : 'sudo systemctl stop ' . escapeArg($service) . ' 2>&1';

    $result = execSecure($cmd);

    return $result['success'];
}

/**
 * Enable systemd linger for user
 */
function enableSystemdLinger(string $user): bool {
    $cmd = 'sudo loginctl enable-linger ' . escapeArg($user) . ' 2>&1';
    $result = execSecure($cmd);

    return $result['success'];
}

/**
 * Check if file/socket exists
 */
function fileOrSocketExists(string $path): bool {
    return file_exists($path) || is_link($path);
}

/**
 * Get current user
 */
function getCurrentUser(): string {
    return trim(shell_exec('whoami'));
}

/**
 * Get current user UID
 */
function getCurrentUid(): int {
    return (int)trim(shell_exec('id -u'));
}

/**
 * Execute command from scripts directory
 * Wraps service-manager.sh and other scripts
 */
function execScriptCommand(string $script, array $args = []): array {
    $scriptPath = '/workspace/scripts/' . $script;

    if (!file_exists($scriptPath)) {
        return [
            'success' => false,
            'output' => "Script not found: {$script}",
            'returnCode' => 127
        ];
    }

    $cmd = $scriptPath;
    foreach ($args as $arg) {
        $cmd .= ' ' . escapeArg($arg);
    }

    return execSecure($cmd);
}

/**
 * Check if running in rootless or rootful mode
 */
function isRootlessPodman(): bool {
    // Check if podman socket is in user directory
    $userSocket = '/run/user/' . getCurrentUid() . '/podman/podman.sock';
    return fileOrSocketExists($userSocket);
}

/**
 * Get active Podman socket path
 */
function getActivePodmanSocket(): ?string {
    $rootlessSocket = '/run/user/' . getCurrentUid() . '/podman/podman.sock';
    $rootfulSocket = '/run/podman/podman.sock';

    if (fileOrSocketExists($rootlessSocket)) {
        return $rootlessSocket;
    } elseif (fileOrSocketExists($rootfulSocket)) {
        return $rootfulSocket;
    }

    return null;
}

/**
 * Test Podman socket connectivity
 */
function testPodmanSocketConnectivity(string $socketPath): bool {
    if (!fileOrSocketExists($socketPath)) {
        return false;
    }

    putenv("CONTAINER_HOST=unix://{$socketPath}");
    $result = execSecure('podman version 2>&1');

    return $result['success'];
}

/**
 * Parse age string to seconds (e.g., "7d" -> 604800)
 */
function parseAgeToSeconds(string $age): int {
    if (preg_match('/^(\d+)([smhdw])$/', $age, $matches)) {
        $value = (int)$matches[1];
        $unit = $matches[2];

        return match($unit) {
            's' => $value,
            'm' => $value * 60,
            'h' => $value * 3600,
            'd' => $value * 86400,
            'w' => $value * 604800,
            default => 0
        };
    }

    return 0;
}
