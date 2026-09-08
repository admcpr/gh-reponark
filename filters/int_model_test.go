package filters

import (
	"fmt"
	"testing"

	"gh-reponark/shared"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewFilterIntModel(t *testing.T) {
	tests := []struct {
		name string
		from int
		to   int
	}{
		{"Test 1", 0, 10},
		{"Test 2", -5, 5},
		{"Test 3", 100, 200},
		{"Test 4", -100, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewIntModel(tt.name, tt.from, tt.to, 60, 40)

			assert.Equal(t, m.Name(), tt.name)
			assert.Equal(t, m.fromInput.Placeholder, fmt.Sprint(tt.from))
			assert.Equal(t, m.toInput.Placeholder, fmt.Sprint(tt.to))
		})
	}
}

func TestIntValidator(t *testing.T) {
	tests := []struct {
		input  string
		prompt string
		want   error
	}{
		{"123", "Test 1", nil},
		{"-5", "Test 2", nil},
		{"abc", "Test 3", fmt.Errorf("please enter an integer for the `Test 3` value")},
		{"1.23", "Test 4", fmt.Errorf("please enter an integer for the `Test 4` value")},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s: %s", tt.prompt, tt.input), func(t *testing.T) {
			got := intValidator(tt.input, tt.prompt)

			assert.Equal(t, got, tt.want, fmt.Sprintf("got %q, want %q", got, tt.want))
		})
	}
}

func TestIntModel_SetDimensions(t *testing.T) {
	m := NewIntModel("Stars", 0, 10, 60, 40)

	m.SetDimensions(100, 50)

	assert.Equal(t, 100, m.width)
	assert.Equal(t, 50, m.height)
}

func TestIntModel_Init(t *testing.T) {
	m := NewIntModel("Stars", 0, 10, 60, 40)
	assert.NotNil(t, m.Init())
}

func TestIntModel_FocusStartsOnFrom(t *testing.T) {
	m := NewIntModel("Stars", 0, 10, 60, 40)

	assert.True(t, m.fromInput.Focused())
	assert.False(t, m.toInput.Focused())
}

func TestIntModel_Value(t *testing.T) {
	tests := []struct {
		name      string
		fromValue string
		toValue   string
		wantFrom  int
		wantTo    int
	}{
		{name: "valid values", fromValue: "5", toValue: "50", wantFrom: 5, wantTo: 50},
		{name: "negative values", fromValue: "-5", toValue: "-1", wantFrom: -5, wantTo: -1},
		{name: "empty values default to zero", fromValue: "", toValue: "", wantFrom: 0, wantTo: 0},
		{name: "invalid values default to zero", fromValue: "abc", toValue: "1.5", wantFrom: 0, wantTo: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewIntModel("Stars", 0, 10, 60, 40)
			m.fromInput.SetValue(tt.fromValue)
			m.toInput.SetValue(tt.toValue)

			from, to := m.Value()

			assert.Equal(t, tt.wantFrom, from)
			assert.Equal(t, tt.wantTo, to)
		})
	}
}

func TestIntModel_Update_TabTogglesFocus(t *testing.T) {
	tests := []struct {
		name string
		key  tea.KeyPressMsg
	}{
		{name: "tab", key: tea.KeyPressMsg{Code: tea.KeyTab}},
		{name: "shift+tab", key: tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewIntModel("Stars", 0, 10, 60, 40)

			m.Update(tt.key)
			assert.False(t, m.fromInput.Focused())
			assert.True(t, m.toInput.Focused())

			m.Update(tt.key)
			assert.True(t, m.fromInput.Focused())
			assert.False(t, m.toInput.Focused())
		})
	}
}

func TestIntModel_Update_TypingGoesToFocusedInput(t *testing.T) {
	m := NewIntModel("Stars", 0, 10, 60, 40)

	m.Update(tea.KeyPressMsg{Code: '4', Text: "4"})
	m.Update(tea.KeyPressMsg{Code: '2', Text: "2"})
	assert.Equal(t, "42", m.fromInput.Value())
	assert.Equal(t, "", m.toInput.Value())

	m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	m.Update(tea.KeyPressMsg{Code: '7', Text: "7"})
	assert.Equal(t, "42", m.fromInput.Value())
	assert.Equal(t, "7", m.toInput.Value())

	from, to := m.Value()
	assert.Equal(t, 42, from)
	assert.Equal(t, 7, to)
}

func TestIntModel_Update_EnterSendsFilter(t *testing.T) {
	m := NewIntModel("Stargazer Count", 0, 10, 60, 40)
	m.fromInput.SetValue("3")
	m.toInput.SetValue("30")

	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	assert.NotNil(t, cmd)

	prev, ok := cmd().(shared.PreviousMsg)
	assert.True(t, ok, "expected PreviousMsg")

	filter, ok := prev.Message.(IntFilter)
	assert.True(t, ok, "expected IntFilter")
	assert.Equal(t, "Stargazer Count", filter.Name())
	assert.Equal(t, 3, filter.From)
	assert.Equal(t, 30, filter.To)
}

func TestIntModel_Update_EscGoesBack(t *testing.T) {
	m := NewIntModel("Stars", 0, 10, 60, 40)

	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	assert.NotNil(t, cmd)

	prev, ok := cmd().(shared.PreviousMsg)
	assert.True(t, ok, "expected PreviousMsg")
	assert.Nil(t, prev.Message)
}

func TestIntModel_Update_IgnoresNonKeyMessages(t *testing.T) {
	m := NewIntModel("Stars", 0, 10, 60, 40)

	_, cmd := m.Update(tea.WindowSizeMsg{Width: 10, Height: 10})

	assert.Nil(t, cmd)
	assert.True(t, m.fromInput.Focused())
}

func TestIntModel_View(t *testing.T) {
	m := NewIntModel("Stargazer Count", 0, 10, 100, 40)

	content := plain(m.View())
	assert.Contains(t, content, "Stargazer Count")
	assert.Contains(t, content, "From")
	assert.Contains(t, content, "To")
	assert.NotContains(t, content, "please enter")
}

func TestIntModel_View_ShowsValidationErrors(t *testing.T) {
	tests := []struct {
		name      string
		fromValue string
		toValue   string
		wantError string
	}{
		{name: "from error", fromValue: "abc", toValue: "1", wantError: "`From` value"},
		{name: "to error", fromValue: "1", toValue: "abc", wantError: "`To` value"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewIntModel("Stargazer Count", 0, 10, 100, 40)
			m.fromInput.SetValue(tt.fromValue)
			m.toInput.SetValue(tt.toValue)

			assert.Contains(t, plain(m.View()), tt.wantError)
		})
	}
}

func TestIntModel_HelpView(t *testing.T) {
	m := NewIntModel("Stars", 0, 10, 120, 40)

	content := plain(m.HelpView())

	for _, want := range []string{"next field", "apply", "back"} {
		assert.Contains(t, content, want)
	}
	assert.NotContains(t, content, "toggle", "the int editor has nothing to toggle")
}

func TestIntModel_SetDimensions_ResizesHelp(t *testing.T) {
	m := NewIntModel("Stars", 0, 10, 60, 40)

	m.SetDimensions(100, 50)

	assert.Equal(t, 100, m.help.Width())
}

func TestIntModel_View_TitleIsJustTheName(t *testing.T) {
	m := NewIntModel("Stargazer Count", 0, 10, 100, 40)

	content := plain(m.View())

	assert.Contains(t, content, "Stargazer Count")
	assert.NotContains(t, content, "w:", "the title should not include debugging dimensions")
	assert.NotContains(t, content, "h:")
}
