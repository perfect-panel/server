package tool

import "testing"

func TestNormalizeCertFingerprintSha256(t *testing.T) {
	input := "SHA256=AA:BB:CC:DD:EE:FF:00:11:22:33:44:55:66:77:88:99:AA:BB:CC:DD:EE:FF:00:11:22:33:44:55:66:77:88:99"
	want := "aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899"

	if got := NormalizeCertFingerprintSha256(input); got != want {
		t.Fatalf("NormalizeCertFingerprintSha256() = %q, want %q", got, want)
	}
	if got := NormalizeCertFingerprintSha256("invalid"); got != "" {
		t.Fatalf("NormalizeCertFingerprintSha256(invalid) = %q, want empty", got)
	}
}
