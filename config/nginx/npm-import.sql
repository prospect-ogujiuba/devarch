-- ============================================================================
-- NGINX PROXY MANAGER - HTTP.CONF IMPORT SCRIPT
-- ============================================================================
-- Purpose: Convert nginx server blocks from http.conf to NPM database entries
-- Source: /home/priz/projects/devarch/config/nginx/custom/http.conf
-- Target: Nginx Proxy Manager v2.x database (MariaDB/MySQL)
--
-- IMPORTANT NOTES:
-- 1. The wildcard server block (lines 1142-1464) with complex routing logic
--    CANNOT be fully represented in NPM's proxy_host table. It requires
--    custom nginx config and should remain as a static configuration file.
--
-- 2. This script creates proxy hosts for 39 explicit server blocks only.
--    The dynamic .test domain routing remains in custom http.conf.
--
-- 3. SSL Certificate: All entries reference certificate_id=1 (local.crt).
--    Ensure this certificate exists in NPM before importing.
--
-- 4. Owner: All entries set owner_user_id=1 (admin). Adjust if needed.
--
-- ROLLBACK INSTRUCTIONS:
-- To remove all imported proxy hosts:
--   DELETE FROM proxy_host WHERE meta LIKE '%devarch_import%';
--
-- ============================================================================

-- ============================================================================
-- PREREQUISITES: Verify Database Schema
-- ============================================================================

-- Verify proxy_host table exists with expected columns
SELECT 'Checking proxy_host table schema...' as status;

-- Expected columns based on NPM migrations:
-- id, created_on, modified_on, owner_user_id, is_deleted,
-- domain_names (json), forward_host, forward_port, forward_scheme,
-- access_list_id, certificate_id, ssl_forced, caching_enabled,
-- block_exploits, allow_websocket_upgrade, http2_support,
-- hsts_enabled, hsts_subdomains, advanced_config (text), meta (json)

-- ============================================================================
-- SECTION 1: NGINX PROXY MANAGER ADMIN INTERFACE
-- ============================================================================
-- Server blocks: nginx.test, npm.test, nginx-proxy-manager.test
-- Upstream: nginx-proxy-manager:81
-- ============================================================================

INSERT INTO proxy_host (
    created_on,
    modified_on,
    owner_user_id,
    is_deleted,
    domain_names,
    forward_scheme,
    forward_host,
    forward_port,
    access_list_id,
    certificate_id,
    ssl_forced,
    caching_enabled,
    block_exploits,
    allow_websocket_upgrade,
    http2_support,
    hsts_enabled,
    hsts_subdomains,
    advanced_config,
    meta
) VALUES (
    NOW(),
    NOW(),
    1, -- owner_user_id (admin)
    0, -- not deleted
    '["nginx.test","npm.test","nginx-proxy-manager.test"]',
    'https',
    'nginx-proxy-manager',
    81,
    0, -- no access list
    1, -- certificate_id (local.crt - ensure this exists!)
    1, -- SSL forced
    0, -- caching disabled
    1, -- block exploits enabled
    1, -- websocket upgrade allowed (for admin interface)
    1, -- HTTP/2 enabled
    1, -- HSTS enabled
    0, -- HSTS subdomains disabled
    '', -- no advanced config needed
    '{"devarch_import":true,"category":"management","service":"nginx-proxy-manager","source_file":"http.conf","log_prefix":"nginx-admin"}'
);

-- ============================================================================
-- SECTION 2: DATABASE MANAGEMENT TOOLS
-- ============================================================================
-- Services: Adminer, phpMyAdmin, MongoDB Express, Cloudbeaver, Metabase,
--           NocoDB, pgAdmin, Redis Commander, DrawDB
-- ============================================================================

-- Adminer (adminer.test -> adminer:8080)
INSERT INTO proxy_host (
    created_on, modified_on, owner_user_id, is_deleted,
    domain_names, forward_scheme, forward_host, forward_port,
    access_list_id, certificate_id, ssl_forced, caching_enabled,
    block_exploits, allow_websocket_upgrade, http2_support,
    hsts_enabled, hsts_subdomains, advanced_config, meta
) VALUES (
    NOW(), NOW(), 1, 0,
    '["adminer.test"]',
    'https', 'adminer', 8080,
    0, 1, 1, 0, 1, 0, 1, 1, 0,
    'error_page 502 503 504 = @fallback;\nlocation @fallback {\n    return 503 "Adminer service is temporarily unavailable";\n    add_header Content-Type text/plain always;\n}',
    '{"devarch_import":true,"category":"dbms","service":"adminer","source_file":"http.conf","log_prefix":"adminer"}'
);

-- phpMyAdmin (phpmyadmin.test -> phpmyadmin:80)
INSERT INTO proxy_host (
    created_on, modified_on, owner_user_id, is_deleted,
    domain_names, forward_scheme, forward_host, forward_port,
    access_list_id, certificate_id, ssl_forced, caching_enabled,
    block_exploits, allow_websocket_upgrade, http2_support,
    hsts_enabled, hsts_subdomains, advanced_config, meta
) VALUES (
    NOW(), NOW(), 1, 0,
    '["phpmyadmin.test"]',
    'https', 'phpmyadmin', 80,
    0, 1, 1, 0, 1, 0, 1, 1, 0,
    'error_page 502 503 504 = @fallback;\nlocation @fallback {\n    return 503 "phpMyAdmin service is temporarily unavailable";\n    add_header Content-Type text/plain always;\n}',
    '{"devarch_import":true,"category":"dbms","service":"phpmyadmin","source_file":"http.conf","log_prefix":"phpmyadmin"}'
);

