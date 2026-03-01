package provider

import (
	"os"
	"strings"
)

func resolveAPIKey(providerName string, configured string) string {
	if key := strings.TrimSpace(configured); key != "" {
		return key
	}
	name := strings.ToUpper(strings.TrimSpace(providerName))
	candidates := []string{
		name + "_API_KEY",
		"RAGODESK_" + name + "_API_KEY",
		"RAGODESK_API_KEY",
	}
	for _, envKey := range candidates {
		if key := strings.TrimSpace(os.Getenv(envKey)); key != "" {
			return key
		}
	}
	return ""
}
