package container

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// DNS-safe name pattern: lowercase alphanumeric with hyphens, 1-63 chars
	dnsNamePattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)

	// Reserved names that cannot be used for stacks or instances
	reservedNames = map[string]bool{
		"default": true,
		"devarch": true,
		"system":  true,
		"none":    true,
		"all":     true,
	}
)

// ValidateName validates a stack or instance name
// Returns nil if valid, or a prescriptive error describing the issue and suggested fix
func ValidateName(name string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	if len(name) > 63 {
		return fmt.Errorf("%q is too long: must be 63 characters or less", name)
	}

	if reservedNames[strings.ToLower(name)] {
		return fmt.Errorf("%q is a reserved name and cannot be used", name)
	}

	if !dnsNamePattern.MatchString(name) {
		suggestion := Slugify(name)
		return fmt.Errorf("%q is not a valid name: must be lowercase alphanumeric with hyphens — try: %s", name, suggestion)
	}

	return nil
}

// Slugify converts a string to a valid DNS-safe name
func Slugify(input string) string {
	// Lowercase
	s := strings.ToLower(input)

	// Replace spaces and underscores with hyphens
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")

	// Strip invalid characters (keep only a-z, 0-9, -)
	var builder strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			builder.WriteRune(r)
		}
	}
	s = builder.String()

	// Collapse multiple hyphens
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}

	// Trim leading/trailing hyphens and spaces
	s = strings.Trim(s, "- ")

	// Truncate to 63 chars
	if len(s) > 63 {
		s = s[:63]
		s = strings.TrimRight(s, "-")
	}

	// If empty after all that, return a default
	if s == "" {
		return "unnamed"
	}

	return s
}
