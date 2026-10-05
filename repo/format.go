package repo

import (
	"fmt"
	"strings"
)

// CompactNumber shortens large counts, e.g. 1204 becomes "1.2k".
func CompactNumber(n int) string {
	if n < 1000 {
		return fmt.Sprint(n)
	}
	s := fmt.Sprintf("%.1f", float64(n)/1000)
	return strings.TrimSuffix(s, ".0") + "k"
}

// Kilobytes formats a size given in kilobytes, as GitHub reports disk usage.
func Kilobytes(kb int) string {
	switch {
	case kb < 1024:
		return fmt.Sprintf("%d KB", kb)
	case kb < 1024*1024:
		return fmt.Sprintf("%.1f MB", float64(kb)/1024)
	default:
		return fmt.Sprintf("%.1f GB", float64(kb)/(1024*1024))
	}
}

// ShortName drops the words every property in a group shares, so "Has Wiki
// Enabled" reads as "Wiki" where the group already says what it is.
func ShortName(name string) string {
	for _, prefix := range []string{"Viewer Can ", "Viewer ", "Is ", "Has "} {
		name = strings.TrimPrefix(name, prefix)
	}
	for _, suffix := range []string{" Enabled", " Allowed", " Count"} {
		name = strings.TrimSuffix(name, suffix)
	}
	return name
}
