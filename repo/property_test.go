package repo

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestToProperties(t *testing.T) {
	repo := Repository{
		Id:            "123",
		Name:          "test-repo",
		Description:   "A test repository",
		NameWithOwner: "test-repo-owner/test-repo",
	}

	properties := ToProperties(repo)

	if len(properties) != 49 {
		t.Fatalf("expected 49 properties, got %d", len(properties))
	}
}

func TestNewRepoConfig(t *testing.T) {
	expectedPropertyCount := 49
	expectedPropertyGroupCount := 7

	repo := Repository{
		Id:            "123",
		Name:          "test-repo",
		Description:   "A test repository",
		NameWithOwner: "test-repo-owner/test-repo",
	}

	repoConfig := NewRepoConfig(repo)

	assert.Equal(t, expectedPropertyCount, len(repoConfig.Properties))
	assert.Equal(t, expectedPropertyGroupCount, len(repoConfig.PropertyGroups))
}

func TestNewRepoConfig_CopiesNameAndUrl(t *testing.T) {
	repoConfig := NewRepoConfig(Repository{Name: "test-repo", Url: "https://github.com/owner/test-repo"})

	assert.Equal(t, "test-repo", repoConfig.Name)
	assert.Equal(t, "https://github.com/owner/test-repo", repoConfig.Url)
}

func TestNewRepoConfig_GroupKeysAreSorted(t *testing.T) {
	repoConfig := NewRepoConfig(Repository{})

	want := []string{
		"1⟭ Overview",
		"2⟭ Status",
		"3⟭ Metrics",
		"4⟭ Features",
		"5⟭ Merge",
		"6⟭ Permissions",
		"6⟭ Security",
	}
	assert.Equal(t, want, repoConfig.GroupKeys)

	for _, key := range want {
		assert.NotEmpty(t, repoConfig.PropertyGroups[key], "group %q should have properties", key)
		for _, p := range repoConfig.PropertyGroups[key] {
			assert.Equal(t, key, p.Group)
		}
	}
}

func TestNewRepoConfig_PropertyGroupsCoverAllProperties(t *testing.T) {
	repoConfig := NewRepoConfig(Repository{})

	total := 0
	for _, group := range repoConfig.PropertyGroups {
		total += len(group)
	}

	assert.Equal(t, len(repoConfig.Properties), total)
}

func TestToProperties_Values(t *testing.T) {
	createdAt := time.Date(2024, 3, 4, 5, 6, 7, 0, time.UTC)
	repo := Repository{
		Name:           "test-repo",
		IsArchived:     true,
		StargazerCount: 42,
		CreatedAt:      createdAt,
	}
	repo.PrimaryLanguage.Name = "Go"
	repo.Issues.TotalCount = 3
	repo.OpenPullRequests.TotalCount = 5
	repo.DefaultBranchRef.Name = "main"

	properties := ToProperties(repo)

	tests := []struct {
		name      string
		wantValue interface{}
		wantType  string
		wantGroup string
	}{
		{name: "Name", wantValue: "test-repo", wantType: "string", wantGroup: "1⟭ Overview"},
		{name: "Is Archived", wantValue: true, wantType: "bool", wantGroup: "2⟭ Status"},
		{name: "Stargazer Count", wantValue: 42, wantType: "int", wantGroup: "3⟭ Metrics"},
		{name: "Created At", wantValue: createdAt, wantType: "time.Time", wantGroup: "3⟭ Metrics"},
		{name: "Primary Language", wantValue: "Go", wantType: "string", wantGroup: "1⟭ Overview"},
		{name: "Open Issues", wantValue: 3, wantType: "int", wantGroup: "3⟭ Metrics"},
		{name: "Open Pull Requests", wantValue: 5, wantType: "int", wantGroup: "3⟭ Metrics"},
		{name: "Default Branch", wantValue: "main", wantType: "string", wantGroup: "5⟭ Merge"},
		{name: "Database ID", wantValue: 0, wantType: "int", wantGroup: "1⟭ Overview"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, ok := properties[tt.name]
			if !assert.True(t, ok, "property %q should exist", tt.name) {
				return
			}
			assert.Equal(t, tt.name, p.Name)
			assert.Equal(t, tt.wantValue, p.Value)
			assert.Equal(t, tt.wantType, p.Type)
			assert.Equal(t, tt.wantGroup, p.Group)
			assert.NotEmpty(t, p.Description)
		})
	}
}

