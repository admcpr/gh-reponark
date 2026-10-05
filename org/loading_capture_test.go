//go:build capture

package org

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"gh-reponark/github"
	"gh-reponark/github/githubtest"
	"gh-reponark/repo"
	"gh-reponark/shared"

	"github.com/charmbracelet/x/ansi"
)

// TestCaptureLoading writes the loading screen part way through a load, for
// looking at rather than asserting on.
func TestCaptureLoading(t *testing.T) {
	dir := os.Getenv("CAPTURE_DIR")
	if dir == "" {
		t.Skip("CAPTURE_DIR not set")
	}
	m := NewModel(&githubtest.Fake{}, shared.OrgKey{Name: "acme-corp"}, 118, 33)
	refs := make([]github.RepositoryRef, 148)
	for i := range refs {
		refs[i] = github.RepositoryRef{Name: fmt.Sprintf("service-%03d", i)}
	}
	m.Update(repositoryPageMsg(github.RepositoryPage{Repositories: refs, TotalCount: 148}))
	batch := make([]repo.Repository, 61)
	for i := range batch {
		batch[i] = repo.Repository{Name: []string{"api-gateway", "billing", "charts", "docs", "ember-forms", "fjord"}[i%6] + fmt.Sprintf("-%d", i)}
	}
	m.Update(repositoryBatchMsg(batch))
	for i := 0; i < 30; i++ {
		m.Update(m.spinner.Tick())
	}
	// Let the bar's spring settle so the capture shows the fill.
	for cmd := m.progress.SetPercent(m.loadedFraction()); cmd != nil && m.progress.IsAnimating(); {
		_, cmd = m.Update(cmd())
	}
	content := fmt.Sprint(m.View().Content)
	os.WriteFile(filepath.Join(dir, "loading.ans"), []byte(content), 0o644)
	os.WriteFile(filepath.Join(dir, "loading.txt"), []byte(ansi.Strip(content)), 0o644)
}
