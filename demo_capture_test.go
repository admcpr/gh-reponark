//go:build capture

package main

// A throwaway capture of the inspector for visual review. Run with:
//
//	CAPTURE_DIR=/path go test -tags capture -run Capture ./ -count=1
//
// It writes <WxH>-<screen>.ans (with ANSI colour) and .txt (plain) into the
// directory named by CAPTURE_DIR, or ./captures by default.

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

type capture struct {
	t      *testing.T
	m      MainModel
	dir    string
	prefix string
}

// drive runs a command and feeds every message it yields back into the
// model, the way the Bubble Tea runtime would, stopping at animation frames.
func (c *capture) drive(cmd tea.Cmd, depth int) {
	if cmd == nil || depth > 50 {
		return
	}
	// Timers such as cursor blinks and progress frames would re-arm forever;
	// anything that does not answer promptly is abandoned.
	done := make(chan tea.Msg, 1)
	go func() { done <- cmd() }()
	var msg tea.Msg
	select {
	case msg = <-done:
	case <-time.After(100 * time.Millisecond):
		return
	}
	switch msg := msg.(type) {
	case nil:
		return
	case tea.BatchMsg:
		for _, sub := range msg {
			c.drive(sub, depth+1)
		}
		return
	case progress.FrameMsg:
		return
	case tea.QuitMsg:
		return
	}
	c.send(msg)
}

// send runs a message, a key press included, through the model.
func (c *capture) send(msg tea.Msg) {
	updated, cmd := c.m.Update(msg)
	c.m = updated.(MainModel)
	c.drive(cmd, 0)
}

func (c *capture) press(keys ...string) {
	for _, k := range keys {
		switch k {
		case "enter":
			c.send(tea.KeyPressMsg{Code: tea.KeyEnter})
		case "esc":
			c.send(tea.KeyPressMsg{Code: tea.KeyEscape})
		case "tab":
			c.send(tea.KeyPressMsg{Code: tea.KeyTab})
		case "space":
			c.send(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
		default:
			for _, r := range k {
				c.send(tea.KeyPressMsg{Code: r, Text: string(r)})
			}
		}
	}
}

func (c *capture) snap(name string) {
	name = c.prefix + "-" + name
	content := fmt.Sprint(c.m.View().Content)
	if !strings.Contains(content, "\x1b[") {
		c.t.Fatalf("%s: no escape codes in the view", name)
	}
	for ext, body := range map[string]string{".ans": content, ".txt": ansi.Strip(content)} {
		path := filepath.Join(c.dir, name+ext)
		if err := os.WriteFile(path, []byte(body+"\n"), 0o644); err != nil {
			c.t.Fatal(err)
		}
	}
	c.t.Logf("wrote %s", filepath.Join(c.dir, name))
}

// repeat presses a key n times.
func (c *capture) repeat(k string, n int) {
	for i := 0; i < n; i++ {
		c.press(k)
	}
}

func TestCapture(t *testing.T) {
	dir := os.Getenv("CAPTURE_DIR")
	if dir == "" {
		dir = "captures"
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	for _, size := range [][2]int{{120, 36}, {160, 44}} {
		width, height := size[0], size[1]
		c := &capture{t: t, m: NewMainModel(demoService()), dir: dir, prefix: fmt.Sprintf("%dx%d", width, height)}
		c.send(tea.WindowSizeMsg{Width: width, Height: height})
		c.drive(c.m.Init(), 0)

		c.snap("picker")

		// Open the user's repositories and wait for them to load.
		c.press("enter")
		// aurora is the third repository alphabetically.
		c.press("j", "j")
		c.snap("list")

		c.press("?")
		c.snap("help")
		c.press("?")

		c.press("v")
		c.press("tab", "tab", "tab") // Features
		c.press("l", "l", "l")
		c.snap("matrix")
		c.press("v")

		c.press("f")
		c.snap("filters")
		// Toggle "Is Archived" to no: the chip, the segmented control and
		// the toggle chart. Then the star count's histogram.
		c.press("/", "archived", "enter")
		c.press("space", "space")
		c.snap("filters-toggle")
		c.press("/", "esc", "/", "stargazer", "enter") // esc clears the old search
		c.snap("filters-count")
		c.press("esc")

		c.press("enter")
		c.press("j", "j")
		c.snap("overview")

		c.press("tab", "tab") // Metrics
		c.press("j", "j")
		c.snap("metrics")

		c.press("tab") // Features
		c.press("j")
		c.snap("features")

		c.press("tab") // Merge
		c.repeat("j", 4)
		c.snap("merge")

		c.press("tab", "tab") // Security
		c.repeat("j", 3)
		c.snap("security")

		// legacy-billing is archived with open alerts and no license. It is
		// the seventeenth repository alphabetically.
		c.press("esc")
		c.press("g")
		c.repeat("j", 16)
		c.press("enter")
		c.press("tab") // Security is still the active tab; one more wraps to Overview
		c.snap("low-overview")
		c.repeat("tab", 6) // back round to Security
		c.press("j", "j")
		c.snap("low-security")

		// A sign-in failure lands on the error screen.
		failing := demoService()
		failing.UserErr = errors.New("fetching the signed-in user: GET https://api.github.com/user: 401 Bad credentials")
		e := &capture{t: t, m: NewMainModel(failing), dir: dir, prefix: c.prefix}
		e.send(tea.WindowSizeMsg{Width: width, Height: height})
		e.drive(e.m.Init(), 0)
		e.snap("error")
	}
}
