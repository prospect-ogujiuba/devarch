# Testing & QA Tools

Browser automation, load testing, and quality assurance services for DevArch.

## Services

### Selenium Grid (Browser automation)
- **Files**: `selenium-hub.yml`, `selenium-chrome.yml`, `selenium-firefox.yml`
- **Port**: 10100 (hub)
- **Dependencies**: None (standalone)
- **Usage**:
  ```bash
  # Start hub and browser nodes
  docker compose \
    -f compose/testing/selenium-hub.yml \
    -f compose/testing/selenium-chrome.yml \
    -f compose/testing/selenium-firefox.yml \
    up -d
  ```
- **Access**: http://localhost:10100

### k6 (Load testing)
- **File**: `k6.yml`
- **Port**: 10110
- **Dependencies**: None (optional: InfluxDB for metrics)
- **Usage**:
  ```bash
  docker compose -f compose/testing/k6.yml up -d
  ```
- **Note**: Place test scripts in k6_scripts volume

### Playwright (Modern browser testing)
- **File**: `playwright.yml`
- **Dependencies**: None (standalone)
- **Usage**:
  ```bash
  docker compose -f compose/testing/playwright.yml up -d
  ```
- **Note**: Interactive container for test execution

### Gatling (Load testing)
- **File**: `gatling.yml`
- **Port**: 10130
- **Dependencies**: None (standalone)
- **Usage**:
  ```bash
  docker compose -f compose/testing/gatling.yml up -d
  ```
- **Note**: Place simulations in gatling_simulations volume

## Port Range
10100-10199
