package updater

import (
	"crypto/sha256"
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetBinaryName(t *testing.T) {
	name := GetBinaryName()

	// Should contain current OS and arch
	assert.Contains(t, name, runtime.GOOS)
	assert.Contains(t, name, runtime.GOARCH)

	// Should start with gdk-
	assert.True(t, len(name) > 4 && name[:4] == "gdk-")

	// Windows should have .exe extension
	if runtime.GOOS == "windows" {
		assert.Contains(t, name, ".exe")
	} else {
		assert.NotContains(t, name, ".exe")
	}
}

func TestGetBinaryNameForPlatform(t *testing.T) {
	tests := []struct {
		goos   string
		goarch string
		want   string
	}{
		{"linux", "amd64", "gdk-linux-amd64"},
		{"linux", "arm64", "gdk-linux-arm64"},
		{"darwin", "amd64", "gdk-darwin-amd64"},
		{"darwin", "arm64", "gdk-darwin-arm64"},
		{"windows", "amd64", "gdk-windows-amd64.exe"},
	}

	for _, tt := range tests {
		t.Run(tt.goos+"-"+tt.goarch, func(t *testing.T) {
			got := GetBinaryNameForPlatform(tt.goos, tt.goarch)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name    string
		current string
		latest  string
		want    int // -1: current < latest, 0: equal, 1: current > latest
	}{
		{"equal versions", "v0.1.0", "v0.1.0", 0},
		{"newer available", "v0.1.0", "v0.2.0", -1},
		{"already latest", "v0.2.0", "v0.1.0", 1},
		{"patch update", "v0.1.0", "v0.1.1", -1},
		{"major update", "v0.1.0", "v1.0.0", -1},
		{"without v prefix", "0.1.0", "0.2.0", -1},
		{"mixed prefix", "v0.1.0", "0.2.0", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CompareVersions(tt.current, tt.latest)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNeedsUpdate(t *testing.T) {
	tests := []struct {
		name    string
		current string
		latest  string
		want    bool
	}{
		{"needs update", "v0.1.0", "v0.2.0", true},
		{"already latest", "v0.2.0", "v0.2.0", false},
		{"ahead of latest", "v0.3.0", "v0.2.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &Updater{CurrentVersion: tt.current}
			got := u.NeedsUpdate(tt.latest)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestVerifyChecksum(t *testing.T) {
	data := []byte("test data for checksum verification")
	hash := sha256.Sum256(data)
	expectedChecksum := fmt.Sprintf("%x", hash)

	t.Run("valid checksum", func(t *testing.T) {
		assert.True(t, VerifyChecksum(data, expectedChecksum))
	})

	t.Run("invalid checksum", func(t *testing.T) {
		assert.False(t, VerifyChecksum(data, "invalid-checksum"))
	})

	t.Run("wrong data", func(t *testing.T) {
		wrongData := []byte("different data")
		assert.False(t, VerifyChecksum(wrongData, expectedChecksum))
	})
}

func TestParseChecksumsFile(t *testing.T) {
	content := `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855  gdk-linux-amd64
a1b2c3d4e5f6789012345678901234567890123456789012345678901234abcd  gdk-linux-arm64
deadbeef12345678901234567890123456789012345678901234567890abcdef  gdk-darwin-amd64
cafebabe12345678901234567890123456789012345678901234567890fedcba  gdk-darwin-arm64
12345678901234567890123456789012345678901234567890123456789abcde  gdk-windows-amd64.exe
`

	checksums := ParseChecksumsFile(content)

	assert.Equal(t, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", checksums["gdk-linux-amd64"])
	assert.Equal(t, "a1b2c3d4e5f6789012345678901234567890123456789012345678901234abcd", checksums["gdk-linux-arm64"])
	assert.Equal(t, "deadbeef12345678901234567890123456789012345678901234567890abcdef", checksums["gdk-darwin-amd64"])
	assert.Equal(t, "cafebabe12345678901234567890123456789012345678901234567890fedcba", checksums["gdk-darwin-arm64"])
	assert.Equal(t, "12345678901234567890123456789012345678901234567890123456789abcde", checksums["gdk-windows-amd64.exe"])
}

func TestParseChecksumsFileEmpty(t *testing.T) {
	checksums := ParseChecksumsFile("")
	assert.Empty(t, checksums)
}

func TestParseChecksumsFileInvalidLines(t *testing.T) {
	content := `valid-checksum  valid-file
invalid line without proper format
another-checksum  another-file
`
	checksums := ParseChecksumsFile(content)
	assert.Len(t, checksums, 2)
	assert.Equal(t, "valid-checksum", checksums["valid-file"])
	assert.Equal(t, "another-checksum", checksums["another-file"])
}

func TestBuildDownloadURL(t *testing.T) {
	u := &Updater{
		RepoOwner: "ygncode",
		RepoName:  "gdk",
	}

	url := u.BuildDownloadURL("v0.2.0", "gdk-linux-amd64")
	expected := "https://github.com/ygncode/gdk/releases/download/v0.2.0/gdk-linux-amd64"
	assert.Equal(t, expected, url)
}

func TestBuildChecksumsURL(t *testing.T) {
	u := &Updater{
		RepoOwner: "ygncode",
		RepoName:  "gdk",
	}

	url := u.BuildChecksumsURL("v0.2.0")
	expected := "https://github.com/ygncode/gdk/releases/download/v0.2.0/checksums.txt"
	assert.Equal(t, expected, url)
}

func TestBuildAPIURL(t *testing.T) {
	u := &Updater{
		RepoOwner: "ygncode",
		RepoName:  "gdk",
	}

	url := u.BuildAPIURL()
	expected := "https://api.github.com/repos/ygncode/gdk/releases/latest"
	assert.Equal(t, expected, url)
}
