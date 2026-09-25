package filters

import (
	"errors"
	"fmt"
	"image/color"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gh-reponark/repo"
	"gh-reponark/shared"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

// editor edits the filter on one property. It always describes a filter, or
// none, so the screen can apply changes as they are typed.
type editor interface {
	// Update handles a message while the editor has focus.
	Update(msg tea.Msg) tea.Cmd
	// Filter returns the filter the editor describes, nil for no filter, or
	// an error when the input cannot be read.
	Filter() (Filter, error)
	Focus() tea.Cmd
	Blur()
	// View renders the editor's controls in width cells.
	View(focused bool, width int) []string
	// Keys lists the bindings the editor handles while focused.
	Keys() []key.Binding
}

// newEditor returns the editor for property, seeded with its current filter.
func newEditor(property repo.PropertySchema, current Filter) editor {
	switch property.Type {
	case "bool":
		return newBoolEditor(property.Name, current)
	case "int":
		return newIntEditor(property.Name, current)
	case "time.Time":
		return newDateEditor(property.Name, current)
	default:
		return newTextEditor(property.Name, current)
	}
}

// isSupportedPropertyType reports whether properties of type t can be
// filtered on.
func isSupportedPropertyType(t string) bool {
	switch t {
	case "bool", "int", "time.Time", "string":
		return true
	default:
		return false
	}
}

// ---- yes / no

var boolChoices = []string{"any", "yes", "no"}

type boolEditor struct {
	name   string
	choice int // index into boolChoices
	keymap boolEditorKeyMap
}

type boolEditorKeyMap struct {
	Prev, Next, Any, Yes, No key.Binding
}

func newBoolEditor(name string, current Filter) *boolEditor {
	e := &boolEditor{name: name, keymap: boolEditorKeyMap{
		Prev: key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/→", "choose")),
		Next: key.NewBinding(key.WithKeys("right", "l", "space")),
		Any:  key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "any")),
		Yes:  key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "yes")),
		No:   key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "no")),
	}}
	if f, ok := current.(BoolFilter); ok {
		e.choice = map[bool]int{true: 1, false: 2}[f.Value]
	}
	return e
}

// Cycle moves to the next choice: any, yes, no and round again.
func (e *boolEditor) Cycle() { e.choice = (e.choice + 1) % len(boolChoices) }

func (e *boolEditor) Update(msg tea.Msg) tea.Cmd {
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(msg, e.keymap.Prev):
			e.choice = (e.choice + len(boolChoices) - 1) % len(boolChoices)
		case key.Matches(msg, e.keymap.Next):
			e.Cycle()
		case key.Matches(msg, e.keymap.Any):
			e.choice = 0
		case key.Matches(msg, e.keymap.Yes):
			e.choice = 1
		case key.Matches(msg, e.keymap.No):
			e.choice = 2
		}
	}
	return nil
}

func (e *boolEditor) Filter() (Filter, error) {
	if e.choice == 0 {
		return nil, nil
	}
	return NewBoolFilter(e.name, e.choice == 1), nil
}

func (e *boolEditor) Focus() tea.Cmd { return nil }
func (e *boolEditor) Blur()          {}

// boolChoiceColors colour each choice by what it means: any is neutral,
// yes and no are the good and bad colours.
var boolChoiceColors = []color.Color{shared.AppColors.Blue, shared.AppColors.Green, shared.AppColors.Red}

func (e *boolEditor) View(focused bool, width int) []string {
	options := make([]string, len(boolChoices))
	for i, choice := range boolChoices {
		label := "  " + choice + "  "
		switch {
		case i == e.choice:
			options[i] = shared.PillStyle(boolChoiceColors[i]).Render(label)
		case focused:
			options[i] = shared.TextBodyStyle.Render(label)
		default:
			options[i] = shared.DimStyle.Render(label)
		}
	}
	return []string{strings.Join(options, " ")}
}

func (e *boolEditor) Keys() []key.Binding {
	return []key.Binding{e.keymap.Prev, e.keymap.Yes, e.keymap.No, e.keymap.Any}
}

// ---- ranges of numbers and dates

// rangeEditor edits a lower and upper bound, either of which may be blank.
type rangeEditor struct {
	labels [2]string
	inputs [2]textinput.Model
	active int
	build  func(lower, upper string) (Filter, error)
	next   key.Binding
}

func newRangeEditor(labels, placeholders, values [2]string, build func(lower, upper string) (Filter, error)) *rangeEditor {
	e := &rangeEditor{
		labels: labels,
		build:  build,
		next:   key.NewBinding(key.WithKeys("tab", "shift+tab", "up", "down"), key.WithHelp("tab", "next field")),
	}
	for i := range e.inputs {
		e.inputs[i] = newInput(placeholders[i], values[i])
	}
	return e
}

func (e *rangeEditor) Update(msg tea.Msg) tea.Cmd {
	if msg, ok := msg.(tea.KeyPressMsg); ok && key.Matches(msg, e.next) {
		e.inputs[e.active].Blur()
		e.active = 1 - e.active
		return e.inputs[e.active].Focus()
	}
	var cmd tea.Cmd
	e.inputs[e.active], cmd = e.inputs[e.active].Update(msg)
	return cmd
}

func (e *rangeEditor) Filter() (Filter, error) {
	return e.build(strings.TrimSpace(e.inputs[0].Value()), strings.TrimSpace(e.inputs[1].Value()))
}

func (e *rangeEditor) Focus() tea.Cmd { return e.inputs[e.active].Focus() }
func (e *rangeEditor) Blur()          { e.inputs[e.active].Blur() }

