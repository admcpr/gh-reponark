package repo

import (
	"reflect"
	"sync"
)

// PropertySchema describes one property of a Repository without its value:
// what it is called, which group it is displayed under, its Go type and a
// human readable description. All of it comes from the struct tags on
// Repository, so this is the single source of truth for both the filter
// search and the detail pane.
type PropertySchema struct {
	Name        string
	Group       string
	Type        string
	Description string
}

var (
	schemaOnce sync.Once
	schema     []PropertySchema
)

// Schema returns every property of Repository in declaration order. It is
// computed once; callers receive a copy they are free to modify.
func Schema() []PropertySchema {
	schemaOnce.Do(func() {
		for _, p := range repositoryProperties(Repository{}) {
			schema = append(schema, PropertySchema{
				Name:        p.Name,
				Group:       p.Group,
				Type:        p.Type,
				Description: p.Description,
			})
		}
	})
	return append([]PropertySchema(nil), schema...)
}

// repositoryProperties flattens a Repository into its properties, in the
// order the fields are declared.
func repositoryProperties(r Repository) []RepoProperty {
	var properties []RepoProperty
	t := reflect.TypeOf(r)
	v := reflect.ValueOf(r)

	for i := 0; i < t.NumField(); i++ {
		properties = append(properties, processField(t.Field(i), v.Field(i))...)
	}

	return properties
}