func TestRepoProperty_String(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  string
	}{
		{name: "true bool", value: true, want: "Yes"},
		{name: "false bool", value: false, want: "No"},
		{name: "string", value: "hello", want: "hello"},
		{name: "empty string", value: "", want: ""},
		{name: "int", value: 42, want: "42"},
		{name: "negative int", value: -7, want: "-7"},
		{name: "time", value: time.Date(2024, 3, 4, 5, 6, 7, 0, time.UTC), want: "2024/03/04"},
		{name: "unsupported float", value: 1.5, want: "Unknown type"},
		{name: "unsupported slice", value: []string{"a"}, want: "Unknown type"},
		{name: "nil", value: nil, want: "Unknown type"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, RepoProperty{Value: tt.value}.String())
		})
	}
}

func TestNewRepoProperty(t *testing.T) {
	p := NewRepoProperty("Name", "Group", "value", "string", "A description")

	assert.Equal(t, "Name", p.Name)
	assert.Equal(t, "Group", p.Group)
	assert.Equal(t, "value", p.Value)
	assert.Equal(t, "string", p.Type)
	assert.Equal(t, "A description", p.Description)
}

func TestSplitCamelCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "Name", want: "Name"},
		{input: "IsArchived", want: "Is Archived"},
		{input: "HasIssuesEnabled", want: "Has Issues Enabled"},
		{input: "DatabaseID", want: "Database ID"},
		{input: "HTMLParser", want: "HTML Parser"},
		{input: "DescriptionHTML", want: "Description HTML"},
		{input: "Field1Name", want: "Field1 Name"},
		{input: "lowercase", want: "lowercase"},
		{input: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, splitCamelCase(tt.input))
		})
	}
}

type processFieldNested struct {
	Inner string `name:"Inner Value" group:"nested" desc:"nested description"`
}

type processFieldSample struct {
	Plain    string `group:"plain" desc:"plain description"`
	Tagged   int    `name:"Custom Name" group:"tagged" desc:"tagged description"`
	Flag     bool
	Stamp    time.Time
	Nested   processFieldNested
	NilPtr   *processFieldNested
	Slice    []string
	Untagged struct {
		DeepValue int
	}
}

func TestProcessField(t *testing.T) {
	sample := processFieldSample{Plain: "text", Tagged: 9, Flag: true}
	sample.Nested.Inner = "inner"
	sample.Untagged.DeepValue = 3

	structType := reflect.TypeOf(sample)
	structValue := reflect.ValueOf(sample)

	field := func(name string) (reflect.StructField, reflect.Value) {
		f, ok := structType.FieldByName(name)
		if !ok {
			t.Fatalf("no field %q", name)
		}
		return f, structValue.FieldByName(name)
	}

	tests := []struct {
		name string
		want []RepoProperty
	}{
		{
			name: "Plain",
			want: []RepoProperty{{Name: "Plain", Group: "plain", Value: "text", Type: "string", Description: "plain description"}},
		},
		{
			name: "Tagged",
			want: []RepoProperty{{Name: "Custom Name", Group: "tagged", Value: 9, Type: "int", Description: "tagged description"}},
		},
		{
			name: "Flag",
			want: []RepoProperty{{Name: "Flag", Value: true, Type: "bool"}},
		},
		{
			name: "Stamp",
			want: []RepoProperty{{Name: "Stamp", Value: time.Time{}, Type: "time.Time"}},
		},
		{
			name: "Nested",
			want: []RepoProperty{{Name: "Inner Value", Group: "nested", Value: "inner", Type: "string", Description: "nested description"}},
		},
		{
			name: "NilPtr",
			want: []RepoProperty{{Name: "Nil Ptr", Value: (*processFieldNested)(nil), Type: "*repo.processFieldNested"}},
		},
		{
			name: "Slice",
			want: nil,
		},
		{
			name: "Untagged",
			want: []RepoProperty{{Name: "Deep Value", Value: 3, Type: "int"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, v := field(tt.name)
			assert.Equal(t, tt.want, processField(f, v))
		})
	}
}
