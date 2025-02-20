package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea/v2"
)

func main() {
	mainModel := NewMainModel()
	// p := tea.NewProgram(mainModel, tea.WithKeyboardEnhancements(), tea.WithAltScreen())
	p := tea.NewProgram(mainModel, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
