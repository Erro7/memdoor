package telemetry

import "testing"

func TestLoadConfigUsesLocalTelemetryEndpointByDefault(t *testing.T) {
	t.Setenv("MEMDOOR_TELEMETRY_URL", "")

	if got := LoadConfig().Endpoint; got != EndpointDefault {
		t.Fatalf("LoadConfig().Endpoint = %q, want %q", got, EndpointDefault)
	}
}

func TestLoadConfigUsesTelemetryEndpointFromEnvironment(t *testing.T) {
	t.Setenv("MEMDOOR_TELEMETRY_URL", " https://monitor.example.test/events/ ")
	t.Setenv("MEMDOOR_REMOTE_URL", "http://localhost:48789/relay")

	if got, want := LoadConfig().Endpoint, "https://monitor.example.test/events"; got != want {
		t.Fatalf("LoadConfig().Endpoint = %q, want %q", got, want)
	}
}

func TestLoadConfigDoesNotUseRemoteURL(t *testing.T) {
	t.Setenv("MEMDOOR_TELEMETRY_URL", "")
	t.Setenv("MEMDOOR_REMOTE_URL", "http://localhost:48789/relay")

	if got := LoadConfig().Endpoint; got != EndpointDefault {
		t.Fatalf("LoadConfig().Endpoint = %q, want %q; MEMDOOR_REMOTE_URL is relay-only", got, EndpointDefault)
	}
}
