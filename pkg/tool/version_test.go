package tool

import (
	"testing"

	"github.com/perfect-panel/ppanel-server/pkg/constant"
)

func TestExtractVersionNumber(t *testing.T) {
	versionNumber := ExtractVersionNumber(constant.Version)
	t.Log(versionNumber)
}
