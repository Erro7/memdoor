package gateway

import (
	"os"
	"strings"

	"memdoor/pkg/shared"
)

// brokerBaseURL is where memdoor.ai's billing service answers: the relay asks
// it whose account a token is, and on which plan (remote_relay_auth.go). The
// gateway calls nothing else there any more — no lease, no seat model, no
// seat decisions (2026-10-04: a seat never supplies a key). Overridable only
// for memdoor.ai or this machine.
func brokerBaseURL() string {
	for _, key := range []string{"MEMDOOR_ACCOUNT_URL", "MEMDOOR_BILLING_URL"} {
		if base := strings.TrimSpace(os.Getenv(key)); base != "" && shared.AccountHostTrusted(base) {
			return strings.TrimRight(base, "/")
		}
	}
	return "http://localhost:18790"
}
