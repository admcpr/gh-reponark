package repo

import (
	"fmt"
	"gh-reponark/shared"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"
)

type RepoConfig struct {
	Name           string
	Url            string
	Properties     map[string]RepoProperty
	PropertyGroups map[string][]RepoProperty
	GroupKeys      []string
}

func NewRepoConfig(r Repository) RepoConfig {
	propertyGroups := make(map[string][]RepoProperty)
	properties := ToProperties(r)
	for _, p := range properties {
		propertyGroups[p.Group] = append(propertyGroups[p.Group], p)
	}
	keys := make([]string, 0, len(propertyGroups))
	for k := range propertyGroups {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return RepoConfig{
		Name:           r.Name,
		Url:            r.Url,
		Properties:     properties,
		PropertyGroups: propertyGroups,
		GroupKeys:      keys,
	}
}

type RepoProperty struct {
	Name        string
	Group       string
	Value       interface{}
	Type        string
	Description string
}

func (s RepoProperty) String() string {
	switch value := s.Value.(type) {
	case bool:
		return shared.YesNo(value)
	case string:
		return value
	case time.Time:
		return value.Format("2006/01/02")
	case int:
		return fmt.Sprint(value)
	default:
		return "Unknown type"
	}
}

func NewRepoProperty(name string, group string, value interface{}, typeStr string, description string) RepoProperty {
	return RepoProperty{Name: name, Group: group, Value: value, Type: typeStr, Description: description}
}

// ToProperties flattens a Repository into its properties keyed by name. The
// names, groups and types match Schema; this adds the values.
func ToProperties(r Repository) map[string]RepoProperty {
	properties := repositoryProperties(r)

	propertiesMap := make(map[string]RepoProperty, len(properties))
	for _, p := range properties {
		propertiesMap[p.Name] = p
	}

	return propertiesMap
}

func processField(field reflect.StructField, value reflect.Value) []RepoProperty {
	var properties []RepoProperty

	name := field.Tag.Get("name")
	if name == "" {
		name = splitCamelCase(field.Name)
	}
	desc := field.Tag.Get("desc")
	group := field.Tag.Get("group")
	typeStr := field.Type.String()

	if typeStr == "time.Time" || typeStr == "int" || typeStr == "string" || typeStr == "bool" {
		properties = append(properties, NewRepoProperty(name, group, value.Interface(), typeStr, desc))
	} else {
		switch field.Type.Kind() {
		case reflect.Struct:
			// Recursively process nested struct
			for i := 0; i < field.Type.NumField(); i++ {
				nestedField := field.Type.Field(i)
				nestedValue := value.Field(i)
				properties = append(properties, processField(nestedField, nestedValue)...)
			}
		case reflect.Ptr:
			// Handle pointer to struct
			if !value.IsNil() && value.Elem().Kind() == reflect.Struct {
				properties = append(properties, processField(field, value.Elem())...)
			} else {
				properties = append(properties, NewRepoProperty(name, group, value.Interface(), typeStr, desc))
			}
		}
	}

	return properties
}

func splitCamelCase(s string) string {
	// Add space before capital letters
	words := regexp.MustCompile(`([a-z0-9])([A-Z])`).ReplaceAllString(s, "$1 $2")
	// Handle consecutive capital letters (like "ID" in "UserID")
	words = regexp.MustCompile(`([A-Z])([A-Z][a-z])`).ReplaceAllString(words, "$1 $2")
	// Capitalize first letter and trim spaces
	return strings.TrimSpace(words)
}
