package filters

import (
	"fmt"
	"testing"
	"time"

	"gh-reponark/shared"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewFilterDateModel(t *testing.T) {
	const name = "Title"
	fromString := "2022-01-01"
	toString := "2022-12-31"
	from, _ := time.Parse("2006-01-02", fromString)
	to, _ := time.Parse("2006-01-02", toString)

	t.Run("NewFilterDateModel", func(t *testing.T) {
		m := NewDateModel(name, from, to, 60, 40)
		assert.Equal(t, m.Name(), name)
		assert.Equal(t, m.fromInput.Placeholder, fromString)
		assert.Equal(t, m.toInput.Placeholder, toString)
	})
}

func TestDateValidator(t *testing.T) {
	errorMessage := fmt.Errorf("please enter a YYYY-MM-DD date for `from`")

	tests := []struct {
		name   string
		input  string
		prompt string
		want   error
	}{
		{name: "Valid date", input: "2022-01-01", prompt: "from", want: nil},
		{name: "Invalid date format", input: "01-01-2022", prompt: "from", want: errorMessage},
		{name: "Invalid date value", input: "2022-13-01", prompt: "from", want: errorMessage},
		{name: "Too long input", input: "2022-01-01T00:00:00Z", prompt: "from", want: errorMessage},
		{name: "Invalid characters", input: "2022-01-0a", prompt: "from", want: errorMessage},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dateValidator(tt.input, tt.prompt)
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestFilterDateModel_GetValue(t *testing.T) {
	tests := []struct {
		name      string
		fromValue string
		toValue   string
		fromError error
		toError   error
		wantFrom  time.Time
		wantTo    time.Time
		wantErr   error
	}{
		{
			name:      "Valid input",
			fromValue: "2022-01-01",
			toValue:   "2022-12-31",
			wantFrom:  time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
			wantTo:    time.Date(2022, 12, 31, 0, 0, 0, 0, time.UTC),
			wantErr:   nil,
		},
		{
			name:      "Invalid from input",
			fromValue: "01-01-2022",
			toValue:   "2022-12-31",
			wantErr:   fmt.Errorf("please enter a YYYY-MM-DD date for `From`"),
		},
		{
			name:      "Invalid to input",
			fromValue: "2022-01-01",
			toValue:   "31-12-2022",
			wantErr:   fmt.Errorf("please enter a YYYY-MM-DD date for `To`"),
		},
		{
			name:      "Invalid from and to input",
			fromValue: "01-01-2022",
			toValue:   "31-12-2022",
			wantErr:   fmt.Errorf("please enter a YYYY-MM-DD date for `From`"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			m := NewDateModel(tt.name, time.Time{}, time.Time{}, 60, 40)

			m.fromInput.SetValue(tt.fromValue)
			m.toInput.SetValue(tt.toValue)

			gotFrom, gotTo, gotErr := m.Value()

			assert.Equal(t, gotFrom, tt.wantFrom)
			assert.Equal(t, gotTo, tt.wantTo)
			assert.Equal(t, gotErr, tt.wantErr)
		})
	}
}

func TestDateModel_SetDimensions(t *testing.T) {
	m := NewDateModel("Created At", time.Time{}, time.Time{}, 60, 40)

	m.SetDimensions(100, 50)

	assert.Equal(t, 100, m.width)
	assert.Equal(t, 50, m.height)
}

func TestDateModel_Init(t *testing.T) {
	m := NewDateModel("Created At", time.Time{}, time.Time{}, 60, 40)
	assert.NotNil(t, m.Init())
}

func TestDateModel_FocusStartsOnFrom(t *testing.T) {
	m := NewDateModel("Created At", time.Time{}, time.Time{}, 60, 40)

	assert.True(t, m.fromInput.Focused())
	assert.False(t, m.toInput.Focused())
}

func TestDateModel_Focus(t *testing.T) {
	m := NewDateModel("Created At", time.Time{}, time.Time{}, 60, 40)
	m.fromInput.Blur()

	m.Focus()

	assert.True(t, m.fromInput.Focused())
}

func TestDateModel_Update_TabTogglesFocus(t *testing.T) {
	m := NewDateModel("Created At", time.Time{}, time.Time{}, 60, 40)
	tab := tea.KeyPressMsg{Code: tea.KeyTab}

	m.Update(tab)
	assert.False(t, m.fromInput.Focused())
	assert.True(t, m.toInput.Focused())

	m.Update(tab)
	assert.True(t, m.fromInput.Focused())
	assert.False(t, m.toInput.Focused())
}

func TestDateModel_Update_TypingGoesToFocusedInput(t *testing.T) {
	m := NewDateModel("Created At", time.Time{}, time.Time{}, 60, 40)

	m.Update(tea.KeyPressMsg{Code: '2', Text: "2"})
	assert.Equal(t, "2", m.fromInput.Value())
	assert.Equal(t, "", m.toInput.Value())

	m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	m.Update(tea.KeyPressMsg{Code: '9', Text: "9"})
	assert.Equal(t, "2", m.fromInput.Value())
	assert.Equal(t, "9", m.toInput.Value())
}

func TestDateModel_Update_EnterSendsFilter(t *testing.T) {
	m := NewDateModel("Created At", time.Time{}, time.Time{}, 60, 40)
	m.fromInput.SetValue("2022-01-01")
	m.toInput.SetValue("2022-12-31")

	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	assert.NotNil(t, cmd)

	prev, ok := cmd().(shared.PreviousMsg)
	assert.True(t, ok, "expected PreviousMsg")

	filter, ok := prev.Message.(DateFilter)
	assert.True(t, ok, "expected DateFilter")
	assert.Equal(t, "Created At", filter.Name())
	assert.Equal(t, time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC), filter.From)
	assert.Equal(t, time.Date(2022, 12, 31, 0, 0, 0, 0, time.UTC), filter.To)
}

func TestDateModel_Update_EscGoesBack(t *testing.T) {
	m := NewDateModel("Created At", time.Time{}, time.Time{}, 60, 40)

	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	assert.NotNil(t, cmd)

	prev, ok := cmd().(shared.PreviousMsg)
	assert.True(t, ok, "expected PreviousMsg")
	assert.Nil(t, prev.Message)
}

func TestDateModel_Update_IgnoresNonKeyMessages(t *testing.T) {
	m := NewDateModel("Created At", time.Time{}, time.Time{}, 60, 40)

	_, cmd := m.Update(tea.WindowSizeMsg{Width: 10, Height: 10})

	assert.Nil(t, cmd)
	assert.True(t, m.fromInput.Focused())
}

func TestDateModel_Value_PartialInputUsesOpenBounds(t *testing.T) {
	m := NewDateModel("Created At", time.Time{}, time.Time{}, 60, 40)
	// These pass validation (digits and dashes, under 10 chars) but are not full dates.
	m.fromInput.SetValue("2022-01")
	m.toInput.SetValue("2022")

	from, to, err := m.Value()

	assert.NoError(t, err)
	assert.True(t, from.IsZero(), "unparseable from should fall back to the zero time")
	assert.True(t, to.After(time.Now().AddDate(1000, 0, 0)), "unparseable to should fall back to the maximum time")
}

func TestDateModel_Value_EmptyInputIsInvalid(t *testing.T) {
	m := NewDateModel("Created At", time.Time{}, time.Time{}, 60, 40)

	_, _, err := m.Value()

	assert.EqualError(t, err, "please enter a YYYY-MM-DD date for `From`")
}

func TestDateModel_View(t *testing.T) {
	m := NewDateModel("Created At", time.Time{}, time.Time{}, 100, 40)

	content := plain(m.View())
	assert.Contains(t, content, "Created At")
	assert.Contains(t, content, "From")
	assert.Contains(t, content, "To")
	assert.NotContains(t, content, "please enter")
}

func TestDateModel_View_ShowsValidationErrors(t *testing.T) {
	tests := []struct {
		name      string
		fromValue string
		toValue   string
		wantError string
	}{
		{name: "from error", fromValue: "bad", toValue: "2022-01-01", wantError: "date for `From`"},
		{name: "to error", fromValue: "2022-01-01", toValue: "bad", wantError: "date for `To`"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewDateModel("Created At", time.Time{}, time.Time{}, 100, 40)
			m.fromInput.SetValue(tt.fromValue)
			m.toInput.SetValue(tt.toValue)

			assert.Contains(t, plain(m.View()), tt.wantError)
		})
	}
}
