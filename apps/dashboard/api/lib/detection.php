<?php
/**
 * DevArch Dashboard - Detection Library
 *
 * Contains all the detection logic for identifying application runtimes,
 * frameworks, and status. Extracted from the original monolithic index.php
 * for better maintainability and reusability.
 *
 * @version 2.0.0
 * @author DevArch
 */

// =============================================================================
// CONFIGURATION
// =============================================================================

// Auto-detect base path (works in container and on host)
$detectedBasePath = dirname(__DIR__, 3);
$baseName = basename($detectedBasePath);
if ($baseName === 'apps' || $baseName === 'html') {
    define('APPS_BASE_PATH', $detectedBasePath);
} else {
    define('APPS_BASE_PATH', '/var/www/html');
}

const SELF_DIR = 'dashboard';
const EXCLUDE_DIRS = ['.', '..', '.idea', SELF_DIR, 'serverinfo'];

// Runtime port mappings (from detect-app-runtime.sh)
const RUNTIME_PORTS = [
    'php' => ['external' => 8100, 'internal' => 8000, 'container' => 'php'],
    'node' => ['external' => 8200, 'internal' => 3000, 'container' => 'node'],
    'python' => ['external' => 8300, 'internal' => 8000, 'container' => 'python'],
    'go' => ['external' => 8400, 'internal' => 8080, 'container' => 'go'],
    'dotnet' => ['external' => 8600, 'internal' => 8080, 'container' => 'dotnet'],
];

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

/**
 * Scan apps directory and return list of app directories
 */
function scanApps(string $basePath, array $exclude = []): array
{
    if (!is_dir($basePath)) {
        return [];
    }

    $apps = [];
    $items = scandir($basePath);

    foreach ($items as $item) {
        if (in_array($item, $exclude)) {
            continue;
        }

        $itemPath = $basePath . '/' . $item;
        if (is_dir($itemPath)) {
            $apps[] = $item;
        }
    }

    sort($apps);
    return $apps;
}

/**
 * Detect the runtime type for an application
 */
function detectRuntime(string $appPath): string
{
    // PHP detection (highest priority)
    if (file_exists("$appPath/composer.json") ||
        file_exists("$appPath/artisan") ||
        file_exists("$appPath/public/wp-config.php") ||
        file_exists("$appPath/wp-config.php") ||
        file_exists("$appPath/index.php") ||
        file_exists("$appPath/public/index.php")) {
        return 'php';
    }

    // Node detection
    if (file_exists("$appPath/package.json")) {
        return 'node';
    }

    // Python detection
    if (file_exists("$appPath/requirements.txt") ||
        file_exists("$appPath/pyproject.toml") ||
        file_exists("$appPath/manage.py")) {
        return 'python';
    }

    // Go detection
    if (file_exists("$appPath/go.mod") ||
        file_exists("$appPath/main.go")) {
        return 'go';
    }

    // .NET detection
    $csprojFiles = glob("$appPath/*.csproj");
    $slnFiles = glob("$appPath/*.sln");
    $fsprojFiles = glob("$appPath/*.fsproj");

    if (!empty($csprojFiles) || !empty($slnFiles) || !empty($fsprojFiles) ||
        (file_exists("$appPath/Program.cs") && file_exists("$appPath/appsettings.json"))) {
        return 'dotnet';
    }

    return 'unknown';
}

/**
 * Detect the framework for an application based on runtime
 */
function detectFramework(string $appPath, string $runtime): string
{
    return match ($runtime) {
        'php' => detectPhpFramework($appPath),
        'node' => detectNodeFramework($appPath),
        'python' => detectPythonFramework($appPath),
        'go' => detectGoFramework($appPath),
        'dotnet' => detectDotnetFramework($appPath),
        default => 'Unknown'
    };
}

/**
 * Detect PHP framework
 */
