package main

import (
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionString(t *testing.T) {
	defer func(v string) { version = v }(version)

	version = "v1.2.3"
	assert.Equal(t, "gh-reponark v1.2.3", versionString(nil), "a release reports its tag")

	version = "dev"
	assert.Equal(t, "gh-reponark dev", versionString(nil))
	assert.Equal(t, "gh-reponark dev", versionString(&debug.BuildInfo{}), "no VCS info")
	assert.Equal(t, "gh-reponark v2.0.0", versionString(&debug.BuildInfo{Main: debug.Module{Version: "v2.0.0"}}), "a go install build knows its module version")
	assert.Equal(t, "gh-reponark dev", versionString(&debug.BuildInfo{Main: debug.Module{Version: "(devel)"}}))

	info := &debug.BuildInfo{Settings: []debug.BuildSetting{
		{Key: "vcs.revision", Value: "0123456789abcdef0123456789abcdef01234567"},
		{Key: "vcs.modified", Value: "true"},
	}}
	assert.Equal(t, "gh-reponark dev (0123456789ab-dirty)", versionString(info))
}
