package accounts

import (
	"hash/fnv"
	"image/color"
	"time"
	"unicode"
	"unicode/utf8"

	"gh-reponark/github"
	"gh-reponark/ui"

	"charm.land/lipgloss/v2"
)

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

// monogramPalette is the set of block colours a monogram is drawn in.
var monogramPalette = []color.Color{
	ui.AppColors.Accent,
	ui.AppColors.Good,
	ui.AppColors.Warn,
	ui.AppColors.Pink,
	ui.AppColors.Purple,
	ui.AppColors.Blue,
	ui.AppColors.BrightYellow,
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
	return lipgloss.NewStyle().Foreground(ui.AppColors.Background).Background(c).Bold(true).Render(initial + " ")
}
