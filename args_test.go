package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseArgs(t *testing.T) {
	for _, args := range [][]string{{"--version"}, {"-v"}, {"-version"}} {
		o, err := parseArgs(args)
		assert.NoError(t, err, args)
		assert.True(t, o.version, args)
	}
	o, err := parseArgs([]string{"--demo", "--version"})
	assert.NoError(t, err)
	assert.Equal(t, options{version: true, demo: true}, o, "flags combine")

	o, err = parseArgs(nil)
	assert.NoError(t, err)
	assert.Equal(t, options{}, o)

	_, err = parseArgs([]string{"version"})
	assert.ErrorContains(t, err, `unknown argument "version"`, "a bare word is not a flag")
	_, err = parseArgs([]string{"--version", "--bogus"})
	assert.ErrorContains(t, err, `unknown flag "--bogus"`, "every argument is checked")
	_, err = parseArgs([]string{"--help"})
	assert.ErrorIs(t, err, errHelp)
}
