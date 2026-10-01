package org

import (
	"testing"

	"gh-reponark/github/githubtest"
	"gh-reponark/shared"

	githubassert "github.com/stretchr/testify/assert"
)

func TestHelpNotEmpty(t *testing.T) {
	m := NewModel(&githubtest.Fake{}, shared.OrgKey{Name: "demo", IsUser: false}, 80, 24)
	m.SetDimensions(80, 24)
	githubassert.NotEmpty(t, m.Help().String())
}
