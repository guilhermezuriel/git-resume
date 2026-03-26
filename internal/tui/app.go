package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	gitpkg "github.com/guilhermezuriel/git-resume/internal/git"
	"github.com/guilhermezuriel/git-resume/internal/storage"
	"github.com/guilhermezuriel/git-resume/internal/tui/screens"
)

// App is the root Bubble Tea model. It owns the active screen model and
// handles all navigation messages.
type App struct {
	current tea.Model
	// Keep repo info for screens that need it.
	repoInfo *gitpkg.RepoInfo
}

// NewApp creates a new App rooted at the main menu.
func NewApp(repoInfo *gitpkg.RepoInfo) App {
	return App{
		repoInfo: repoInfo,
		current:  screens.NewMainMenu(repoInfo.Name),
	}
}

func (a App) Init() tea.Cmd {
	return a.current.Init()
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// ---- navigation ---------------------------------------------------------
	case screens.NavigateMsg:
		return a.navigate(msg)

	// ---- menu selection → map to navigation ---------------------------------
	case screens.MenuSelectedMsg:
		switch msg.Index {
		case 0: // Generate new summary
			flow := screens.NewGenerateFlow(a.repoInfo)
			a.current = flow
			return a, flow.Init()
		case 1: // View summaries
			repoDir := storage.RepoDirFor(a.repoInfo.ID)
			list := screens.NewSummaryList(a.repoInfo.Name, repoDir, screens.ScreenMenu)
			a.current = list
			return a, list.Init()
		case 2: // Browse all repositories
			browser := screens.NewRepoBrowser()
			a.current = browser
			return a, browser.Init()
		case 3: // Settings
			s := screens.NewSettings()
			a.current = s
			return a, s.Init()
		case 4: // Exit
			return a, tea.Quit
		}

	// ---- open a file in the viewer ------------------------------------------
	case screens.OpenFileMsg:
		v := screens.NewSummaryViewer(msg.RepoName, msg.Path)
		a.current = v
		return a, v.Init()

	// ---- open a repo's summary list from the browser ------------------------
	case screens.OpenRepoDirMsg:
		list := screens.NewSummaryList(msg.RepoName, msg.RepoDir, screens.ScreenRepos)
		a.current = list
		return a, list.Init()
	}

	// Delegate all other messages to the active screen.
	var cmd tea.Cmd
	a.current, cmd = a.current.Update(msg)
	return a, cmd
}

func (a App) View() string {
	return a.current.View()
}

func (a App) navigate(msg screens.NavigateMsg) (tea.Model, tea.Cmd) {
	switch msg.To {
	case screens.ScreenMenu:
		m := screens.NewMainMenu(a.repoInfo.Name)
		a.current = m
		return a, m.Init()
	case screens.ScreenGenerate:
		flow := screens.NewGenerateFlow(a.repoInfo)
		a.current = flow
		return a, flow.Init()
	case screens.ScreenSummaries:
		repoDir := storage.RepoDirFor(a.repoInfo.ID)
		list := screens.NewSummaryList(a.repoInfo.Name, repoDir, screens.ScreenMenu)
		a.current = list
		return a, list.Init()
	case screens.ScreenRepos:
		browser := screens.NewRepoBrowser()
		a.current = browser
		return a, browser.Init()
	case screens.ScreenSettings:
		s := screens.NewSettings()
		a.current = s
		return a, s.Init()
	case screens.ScreenViewer:
		v := screens.NewSummaryViewer(msg.RepoName, msg.FilePath)
		a.current = v
		return a, v.Init()
	}
	return a, nil
}

// Run starts the Bubble Tea program.
func Run(repoInfo *gitpkg.RepoInfo) error {
	app := NewApp(repoInfo)
	p := tea.NewProgram(app, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