function detectPhpFramework(string $appPath): string
{
    // Laravel detection
    if (file_exists("$appPath/artisan") && file_exists("$appPath/app")) {
        return 'Laravel';
    }

    // WordPress detection (check both root and public directory)
    if (file_exists("$appPath/wp-config.php") ||
        file_exists("$appPath/public/wp-config.php")) {
        return 'WordPress';
    }

    // Symfony detection
    if (file_exists("$appPath/composer.json")) {
        $composerPath = "$appPath/composer.json";
        $composer = json_decode(file_get_contents($composerPath), true);

        if (isset($composer['require'])) {
            if (isset($composer['require']['symfony/framework-bundle'])) {
                return 'Symfony';
            }
            if (isset($composer['require']['laravel/framework'])) {
                return 'Laravel';
            }
            if (isset($composer['require']['slim/slim'])) {
                return 'Slim';
            }
        }
    }

    // Generic PHP
    if (file_exists("$appPath/composer.json")) {
        return 'PHP (Composer)';
    }

    return 'PHP';
}

/**
 * Detect Node.js framework
 */
function detectNodeFramework(string $appPath): string
{
    if (file_exists("$appPath/next.config.js") || file_exists("$appPath/next.config.mjs")) {
        return 'Next.js';
    }

    if (file_exists("$appPath/package.json")) {
        $packagePath = "$appPath/package.json";
        $package = json_decode(file_get_contents($packagePath), true);

        if (isset($package['dependencies'])) {
            $deps = $package['dependencies'];

            if (isset($deps['next'])) return 'Next.js';
            if (isset($deps['nuxt'])) return 'Nuxt.js';
            if (isset($deps['react'])) return 'React';
            if (isset($deps['vue'])) return 'Vue.js';
            if (isset($deps['@angular/core'])) return 'Angular';
            if (isset($deps['express'])) return 'Express';
            if (isset($deps['fastify'])) return 'Fastify';
            if (isset($deps['koa'])) return 'Koa';
        }
    }

    return 'Node.js';
}

/**
 * Detect Python framework
 */
function detectPythonFramework(string $appPath): string
{
    // Django detection
    if (file_exists("$appPath/manage.py")) {
        return 'Django';
    }

    // Check requirements.txt
    if (file_exists("$appPath/requirements.txt")) {
        $requirements = strtolower(file_get_contents("$appPath/requirements.txt"));

        if (stripos($requirements, 'django') !== false) return 'Django';
        if (stripos($requirements, 'flask') !== false) return 'Flask';
        if (stripos($requirements, 'fastapi') !== false) return 'FastAPI';
        if (stripos($requirements, 'tornado') !== false) return 'Tornado';
        if (stripos($requirements, 'bottle') !== false) return 'Bottle';
    }

    // Check pyproject.toml
    if (file_exists("$appPath/pyproject.toml")) {
        $pyproject = strtolower(file_get_contents("$appPath/pyproject.toml"));

        if (stripos($pyproject, 'django') !== false) return 'Django';
        if (stripos($pyproject, 'flask') !== false) return 'Flask';
        if (stripos($pyproject, 'fastapi') !== false) return 'FastAPI';
    }

    return 'Python';
}

/**
 * Detect Go framework
 */
function detectGoFramework(string $appPath): string
{
    if (file_exists("$appPath/go.mod")) {
        $gomod = file_get_contents("$appPath/go.mod");

        if (stripos($gomod, 'gin-gonic/gin') !== false) return 'Gin';
        if (stripos($gomod, 'labstack/echo') !== false) return 'Echo';
        if (stripos($gomod, 'gofiber/fiber') !== false) return 'Fiber';
        if (stripos($gomod, 'gorilla/mux') !== false) return 'Gorilla Mux';
        if (stripos($gomod, 'chi') !== false) return 'Chi';
    }

    return 'Go';
}

/**
 * Detect .NET framework type
 */
