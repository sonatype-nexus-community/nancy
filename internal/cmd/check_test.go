package cmd

import (
	"github.com/blang/semver"
	"github.com/sonatype-nexus-community/nancy/buildversion"
	"testing"
)

func TestGetVersionNumberSemver(t *testing.T) {
	origBuildVersion := buildversion.BuildVersion
	defer func() {
		buildversion.BuildVersion = origBuildVersion
	}()

	// check default ("development")
	semver.MustParse(getVersionNumberSemver())

	// check explicit "development"
	buildversion.BuildVersion = "development"
	semver.MustParse(getVersionNumberSemver())

	buildversion.BuildVersion = "1.2.3"
	semver.MustParse(getVersionNumberSemver())
}
