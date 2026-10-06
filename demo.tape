# A tour of gh-reponark against its built-in demo data, recorded with VHS
# (https://github.com/charmbracelet/vhs). The release workflow records it and
# attaches demo.gif to the release; locally, put a `gh` that runs the built
# binary on the PATH and run `vhs demo.tape`.

Output demo.gif

Set Shell "bash"
Set FontSize 15
Set Width 1400
Set Height 820
Set Padding 12
Set Theme "Catppuccin Mocha"
Set TypingSpeed 40ms
Set Framerate 30
Set PlaybackSpeed 1.0

# ttyd does not advertise truecolour; without this the palette is downsampled.
Env COLORTERM "truecolor"

Hide
Type "clear" Enter
Show

# Sign in and pick an account.
Type "gh reponark --demo" Sleep 500ms Enter
Sleep 2s
Down Sleep 400ms
Down Sleep 800ms
Up Sleep 400ms
Up Sleep 600ms

# Open the demo user's repositories; the loading screen is brief.
Enter
Sleep 2.5s

# Browse the list, then step into the inspector.
Down Sleep 300ms
Down Sleep 800ms
Enter
Sleep 1.5s
Tab Sleep 900ms
Tab Sleep 900ms
Down Sleep 300ms
Down Sleep 300ms
Down Sleep 1.2s
Tab Sleep 1.2s

# The matrix view: every repo against one group of settings.
Type "v"
Sleep 1.5s
Right Sleep 400ms
Right Sleep 400ms
Down Sleep 300ms
Down Sleep 1.2s
Type "v"
Sleep 800ms

# Filter down to the archived repositories.
Type "f"
Sleep 1.2s
Type "/"
Type "archived" Sleep 600ms
Enter
Sleep 800ms
Space Sleep 900ms
Space Sleep 1.2s
Escape
Sleep 2.5s
