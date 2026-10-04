package user

import (
	"fmt"
	"hash/fnv"
	"image/color"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"gh-reponark/github"
	"gh-reponark/repo"
	"gh-reponark/shared"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// userLoadedMsg carries the authenticated user once the API call completes.
type userLoadedMsg github.User

// account is one card on the picker: the signed-in user or one of their
// organizations, reduced to what the card shows.
type account struct {
	login       string
	name        string
	description string
	url         string
	repos       int
	public      int
	members     int
	admin       bool
	verified    bool
	isUser      bool
	created     time.Time
}

// userAccount describes the signed-in user as a card.
func userAccount(u github.User) account {
	return account{
		login:       u.Login,
		name:        u.Name,
		description: u.Description,
		url:         u.Url,
		repos:       u.Repositories,
		public:      u.PublicRepositories,
		members:     u.Members,
		isUser:      true,
		created:     u.CreatedAt,
	}
}

// orgAccount describes an organization as a card.
func orgAccount(o github.Organization) account {
	return account{
		login:       o.Login,
		name:        o.Name,
		description: o.Description,
		url:         o.Url,
		repos:       o.Repositories,
		public:      o.PublicRepositories,
		members:     o.Members,
		admin:       o.ViewerCanAdminister,
		verified:    o.IsVerified,
		created:     o.CreatedAt,
	}
}

type Model struct {
	svc   github.Service
	login string
	// items are the accounts to pick from: the user first, then their
	// organizations. cursor is the selected one and offset the first card
	// scrolled into view.
	items  []account
	cursor int
	offset int
	width  int
	height int
	keymap userKeyMap
}

func NewModel(svc github.Service, width, height int) *Model {
	return &Model{svc: svc, width: width, height: height, keymap: newUserKeyMap()}
}

func (m *Model) SetDimensions(width, height int) {
	m.width = width
	m.height = height
	m.scrollToCursor()
}

func (m *Model) Init() tea.Cmd {
	return m.loadUser
}

// loadUser fetches the authenticated user and their organizations.
func (m *Model) loadUser() tea.Msg {
	user, err := m.svc.CurrentUser()
	if err != nil {
		return shared.ErrorMsg{Err: err}
	}
	return userLoadedMsg(user)
}

// SetUser populates the list with the user followed by their organizations.
func (m *Model) SetUser(user github.User) {
	m.login = user.Login
	items := make([]account, len(user.Organizations))
	for i, org := range user.Organizations {
		items[i] = orgAccount(org)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].login < items[j].login
	})

	// Add the user to the top of the list
	// They're not an organization but they also have repositories
	m.items = append([]account{userAccount(user)}, items...)
	m.moveCursor(0)
}

// selected returns the highlighted account, if there is one.
func (m *Model) selected() (account, bool) {
	if m.cursor < 0 || m.cursor >= len(m.items) {
		return account{}, false
	}
	return m.items[m.cursor], true
}

// moveCursor selects the account delta cards away, stopping at either end.
func (m *Model) moveCursor(delta int) {
	if len(m.items) == 0 {
		return
	}
	m.cursor = shared.Max(0, shared.Min(m.cursor+delta, len(m.items)-1))
	m.scrollToCursor()
}

func (m *Model) scrollToCursor() {
	m.offset = shared.ScrollOffset(m.offset, m.cursor, m.rows(), len(m.items))
}

// pickerChrome is the heading line above the cards.
const pickerChrome = 1

// Layout thresholds. Below shortHeight the cards shrink to two lines and
// below narrowWidth they lose the monogram and the URL line.
const (
	shortHeight = 16
	narrowWidth = 50
)

// short is true when the pane is too low for four-line cards.
func (m Model) short() bool { return m.height < shortHeight }

// narrow is true when the pane is too narrow for the monogram and URL.
func (m Model) narrow() bool { return m.width < narrowWidth }

// cardLines is how many lines one card takes at the current size.
func (m Model) cardLines() int {
	switch {
	case m.short():
		return 2
	case m.narrow():
		return 3
	default:
		return 4
	}
}

