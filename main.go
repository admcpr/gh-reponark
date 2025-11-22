package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
)

func main() {
	mainModel := NewMainModel()
	// p := tea.NewProgram(mainModel, tea.WithKeyboardEnhancements())
	p := tea.NewProgram(mainModel)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