-- MongoDB Express (mongodb.test -> mongo-express:8081)
INSERT INTO proxy_host (
    created_on, modified_on, owner_user_id, is_deleted,
    domain_names, forward_scheme, forward_host, forward_port,
    access_list_id, certificate_id, ssl_forced, caching_enabled,
    block_exploits, allow_websocket_upgrade, http2_support,
    hsts_enabled, hsts_subdomains, advanced_config, meta
) VALUES (
    NOW(), NOW(), 1, 0,
    '["mongodb.test"]',
    'https', 'mongo-express', 8081,
    0, 1, 1, 0, 1, 0, 1, 1, 0,
    'error_page 502 503 504 = @fallback;\nlocation @fallback {\n    return 503 "MongoDB Express service is temporarily unavailable";\n    add_header Content-Type text/plain always;\n}',
    '{"devarch_import":true,"category":"dbms","service":"mongodb-express","source_file":"http.conf","log_prefix":"mongodb"}'
);

-- Cloudbeaver (cloudbeaver.test -> cloudbeaver:8978)
INSERT INTO proxy_host (
    created_on, modified_on, owner_user_id, is_deleted,
    domain_names, forward_scheme, forward_host, forward_port,
    access_list_id, certificate_id, ssl_forced, caching_enabled,
    block_exploits, allow_websocket_upgrade, http2_support,
    hsts_enabled, hsts_subdomains, advanced_config, meta
) VALUES (
    NOW(), NOW(), 1, 0,
    '["cloudbeaver.test"]',
    'https', 'cloudbeaver', 8978,
    0, 1, 1, 0, 1, 0, 1, 1, 0,
    'error_page 502 503 504 = @fallback;\nlocation @fallback {\n    return 503 "Cloudbeaver service is temporarily unavailable";\n    add_header Content-Type text/plain always;\n}',
    '{"devarch_import":true,"category":"dbms","service":"cloudbeaver","source_file":"http.conf","log_prefix":"cloudbeaver"}'
);

-- Metabase (metabase.test -> metabase:3000)
INSERT INTO proxy_host (
    created_on, modified_on, owner_user_id, is_deleted,
    domain_names, forward_scheme, forward_host, forward_port,
    access_list_id, certificate_id, ssl_forced, caching_enabled,
    block_exploits, allow_websocket_upgrade, http2_support,
    hsts_enabled, hsts_subdomains, advanced_config, meta
) VALUES (
    NOW(), NOW(), 1, 0,
    '["metabase.test"]',
    'https', 'metabase', 3000,
    0, 1, 1, 0, 1, 0, 1, 1, 0,
    'error_page 502 503 504 = @fallback;\nlocation @fallback {\n    return 503 "Metabase service is temporarily unavailable";\n    add_header Content-Type text/plain always;\n}',
    '{"devarch_import":true,"category":"dbms","service":"metabase","source_file":"http.conf","log_prefix":"metabase"}'
);

-- NocoDB (nocodb.test -> nocodb:8080)
INSERT INTO proxy_host (
    created_on, modified_on, owner_user_id, is_deleted,
    domain_names, forward_scheme, forward_host, forward_port,
    access_list_id, certificate_id, ssl_forced, caching_enabled,
    block_exploits, allow_websocket_upgrade, http2_support,
    hsts_enabled, hsts_subdomains, advanced_config, meta
) VALUES (
    NOW(), NOW(), 1, 0,
    '["nocodb.test"]',
    'https', 'nocodb', 8080,
    0, 1, 1, 0, 1, 0, 1, 1, 0,
    'error_page 502 503 504 = @fallback;\nlocation @fallback {\n    return 503 "NocoDB service is temporarily unavailable";\n    add_header Content-Type text/plain always;\n}',
    '{"devarch_import":true,"category":"dbms","service":"nocodb","source_file":"http.conf","log_prefix":"nocodb"}'
);

-- pgAdmin (pgadmin.test -> pgadmin:80)
INSERT INTO proxy_host (
    created_on, modified_on, owner_user_id, is_deleted,
    domain_names, forward_scheme, forward_host, forward_port,
    access_list_id, certificate_id, ssl_forced, caching_enabled,
    block_exploits, allow_websocket_upgrade, http2_support,
    hsts_enabled, hsts_subdomains, advanced_config, meta
) VALUES (
    NOW(), NOW(), 1, 0,
    '["pgadmin.test"]',
    'https', 'pgadmin', 80,
    0, 1, 1, 0, 1, 0, 1, 1, 0,
    'error_page 502 503 504 = @fallback;\nlocation @fallback {\n    return 503 "pgAdmin service is temporarily unavailable";\n    add_header Content-Type text/plain always;\n}',
    '{"devarch_import":true,"category":"dbms","service":"pgadmin","source_file":"http.conf","log_prefix":"pgadmin"}'
);

-- Redis Commander (redis.test -> redis-commander:8081)
INSERT INTO proxy_host (
    created_on, modified_on, owner_user_id, is_deleted,
    domain_names, forward_scheme, forward_host, forward_port,
    access_list_id, certificate_id, ssl_forced, caching_enabled,
    block_exploits, allow_websocket_upgrade, http2_support,
    hsts_enabled, hsts_subdomains, advanced_config, meta
) VALUES (
    NOW(), NOW(), 1, 0,
    '["redis.test"]',
    'https', 'redis-commander', 8081,
    0, 1, 1, 0, 1, 0, 1, 1, 0,
    'error_page 502 503 504 = @fallback;\nlocation @fallback {\n    return 503 "Redis Commander service is temporarily unavailable";\n    add_header Content-Type text/plain always;\n}',
    '{"devarch_import":true,"category":"dbms","service":"redis-commander","source_file":"http.conf","log_prefix":"redis"}'
);

