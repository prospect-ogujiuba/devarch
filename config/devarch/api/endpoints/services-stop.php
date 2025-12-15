<?php
/**
 * POST /api/services/stop
 * Advanced service stop with cleanup options
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
    'timeout' => 30,
    'removeVolumes' => false,
    'cleanupOrphans' => false,
    'preserveData' => true,
    'dryRun' => false
];

$options = array_merge($defaultOptions, $options);

// Safety check: prevent accidental data loss
if ($options['removeVolumes'] && $options['preserveData']) {
    errorResponse(
        'Conflicting options',
        'Cannot set both removeVolumes and preserveData to true',
        400
    );
}

// Validate each target
foreach ($targets as $target) {
    if (!is_string($target) || empty($target)) {
        errorResponse('Invalid target', 'Each target must be a non-empty string', 400);
    }
}

// Dry run mode
if ($options['dryRun']) {
    $servicesToStop = [];

    foreach ($targets as $target) {
        if (validateCategoryName($target)) {
            $services = getServicesInCategory($target);
            $servicesToStop = array_merge($servicesToStop, $services);
        } else {
            $servicesToStop[] = $target;
        }
    }

    successResponse([
        'dryRun' => true,
        'servicesToStop' => $servicesToStop,
        'estimatedCount' => count($servicesToStop),
        'options' => $options,
        'warning' => $options['removeVolumes'] ? 'Will remove volumes (data loss)' : null
    ]);
}

// Map options to service stop options
$serviceOptions = [
    'timeout' => $options['timeout'],
    'removeVolumes' => $options['removeVolumes']
];

// Stop services
try {
    $result = stopServices($targets, $serviceOptions);

    // Add cleanup info if applicable
    $result['cleanup'] = [
        'orphanContainersRemoved' => 0,
        'imagesRemoved' => 0,
        'volumesRemoved' => $options['removeVolumes'] ? count($result['stopped']) : 0
    ];

    if ($result['success']) {
        successResponse($result);
    } else {
        errorResponse(
            'Some services failed to stop',
            count($result['failed']) . ' service(s) failed',
            500
        );
    }
} catch (Exception $e) {
    errorResponse('Operation failed', $e->getMessage(), 500);
}
