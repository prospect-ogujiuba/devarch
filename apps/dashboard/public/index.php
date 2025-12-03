<?php
/**
 * DevArch Dashboard
 *
 * A simple, self-contained PHP application that automatically scans the apps/
 * directory, detects application types, categorizes them by runtime and framework,
 * and displays their status in a clean minimal UI.
 *
 * @version 1.0.0
 * @author DevArch
 */

// =============================================================================
// CONFIGURATION
// =============================================================================

// Auto-detect base path (works in container and on host)
// Check if we're in the apps/dashboard or html/dashboard structure
$detectedBasePath = dirname(__DIR__, 2);
$baseName = basename($detectedBasePath);
if ($baseName === 'apps' || $baseName === 'html') {
    // We're in apps/dashboard/public or /var/www/html/dashboard/public
    define('APPS_BASE_PATH', $detectedBasePath);
} else {
    // Fallback to /var/www/html (container path)
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
    // For now, return 'unknown' - can be enhanced to check container status
    // or ping ports in future versions

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

// =============================================================================
// MAIN LOGIC
// =============================================================================

// Get filter from query parameter
$filter = $_GET['filter'] ?? 'all';

// Scan apps directory
$appNames = scanApps(APPS_BASE_PATH, EXCLUDE_DIRS);

// Build app data
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
    ];

    $stats['total']++;
    if (isset($stats[$runtime])) {
        $stats[$runtime]++;
    }
}

// Filter apps if needed
$filteredApps = $filter === 'all'
    ? $apps
    : array_filter($apps, fn($app) => $app['runtime'] === $filter);

?>
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>DevArch Dashboard</title>
    <style>
        :root {
            --color-php: #8892BF;
            --color-node: #68A063;
            --color-python: #3776AB;
            --color-go: #00ADD8;
            --color-dotnet: #512BD4;
            --color-active: #22c55e;
            --color-ready: #eab308;
            --color-stopped: #94a3b8;
            --color-unknown: #cbd5e1;
            --bg-primary: #ffffff;
            --bg-secondary: #f8fafc;
            --bg-tertiary: #f1f5f9;
            --text-primary: #0f172a;
            --text-secondary: #64748b;
            --text-tertiary: #94a3b8;
            --border: #e2e8f0;
            --shadow: 0 1px 3px 0 rgb(0 0 0 / 0.1);
            --shadow-lg: 0 10px 15px -3px rgb(0 0 0 / 0.1);
        }

        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            background: var(--bg-secondary);
            color: var(--text-primary);
            line-height: 1.6;
        }

        .container {
            max-width: 1400px;
            margin: 0 auto;
            padding: 2rem;
        }

        header {
            background: var(--bg-primary);
            border-bottom: 1px solid var(--border);
            margin-bottom: 2rem;
            box-shadow: var(--shadow);
        }

        .header-content {
            max-width: 1400px;
            margin: 0 auto;
            padding: 1.5rem 2rem;
        }

        h1 {
            font-size: 1.875rem;
            font-weight: 700;
            color: var(--text-primary);
            margin-bottom: 0.5rem;
        }

        .subtitle {
            color: var(--text-secondary);
            font-size: 0.875rem;
        }

        .stats {
            background: var(--bg-primary);
            border-radius: 0.5rem;
            padding: 1.5rem;
            margin-bottom: 1.5rem;
            box-shadow: var(--shadow);
            border: 1px solid var(--border);
        }

        .stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
            gap: 1.5rem;
        }

        .stat-item {
            text-align: center;
        }

        .stat-value {
            font-size: 2rem;
            font-weight: 700;
            color: var(--text-primary);
            display: block;
        }

        .stat-label {
            font-size: 0.875rem;
            color: var(--text-secondary);
            text-transform: uppercase;
            letter-spacing: 0.05em;
            margin-top: 0.25rem;
        }

        .filters {
            display: flex;
            gap: 0.5rem;
            margin-bottom: 2rem;
            flex-wrap: wrap;
        }

        .filter-btn {
            padding: 0.5rem 1rem;
            border: 2px solid var(--border);
            background: var(--bg-primary);
            color: var(--text-secondary);
            border-radius: 0.375rem;
            cursor: pointer;
            font-size: 0.875rem;
            font-weight: 500;
            text-decoration: none;
            transition: all 0.2s;
        }

        .filter-btn:hover {
            border-color: var(--text-secondary);
            color: var(--text-primary);
        }

        .filter-btn.active {
            background: var(--text-primary);
            color: white;
            border-color: var(--text-primary);
        }

        .apps-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
            gap: 1.5rem;
        }

        .app-card {
            background: var(--bg-primary);
            border: 1px solid var(--border);
            border-radius: 0.5rem;
            padding: 1.5rem;
            box-shadow: var(--shadow);
            transition: all 0.2s;
        }

        .app-card:hover {
            box-shadow: var(--shadow-lg);
            transform: translateY(-2px);
        }

        .app-header {
            display: flex;
            justify-content: space-between;
            align-items: start;
            margin-bottom: 1rem;
        }

        .app-name {
            font-size: 1.25rem;
            font-weight: 600;
            color: var(--text-primary);
            font-family: 'Monaco', 'Menlo', 'Courier New', monospace;
        }

        .status-indicator {
            width: 12px;
            height: 12px;
            border-radius: 50%;
            margin-top: 0.25rem;
        }

        .app-meta {
            display: flex;
            gap: 0.5rem;
            margin-bottom: 1rem;
            flex-wrap: wrap;
        }

        .badge {
            display: inline-block;
            padding: 0.25rem 0.75rem;
            border-radius: 0.25rem;
            font-size: 0.75rem;
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 0.025em;
        }

        .runtime-badge {
            color: white;
        }

        .framework-badge {
            background: var(--bg-tertiary);
            color: var(--text-secondary);
        }

        .app-url {
            font-size: 0.875rem;
            color: var(--text-secondary);
            font-family: 'Monaco', 'Menlo', 'Courier New', monospace;
            margin-bottom: 1rem;
            word-break: break-all;
        }

        .app-actions {
            display: flex;
            gap: 0.5rem;
        }

        .action-link {
            flex: 1;
            text-align: center;
            padding: 0.5rem;
            background: var(--bg-tertiary);
            color: var(--text-primary);
            text-decoration: none;
            border-radius: 0.25rem;
            font-size: 0.875rem;
            font-weight: 500;
            transition: all 0.2s;
        }

        .action-link:hover {
            background: var(--text-primary);
            color: white;
        }

        .action-link.primary {
            background: var(--text-primary);
            color: white;
        }

        .action-link.primary:hover {
            background: #1e293b;
        }

        .empty-state {
            text-align: center;
            padding: 4rem 2rem;
            color: var(--text-secondary);
        }

        .empty-state-icon {
            font-size: 4rem;
            margin-bottom: 1rem;
        }

        .empty-state-title {
            font-size: 1.5rem;
            font-weight: 600;
            margin-bottom: 0.5rem;
            color: var(--text-primary);
        }

        @media (max-width: 768px) {
            .container {
                padding: 1rem;
            }

            .header-content {
                padding: 1rem;
            }

            h1 {
                font-size: 1.5rem;
            }

            .apps-grid {
                grid-template-columns: 1fr;
            }

            .stats-grid {
                grid-template-columns: repeat(2, 1fr);
            }
        }

        @media (max-width: 480px) {
            .stat-value {
                font-size: 1.5rem;
            }

            .stat-label {
                font-size: 0.75rem;
            }
        }
    </style>
