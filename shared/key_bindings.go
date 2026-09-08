package shared

import "charm.land/bubbles/v2/key"

// KeyBindings is an ordered set of bindings that satisfies help.KeyMap, so a
// screen can hand the help footer exactly the keys it handles.
type KeyBindings []key.Binding

func (k KeyBindings) ShortHelp() []key.Binding  { return k }
func (k KeyBindings) FullHelp() [][]key.Binding { return [][]key.Binding{k} }
