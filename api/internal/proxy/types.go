package proxy

// ProxyType identifies a supported reverse proxy.
type ProxyType string

const (
	Nginx   ProxyType = "nginx"
	Caddy   ProxyType = "caddy"
	Traefik ProxyType = "traefik"
	HAProxy ProxyType = "haproxy"
	Apache  ProxyType = "apache"
)

// SupportedTypes returns every proxy type the generator can produce config for.
func SupportedTypes() []ProxyType {
	return []ProxyType{Nginx, Caddy, Traefik, HAProxy, Apache}
}

// ProxyTarget is one routable service with at least one domain.
type ProxyTarget struct {
	Name          string `json:"name"`
	Domain        string `json:"domain"`
	TargetHost    string `json:"target_host"`
	TargetPort    int    `json:"target_port"`
	HTTPS         bool   `json:"https"`
	WebSocket     bool   `json:"websocket"`
	ClientMaxBody string `json:"client_max_body,omitempty"` // e.g. "100M"
}

// ProxyConfigResult is the API response for a generated config.
type ProxyConfigResult struct {
	ProxyType ProxyType     `json:"proxy_type"`
	Scope     string        `json:"scope"`  // "service", "stack", "project"
	Name      string        `json:"name"`   // name of the service/stack/project
	Config    string        `json:"config"` // generated configuration text
	Targets   []ProxyTarget `json:"targets"`
}
