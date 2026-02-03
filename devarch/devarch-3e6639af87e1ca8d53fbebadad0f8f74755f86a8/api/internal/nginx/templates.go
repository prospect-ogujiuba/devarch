package nginx

const projectTemplate = `server {
    listen 80;
    http2 on;
    listen 443 ssl;
    server_name {{.Domain}};

    ssl_certificate /etc/ssl/certs/local.crt;
    ssl_certificate_key /etc/ssl/private/local.key;

    if ($scheme != "https" ) {
        return 301 https://$host$request_uri;
    }
{{if .ClientMaxBody}}
    client_max_body_size {{.ClientMaxBody}};
{{end}}
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_set_header X-Forwarded-Host $host;
    proxy_set_header X-Forwarded-Port $server_port;
{{if .WebSocketUpstream}}
    location /app {
        set $upstream_ws {{.WebSocketUpstream}};
        proxy_pass http://$upstream_ws;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_read_timeout 86400;
        proxy_send_timeout 86400;
        error_page 502 503 504 = @fallback;
    }
{{end}}
    location / {
        set $upstream {{.Upstream}};
        proxy_pass http://$upstream;
        include /data/nginx/custom/server_proxy.conf;
        error_page 502 503 504 = @fallback;
    }

    location @fallback {
        return 503 "{{.Name}} is temporarily unavailable";
        add_header Content-Type text/plain always;
    }

    error_log /var/log/nginx/{{.LogName}}.error.log;
    access_log /var/log/nginx/{{.LogName}}.access.log;
}
`

const serviceTemplate = `server {
    listen 80;
    http2 on;
    listen 443 ssl;
    server_name {{.Domain}};

    ssl_certificate /etc/ssl/certs/local.crt;
    ssl_certificate_key /etc/ssl/private/local.key;

    if ($scheme != "https" ) {
        return 301 https://$host$request_uri;
    }

    location / {
        set $upstream_{{.VarName}} {{.Upstream}};
        proxy_pass http://$upstream_{{.VarName}};
        include /data/nginx/custom/server_proxy.conf;
        error_page 502 503 504 = @fallback_{{.VarName}};
    }

    location @fallback_{{.VarName}} {
        return 503 "{{.DisplayName}} service is temporarily unavailable";
        add_header Content-Type text/plain always;
    }

    error_log /var/log/nginx/{{.LogName}}.error.log;
    access_log /var/log/nginx/{{.LogName}}.access.log;
}
`