// rows is how many whole cards are shown at once. Cards are separated by a
// blank line, which the last card on screen does not need.
func (m *Model) rows() int {
	return shared.Max(1, (m.height-pickerChrome+1)/(m.cardLines()+1))
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case userLoadedMsg:
		m.SetUser(github.User(msg))

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keymap.Select):
			item, ok := m.selected()
			if !ok {
				return m, nil
			}
			return m, func() tea.Msg {
				orgKey := shared.OrgKey{
					Name:   item.login,
					IsUser: item.isUser,
				}
				return shared.OpenOrgMsg{Key: orgKey}
			}
		case key.Matches(msg, m.keymap.Back):
			return m, func() tea.Msg { return shared.PreviousMsg{} }
		case key.Matches(msg, m.keymap.Up):
			m.moveCursor(-1)
		case key.Matches(msg, m.keymap.Down):
			m.moveCursor(1)
		case key.Matches(msg, m.keymap.PageUp):
			m.moveCursor(-m.rows())
		case key.Matches(msg, m.keymap.PageDown):
			m.moveCursor(m.rows())
		case key.Matches(msg, m.keymap.Top):
			m.moveCursor(-len(m.items))
		case key.Matches(msg, m.keymap.Bottom):
			m.moveCursor(len(m.items))
		}
	}
	return m, nil
}

func (m Model) View() tea.View {
	width, height := shared.Max(1, m.width), shared.Max(1, m.height)
	if m.login == "" {
		// Nothing to pick from yet: say so in the middle of the pane.
		return tea.NewView(lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center,
			shared.DimStyle.Render("Signing in…")))
	}

	lines := []string{m.heading(width)}
	// The pane may have been resized since the cursor last moved.
	rows := m.rows()
	offset := shared.ScrollOffset(m.offset, m.cursor, rows, len(m.items))
	for i := offset; i < len(m.items) && i < offset+rows; i++ {
		if i > offset {
			lines = append(lines, "")
		}
		lines = append(lines, m.card(m.items[i], i == m.cursor, width)...)
	}
	return tea.NewView(shared.Lines(lines, width, height))
}

// heading is the one-line strip above the cards: the column heading and a
// chip counting the organizations.
func (m Model) heading(width int) string {
	count := len(m.items) - 1
	var chip string
	switch count {
	case 0:
		chip = shared.GhostPill("no organisations")
	case 1:
		chip = shared.TintedPill("1 organisation", shared.AppColors.Accent, shared.AppColors.AccentTint)
	default:
		chip = shared.TintedPill(fmt.Sprintf("%d organisations", count), shared.AppColors.Accent, shared.AppColors.AccentTint)
	}
	return shared.Fit(shared.ColumnHeading.Render("  ACCOUNTS")+"  "+chip, width)
}

// card renders one account. The selected card carries the accent marker on
// every line and sits on the selection band; the others are quieter.
func (m Model) card(a account, selected bool, width int) []string {
	marker := "  "
	if selected {
		marker = shared.AccentStyle.Render("▌ ")
	}
	// Text lines start under the name, which sits after the monogram.
	indent := marker
	if !m.narrow() {
		indent += "   "
	}
	inner := width - lipgloss.Width(indent)

	lines := []string{marker + m.nameLine(a, selected, width-lipgloss.Width(marker))}
	if !m.short() {
		lines = append(lines, indent+m.descriptionLine(a, inner))
	}
	lines = append(lines, indent+m.factsLine(a, inner))
	if !m.short() && !m.narrow() {
		url := shared.DimStyle
		if selected {
			url = shared.LinkTextStyle
		}
		lines = append(lines, indent+url.Render(shared.Fit(a.url, inner)))
	}

	for i, line := range lines {
		if selected {
			lines[i] = shared.HighlightRow(line, width, true)
		} else {
			lines[i] = shared.Fit(line, width)
		}
	}
	return lines
}

// nameLine is the card's first line: the monogram, the display name with
// the login beside it, and the account's pills.
func (m Model) nameLine(a account, selected bool, width int) string {
	var pills []string
	if a.isUser {
		pills = append(pills, shared.Pill("you", shared.AppColors.Accent))
	} else {
		pills = append(pills, shared.GhostPill("org"))
	}
	if a.admin {
		pills = append(pills, shared.TintedPill("admin", shared.AppColors.Good, shared.AppColors.GoodTint))
	}
	if a.verified {
		pills = append(pills, shared.TintedPill("verified", shared.AppColors.Link, verifiedTint))
	}
	tags := strings.Join(pills, " ")

	prefix := ""
	if !m.narrow() {
		prefix = monogram(a.login) + " "
	}

	name, login := a.name, ""
	if name == "" {
		name = a.login
	} else if a.name != a.login {
		login = a.login
	}
	nameStyle := shared.TextBodyStyle.Bold(true)
	if selected {
		nameStyle = shared.StrongStyle
	}
	title := nameStyle.Render(name)
	if login != "" {
		title += "  " + shared.DimStyle.Render(login)
	}

	// The pills are never cut; the title gives way first, losing the login
	// whole before the name is truncated.
	room := width - lipgloss.Width(prefix) - lipgloss.Width(tags) - 2
	if room < 1 {
		return shared.Fit(prefix+title, width)
	}
	if lipgloss.Width(title) > room && login != "" {
		title = nameStyle.Render(name)
	}
	if lipgloss.Width(title) > room {
		title = shared.Fit(title, room)
	}
	return prefix + title + "  " + tags
}

