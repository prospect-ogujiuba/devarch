<?php
/**
 * GET /api/services/status
 * Enhanced service status with categories and health information
 */

require_once __DIR__ . '/../lib/common.php';
require_once __DIR__ . '/../lib/services.php';

try {
    $statusData = getEnhancedServiceStatus();
    successResponse($statusData);
} catch (Exception $e) {
    errorResponse('Failed to get status', $e->getMessage(), 500);
}
