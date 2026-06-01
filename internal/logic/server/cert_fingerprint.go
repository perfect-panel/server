package server

import (
	"github.com/perfect-panel/server/internal/model/node"
	"github.com/perfect-panel/server/pkg/tool"
)

func updateReportedCertFingerprintSha256(protocols []node.Protocol, protocolType, fingerprint string) bool {
	fingerprint = tool.NormalizeCertFingerprintSha256(fingerprint)
	if protocolType == "" || fingerprint == "" {
		return false
	}

	for i := range protocols {
		if protocols[i].Type != protocolType {
			continue
		}
		if protocols[i].ReportedCertFingerprintSha256 == fingerprint {
			return false
		}
		protocols[i].ReportedCertFingerprintSha256 = fingerprint
		return true
	}
	return false
}

func effectiveCertFingerprintSha256(protocol node.Protocol) string {
	if protocol.CertFingerprintSha256 != "" {
		return tool.NormalizeCertFingerprintSha256(protocol.CertFingerprintSha256)
	}
	return tool.NormalizeCertFingerprintSha256(protocol.ReportedCertFingerprintSha256)
}
