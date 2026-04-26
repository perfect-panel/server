package tool

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestGenerateDeviceToken_Length(t *testing.T) {
	// 32 bytes → base64url no-padding = 43 chars
	tok := GenerateDeviceToken()
	if len(tok) != 43 {
		t.Errorf("expected 43-char token, got %d (%q)", len(tok), tok)
	}
	if strings.ContainsAny(tok, "+/=") {
		t.Errorf("token must be url-safe (no '+/='): %q", tok)
	}
}

func TestGenerateDeviceToken_Random(t *testing.T) {
	seen := make(map[string]struct{}, 1000)
	for i := 0; i < 1000; i++ {
		tok := GenerateDeviceToken()
		if _, dup := seen[tok]; dup {
			t.Fatalf("token collision at iter %d: %q", i, tok)
		}
		seen[tok] = struct{}{}
	}
}

func TestGenerateUUIDv4_ValidAndRandom(t *testing.T) {
	a := GenerateUUIDv4()
	b := GenerateUUIDv4()
	if a == b {
		t.Errorf("two consecutive UUIDs match: %q", a)
	}
	if _, err := uuid.Parse(a); err != nil {
		t.Errorf("UUID %q not parseable: %v", a, err)
	}
}

func TestDerivePasswordFromUUID_DeterministicAndHexLength(t *testing.T) {
	id := "00000000-0000-0000-0000-000000000000"
	a := DerivePasswordFromUUID(id)
	b := DerivePasswordFromUUID(id)
	if a != b {
		t.Errorf("derive must be deterministic; got %q vs %q", a, b)
	}
	if len(a) != 16 {
		t.Errorf("expected 16-hex-char password, got %d (%q)", len(a), a)
	}
	for _, ch := range a {
		if !(ch >= '0' && ch <= '9' || ch >= 'a' && ch <= 'f') {
			t.Errorf("non-hex char in password: %q", a)
			break
		}
	}
}

func TestDerivePasswordFromUUID_ChangesWithInput(t *testing.T) {
	a := DerivePasswordFromUUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	b := DerivePasswordFromUUID("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	if a == b {
		t.Errorf("different UUIDs produced same password: %q", a)
	}
}
