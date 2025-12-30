package org

import (
	"testing"

	"gh-reponark/shared"

	githubassert "github.com/stretchr/testify/assert"
)

func TestHelpViewNotEmpty(t *testing.T) {
	m := NewModel(shared.OrgKey{Name: "demo", IsUser: false}, 80, 24)
	m.SetDimensions(80, 24)
	content := m.HelpView().Content
	githubassert.NotEmpty(t, content)
}
