package filters

import "strings"

func containsCaseInsensitive(haystack, needle string) bool {
	haystackLower := strings.ToLower(haystack)
	needleLower := strings.ToLower(needle)
	return strings.Contains(haystackLower, needleLower)
}