-- DrawDB (drawdb.test -> drawdb:80)
INSERT INTO proxy_host (
    created_on, modified_on, owner_user_id, is_deleted,
    domain_names, forward_scheme, forward_host, forward_port,
    access_list_id, certificate_id, ssl_forced, caching_enabled,
    block_exploits, allow_websocket_upgrade, http2_support,
    hsts_enabled, hsts_subdomains, advanced_config, meta
) VALUES (
    NOW(), NOW(), 1, 0,
    '["drawdb.test"]',
    'https', 'drawdb', 80,
    0, 1, 1, 0, 1, 0, 1, 1, 0,
    'error_page 502 503 504 = @fallback;\nlocation @fallback {\n    return 503 "DrawDB service is temporarily unavailable";\n    add_header Content-Type text/plain always;\n}',
    '{"devarch_import":true,"category":"dbms","service":"drawdb","source_file":"http.conf","log_prefix":"drawdb"}'
);

-- ============================================================================
-- SECTION 3: BACKEND RUNTIME SERVICES
-- ============================================================================
-- Services: .NET, Go, Node.js, Python
-- Note: These are direct backend proxies, not file serving
-- ============================================================================

-- .NET Runtime (dotnet.test -> dotnet:80)
INSERT INTO proxy_host (
    created_on, modified_on, owner_user_id, is_deleted,
    domain_names, forward_scheme, forward_host, forward_port,
    access_list_id, certificate_id, ssl_forced, caching_enabled,
    block_exploits, allow_websocket_upgrade, http2_support,
    hsts_enabled, hsts_subdomains, advanced_config, meta
) VALUES (
    NOW(), NOW(), 1, 0,
    '["dotnet.test"]',
    'https', 'dotnet', 80,
    0, 1, 1, 0, 1, 0, 1, 1, 0,
    'error_page 502 503 504 = @fallback;\nlocation @fallback {\n    return 503 ".NET service is temporarily unavailable";\n    add_header Content-Type text/plain always;\n}',
    '{"devarch_import":true,"category":"backend","service":"dotnet","source_file":"http.conf","log_prefix":"dotnet"}'
);

-- Go Runtime (go.test -> go:8080)
INSERT INTO proxy_host (
    created_on, modified_on, owner_user_id, is_deleted,
    domain_names, forward_scheme, forward_host, forward_port,
    access_list_id, certificate_id, ssl_forced, caching_enabled,
    block_exploits, allow_websocket_upgrade, http2_support,
    hsts_enabled, hsts_subdomains, advanced_config, meta
) VALUES (
    NOW(), NOW(), 1, 0,
    '["go.test"]',
    'https', 'go', 8080,
    0, 1, 1, 0, 1, 0, 1, 1, 0,
    'error_page 502 503 504 = @fallback;\nlocation @fallback {\n    return 503 "Go service is temporarily unavailable";\n    add_header Content-Type text/plain always;\n}',
    '{"devarch_import":true,"category":"backend","service":"go","source_file":"http.conf","log_prefix":"go"}'
);

-- Node.js Runtime (node.test -> node:3000)
INSERT INTO proxy_host (
    created_on, modified_on, owner_user_id, is_deleted,
    domain_names, forward_scheme, forward_host, forward_port,
    access_list_id, certificate_id, ssl_forced, caching_enabled,
    block_exploits, allow_websocket_upgrade, http2_support,
    hsts_enabled, hsts_subdomains, advanced_config, meta
) VALUES (
    NOW(), NOW(), 1, 0,
    '["node.test"]',
    'https', 'node', 3000,
    0, 1, 1, 0, 1, 0, 1, 1, 0,
    'error_page 502 503 504 = @fallback;\nlocation @fallback {\n    return 503 "Node.js service is temporarily unavailable";\n    add_header Content-Type text/plain always;\n}',
    '{"devarch_import":true,"category":"backend","service":"node","source_file":"http.conf","log_prefix":"node"}'
);

-- Python Runtime (python.test -> python:8000)
INSERT INTO proxy_host (
    created_on, modified_on, owner_user_id, is_deleted,
    domain_names, forward_scheme, forward_host, forward_port,
    access_list_id, certificate_id, ssl_forced, caching_enabled,
    block_exploits, allow_websocket_upgrade, http2_support,
    hsts_enabled, hsts_subdomains, advanced_config, meta
) VALUES (
    NOW(), NOW(), 1, 0,
    '["python.test"]',
    'https', 'python', 8000,
    0, 1, 1, 0, 1, 0, 1, 1, 0,
    'error_page 502 503 504 = @fallback;\nlocation @fallback {\n    return 503 "Python service is temporarily unavailable";\n    add_header Content-Type text/plain always;\n}',
    '{"devarch_import":true,"category":"backend","service":"python","source_file":"http.conf","log_prefix":"python"}'
);

-- ============================================================================
-- SECTION 4: ANALYTICS & MONITORING STACK
-- ============================================================================
-- Services: Elasticsearch, Kibana, Logstash (ELK), Grafana, Prometheus,
--           Matomo, cAdvisor, OpenTelemetry Collector
-- ============================================================================