// descriptionLine is the bio or description on one line, or a quiet note
// that there is none.
func (m Model) descriptionLine(a account, width int) string {
	if strings.TrimSpace(a.description) == "" {
		return shared.DimStyle.Italic(true).Render("No description")
	}
	return shared.TextBodyStyle.Render(shared.Fit(strings.Join(strings.Fields(a.description), " "), width))
}

// factsLine is the chips line: counts in the foreground colour with what
// they count in dim, shed from the end when the pane is narrow.
func (m Model) factsLine(a account, width int) string {
	fact := func(n int, word string) string {
		return shared.TextBodyStyle.Render(repo.CompactNumber(n)) + shared.DimStyle.Render(" "+word)
	}
	facts := []string{fact(a.repos, plural(a.repos, "repo", "repos"))}
	if a.repos > 0 {
		facts = append(facts, fact(a.public, "public"))
		if private := a.repos - a.public; private > 0 {
			facts = append(facts, fact(private, "private"))
		}
	}
	if a.members > 0 {
		word := plural(a.members, "member", "members")
		if a.isUser {
			word = plural(a.members, "follower", "followers")
		}
		facts = append(facts, fact(a.members, word))
	}
	if !a.created.IsZero() {
		facts = append(facts, shared.DimStyle.Render("since ")+shared.TextBodyStyle.Render(fmt.Sprint(a.created.Year())))
	}
	line, _ := shared.JoinFit(facts, shared.DimStyle.Render(" · "), width)
	return line
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

// verifiedTint is the dark version of the link colour, the background of
// the verified pill.
var verifiedTint = lipgloss.Darken(shared.AppColors.Link, 0.7)

// monogramPalette is the set of block colours a monogram is drawn in.
var monogramPalette = []color.Color{
	shared.AppColors.Accent,
	shared.AppColors.Good,
	shared.AppColors.Warn,
	shared.AppColors.Pink,
	shared.AppColors.Purple,
	shared.AppColors.Blue,
	shared.AppColors.BrightYellow,
}

// monogram is a two-cell block with the account's initial on a colour
// chosen by hashing the login, so an account keeps its colour between runs.
func monogram(login string) string {
	initial := "?"
	if r, _ := utf8.DecodeRuneInString(login); r != utf8.RuneError && r != 0 {
		initial = string(unicode.ToUpper(r))
	}
	h := fnv.New32a()
	h.Write([]byte(login))
	c := monogramPalette[int(h.Sum32()%uint32(len(monogramPalette)))]
	return lipgloss.NewStyle().Foreground(shared.AppColors.Background).Background(c).Bold(true).Render(initial + " ")
}

// Status says who is signed in. The screen adds nothing to the breadcrumb:
// it is the root the others open from.
func (m Model) Status() string {
	if m.login == "" {
		return "signing in"
	}
	return "signed in as " + m.login
}

// Help lists the picker's keys, including the paging keys.
func (m Model) Help() shared.Help {
	pages := shared.Combine("←/→", "page", m.keymap.PageUp, m.keymap.PageDown)
	ends := shared.Combine("g/G", "first/last", m.keymap.Top, m.keymap.Bottom)
	return shared.Help{
		Short: []key.Binding{
			shared.Combine("j/k", "org", m.keymap.Down, m.keymap.Up),
			m.keymap.Select,
			m.keymap.Back,
		},
		Full: [][]key.Binding{
			{m.keymap.Up, m.keymap.Down, pages, ends},
			{m.keymap.Select, m.keymap.Back},
		},
	}
}

// userKeyMap holds the bindings for the organization picker. Update matches
// against these and the help footer renders them.
type userKeyMap struct {
	Up       key.Binding
	Down     key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Top      key.Binding
	Bottom   key.Binding
	Select   key.Binding
	Back     key.Binding
}

func newUserKeyMap() userKeyMap {
	return userKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		// The paging keys are the ones the list component used to answer to.
		PageUp: key.NewBinding(
			key.WithKeys("left", "h", "pgup", "b", "u"),
			key.WithHelp("←/h/pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("right", "l", "pgdown", "f", "d"),
			key.WithHelp("→/l/pgdn", "page down"),
		),
		Top: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("g/home", "first"),
		),
		Bottom: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("G/end", "last"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "open"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "quit"),
		),
	}
}
