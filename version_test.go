package succincter

import "testing"

func TestVersion(t *testing.T) {
	if Version == "" {
		t.Error("Version should not be empty")
	}

	if VersionMajor < 0 || VersionMinor < 0 || VersionPatch < 0 {
		t.Error("Version components should be non-negative")
	}
}

func TestFullVersion(t *testing.T) {
	v := FullVersion()
	if v == "" {
		t.Error("FullVersion() should not return empty string")
	}

	// When no prerelease, should equal Version
	if VersionPrerelease == "" && v != Version {
		t.Errorf("FullVersion() = %q; want %q", v, Version)
	}
}
