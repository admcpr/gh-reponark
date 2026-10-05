package main

import (
	"errors"
	"fmt"
	"strings"
)

// options is what the command line asks for.
type options struct {
	version bool // print the version and exit
	demo    bool // run against the built-in sample data instead of GitHub
}

// errHelp is returned by parseArgs when help was asked for, so main can
// print the usage as output rather than as an error.
var errHelp = errors.New(usage)

const usage = `usage: gh reponark [--demo] [--version]

  --demo     explore the built-in sample data instead of your GitHub account
  --version  print the installed version`

// parseArgs reads the command line. Flags may be given as --name or -name;
// anything else is an error whose message carries the usage.
func parseArgs(args []string) (options, error) {
	var o options
	for _, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			return o, fmt.Errorf("unknown argument %q\n%s", arg, usage)
		}
		switch strings.TrimLeft(arg, "-") {
		case "version", "v":
			o.version = true
		case "demo":
			o.demo = true
		case "help", "h":
			return o, errHelp
		default:
			return o, fmt.Errorf("unknown flag %q\n%s", arg, usage)
		}
	}
	return o, nil
}