func (e *rangeEditor) View(focused bool, width int) []string {
	lines := make([]string, 0, 3)
	for i := range e.inputs {
		label := shared.TextBodyStyle.Render(shared.Fit(e.labels[i], 10))
		if focused && i == e.active {
			label = shared.AccentStyle.Render(shared.Fit(e.labels[i], 10))
		}
		e.inputs[i].SetWidth(shared.Max(1, width-11))
		lines = append(lines, label+e.inputs[i].View())
	}
	if _, err := e.Filter(); err != nil {
		lines = append(lines, shared.ErrorStyle.Render(err.Error()))
	}
	return lines
}

func (e *rangeEditor) Keys() []key.Binding { return []key.Binding{e.next} }

func newIntEditor(name string, current Filter) *rangeEditor {
	var values [2]string
	if f, ok := current.(IntFilter); ok {
		if f.From != NoMin {
			values[0] = strconv.Itoa(f.From)
		}
		if f.To != NoMax {
			values[1] = strconv.Itoa(f.To)
		}
	}
	return newRangeEditor([2]string{"at least", "at most"}, [2]string{"no minimum", "no maximum"}, values,
		func(lower, upper string) (Filter, error) {
			if lower == "" && upper == "" {
				return nil, nil
			}
			from, to := NoMin, NoMax
			var err error
			if lower != "" {
				if from, err = strconv.Atoi(lower); err != nil {
					return nil, fmt.Errorf("%q is not a whole number", lower)
				}
			}
			if upper != "" {
				if to, err = strconv.Atoi(upper); err != nil {
					return nil, fmt.Errorf("%q is not a whole number", upper)
				}
			}
			if from > to {
				return nil, errors.New("the minimum is above the maximum")
			}
			return NewIntFilter(name, from, to), nil
		})
}

// now is the clock relative dates are measured from; tests replace it.
var now = time.Now

var relativeDate = regexp.MustCompile(`^(\d+)\s*([dwmy])$`)

// parseDate reads a YYYY-MM-DD date, or a time ago such as "30d", "6w",
// "6m" or "2y".
func parseDate(s string) (time.Time, error) {
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}
	if m := relativeDate.FindStringSubmatch(strings.ToLower(s)); m != nil {
		n, _ := strconv.Atoi(m[1])
		today := now()
		switch m[2] {
		case "d":
			return today.AddDate(0, 0, -n), nil
		case "w":
			return today.AddDate(0, 0, -7*n), nil
		case "m":
			return today.AddDate(0, -n, 0), nil
		default:
			return today.AddDate(-n, 0, 0), nil
		}
	}
	return time.Time{}, fmt.Errorf("%q is not a date like 2024-06-30 or a time ago like 6m", s)
}

func newDateEditor(name string, current Filter) *rangeEditor {
	var values [2]string
	if f, ok := current.(DateFilter); ok {
		if !f.From.IsZero() {
			values[0] = f.From.Format("2006-01-02")
		}
		if !f.To.IsZero() {
			values[1] = f.To.Format("2006-01-02")
		}
	}
	placeholder := "YYYY-MM-DD, or 30d, 6m, 1y ago"
	return newRangeEditor([2]string{"after", "before"}, [2]string{placeholder, placeholder}, values,
		func(lower, upper string) (Filter, error) {
			if lower == "" && upper == "" {
				return nil, nil
			}
			var from, to time.Time
			var err error
			if lower != "" {
				if from, err = parseDate(lower); err != nil {
					return nil, err
				}
			}
			if upper != "" {
				if to, err = parseDate(upper); err != nil {
					return nil, err
				}
			}
			if !from.IsZero() && !to.IsZero() && from.After(to) {
				return nil, errors.New("the start is after the end")
			}
			return NewDateFilter(name, from, to), nil
		})
}

// ---- text

type textEditor struct {
	name  string
	input textinput.Model
}

func newTextEditor(name string, current Filter) *textEditor {
	value := ""
	if f, ok := current.(StringFilter); ok {
		value = f.Value()
	}
	return &textEditor{name: name, input: newInput("any text", value)}
}

func (e *textEditor) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	e.input, cmd = e.input.Update(msg)
	return cmd
}

func (e *textEditor) Filter() (Filter, error) {
	if strings.TrimSpace(e.input.Value()) == "" {
		return nil, nil
	}
	return NewStringFilter(e.name, e.input.Value()), nil
}

func (e *textEditor) Focus() tea.Cmd { return e.input.Focus() }
func (e *textEditor) Blur()          { e.input.Blur() }

func (e *textEditor) View(focused bool, width int) []string {
	label := shared.TextBodyStyle.Render(shared.Fit("contains", 10))
	if focused {
		label = shared.AccentStyle.Render(shared.Fit("contains", 10))
	}
	e.input.SetWidth(shared.Max(1, width-11))
	return []string{label + e.input.View()}
}

func (e *textEditor) Keys() []key.Binding { return nil }

// newInput builds a borderless text field in the browser's colours.
func newInput(placeholder, value string) textinput.Model {
	input := textinput.New()
	input.Prompt = ""
	input.Placeholder = placeholder
	input.CharLimit = 100
	input.SetValue(value)
	styles := textinput.DefaultStyles(true)
	styles.Focused.Text = shared.ValueStyle
	styles.Blurred.Text = shared.ValueStyle
	styles.Focused.Placeholder = shared.DimStyle
	styles.Blurred.Placeholder = shared.DimStyle
	styles.Cursor.Color = shared.AppColors.Cyan
	input.SetStyles(styles)
	return input
}