-- Elasticsearch (elasticsearch.test -> elasticsearch:9200)
INSERT INTO proxy_host (
    created_on, modified_on, owner_user_id, is_deleted,
    domain_names, forward_scheme, forward_host, forward_port,
    access_list_id, certificate_id, ssl_forced, caching_enabled,
    block_exploits, allow_websocket_upgrade, http2_support,
    hsts_enabled, hsts_subdomains, advanced_config, meta
) VALUES (
    NOW(), NOW(), 1, 0,
    '["elasticsearch.test"]',
    'https', 'elasticsearch', 9200,
    0, 1, 1, 0, 1, 0, 1, 1, 0,
    'error_page 502 503 504 = @fallback;\nlocation @fallback {\n    return 503 "Elasticsearch service is temporarily unavailable";\n    add_header Content-Type text/plain always;\n}',
    '{"devarch_import":true,"category":"analytics","service":"elasticsearch","source_file":"http.conf","log_prefix":"elasticsearch"}'
);

-- Kibana (kibana.test -> kibana:5601)
INSERT INTO proxy_host (
    created_on, modified_on, owner_user_id, is_deleted,
    domain_names, forward_scheme, forward_host, forward_port,
    access_list_id, certificate_id, ssl_forced, caching_enabled,
    block_exploits, allow_websocket_upgrade, http2_support,
    hsts_enabled, hsts_subdomains, advanced_config, meta
) VALUES (
    NOW(), NOW(), 1, 0,
    '["kibana.test"]',
    'https', 'kibana', 5601,
    0, 1, 1, 0, 1, 0, 1, 1, 0,
    'error_page 502 503 504 = @fallback;\nlocation @fallback {\n    return 503 "Kibana service is temporarily unavailable";\n    add_header Content-Type text/plain always;\n}',
    '{"devarch_import":true,"category":"analytics","service":"kibana","source_file":"http.conf","log_prefix":"kibana"}'
);

-- Logstash (logstash.test -> logstash:9600)
INSERT INTO proxy_host (
    created_on, modified_on, owner_user_id, is_deleted,
    domain_names, forward_scheme, forward_host, forward_port,
    access_list_id, certificate_id, ssl_forced, caching_enabled,
    block_exploits, allow_websocket_upgrade, http2_support,
    hsts_enabled, hsts_subdomains, advanced_config, meta
) VALUES (
    NOW(), NOW(), 1, 0,
    '["logstash.test"]',
    'https', 'logstash', 9600,
    0, 1, 1, 0, 1, 0, 1, 1, 0,
    'error_page 502 503 504 = @fallback;\nlocation @fallback {\n    return 503 "Logstash service is temporarily unavailable";\n    add_header Content-Type text/plain always;\n}',
    '{"devarch_import":true,"category":"analytics","service":"logstash","source_file":"http.conf","log_prefix":"logstash"}'
);

-- Grafana (grafana.test -> grafana:3000)
INSERT INTO proxy_host (
    created_on, modified_on, owner_user_id, is_deleted,
    domain_names, forward_scheme, forward_host, forward_port,
    access_list_id, certificate_id, ssl_forced, caching_enabled,
    block_exploits, allow_websocket_upgrade, http2_support,
    hsts_enabled, hsts_subdomains, advanced_config, meta
) VALUES (
    NOW(), NOW(), 1, 0,
    '["grafana.test"]',
    'https', 'grafana', 3000,
    0, 1, 1, 0, 1, 0, 1, 1, 0,
    'error_page 502 503 504 = @fallback;\nlocation @fallback {\n    return 503 "Grafana service is temporarily unavailable";\n    add_header Content-Type text/plain always;\n}',
    '{"devarch_import":true,"category":"analytics","service":"grafana","source_file":"http.conf","log_prefix":"grafana"}'
);

-- Prometheus (prometheus.test -> prometheus:9090)
INSERT INTO proxy_host (
    created_on, modified_on, owner_user_id, is_deleted,
    domain_names, forward_scheme, forward_host, forward_port,
    access_list_id, certificate_id, ssl_forced, caching_enabled,
    block_exploits, allow_websocket_upgrade, http2_support,
    hsts_enabled, hsts_subdomains, advanced_config, meta
) VALUES (
    NOW(), NOW(), 1, 0,
    '["prometheus.test"]',
    'https', 'prometheus', 9090,
    0, 1, 1, 0, 1, 0, 1, 1, 0,
    'error_page 502 503 504 = @fallback;\nlocation @fallback {\n    return 503 "Prometheus service is temporarily unavailable";\n    add_header Content-Type text/plain always;\n}',
    '{"devarch_import":true,"category":"analytics","service":"prometheus","source_file":"http.conf","log_prefix":"prometheus"}'
);

-- Matomo (matomo.test -> matomo:80)
INSERT INTO proxy_host (
    created_on, modified_on, owner_user_id, is_deleted,
    domain_names, forward_scheme, forward_host, forward_port,
    access_list_id, certificate_id, ssl_forced, caching_enabled,
    block_exploits, allow_websocket_upgrade, http2_support,
    hsts_enabled, hsts_subdomains, advanced_config, meta
) VALUES (
    NOW(), NOW(), 1, 0,
    '["matomo.test"]',
    'https', 'matomo', 80,
    0, 1, 1, 0, 1, 0, 1, 1, 0,
    'error_page 502 503 504 = @fallback;\nlocation @fallback {\n    return 503 "Matomo service is temporarily unavailable";\n    add_header Content-Type text/plain always;\n}',
    '{"devarch_import":true,"category":"analytics","service":"matomo","source_file":"http.conf","log_prefix":"matomo"}'
);

