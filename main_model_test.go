package main

import (
	"reflect"
	"testing"

	"gh-reponark/filters"
	"gh-reponark/org"
	"gh-reponark/shared"
	"gh-reponark/user"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
)

func TestValidateTransition_AllowsExpectedFlow(t *testing.T) {
	tests := []struct {
		name    string
		current tea.Model
		next    tea.Model
		wantErr bool
	}{
		{"user -> org ok", &user.Model{}, &org.Model{}, false},
		{"org -> filters ok", &org.Model{}, &filters.Model{}, false},
		{"filters -> filter detail ok", &filters.Model{}, &filters.BoolModel{}, false},
		{"detail -> anything ok", &filters.BoolModel{}, &user.Model{}, false},
		{"org -> detail blocked", &org.Model{}, &filters.BoolModel{}, true},
		{"user -> filters blocked", &user.Model{}, &filters.Model{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTransition(tt.current, tt.next)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMainModel_Next_FromFiltersCreatesDetail(t *testing.T) {
	m := MainModel{nav: shared.NewNavigator()}

	base := filters.NewModel(nil, 20, 10)
	assert.NoError(t, m.nav.Push(base))

	prop := filters.Property{Name: "is archived", Type: "bool"}
	cmd := m.Next(shared.NextMsg{ModelData: prop})

	assert.NotNil(t, cmd)
	assert.Equal(t, 2, m.nav.Len())

	top, _ := m.nav.Current()
	_, ok := top.(*filters.BoolModel)
	assert.True(t, ok, "top of stack should be BoolModel")
}

func TestMainModel_Previous_EmptyNavQuits(t *testing.T) {
	m := MainModel{nav: shared.NewNavigator()}

	cmd := m.Previous(shared.PreviousMsg{})
	assert.Equal(t, reflect.ValueOf(tea.Quit).Pointer(), reflect.ValueOf(cmd).Pointer())
}
