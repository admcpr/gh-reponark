# A still of each main view of gh-reponark against its built-in demo data,
# for the README, taken with VHS (https://github.com/charmbracelet/vhs). The
# Demo workflow runs it with demo.tape and attaches the PNGs to the release;
# locally, put a `gh` that runs the built binary on the PATH and run
# `vhs screenshots.tape` in an empty directory.

Output screenshots.gif

Set Shell "bash"
Set FontSize 15
Set Width 1400
Set Height 820
Set Padding 12
Set Theme "Catppuccin Mocha"
Set TypingSpeed 0ms

# ttyd does not advertise truecolour; without this the palette is downsampled.
Env COLORTERM "truecolor"

Hide
Type "clear" Enter
Type "gh reponark --demo" Enter
Sleep 2s
Show

# The account picker: the signed-in user and their organisations.
Sleep 500ms
Screenshot view-accounts.png

# The repository list beside the inspector for aurora, the third repo, on
# the Security group.
Hide
Enter
Sleep 3s
Down Down
Tab@100ms 6
Sleep 500ms
Show
Sleep 300ms
Screenshot view-list.png

# The matrix on the Branch rules group, the next one along: every repo
# against every rule.
Hide
Type "v"
Tab
Sleep 500ms
Show
Sleep 300ms
Screenshot view-matrix.png

# The overview: the account as a grid of cards, with the repos behind the
# focused slice listed underneath.
Hide
Type "v"
Sleep 500ms
Show
Sleep 300ms
Screenshot view-overview.png

# The filters screen, on the Is Archived toggle with its chart.
Hide
Type "f"
Sleep 500ms
Type "/archived"
Enter
Sleep 300ms
Space Space
Sleep 500ms
Show
Sleep 300ms
Screenshot view-filters.png
