package screens

// Screen identifies which TUI screen is currently active.
type Screen int

const (
	ScreenMenu Screen = iota
	ScreenGenerate
	ScreenSummaries
	ScreenViewer
	ScreenRepos
	ScreenSettings
)

// NavigateMsg is sent to the root model to switch screens.
type NavigateMsg struct {
	To      Screen
	// Extra context payloads used by some transitions.
	RepoName string
	RepoDir  string
	FilePath string
}
