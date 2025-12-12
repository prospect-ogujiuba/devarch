<?php
/**
 * DevArch API - Domains Endpoint
 * Returns .test domains parsed from nginx config
 */

require_once __DIR__ . '/../lib/common.php';
require_once __DIR__ . '/../lib/containers.php';

try {
    $domains = parseNginxDomains();

    successResponse([
        'domains' => $domains
    ], [
        'timestamp' => time()
    ]);

} catch (Exception $e) {
    errorResponse('Internal server error', $e->getMessage(), 500);
}
