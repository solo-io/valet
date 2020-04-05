package version

var UndefinedVersion = "undefined"
var DevVersion = "dev"

// This will be set by the linker during build
var Version = UndefinedVersion

func IsReleaseVersion() bool {
	return Version != UndefinedVersion && Version != DevVersion
}