-- cAdvisor (cadvisor.test -> cadvisor:8080)
INSERT INTO proxy_host (
    created_on, modified_on, owner_user_id, is_deleted,
    domain_names, forward_scheme, forward_host, forward_port,
    access_list_id, certificate_id, ssl_forced, caching_enabled,
    block_exploits, allow_websocket_upgrade, http2_support,
    hsts_enabled, hsts_subdomains, advanced_config, meta
) VALUES (
    NOW(), NOW(), 1, 0,
    '["cadvisor.test"]',
    'https', 'cadvisor', 8080,
    0, 1, 1, 0, 1, 0, 1, 1, 0,
    'error_page 502 503 504 = @fallback;\nlocation @fallback {\n    return 503 "cAdvisor service is temporarily unavailable";\n    add_header Content-Type text/plain always;\n}',
    '{"devarch_import":true,"category":"analytics","service":"cadvisor","source_file":"http.conf","log_prefix":"cadvisor"}'
);

-- OpenTelemetry Collector (otel.test -> otel-collector:8889)
INSERT INTO proxy_host (
    created_on, modified_on, owner_user_id, is_deleted,
    domain_names, forward_scheme, forward_host, forward_port,
    access_list_id, certificate_id, ssl_forced, caching_enabled,
    block_exploits, allow_websocket_upgrade, http2_support,
    hsts_enabled, hsts_subdomains, advanced_config, meta
) VALUES (
    NOW(), NOW(), 1, 0,
    '["otel.test"]',
    'https', 'otel-collector', 8889,
    0, 1, 1, 0, 1, 0, 1, 1, 0,
    'error_page 502 503 504 = @fallback;\nlocation @fallback {\n    return 503 "OpenTelemetry Collector service is temporarily unavailable";\n    add_header Content-Type text/plain always;\n}',
    '{"devarch_import":true,"category":"analytics","service":"otel-collector","source_file":"http.conf","log_prefix":"otel"}'
);

-- ============================================================================
-- SECTION 5: AI / AUTOMATION SERVICES
-- ============================================================================
-- Services: Langflow, n8n
-- ============================================================================

-- Langflow (langflow.test -> langflow:7860)
INSERT INTO proxy_host (
    created_on, modified_on, owner_user_id, is_deleted,
    domain_names, forward_scheme, forward_host, forward_port,
    access_list_id, certificate_id, ssl_forced, caching_enabled,
    block_exploits, allow_websocket_upgrade, http2_support,
    hsts_enabled, hsts_subdomains, advanced_config, meta
) VALUES (
    NOW(), NOW(), 1, 0,
    '["langflow.test"]',
    'https', 'langflow', 7860,
    0, 1, 1, 0, 1, 1, 1, 1, 0,
    'error_page 502 503 504 = @fallback;\nlocation @fallback {\n    return 503 "Langflow service is temporarily unavailable";\n    add_header Content-Type text/plain always;\n}',
    '{"devarch_import":true,"category":"ai","service":"langflow","source_file":"http.conf","log_prefix":"langflow"}'
);

-- n8n (n8n.test -> n8n:5678)
INSERT INTO proxy_host (
    created_on, modified_on, owner_user_id, is_deleted,
    domain_names, forward_scheme, forward_host, forward_port,
    access_list_id, certificate_id, ssl_forced, caching_enabled,
    block_exploits, allow_websocket_upgrade, http2_support,
    hsts_enabled, hsts_subdomains, advanced_config, meta
) VALUES (
    NOW(), NOW(), 1, 0,
    '["n8n.test"]',
    'https', 'n8n', 5678,
    0, 1, 1, 0, 1, 1, 1, 1, 0,
    'error_page 502 503 504 = @fallback;\nlocation @fallback {\n    return 503 "n8n service is temporarily unavailable";\n    add_header Content-Type text/plain always;\n}',
    '{"devarch_import":true,"category":"ai","service":"n8n","source_file":"http.conf","log_prefix":"n8n"}'
);

-- ============================================================================
-- SECTION 6: PROJECT MANAGEMENT
-- ============================================================================
-- Services: Gitea, OpenProject
-- ============================================================================

-- Gitea (gitea.test -> gitea:3000)
INSERT INTO proxy_host (
    created_on, modified_on, owner_user_id, is_deleted,
    domain_names, forward_scheme, forward_host, forward_port,
    access_list_id, certificate_id, ssl_forced, caching_enabled,
    block_exploits, allow_websocket_upgrade, http2_support,
    hsts_enabled, hsts_subdomains, advanced_config, meta
) VALUES (
    NOW(), NOW(), 1, 0,
    '["gitea.test"]',
    'https', 'gitea', 3000,
    0, 1, 1, 0, 1, 0, 1, 1, 0,
    'error_page 502 503 504 = @fallback;\nlocation @fallback {\n    return 503 "Gitea service is temporarily unavailable";\n    add_header Content-Type text/plain always;\n}',
    '{"devarch_import":true,"category":"project","service":"gitea","source_file":"http.conf","log_prefix":"gitea"}'
);

