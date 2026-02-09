package proxy

var templates = map[ProxyType]string{
	Nginx:   nginxTemplate,
	Caddy:   caddyTemplate,
	Traefik: traefikTemplate,
	HAProxy: haproxyTemplate,
	Apache:  apacheTemplate,
}

// ── Nginx ──────────────────────────────────────────────────────────────────

const nginxTemplate = `{{- range . }}
upstream {{ .Name }}_backend {
    server {{ .TargetHost }}:{{ .TargetPort }};
}

server {
    listen 80;
    server_name {{ .Domain }};
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name {{ .Domain }};

    ssl_certificate     /etc/ssl/certs/{{ .Domain }}.crt;
    ssl_certificate_key /etc/ssl/private/{{ .Domain }}.key;
{{ if .ClientMaxBody }}
    client_max_body_size {{ .ClientMaxBody }};
{{ end }}
    proxy_set_header Host              $host;
    proxy_set_header X-Real-IP         $remote_addr;
    proxy_set_header X-Forwarded-For   $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
{{ if .WebSocket }}
    location /ws {
        proxy_pass http://{{ .Name }}_backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade    $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_read_timeout  86400;
        proxy_send_timeout  86400;
    }
{{ end }}
    location / {
        proxy_pass http://{{ .Name }}_backend;
    }
}
{{ end }}`

// ── Caddy ──────────────────────────────────────────────────────────────────

const caddyTemplate = `{{- range . }}
{{ .Domain }} {
    reverse_proxy {{ .TargetHost }}:{{ .TargetPort }}{{ if .WebSocket }} {
        transport http {
            versions h2c 2
        }
    }{{ end }}
{{ if .ClientMaxBody }}    request_body {
        max_size {{ .ClientMaxBody }}
    }
{{ end }}    header {
        X-Forwarded-Proto {scheme}
        -Server
    }

    encode gzip zstd

    log {
        output file /var/log/caddy/{{ .Name }}.log
    }
}
{{ end }}`

// ── Traefik (dynamic YAML) ────────────────────────────────────────────────

const traefikTemplate = `http:
  routers:
{{- range . }}
    {{ .Name }}-router:
      rule: "Host(` + "`" + `{{ .Domain }}` + "`" + `)"
      entryPoints:
        - websecure
      service: {{ .Name }}-service
      tls:
        certResolver: letsencrypt
{{- end }}

  services:
{{- range . }}
    {{ .Name }}-service:
      loadBalancer:
        servers:
          - url: "http://{{ .TargetHost }}:{{ .TargetPort }}"
{{- end }}
`

// ── HAProxy ────────────────────────────────────────────────────────────────

const haproxyTemplate = `global
    log stdout format raw local0
    maxconn 4096

defaults
    log     global
    mode    http
    option  httplog
    option  dontlognull
    timeout connect 5s
    timeout client  30s
    timeout server  30s

frontend http-in
    bind *:80
    redirect scheme https code 301 if !{ ssl_fc }

frontend https-in
    bind *:443 ssl crt /etc/haproxy/certs/
    http-request set-header X-Forwarded-Proto https
{{- range . }}
    acl host_{{ .Name }} hdr(host) -i {{ .Domain }}
    use_backend {{ .Name }}_backend if host_{{ .Name }}
{{- end }}
{{ range . }}
backend {{ .Name }}_backend
    server {{ .Name }}1 {{ .TargetHost }}:{{ .TargetPort }} check
{{ end }}`

// ── Apache ─────────────────────────────────────────────────────────────────

const apacheTemplate = `{{- range . }}
<VirtualHost *:80>
    ServerName {{ .Domain }}
    Redirect permanent / https://{{ .Domain }}/
</VirtualHost>

<VirtualHost *:443>
    ServerName {{ .Domain }}

    SSLEngine on
    SSLCertificateFile    /etc/ssl/certs/{{ .Domain }}.crt
    SSLCertificateKeyFile /etc/ssl/private/{{ .Domain }}.key

    ProxyPreserveHost On
    ProxyRequests Off
{{ if .WebSocket }}
    RewriteEngine On
    RewriteCond %{HTTP:Upgrade} websocket [NC]
    RewriteCond %{HTTP:Connection} upgrade [NC]
    RewriteRule ^/ws(.*)$ ws://{{ .TargetHost }}:{{ .TargetPort }}/ws$1 [P,L]
{{ end }}
    ProxyPass        / http://{{ .TargetHost }}:{{ .TargetPort }}/
    ProxyPassReverse / http://{{ .TargetHost }}:{{ .TargetPort }}/
{{ if .ClientMaxBody }}
    LimitRequestBody {{ .ClientMaxBody }}
{{ end }}
    RequestHeader set X-Forwarded-Proto "https"

    ErrorLog  /var/log/apache2/{{ .Name }}-error.log
    CustomLog /var/log/apache2/{{ .Name }}-access.log combined
</VirtualHost>
{{ end }}`
