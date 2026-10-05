package repo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompactNumber(t *testing.T) {
	assert.Equal(t, "0", CompactNumber(0))
	assert.Equal(t, "999", CompactNumber(999))
	assert.Equal(t, "1k", CompactNumber(1000))
	assert.Equal(t, "1.2k", CompactNumber(1204))
	assert.Equal(t, "15.3k", CompactNumber(15300))
}

func TestKilobytes(t *testing.T) {
	assert.Equal(t, "512 KB", Kilobytes(512))
	assert.Equal(t, "1.5 MB", Kilobytes(1536))
	assert.Equal(t, "2.0 GB", Kilobytes(2*1024*1024))
}

func TestShortName(t *testing.T) {
	assert.Equal(t, "Wiki", ShortName("Has Wiki Enabled"))
	assert.Equal(t, "Archived", ShortName("Is Archived"))
	assert.Equal(t, "Administer", ShortName("Viewer Can Administer"))
	assert.Equal(t, "Stargazer", ShortName("Stargazer Count"))
	assert.Equal(t, "Default Branch", ShortName("Default Branch"))
}
