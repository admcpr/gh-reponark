package main

import (
	"fmt"
	"runtime/debug"
)

// version is set at release time with
//
//	go build -ldflags "-X main.version=v1.2.3"
//
// and stays "dev" for local builds, which then report the commit they were
// built from instead.
var version = "dev"

// versionString is what `gh reponark --version` prints: the release tag, or
// for a development build the VCS revision and whether the tree was dirty.
func versionString(info *debug.BuildInfo) string {
	if version != "dev" || info == nil {
		return "gh-reponark " + version
	}
	// A `go install module@version` build carries the module version even
	// though nothing set the ldflag.
	if v := info.Main.Version; v != "" && v != "(devel)" {
		return "gh-reponark " + v
	}
	revision, dirty := "", false
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			revision = s.Value
		case "vcs.modified":
			dirty = s.Value == "true"
		}
	}
	if revision == "" {
		return "gh-reponark dev"
	}
	if len(revision) > 12 {
		revision = revision[:12]
	}
	if dirty {
		revision += "-dirty"
	}
	return "gh-reponark dev (" + revision + ")"
}

func printVersion() {
	info, _ := debug.ReadBuildInfo()
	fmt.Println(versionString(info))
}
