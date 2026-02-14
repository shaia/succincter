package succincter

// Version information for the Succincter library.
const (
	// Version is the current semantic version of the library.
	Version = "0.1.0"

	// VersionMajor is the major version component.
	VersionMajor = 0

	// VersionMinor is the minor version component.
	VersionMinor = 1

	// VersionPatch is the patch version component.
	VersionPatch = 0

	// VersionPrerelease is the pre-release identifier (empty for stable releases).
	VersionPrerelease = ""
)

// FullVersion returns the complete version string including pre-release if present.
func FullVersion() string {
	if VersionPrerelease != "" {
		return Version + "-" + VersionPrerelease
	}
	return Version
}
