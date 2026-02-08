package export

import (
	"fmt"
	"strings"
)

var secretKeywords = []string{
	"password", "secret", "key", "token", "api_key", "apikey",
	"auth", "private", "credential", "passwd",
}

func IsSecretKey(key string) bool {
	lower := strings.ToLower(key)
	for _, keyword := range secretKeywords {
		if strings.Contains(lower, keyword) {
			return true
		}
	}
	return false
}

func RedactSecrets(env map[string]string) map[string]string {
	result := make(map[string]string, len(env))
	for k, v := range env {
		if IsSecretKey(k) {
			result[k] = fmt.Sprintf("${SECRET:%s}", k)
		} else {
			result[k] = v
		}
	}
	return result
}
