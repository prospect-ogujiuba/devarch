<?php
/**
 * POST /api/services/start
 * Advanced service start with dependency ordering and health checks
 */

require_once __DIR__ . '/../lib/common.php';
require_once __DIR__ . '/../lib/services.php';

// Parse JSON body
$input = getJsonBody();

// Validate targets
$targets = $input['targets'] ?? [];
if (empty($targets) || !is_array($targets)) {
    errorResponse('Invalid targets', 'Targets must be a non-empty array of categories or service names', 400);
}

// Parse options with defaults
$options = $input['options'] ?? [];
$defaultOptions = [
    'parallel' => false,
    'waitHealthy' => false,
    'healthTimeout' => 120,
    'rebuildServices' => false,
    'forceRecreate' => false,
    'noCache' => false,
    'restartPolicy' => null,
    'excludeCategories' => [],
    'exceptServices' => [],
    'dryRun' => false
];

$options = array_merge($defaultOptions, $options);

// Validate each target
foreach ($targets as $target) {
    if (!is_string($target) || empty($target)) {
        errorResponse('Invalid target', 'Each target must be a non-empty string', 400);
    }
}

// Dry run mode - just return what would be executed
if ($options['dryRun']) {
    $servicesToStart = [];

    foreach ($targets as $target) {
        if (validateCategoryName($target)) {
            $services = getServicesInCategory($target);
            foreach ($services as $service) {
                if (!in_array($service, $options['exceptServices'])) {
                    $servicesToStart[] = $service;
                }
            }
        } else {
            $servicesToStart[] = $target;
        }
    }

    successResponse([
        'dryRun' => true,
        'servicesToStart' => $servicesToStart,
        'estimatedCount' => count($servicesToStart),
        'options' => $options
    ]);
}

// Map options to service start options
$serviceOptions = [
    'waitHealthy' => $options['waitHealthy'],
    'healthTimeout' => $options['healthTimeout'],
    'rebuild' => $options['rebuildServices'],
    'forceRecreate' => $options['forceRecreate'],
    'noCache' => $options['noCache'],
    'exceptServices' => $options['exceptServices']
];

// Start services
try {
    $result = startServices($targets, $serviceOptions);

    if ($result['success']) {
        successResponse($result);
    } else {
        // Partial failure - some services started, some failed
        errorResponse(
            'Some services failed to start',
            count($result['failed']) . ' service(s) failed',
            500,
        );
    }
} catch (Exception $e) {
    errorResponse('Operation failed', $e->getMessage(), 500);
}
