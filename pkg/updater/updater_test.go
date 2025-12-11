package updater

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUpdater(t *testing.T) {
	u := NewUpdater()
	assert.NotNil(t, u)
	assert.Equal(t, "OmnTeam", u.Owner)
	assert.Equal(t, "server", u.Repo)
	assert.NotNil(t, u.HTTPClient)
}

func TestCompareVersions(t *testing.T) {
	u := NewUpdater()

	tests := []struct {
		name           string
		newVersion     string
		currentVersion string
		expected       bool
	}{
		{
			name:           "same version",
			newVersion:     "v1.0.0",
			currentVersion: "v1.0.0",
			expected:       false,
		},
		{
			name:           "different version",
			newVersion:     "v1.1.0",
			currentVersion: "v1.0.0",
			expected:       true,
		},
		{
			name:           "unknown current version",
			newVersion:     "v1.0.0",
			currentVersion: "unknown version",
			expected:       true,
		},
		{
			name:           "version without v prefix",
			newVersion:     "1.1.0",
			currentVersion: "1.0.0",
			expected:       true,
		},
		{
			name:           "empty current version",
			newVersion:     "v1.0.0",
			currentVersion: "",
			expected:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := u.compareVersions(tt.newVersion, tt.currentVersion)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetAssetName(t *testing.T) {
	u := NewUpdater()
	u.CurrentVersion = "v1.0.0"

	assetName := u.getAssetName()
	assert.NotEmpty(t, assetName)
	assert.Contains(t, assetName, "ppanel-server")
	assert.Contains(t, assetName, "v1.0.0")
}
