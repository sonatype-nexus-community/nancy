package useragent

import (
	"fmt"
	"os"
	"runtime"

	"github.com/sonatype-nexus-community/nancy/buildversion"
)

func GetUserAgent() (useragent string) {
	useragent = fmt.Sprintf("nancy-client/%s", buildversion.BuildVersion)
	if checkForCIEnvironment() {
		callerInfo := getCallerInfo()
		if callerInfo == "" && checkForCircleCI() {
			useragent = useragent + fmt.Sprintf(" (%s; %s %s)", "circleci", runtime.GOOS, runtime.GOARCH)
		}
		if callerInfo == "" && checkForBitBucket() {
			useragent = useragent + fmt.Sprintf(" (%s; %s %s)", "bitbucket", runtime.GOOS, runtime.GOARCH)
		} else {
			useragent = useragent + fmt.Sprintf(" (%s; %s %s)", callerInfo, runtime.GOOS, runtime.GOARCH)
		}
	} else {
		useragent = useragent + fmt.Sprintf(" (%s; %s %s)", "non ci usage", runtime.GOOS, runtime.GOARCH)
	}

	return
}

func checkForCIEnvironment() bool {
	s := os.Getenv("CI")
	if s != "" {
		return true
	}
	return false
}

// Returns info from SC_CALLER_INFO, example: bitbucket-nancy-pipe-0.1.9
func getCallerInfo() string {
	s := os.Getenv("SC_CALLER_INFO")
	if s != "" {
		return s
	}
	return ""
}

func checkForCircleCI() bool {
	s := os.Getenv("CIRCLECI")
	if s != "" {
		return true
	}
	return false
}

func checkForBitBucket() bool {
	s := os.Getenv("BITBUCKET_BUILD_NUMBER")
	if s != "" {
		return true
	}
	return false
}
