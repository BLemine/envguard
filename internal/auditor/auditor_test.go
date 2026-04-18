package auditor

import "testing"

func TestRedactSecretLineMasksValue(t *testing.T) {
	t.Parallel()

	got := redactSecretLine(`API_KEY=supersecretapikeyvalue1234567890`)

	if got == `API_KEY=supersecretapikeyvalue1234567890` {
		t.Fatalf("secret value was not redacted: %q", got)
	}
	if got != `API_KEY=[REDACTED]` {
		t.Fatalf("unexpected redacted line: %q", got)
	}
}

func TestRedactSecretLineHandlesQuotedSecret(t *testing.T) {
	t.Parallel()

	got := redactSecretLine(`token = "abc123456789"`)

	if got != `token = [REDACTED]` {
		t.Fatalf("unexpected redacted line: %q", got)
	}
}