-- OpenProject (openproject.test -> openproject-web:8080)
INSERT INTO proxy_host (
    created_on, modified_on, owner_user_id, is_deleted,
    domain_names, forward_scheme, forward_host, forward_port,
    access_list_id, certificate_id, ssl_forced, caching_enabled,
    block_exploits, allow_websocket_upgrade, http2_support,
    hsts_enabled, hsts_subdomains, advanced_config, meta
) VALUES (
    NOW(), NOW(), 1, 0,
    '["openproject.test"]',
    'https', 'openproject-web', 8080,
    0, 1, 1, 0, 1, 0, 1, 1, 0,
    'error_page 502 503 504 = @fallback;\nlocation @fallback {\n    return 503 "OpenProject service is temporarily unavailable";\n    add_header Content-Type text/plain always;\n}',
    '{"devarch_import":true,"category":"project","service":"openproject","source_file":"http.conf","log_prefix":"openproject"}'
);

-- ============================================================================
-- SECTION 7: MAIL SERVICES
-- ============================================================================
-- Services: Mailpit
-- ============================================================================

-- Mailpit (mailpit.test -> mailpit:8025)
INSERT INTO proxy_host (
    created_on, modified_on, owner_user_id, is_deleted,
    domain_names, forward_scheme, forward_host, forward_port,
    access_list_id, certificate_id, ssl_forced, caching_enabled,
    block_exploits, allow_websocket_upgrade, http2_support,
    hsts_enabled, hsts_subdomains, advanced_config, meta
) VALUES (
    NOW(), NOW(), 1, 0,
    '["mailpit.test"]',
    'https', 'mailpit', 8025,
    0, 1, 1, 0, 1, 1, 1, 1, 0,
    'error_page 502 503 504 = @fallback;\nlocation @fallback {\n    return 503 "Mailpit service is temporarily unavailable";\n    add_header Content-Type text/plain always;\n}',
    '{"devarch_import":true,"category":"mail","service":"mailpit","source_file":"http.conf","log_prefix":"mailpit"}'
);

-- ============================================================================
-- SECTION 8: CONTAINER MANAGEMENT
-- ============================================================================
-- Services: Portainer
-- ============================================================================

-- Portainer (portainer.test -> portainer:9000)
INSERT INTO proxy_host (
    created_on, modified_on, owner_user_id, is_deleted,
    domain_names, forward_scheme, forward_host, forward_port,
    access_list_id, certificate_id, ssl_forced, caching_enabled,
    block_exploits, allow_websocket_upgrade, http2_support,
    hsts_enabled, hsts_subdomains, advanced_config, meta
) VALUES (
    NOW(), NOW(), 1, 0,
    '["portainer.test"]',
    'https', 'portainer', 9000,
    0, 1, 1, 0, 1, 1, 1, 1, 0,
    'error_page 502 503 504 = @fallback;\nlocation @fallback {\n    return 503 "Portainer service is temporarily unavailable";\n    add_header Content-Type text/plain always;\n}',
    '{"devarch_import":true,"category":"management","service":"portainer","source_file":"http.conf","log_prefix":"portainer"}'
);

-- ============================================================================
-- SECTION 9: API GATEWAY SERVICES
-- ============================================================================
-- Services: KrakenD, Kong, Traefik, Envoy
-- ============================================================================

-- KrakenD Designer (krakend.test -> krakend-designer:80)
INSERT INTO proxy_host (
    created_on, modified_on, owner_user_id, is_deleted,
    domain_names, forward_scheme, forward_host, forward_port,
    access_list_id, certificate_id, ssl_forced, caching_enabled,
    block_exploits, allow_websocket_upgrade, http2_support,
    hsts_enabled, hsts_subdomains, advanced_config, meta
) VALUES (
    NOW(), NOW(), 1, 0,
    '["krakend.test"]',
    'https', 'krakend-designer', 80,
    0, 1, 1, 0, 1, 0, 1, 1, 0,
    'error_page 502 503 504 = @fallback;\nlocation @fallback {\n    return 503 "KrakenD service is temporarily unavailable";\n    add_header Content-Type text/plain always;\n}',
    '{"devarch_import":true,"category":"gateway","service":"krakend","source_file":"http.conf","log_prefix":"krakend"}'
);

-- Kong (kong.test -> kong:8001 Admin API)
INSERT INTO proxy_host (
    created_on, modified_on, owner_user_id, is_deleted,
    domain_names, forward_scheme, forward_host, forward_port,
    access_list_id, certificate_id, ssl_forced, caching_enabled,
    block_exploits, allow_websocket_upgrade, http2_support,
    hsts_enabled, hsts_subdomains, advanced_config, meta
) VALUES (
    NOW(), NOW(), 1, 0,
    '["kong.test"]',
    'https', 'kong', 8001,
    0, 1, 1, 0, 1, 0, 1, 1, 0,
    'error_page 502 503 504 = @fallback;\nlocation @fallback {\n    return 503 "Kong service is temporarily unavailable";\n    add_header Content-Type text/plain always;\n}',
    '{"devarch_import":true,"category":"gateway","service":"kong","source_file":"http.conf","log_prefix":"kong"}'
);

-- Traefik (traefik.test -> traefik:8082)
-- Note: Duplicate definition in http.conf (lines 853 and 1087) - using port 8082
INSERT INTO proxy_host (
    created_on, modified_on, owner_user_id, is_deleted,
    domain_names, forward_scheme, forward_host, forward_port,
    access_list_id, certificate_id, ssl_forced, caching_enabled,
    block_exploits, allow_websocket_upgrade, http2_support,
    hsts_enabled, hsts_subdomains, advanced_config, meta
) VALUES (
    NOW(), NOW(), 1, 0,
    '["traefik.test"]',
    'https', 'traefik', 8082,
    0, 1, 1, 0, 1, 0, 1, 1, 0,
    'error_page 502 503 504 = @fallback;\nlocation @fallback {\n    return 503 "Traefik service is temporarily unavailable";\n    add_header Content-Type text/plain always;\n}',
    '{"devarch_import":true,"category":"gateway","service":"traefik","source_file":"http.conf","log_prefix":"traefik","note":"duplicate_resolved"}'
);

