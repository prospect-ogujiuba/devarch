package catalog

import (
	"fmt"
	"strconv"
	"strings"
)

// Port is a published compose port. Published is 0 when the port is not
// published to the host or uses unresolved interpolation.
type Port struct {
	Raw       string `json:"raw"`
	HostIP    string `json:"host_ip,omitempty"`
	Published int    `json:"published,omitempty"`
	Target    string `json:"target,omitempty"`
	Protocol  string `json:"protocol,omitempty"`
}

// ParsePort reads the short ("[ip:]host:container[/proto]") and long
// (mapping) compose port syntaxes.
func ParsePort(v any) Port {
	switch p := v.(type) {
	case map[string]any:
		port := Port{Protocol: "tcp"}
		port.HostIP, _ = p["host_ip"].(string)
		port.Target = fmt.Sprint(p["target"])
		if proto, ok := p["protocol"].(string); ok && proto != "" {
			port.Protocol = proto
		}
		switch pub := p["published"].(type) {
		case int:
			port.Published = pub
		case string:
			port.Published, _ = strconv.Atoi(pub)
		}
		port.Raw = fmt.Sprintf("%s:%d:%s/%s", port.HostIP, port.Published, port.Target, port.Protocol)
		return port
	default:
		raw := fmt.Sprint(v)
		port := Port{Raw: raw, Protocol: "tcp"}
		spec := raw
		if i := strings.LastIndexByte(spec, '/'); i >= 0 {
			port.Protocol, spec = spec[i+1:], spec[:i]
		}
		if strings.Contains(spec, "${") {
			// Unresolved interpolation: only the container side is known.
			port.Target = spec[strings.LastIndexByte(spec, ':')+1:]
			return port
		}
		parts := strings.Split(spec, ":")
		switch len(parts) {
		case 1:
			port.Target = parts[0]
			return port
		case 2:
			port.Target = parts[1]
		default:
			port.HostIP = strings.Join(parts[:len(parts)-2], ":")
			port.Target = parts[len(parts)-1]
		}
		host := parts[len(parts)-2]
		if i := strings.IndexByte(host, '-'); i >= 0 {
			host = host[:i] // ranges: the first port is enough for conflict checks
		}
		port.Published, _ = strconv.Atoi(host)
		return port
	}
}
