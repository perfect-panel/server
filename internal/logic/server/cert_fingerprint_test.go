package server

import (
	"testing"

	"github.com/perfect-panel/server/internal/model/node"
	"github.com/perfect-panel/server/pkg/tool"
)

func TestNormalizeCertFingerprintSha256(t *testing.T) {
	input := "SHA256:AA:BB:CC:DD:EE:FF:00:11:22:33:44:55:66:77:88:99:AA:BB:CC:DD:EE:FF:00:11:22:33:44:55:66:77:88:99"
	want := "aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899"

	if got := tool.NormalizeCertFingerprintSha256(input); got != want {
		t.Fatalf("NormalizeCertFingerprintSha256() = %q, want %q", got, want)
	}
	if got := tool.NormalizeCertFingerprintSha256("not-a-sha256"); got != "" {
		t.Fatalf("NormalizeCertFingerprintSha256(invalid) = %q, want empty", got)
	}
}

func TestUpdateReportedCertFingerprintSha256(t *testing.T) {
	protocols := []node.Protocol{
		{Type: "vless"},
		{Type: "trojan"},
	}
	fingerprint := "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff"

	if !updateReportedCertFingerprintSha256(protocols, "trojan", fingerprint) {
		t.Fatal("updateReportedCertFingerprintSha256() = false, want true")
	}
	if protocols[1].ReportedCertFingerprintSha256 != fingerprint {
		t.Fatalf("reported fingerprint = %q, want %q", protocols[1].ReportedCertFingerprintSha256, fingerprint)
	}
	if updateReportedCertFingerprintSha256(protocols, "trojan", fingerprint) {
		t.Fatal("second updateReportedCertFingerprintSha256() = true, want false")
	}
	if updateReportedCertFingerprintSha256(protocols, "vmess", fingerprint) {
		t.Fatal("missing protocol updateReportedCertFingerprintSha256() = true, want false")
	}
}
