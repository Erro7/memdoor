package cmd

import "testing"

func TestBillingBaseUsesLocalAccountServiceByDefault(t *testing.T) {
	t.Setenv("MEMDOOR_ACCOUNT_URL", "")
	t.Setenv("MEMDOOR_BILLING_URL", "")

	if got := billingBase(); got != defaultBillingBase {
		t.Fatalf("billingBase() = %q, want %q", got, defaultBillingBase)
	}
}

func TestBillingBaseUsesAccountServiceFromEnvironment(t *testing.T) {
	t.Setenv("MEMDOOR_ACCOUNT_URL", "http://localhost:28790/account/")
	t.Setenv("MEMDOOR_BILLING_URL", "http://localhost:38790/legacy")
	t.Setenv("MEMDOOR_REMOTE_URL", "http://localhost:48789/relay")

	if got, want := billingBase(), "http://localhost:28790/account"; got != want {
		t.Fatalf("billingBase() = %q, want %q", got, want)
	}
}

func TestBillingBaseDoesNotUseRemoteURL(t *testing.T) {
	t.Setenv("MEMDOOR_ACCOUNT_URL", "")
	t.Setenv("MEMDOOR_BILLING_URL", "")
	t.Setenv("MEMDOOR_REMOTE_URL", "http://localhost:48789/relay")

	if got := billingBase(); got != defaultBillingBase {
		t.Fatalf("billingBase() = %q, want %q; MEMDOOR_REMOTE_URL is relay-only", got, defaultBillingBase)
	}
}
