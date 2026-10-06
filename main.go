package main

import (
	"errors"
	"fmt"
	"os"

	"gh-reponark/github"

	tea "charm.land/bubbletea/v2"
)

func main() {
	opts, err := parseArgs(os.Args[1:])
	if errors.Is(err, errHelp) {
		fmt.Println(usage)
		return
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if opts.version {
		printVersion()
		return
	}

	var svc github.Service
	if opts.demo {
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
