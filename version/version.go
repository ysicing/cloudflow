// AGPL License
// Copyright 2022 ysicing(i@ysicing.me).

package version

import "fmt"

var (
	Version       string
	BuildDate     string
	GitCommitHash string
)

const (
	defaultVersion       = "0.0.0"
	defaultGitCommitHash = "a1b2c3d4"
	defaultBuildDate     = "20220805"
)

// GetVersion returns the version of the application.
func GetVersion() string {
	if Version == "" {
		Version = defaultVersion
	}
	if BuildDate == "" {
		BuildDate = defaultBuildDate
	}
	if GitCommitHash == "" {
		GitCommitHash = defaultGitCommitHash
	}
	return fmt.Sprintf("build version: %s-%s-%s", Version, BuildDate, GitCommitHash)
}