-- Envoy (envoy.test -> envoy:9901 Admin UI)
INSERT INTO proxy_host (
    created_on, modified_on, owner_user_id, is_deleted,
    domain_names, forward_scheme, forward_host, forward_port,
    access_list_id, certificate_id, ssl_forced, caching_enabled,
    block_exploits, allow_websocket_upgrade, http2_support,
    hsts_enabled, hsts_subdomains, advanced_config, meta
) VALUES (
    NOW(), NOW(), 1, 0,
    '["envoy.test"]',
    'https', 'envoy', 9901,
    0, 1, 1, 0, 1, 0, 1, 1, 0,
    'error_page 502 503 504 = @fallback;\nlocation @fallback {\n    return 503 "Envoy service is temporarily unavailable";\n    add_header Content-Type text/plain always;\n}',
    '{"devarch_import":true,"category":"gateway","service":"envoy","source_file":"http.conf","log_prefix":"envoy"}'
);

-- ============================================================================
-- SECTION 10: PROMETHEUS EXPORTERS
-- ============================================================================
-- Services: Blackbox, MongoDB Exporter, MySQL Exporter, Node Exporter,
--           PostgreSQL Exporter, Redis Exporter
-- ============================================================================

-- Blackbox Exporter (blackbox.test -> blackbox-exporter:9115)
INSERT INTO proxy_host (
    created_on, modified_on, owner_user_id, is_deleted,
    domain_names, forward_scheme, forward_host, forward_port,
    access_list_id, certificate_id, ssl_forced, caching_enabled,
    block_exploits, allow_websocket_upgrade, http2_support,
    hsts_enabled, hsts_subdomains, advanced_config, meta
) VALUES (
    NOW(), NOW(), 1, 0,
    '["blackbox.test"]',
    'https', 'blackbox-exporter', 9115,
    0, 1, 1, 0, 1, 0, 1, 1, 0,
    'error_page 502 503 504 = @fallback;\nlocation @fallback {\n    return 503 "Blackbox Exporter service is temporarily unavailable";\n    add_header Content-Type text/plain always;\n}',
    '{"devarch_import":true,"category":"exporters","service":"blackbox-exporter","source_file":"http.conf","log_prefix":"blackbox"}'
);

-- MongoDB Exporter (mongodb-exporter.test -> mongodb-exporter:9216)
INSERT INTO proxy_host (
    created_on, modified_on, owner_user_id, is_deleted,
    domain_names, forward_scheme, forward_host, forward_port,
    access_list_id, certificate_id, ssl_forced, caching_enabled,
    block_exploits, allow_websocket_upgrade, http2_support,
    hsts_enabled, hsts_subdomains, advanced_config, meta
) VALUES (
    NOW(), NOW(), 1, 0,
    '["mongodb-exporter.test"]',
    'https', 'mongodb-exporter', 9216,
    0, 1, 1, 0, 1, 0, 1, 1, 0,
    'error_page 502 503 504 = @fallback;\nlocation @fallback {\n    return 503 "MongoDB Exporter service is temporarily unavailable";\n    add_header Content-Type text/plain always;\n}',
    '{"devarch_import":true,"category":"exporters","service":"mongodb-exporter","source_file":"http.conf","log_prefix":"mongodb-exporter"}'
);

-- MySQL Exporter (mysql-exporter.test -> mysqld-exporter:9104)
INSERT INTO proxy_host (
    created_on, modified_on, owner_user_id, is_deleted,
    domain_names, forward_scheme, forward_host, forward_port,
    access_list_id, certificate_id, ssl_forced, caching_enabled,
    block_exploits, allow_websocket_upgrade, http2_support,
    hsts_enabled, hsts_subdomains, advanced_config, meta
) VALUES (
    NOW(), NOW(), 1, 0,
    '["mysql-exporter.test"]',
    'https', 'mysqld-exporter', 9104,
    0, 1, 1, 0, 1, 0, 1, 1, 0,
    'error_page 502 503 504 = @fallback;\nlocation @fallback {\n    return 503 "MySQL Exporter service is temporarily unavailable";\n    add_header Content-Type text/plain always;\n}',
    '{"devarch_import":true,"category":"exporters","service":"mysqld-exporter","source_file":"http.conf","log_prefix":"mysql-exporter"}'
);

-- Node Exporter (node-exporter.test -> node-exporter:9100)
INSERT INTO proxy_host (
    created_on, modified_on, owner_user_id, is_deleted,
    domain_names, forward_scheme, forward_host, forward_port,
    access_list_id, certificate_id, ssl_forced, caching_enabled,
    block_exploits, allow_websocket_upgrade, http2_support,
    hsts_enabled, hsts_subdomains, advanced_config, meta
) VALUES (
    NOW(), NOW(), 1, 0,
    '["node-exporter.test"]',
    'https', 'node-exporter', 9100,
    0, 1, 1, 0, 1, 0, 1, 1, 0,
    'error_page 502 503 504 = @fallback;\nlocation @fallback {\n    return 503 "Node Exporter service is temporarily unavailable";\n    add_header Content-Type text/plain always;\n}',
    '{"devarch_import":true,"category":"exporters","service":"node-exporter","source_file":"http.conf","log_prefix":"node-exporter"}'
);

