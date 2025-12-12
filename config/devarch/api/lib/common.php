<?php
/**
 * Common utilities for DevArch API
 */

// Constants
define('COMPOSE_BASE_PATH', '/workspace/compose');
define('APPS_BASE_PATH', '/workspace/apps');
define('NGINX_CONFIG_PATH', '/workspace/config/nginx/custom/http.conf');

/**
 * Send JSON response
 */
function jsonResponse(array $data, int $statusCode = 200): void {
    http_response_code($statusCode);
    echo json_encode($data, JSON_PRETTY_PRINT);
    exit;
}

/**
 * Send success response
 */
function successResponse(array $data, array $meta = []): void {
    $response = [
        'success' => true,
        'data' => $data
    ];

    if (!empty($meta)) {
        $response['meta'] = $meta;
    }

    jsonResponse($response);
}

/**
 * Send error response
 */
function errorResponse(string $error, string $message = '', int $statusCode = 400): void {
    jsonResponse([
        'success' => false,
        'error' => $error,
        'message' => $message
    ], $statusCode);
}

/**
 * Execute shell command safely
 */
function execCommand(string $command): array {
    $output = [];
    $returnCode = 0;

    exec($command . ' 2>&1', $output, $returnCode);

    return [
        'success' => $returnCode === 0,
        'output' => implode("\n", $output),
        'returnCode' => $returnCode
    ];
}

/**
 * Parse POST JSON body
 */
function getJsonBody(): array {
    $json = file_get_contents('php://input');
    $data = json_decode($json, true);

    if (json_last_error() !== JSON_ERROR_NONE) {
        errorResponse('Invalid JSON', json_last_error_msg());
    }

    return $data ?? [];
}
