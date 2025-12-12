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

// Open process pipe for streaming
$handle = popen($cmd, 'r');
if (!$handle) {
    echo "data: " . json_encode(['error' => 'Failed to start log stream']) . "\n\n";
    flush();
    exit;
}

// Stream logs line by line
$lineCount = 0;
$maxLines = 1000; // Prevent memory exhaustion

// Set stream to non-blocking for better performance
stream_set_blocking($handle, false);

while (!feof($handle)) {
    $line = fgets($handle);

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

    // Small delay to prevent CPU spinning on non-blocking stream
    if (!$follow && $line === false) {
        break; // No more data and not following
    }
    usleep(10000); // 10ms delay

    // Check if client disconnected
    if (connection_status() != CONNECTION_NORMAL) {
        break;
    }
}

// Cleanup
pclose($handle);

// Send completion message
echo "data: " . json_encode(['complete' => true]) . "\n\n";
flush();
