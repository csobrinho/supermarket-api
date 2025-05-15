package supermarket

import (
	"os"
	"strconv"
	"strings"

	"github.com/google/logger"
)

func LookupEnv(key string, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}
func LookupEnvInt(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if vi, err := strconv.Atoi(v); err == nil {
			return vi
		}
		logger.Warningf("supermarket: %q is not a valid int (%q), using default value %d", key, v, def)
	}
	return def
}
func LookupEnvBool(key string, def bool) bool {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		switch strings.ToLower(v) {
		case "true", "1", "yes", "on":
			return true
		case "false", "0", "no", "off":
			return false
		default:
			logger.Warningf("supermarket: %q is not a valid bool (%q), using default value %t", key, v, def)
		}
	}
	return def
}
