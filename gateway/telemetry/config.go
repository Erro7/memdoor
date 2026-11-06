package telemetry

import (
	"os"
	"strings"

	"memdoor/gateway/logs"
)

// Config is the operator-facing configuration of the telemetry stack.
// Built from env vars at gateway boot (LoadConfig) and consumed by
// the wiring helper Install. All fields have safe defaults so an
// install with telemetry OFF needs zero environment setup.
type Config struct {
	// Enabled gates the entire stack. When false, the gateway uses
	// the bare logs.Storage with no wrapper and no flusher goroutine
	// — zero per-event cost.
	Enabled bool

	// Endpoint is the URL the flusher POSTs to. Defaults to the local
	// gateway; a deployed monitoring receiver overrides it via env.
	Endpoint string

	// BearerToken authenticates the POST request to deter random
	// spam against a public endpoint. Empty is allowed (anonymous
	// posting) but recommended only for self-hosted setups.
	BearerToken string

	// MinLevel filters events before they enter the ring. Default
	// LevelWarn keeps the noise floor low; raising to LevelError
	// gives only failures, lowering to LevelInfo gives operational
	// events too. DEBUG is never appropriate for continuous shipping.
	MinLevel logs.Level

	// RingCapacity bounds the in-memory pending buffer. Default 200
	// covers ~24h of typical WARN+ERROR rates with margin.
	RingCapacity int
}

// EndpointDefault is the local monitoring URL. Surfaced as a
// package-level constant so tests can compare against the same
// value without duplicating the literal.
const EndpointDefault = "http://localhost:18789/api/telemetry"

// LoadConfig reads the MEMDOOR_TELEMETRY_* env vars and returns the
// resolved Config. Unset vars use the defaults documented above.
//
// Env variable contract:
//
//	MEMDOOR_TELEMETRY_ENABLED       "1" / "true" → enabled; anything else → off
//	MEMDOOR_TELEMETRY_URL           receiver URL; defaults to the local gateway
//	MEMDOOR_TELEMETRY_TOKEN         bearer token; empty allowed
func LoadConfig() Config {
	cfg := Config{
		Enabled:      envBool("MEMDOOR_TELEMETRY_ENABLED"),
		Endpoint:     envOrDefault("MEMDOOR_TELEMETRY_URL", EndpointDefault),
		BearerToken:  os.Getenv("MEMDOOR_TELEMETRY_TOKEN"),
		MinLevel:     logs.LevelWarn,
		RingCapacity: 200,
	}
	return cfg
}

func envBool(key string) bool {
	switch os.Getenv(key) {
	case "1", "true", "TRUE", "True":
		return true
	default:
		return false
	}
}

func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return strings.TrimRight(value, "/")
	}
	return fallback
}