</head>
<body>
<header>
    <div class="header-content">
        <h1>DevArch Dashboard</h1>
        <p class="subtitle">Application runtime detection and monitoring</p>
    </div>
</header>

<div class="container">
    <!-- Statistics -->
    <div class="stats">
        <div class="stats-grid">
            <div class="stat-item">
                <span class="stat-value"><?= $stats['total'] ?></span>
                <span class="stat-label">Total Apps</span>
            </div>
            <div class="stat-item">
                <span class="stat-value"
                      style="color: var(--color-php)"><?= $stats['php'] ?></span>
                <span class="stat-label">PHP</span>
            </div>
            <div class="stat-item">
                <span class="stat-value"
                      style="color: var(--color-node)"><?= $stats['node'] ?></span>
                <span class="stat-label">Node.js</span>
            </div>
            <div class="stat-item">
                <span class="stat-value"
                      style="color: var(--color-python)"><?= $stats['python'] ?></span>
                <span class="stat-label">Python</span>
            </div>
            <div class="stat-item">
                <span class="stat-value"
                      style="color: var(--color-go)"><?= $stats['go'] ?></span>
                <span class="stat-label">Go</span>
            </div>
            <div class="stat-item">
                <span class="stat-value"
                      style="color: var(--color-dotnet)"><?= $stats['dotnet'] ?></span>
                <span class="stat-label">.NET</span>
            </div>
        </div>
    </div>

    <!-- Filters -->
    <div class="filters">
        <a href="?filter=all" class="filter-btn <?= $filter === 'all' ? 'active' : '' ?>">All
            Apps</a>
        <a href="?filter=php" class="filter-btn <?= $filter === 'php' ? 'active' : '' ?>">PHP</a>
        <a href="?filter=node"
           class="filter-btn <?= $filter === 'node' ? 'active' : '' ?>">Node.js</a>
        <a href="?filter=python"
           class="filter-btn <?= $filter === 'python' ? 'active' : '' ?>">Python</a>
        <a href="?filter=go"
           class="filter-btn <?= $filter === 'go' ? 'active' : '' ?>">Go</a>
        <a href="?filter=dotnet"
           class="filter-btn <?= $filter === 'dotnet' ? 'active' : '' ?>">.NET</a>
    </div>

    <!-- Apps Grid -->
    <?php if (empty($filteredApps)): ?>
        <div class="empty-state">
            <div class="empty-state-icon">📦</div>
            <div class="empty-state-title">No applications found</div>
            <p>No applications matching the selected filter were detected.</p>
        </div>
    <?php else: ?>
        <div class="apps-grid">
            <?php foreach ($filteredApps as $app): ?>
                <div class="app-card">
                    <div class="app-header">
                        <div class="app-name"><?= htmlspecialchars($app['name']) ?></div>
                        <div class="status-indicator"
                             style="background-color: <?= $app['statusColor'] ?>"
                             title="Status: <?= ucfirst($app['status']) ?>"></div>
                    </div>

                    <div class="app-meta">
                            <span class="badge runtime-badge"
                                  style="background-color: <?= $app['color'] ?>">
                                <?= strtoupper($app['runtime']) ?>
                            </span>
                        <span class="badge framework-badge">
                                <?= htmlspecialchars($app['framework']) ?>
                            </span>
                    </div>

                    <div class="app-url"><?= htmlspecialchars($app['url']) ?></div>

                    <div class="app-actions">
                        <a href="<?= htmlspecialchars($app['url']) ?>"
                           class="action-link primary"
                           target="_blank"
                           rel="noopener noreferrer">Open</a>
                    </div>
                </div>
            <?php endforeach; ?>
        </div>
    <?php endif; ?>
</div>
</body>
</html>
