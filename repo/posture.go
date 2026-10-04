package repo

import (
	"fmt"
	"image/color"

	"gh-reponark/shared"
)

// PostureState says whether a posture chip is fine, worth a look or wrong.
type PostureState int

const (
	PosturePass PostureState = iota
	PostureWarn
	PostureFail
)

// PostureChip is one short status for the chip row under the repo name.
type PostureChip struct {
	Label string
	State PostureState
}

// Posture summarises the security and upkeep of c as a row of chips. The
// labels state the finding, so a chip reads correctly without its colour.
func Posture(c RepoConfig) []PostureChip {
	// chip passes when ok, and otherwise takes the given failure state.
	chip := func(ok bool, pass, fail string, failure PostureState) PostureChip {
		if ok {
			return PostureChip{pass, PosturePass}
		}
		return PostureChip{fail, failure}
	}

	chips := []PostureChip{
		chip(c.Int("Branch Protection Rules") > 0 || c.Int("Rulesets") > 0, "Protected", "Unprotected", PostureFail),
		chip(c.Bool("Is Security Policy Enabled"), "Policy", "No policy", PostureFail),
	}

	switch alerts := c.Int("Vulnerability Alerts"); {
	case alerts > 0:
		chips = append(chips, PostureChip{fmt.Sprintf("%d alerts", alerts), PostureFail})
	case !c.Bool("Has Vulnerability Alerts Enabled"):
		chips = append(chips, PostureChip{"Alerts off", PostureWarn})
	default:
		chips = append(chips, PostureChip{"0 alerts", PosturePass})
	}

	chips = append(chips, chip(c.Text("License Name") != "", "Licensed", "No license", PostureWarn))

	switch pushed := c.Time("Pushed At"); {
	case c.Bool("Is Archived"):
		chips = append(chips, PostureChip{"Archived", PostureFail})
	case pushed.IsZero() || now().Sub(pushed) > staleAfter:
		chips = append(chips, PostureChip{"Stale", PostureWarn})
	default:
		chips = append(chips, PostureChip{"Active", PosturePass})
	}

	return append(chips, chip(c.Bool("Delete Branch On Merge"), "Cleanup", "No cleanup", PostureWarn))
}

// postureColors maps a chip's state to its capsule colour and tint.
func postureColors(state PostureState) (color.Color, color.Color) {
	switch state {
	case PostureFail:
		return shared.AppColors.Bad, shared.AppColors.BadTint
	case PostureWarn:
		return shared.AppColors.Warn, shared.AppColors.WarnTint
	default:
		return shared.AppColors.Good, shared.AppColors.GoodTint
	}
}

// PostureRows flows chips into exactly maxRows lines of width cells. Passing
// chips are quiet tinted capsules; warnings and failures are solid, so the
// problems are what the eye lands on. A chip is never cut: each row takes
// the chips that fit, in order, and whatever is left after the last row is
// dropped.
func PostureRows(chips []PostureChip, width, maxRows int) []string {
	pills := make([]string, len(chips))
	for i, chip := range chips {
		c, tint := postureColors(chip.State)
		if chip.State == PosturePass {
			pills[i] = shared.TintedPill(chip.Label, c, tint)
		} else {
			pills[i] = shared.Pill(chip.Label, c)
		}
	}
	var rows []string
	for len(rows) < maxRows {
		row, dropped := shared.JoinFit(pills, " ", width)
		rows = append(rows, row)
		pills = pills[len(pills)-dropped:]
	}
	return rows
}
