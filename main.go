package main

import (
	"fmt"
	"os"

	"gh-reponark/github"

	tea "charm.land/bubbletea/v2"
)

func main() {
	var svc github.Service
	if os.Getenv("REPONARK_DEMO") != "" {
		// Demo mode runs against built-in sample data instead of GitHub.
		svc = demoService()
	} else {
		client, err := github.NewClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "gh-reponark: %v\n", err)
			os.Exit(1)
		}
		svc = client
	}

	mainModel := NewMainModel(svc)
	// p := tea.NewProgram(mainModel, tea.WithKeyboardEnhancements())
	p := tea.NewProgram(mainModel)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
