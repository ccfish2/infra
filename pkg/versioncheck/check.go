package versioncheck

import (
	"fmt"
	"strconv"
	"strings"

	semver "github.com/blang/semver/v4"
)

func MustVersion(version string) semver.Version {
	ver, err := Version(version)
	if err != nil {
		panic(fmt.Errorf("cannot compile go-version version '%s': %w", version, err))
	}
	return ver
}

func Version(version string) (semver.Version, error) {
	ver, err := semver.ParseTolerant(version)
	if err != nil {
		return ver, err
	}

	if len(ver.Pre) == 0 {
		return ver, nil
	}

	for _, pre := range ver.Pre {
		if strings.Contains(pre.VersionStr, "rc") ||
			strings.Contains(pre.VersionStr, "beta") ||
			strings.Contains(pre.VersionStr, "alpha") ||
			strings.Contains(pre.VersionStr, "snapshot") {
			return ver, nil
		}
	}

	strSegments := make([]string, 3)
	strSegments[0] = strconv.Itoa(int(ver.Major))
	strSegments[1] = strconv.Itoa(int(ver.Minor))
	strSegments[2] = strconv.Itoa(int(ver.Patch))
	verStr := strings.Join(strSegments, ".")
	return semver.ParseTolerant(verStr)
}
