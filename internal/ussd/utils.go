package ussd

import (
	"strings"
)

func firstNonEmpty(data map[string]string, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(data[key]); value != "" {
			return value
		}
	}
	return ""
}

func cloneMap(src map[string]string) map[string]string {
	if src == nil {
		return map[string]string{}
	}
	cloned := make(map[string]string, len(src))
	for key, value := range src {
		cloned[key] = value
	}
	return cloned
}

func mergeValues(dst map[string]string, src map[string]string) {
	for key, value := range src {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			dst[key] = trimmed
		}
	}
}

func maskSensitive(value string) string {
	value = strings.TrimSpace(value)
	if len(value) <= 2 {
		return strings.Repeat("*", len(value))
	}
	return value[:1] + strings.Repeat("*", len(value)-2) + value[len(value)-1:]
}

func maskVendorCode(code string) string {
	code = strings.TrimSpace(code)
	if len(code) <= 4 {
		return code
	}
	return strings.Repeat("*", len(code)-4) + code[len(code)-4:]
}
