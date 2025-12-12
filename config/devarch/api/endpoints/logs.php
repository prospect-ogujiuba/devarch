<?php
/**
 * DevArch API - Container Logs Endpoint
 * Server-Sent Events (SSE) endpoint for streaming container logs
 */

require_once __DIR__ . '/../lib/containers.php';

// Enable SSE headers
header('Content-Type: text/event-stream');
header('Cache-Control: no-cache');
header('Connection: keep-alive');
header('Access-Control-Allow-Origin: *');
header('X-Accel-Buffering: no'); // Disable nginx buffering

// Flush output immediately
ob_implicit_flush(true);
if (ob_get_level()) ob_end_clean();

// Disable script timeout for long-running streams
set_time_limit(0);

// Abort script if client disconnects
ignore_user_abort(false);

// Direct Podman access in devarch container
putenv('CONTAINER_HOST=unix:///run/podman/podman.sock');
$podmanBinary = PODMAN_BINARY;

// Get parameters
$container = $_GET['container'] ?? '';
$follow = ($_GET['follow'] ?? 'false') === 'true';
$tail = intval($_GET['tail'] ?? 100);

// Validate container name
if (!preg_match('/^[a-zA-Z0-9_-]+$/', $container)) {
    echo "data: " . json_encode(['error' => 'Invalid container name']) . "\n\n";
    flush();
    exit;
}

// Log access for security audit
error_log("Container logs: streaming from {$container} (follow: " . ($follow ? 'yes' : 'no') . ")");

// Build podman logs command
$cmd = "{$podmanBinary} logs --tail {$tail}";
if ($follow) {
    $cmd .= " -f";
}
$cmd .= " " . escapeshellarg($container) . " 2>&1";

// Process handle for cleanup
$process = null;
$pipes = null;

// Cleanup function to kill process on shutdown
$cleanup = function() use (&$process, &$pipes) {
    if (is_resource($pipes[1] ?? null)) {
        fclose($pipes[1]);
    }
    if (is_resource($process)) {
        $status = proc_get_status($process);
        if ($status['running']) {
            // Send SIGTERM to gracefully stop podman logs
            proc_terminate($process, 15);
            // Wait up to 2 seconds for graceful shutdown
            $timeout = 2;
            $start = time();
            while (proc_get_status($process)['running'] && (time() - $start) < $timeout) {
                usleep(100000); // 100ms
            }
            // Force kill if still running
            if (proc_get_status($process)['running']) {
                proc_terminate($process, 9); // SIGKILL
            }
        }
        proc_close($process);
    }
};

// Register shutdown function to ensure cleanup
register_shutdown_function($cleanup);

// Open process with proc_open for better control
$descriptorspec = [
    0 => ['pipe', 'r'],  // stdin
    1 => ['pipe', 'w'],  // stdout
    2 => ['pipe', 'w']   // stderr
];

$process = proc_open($cmd, $descriptorspec, $pipes);

if (!is_resource($process)) {
    echo "data: " . json_encode(['error' => 'Failed to start log stream']) . "\n\n";
    flush();
    exit;
}

// Close stdin - we don't need it
fclose($pipes[0]);

// Merge stderr into stdout for error messages
stream_set_blocking($pipes[1], false);
stream_set_blocking($pipes[2], false);

// Stream logs line by line
$lineCount = 0;
$maxLines = 1000; // Prevent memory exhaustion

while (true) {
    // Check if client disconnected
    if (connection_aborted() || connection_status() != CONNECTION_NORMAL) {
        error_log("Container logs: client disconnected from {$container}");
        break;
    }

    // Check if process is still running
    $status = proc_get_status($process);
    if (!$status['running'] && feof($pipes[1])) {
        break;
    }

    // Read from stdout
    $line = fgets($pipes[1]);

    // Also check stderr for errors
    if ($line === false) {
        $errLine = fgets($pipes[2]);
        if ($errLine !== false) {
            echo "data: " . json_encode(['line' => '[ERROR] ' . rtrim($errLine)]) . "\n\n";
            flush();
        }
    }

    if ($line !== false) {
        // Send log line as SSE event
        echo "data: " . json_encode(['line' => rtrim($line)]) . "\n\n";
        flush();

        $lineCount++;

        // Stop after max lines to prevent memory issues
        if ($lineCount > $maxLines) {
            echo "data: " . json_encode([
                'warning' => "Maximum line limit ({$maxLines}) reached. Stream stopped."
            ]) . "\n\n";
            flush();
            break;
        }
    }

    // Small delay to prevent CPU spinning
    if (!$follow && $line === false) {
        break; // No more data and not following
    }
    usleep(10000); // 10ms delay
}

// Cleanup will be called by shutdown function
// Send completion message
echo "data: " . json_encode(['complete' => true]) . "\n\n";
flush();
