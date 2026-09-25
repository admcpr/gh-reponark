package repo

import (
	"sort"
	"strings"
	"sync"
)

// Group is one tab of related properties, in the order they are declared on
// Repository.
type Group struct {
	Key        string
	Title      string
	Properties []PropertySchema
}

var (
	groupsOnce sync.Once
	groups     []Group
)

// Groups returns the property groups sorted by key. Every repository has the
// same groups, so screens can page through them before any repo is loaded.
func Groups() []Group {
	groupsOnce.Do(func() {
		byKey := map[string]*Group{}
		var keys []string
		for _, p := range Schema() {
			g, ok := byKey[p.Group]
			if !ok {
				g = &Group{Key: p.Group, Title: GroupTitle(p.Group)}
				byKey[p.Group] = g
				keys = append(keys, p.Group)
			}
			g.Properties = append(g.Properties, p)
		}
		sort.Strings(keys)
		for _, key := range keys {
			groups = append(groups, *byKey[key])
		}
	})
	return groups
}

// GroupTitles returns the display title of every group, in tab order.
func GroupTitles() []string {
	titles := make([]string, len(Groups()))
	for i, g := range Groups() {
		titles[i] = g.Title
	}
	return titles
}

// GroupTitle strips the ordering prefix from a group key, so "3⟭ Metrics"
// is shown as "Metrics".
func GroupTitle(key string) string {
	if _, title, ok := strings.Cut(key, "⟭"); ok {
		return strings.TrimSpace(title)
	}
	return key
}