-- PostgreSQL Exporter (postgres-exporter.test -> postgres-exporter:9187)
INSERT INTO proxy_host (
    created_on, modified_on, owner_user_id, is_deleted,
    domain_names, forward_scheme, forward_host, forward_port,
    access_list_id, certificate_id, ssl_forced, caching_enabled,
    block_exploits, allow_websocket_upgrade, http2_support,
    hsts_enabled, hsts_subdomains, advanced_config, meta
) VALUES (
    NOW(), NOW(), 1, 0,
    '["postgres-exporter.test"]',
    'https', 'postgres-exporter', 9187,
    0, 1, 1, 0, 1, 0, 1, 1, 0,
    'error_page 502 503 504 = @fallback;\nlocation @fallback {\n    return 503 "PostgreSQL Exporter service is temporarily unavailable";\n    add_header Content-Type text/plain always;\n}',
    '{"devarch_import":true,"category":"exporters","service":"postgres-exporter","source_file":"http.conf","log_prefix":"postgres-exporter"}'
);

-- Redis Exporter (redis-exporter.test -> redis-exporter:9121)
INSERT INTO proxy_host (
    created_on, modified_on, owner_user_id, is_deleted,
    domain_names, forward_scheme, forward_host, forward_port,
    access_list_id, certificate_id, ssl_forced, caching_enabled,
    block_exploits, allow_websocket_upgrade, http2_support,
    hsts_enabled, hsts_subdomains, advanced_config, meta
) VALUES (
    NOW(), NOW(), 1, 0,
    '["redis-exporter.test"]',
    'https', 'redis-exporter', 9121,
    0, 1, 1, 0, 1, 0, 1, 1, 0,
    'error_page 502 503 504 = @fallback;\nlocation @fallback {\n    return 503 "Redis Exporter service is temporarily unavailable";\n    add_header Content-Type text/plain always;\n}',
    '{"devarch_import":true,"category":"exporters","service":"redis-exporter","source_file":"http.conf","log_prefix":"redis-exporter"}'
);

-- ============================================================================
-- IMPORT COMPLETE
-- ============================================================================

SELECT '============================================================' as '';
SELECT 'Import Summary' as '';
SELECT '============================================================' as '';
SELECT COUNT(*) as total_imported FROM proxy_host WHERE meta LIKE '%devarch_import%';

-- ============================================================================
-- VERIFICATION QUERIES
-- ============================================================================

-- Show all imported proxy hosts grouped by category
SELECT
    JSON_EXTRACT(meta, '$.category') as category,
    COUNT(*) as count
FROM proxy_host
WHERE meta LIKE '%devarch_import%'
GROUP BY JSON_EXTRACT(meta, '$.category')
ORDER BY category;

-- List all imported domains
SELECT
    id,
    domain_names,
    CONCAT(forward_scheme, '://', forward_host, ':', forward_port) as upstream,
    JSON_EXTRACT(meta, '$.category') as category,
    JSON_EXTRACT(meta, '$.service') as service
FROM proxy_host
WHERE meta LIKE '%devarch_import%'
ORDER BY domain_names;

-- Check for any missing certificate references
SELECT
    id,
    domain_names,
    certificate_id
FROM proxy_host
WHERE meta LIKE '%devarch_import%'
  AND certificate_id NOT IN (SELECT id FROM certificate);

-- Summary by SSL configuration
SELECT
    ssl_forced,
    http2_support,
    hsts_enabled,
    COUNT(*) as count
FROM proxy_host
WHERE meta LIKE '%devarch_import%'
GROUP BY ssl_forced, http2_support, hsts_enabled;

-- ============================================================================
-- POST-IMPORT NOTES
-- ============================================================================
--
-- TOTAL SERVER BLOCKS: 40 in http.conf
-- IMPORTED: 39 explicit proxy hosts
-- NOT IMPORTED: 1 wildcard server block (lines 1142-1464)
--
-- The wildcard server block (~^(?<appname>[^.]+)\.test$) handles:
-- - Dynamic .test domain routing
-- - Smart document root detection (dist/, build/, out/, public/)
-- - Language runtime detection (PHP, Node, Python, Go, .NET, Rust)
-- - WebSocket support for development frameworks
-- - FastCGI PHP processing with Podman environment variables
--
-- This complex routing logic CANNOT be represented in NPM's proxy_host table.
-- RECOMMENDATION: Keep the wildcard block in custom http.conf file.
--
-- TO APPLY THIS IMPORT:
-- 1. Ensure local.crt certificate exists in NPM (certificate_id=1)
-- 2. Backup NPM database: mysqldump -u root -p npm > npm_backup.sql
-- 3. Run this script: mysql -u root -p npm < npm-import.sql
-- 4. Verify in NPM web interface: https://npm.test
-- 5. Test a few proxy hosts to ensure routing works
-- 6. Keep http.conf for wildcard routing (or create NPM custom locations)
--
-- ROLLBACK IF NEEDED:
-- DELETE FROM proxy_host WHERE meta LIKE '%devarch_import%';
--
-- CERTIFICATE SETUP:
-- If certificate_id=1 doesn't exist, either:
-- 1. Upload local.crt/local.key through NPM interface, then update IDs
-- 2. OR modify all certificate_id values to match existing cert
--
-- ============================================================================
