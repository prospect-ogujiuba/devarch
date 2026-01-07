# Testing Patterns

**Analysis Date:** 2026-01-07

## Test Framework

**PHP (Laravel - mxcros):**
- Runner: PHPUnit 11.5.3
- Config: `apps/mxcros/phpunit.xml`
- Dependencies: fakerphp/faker, mockery/mockery, laravel/pail

**PHP (WordPress - makermaker):**
- Runner: Pest 2.34+
- Config: `apps/b2bcnc/wp-content/plugins/makermaker/phpunit.xml`
- Mocking: Brain Monkey 2.6+ for WordPress functions

**JavaScript (Dashboard):**
- No test framework configured
- ESLint only for linting

**Python (config/python):**
- Runner: pytest 7.4+
- Config: Not detected
- Dependencies: pytest-django, pytest-asyncio, pytest-cov

**Run Commands:**
```bash
# Laravel (mxcros)
cd apps/mxcros && vendor/bin/phpunit           # All tests
cd apps/mxcros && vendor/bin/phpunit --filter=ExampleTest  # Single test

# WordPress (makermaker)
cd apps/b2bcnc/wp-content/plugins/makermaker && vendor/bin/pest

# Python (if configured)
cd config/python && pytest                     # All tests
cd config/python && pytest --cov              # With coverage
```

## Test File Organization

**Location:**
- Laravel: `apps/mxcros/tests/` (Unit, Feature subdirectories)
- WordPress: `apps/b2bcnc/wp-content/plugins/makermaker/tests/`
- Dashboard: No tests

**Naming:**
- PHP: `*Test.php` suffix (`ExampleTest.php`)
- Test methods: `test_*` with snake_case (`test_that_true_is_true`)

**Structure:**
```
apps/mxcros/tests/
├── Feature/
│   └── ExampleTest.php
├── Unit/
│   └── ExampleTest.php
└── TestCase.php
```

## Test Structure

**Suite Organization (PHPUnit):**
```php
<?php

namespace Tests\Unit;

use PHPUnit\Framework\TestCase;

class ExampleTest extends TestCase
{
    public function test_that_true_is_true(): void
    {
        $this->assertTrue(true);
    }
}
```

**Feature Test Pattern:**
```php
<?php

namespace Tests\Feature;

use Tests\TestCase;

class ExampleTest extends TestCase
{
    public function test_the_application_returns_a_successful_response(): void
    {
        $response = $this->get('/');
        $response->assertStatus(200);
    }
}
```

**Patterns:**
- Use `setUp()` for per-test initialization
- Use `tearDown()` for cleanup
- One assertion focus per test
- Descriptive snake_case method names

## Mocking

**Framework:**
- Laravel: Mockery (built-in)
- WordPress: Brain Monkey for WP functions

**Patterns (Laravel):**
```php
use Mockery;

public function test_with_mock(): void
{
    $mock = Mockery::mock(SomeClass::class);
    $mock->shouldReceive('method')->once()->andReturn('value');

    // Use mock in test
}
```

**WordPress Mocking (Brain Monkey):**
```php
use Brain\Monkey\Functions;

public function test_wp_function(): void
{
    Functions\expect('get_option')
        ->once()
        ->with('my_option')
        ->andReturn('value');
}
```

**What to Mock:**
- External services (HTTP calls, APIs)
- WordPress functions (in unit tests)
- File system operations
- Time-dependent functions

**What NOT to Mock:**
- Simple utility functions
- Value objects
- The code under test

## Fixtures and Factories

**Test Data (Laravel):**
```php
// Factory usage
$user = User::factory()->create();
$users = User::factory()->count(5)->create();

// Inline test data
$data = [
    'name' => 'Test User',
    'email' => 'test@example.com',
];
```

**Location:**
- Laravel factories: `apps/mxcros/database/factories/`
- Fixtures: Inline or in test file

## Coverage

**Requirements:**
- No enforced coverage target
- Coverage tracked for awareness

**Configuration (Laravel):**
```xml
<!-- phpunit.xml -->
<source>
    <include>
        <directory>app</directory>
    </include>
</source>
```

**View Coverage:**
```bash
cd apps/mxcros && vendor/bin/phpunit --coverage-html coverage/
open coverage/index.html
```

## Test Types

**Unit Tests:**
- Scope: Single class/function in isolation
- Mocking: Mock all dependencies
- Location: `tests/Unit/`
- Speed: Fast (<100ms per test)

**Feature Tests:**
- Scope: HTTP endpoints, workflows
- Mocking: Mock external services only
- Location: `tests/Feature/`
- Setup: Use test database

**Integration Tests:**
- Scope: Multiple components together
- Not formally separated in this codebase

**E2E Tests:**
- Not configured
- Dashboard has no test coverage

## Common Patterns

**Async Testing (PHP):**
```php
// Laravel handles async via queue testing
Queue::fake();

// Trigger action that queues job
$this->post('/api/action');

Queue::assertPushed(SomeJob::class);
```

**Error Testing:**
```php
public function test_throws_on_invalid_input(): void
{
    $this->expectException(ValidationException::class);

    // Call code that should throw
    validateInput(null);
}
```

**HTTP Testing (Laravel):**
```php
public function test_api_endpoint(): void
{
    $response = $this->getJson('/api/users');

    $response->assertStatus(200)
             ->assertJsonStructure(['data' => ['users']]);
}
```

**Snapshot Testing:**
- Not used in this codebase

## Test Gaps

**Untested Areas:**
- Dashboard React components (no tests)
- PHP API endpoints (`config/devarch/api/`)
- Shell scripts (`scripts/*.sh`)
- Service orchestration flows

**Recommended Additions:**
- React Testing Library for dashboard
- PHPUnit for API endpoints
- Integration tests for service start/stop
- Shell script tests (bats or shunit2)

---

*Testing analysis: 2026-01-07*
*Update when test patterns change*
