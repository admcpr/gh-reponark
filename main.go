package main

import (
	"fmt"
	"os"

	"gh-reponark/github"

	tea "charm.land/bubbletea/v2"
)

func main() {
	client, err := github.NewClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "gh-reponark: %v\n", err)
		os.Exit(1)
	}

	mainModel := NewMainModel(client)
	// p := tea.NewProgram(mainModel, tea.WithKeyboardEnhancements())
	p := tea.NewProgram(mainModel)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