function detectDotnetFramework(string $appPath): string
{
    $csprojFiles = glob("$appPath/*.csproj");

    if (empty($csprojFiles)) {
        if (file_exists("$appPath/appsettings.json") && file_exists("$appPath/Program.cs")) {
            return 'ASP.NET Core';
        }
        return '.NET';
    }

    $csproj = file_get_contents($csprojFiles[0]);

    if (stripos($csproj, 'Microsoft.AspNetCore.Components.WebAssembly') !== false) {
        return 'Blazor WebAssembly';
    }
    if (stripos($csproj, 'Microsoft.AspNetCore.Components') !== false) {
        return 'Blazor Server';
    }
    if (stripos($csproj, 'Sdk="Microsoft.NET.Sdk.Web"') !== false) {
        if (stripos($csproj, 'Swashbuckle.AspNetCore') !== false) {
            return 'ASP.NET Core Web API';
        }
        return 'ASP.NET Core';
    }
    if (stripos($csproj, 'Microsoft.Extensions.Hosting') !== false) {
        return '.NET Worker Service';
    }

    return '.NET';
}

/**
 * Get app status (simplified - based on file existence)
 */
function getAppStatus(string $appPath, string $runtime): string
{
    // Basic heuristic: if the app has proper structure files, assume it's ready
    $hasMainFiles = false;

    switch ($runtime) {
        case 'php':
            $hasMainFiles = file_exists("$appPath/public/index.php") ||
                file_exists("$appPath/index.php") ||
                file_exists("$appPath/public/wp-config.php");
            break;
        case 'node':
            $hasMainFiles = file_exists("$appPath/package.json");
            break;
        case 'python':
            $hasMainFiles = file_exists("$appPath/main.py") ||
                file_exists("$appPath/manage.py") ||
                file_exists("$appPath/app.py");
            break;
        case 'go':
            $hasMainFiles = file_exists("$appPath/main.go");
            break;
        case 'dotnet':
            $hasMainFiles = !empty(glob("$appPath/*.csproj")) ||
                file_exists("$appPath/Program.cs");
            break;
    }

    return $hasMainFiles ? 'ready' : 'unknown';
}

/**
 * Get app URL (.test domain)
 */
function getAppUrl(string $appName): string
{
    return "https://{$appName}.test";
}

/**
 * Get runtime color for badge
 */
function getRuntimeColor(string $runtime): string
{
    return match ($runtime) {
        'php' => '#8892BF',
        'node' => '#68A063',
        'python' => '#3776AB',
        'go' => '#00ADD8',
        'dotnet' => '#512BD4',
        default => '#94a3b8'
    };
}

/**
 * Get status color
 */
function getStatusColor(string $status): string
{
    return match ($status) {
        'active' => '#eab308',
        'ready' => '#22c55e',
        'stopped' => '#94a3b8',
        default => '#cbd5e1'
    };
}

/**
 * Get all apps with metadata
 */
function getAllApps(): array
{
    $appNames = scanApps(APPS_BASE_PATH, EXCLUDE_DIRS);
    $apps = [];
    $stats = [
        'total' => 0,
        'php' => 0,
        'node' => 0,
        'python' => 0,
        'go' => 0,
        'dotnet' => 0,
        'unknown' => 0,
    ];

    foreach ($appNames as $appName) {
        $appPath = APPS_BASE_PATH . '/' . $appName;
        $runtime = detectRuntime($appPath);
        $framework = detectFramework($appPath, $runtime);
        $status = getAppStatus($appPath, $runtime);
        $url = getAppUrl($appName);

        $apps[] = [
            'name' => $appName,
            'runtime' => $runtime,
            'framework' => $framework,
            'status' => $status,
            'url' => $url,
            'color' => getRuntimeColor($runtime),
            'statusColor' => getStatusColor($status),
            'path' => $appPath,
        ];

        $stats['total']++;
        if (isset($stats[$runtime])) {
            $stats[$runtime]++;
        }
    }

    return [
        'apps' => $apps,
        'stats' => $stats,
    ];
}
