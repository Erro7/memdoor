package gateway

import "testing"

func TestBrokerBaseURLUsesLocalAccountServiceByDefault(t *testing.T) {
	t.Setenv("MEMDOOR_ACCOUNT_URL", "")
	t.Setenv("MEMDOOR_BILLING_URL", "")

	if got, want := brokerBaseURL(), "http://localhost:18790"; got != want {
		t.Fatalf("brokerBaseURL() = %q, want %q", got, want)
	}
}

func TestBrokerBaseURLUsesAccountServiceFromEnvironment(t *testing.T) {
	t.Setenv("MEMDOOR_ACCOUNT_URL", "http://localhost:28790/account/")
	t.Setenv("MEMDOOR_BILLING_URL", "http://localhost:38790/legacy")

	if got, want := brokerBaseURL(), "http://localhost:28790/account"; got != want {
		t.Fatalf("brokerBaseURL() = %q, want %q", got, want)
	}
}
