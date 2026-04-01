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
	// Remember terminal size so new screens receive it immediately.
	width  int
	height int
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

	// ---- remember terminal size so new screens get it on creation -----------
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height

	// ---- navigation ---------------------------------------------------------
	case screens.NavigateMsg:
		return a.navigate(msg)

	// ---- menu selection → map to navigation ---------------------------------
	case screens.MenuSelectedMsg:
		switch msg.Index {
		case 0: // Generate new summary
			flow := screens.NewGenerateFlow(a.repoInfo)
			a.current, _ = a.initScreen(flow)
			return a, flow.Init()
		case 1: // View summaries
			repoDir := storage.RepoDirFor(a.repoInfo.ID)
			list := screens.NewSummaryList(a.repoInfo.Name, repoDir, screens.ScreenMenu)
			a.current, _ = a.initScreen(list)
			return a, list.Init()
		case 2: // Browse all repositories
			browser := screens.NewRepoBrowser()
			a.current, _ = a.initScreen(browser)
			return a, browser.Init()
		case 3: // Settings
			s := screens.NewSettings()
			a.current, _ = a.initScreen(s)
			return a, s.Init()
		case 4: // Exit
			return a, tea.Quit
		}

	// ---- open a file in the viewer ------------------------------------------
	case screens.OpenFileMsg:
		v := screens.NewSummaryViewer(msg.RepoName, msg.Path)
		a.current, _ = a.initScreen(v)
		return a, v.Init()

	// ---- open a repo's summary list from the browser ------------------------
	case screens.OpenRepoDirMsg:
		list := screens.NewSummaryList(msg.RepoName, msg.RepoDir, screens.ScreenRepos)
		a.current, _ = a.initScreen(list)
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

// initScreen sends a WindowSizeMsg to a newly created screen model so it can
// lay itself out correctly before the first real render.
func (a App) initScreen(m tea.Model) (tea.Model, tea.Cmd) {
	if a.width > 0 && a.height > 0 {
		m, _ = m.Update(tea.WindowSizeMsg{Width: a.width, Height: a.height})
	}
	return m, nil
}

func (a App) navigate(msg screens.NavigateMsg) (tea.Model, tea.Cmd) {
	switch msg.To {
	case screens.ScreenMenu:
		m := screens.NewMainMenu(a.repoInfo.Name)
		a.current, _ = a.initScreen(m)
		return a, m.Init()
	case screens.ScreenGenerate:
		flow := screens.NewGenerateFlow(a.repoInfo)
		a.current, _ = a.initScreen(flow)
		return a, flow.Init()
	case screens.ScreenSummaries:
		repoDir := storage.RepoDirFor(a.repoInfo.ID)
		list := screens.NewSummaryList(a.repoInfo.Name, repoDir, screens.ScreenMenu)
		a.current, _ = a.initScreen(list)
		return a, list.Init()
	case screens.ScreenRepos:
		browser := screens.NewRepoBrowser()
		a.current, _ = a.initScreen(browser)
		return a, browser.Init()
	case screens.ScreenSettings:
		s := screens.NewSettings()
		a.current, _ = a.initScreen(s)
		return a, s.Init()
	case screens.ScreenViewer:
		v := screens.NewSummaryViewer(msg.RepoName, msg.FilePath)
		a.current, _ = a.initScreen(v)
		return a, v.Init()
	}
	return a, nil
}

// Run starts the Bubble Tea program with alt-screen and mouse support.
func Run(repoInfo *gitpkg.RepoInfo) error {
	app := NewApp(repoInfo)
	p := tea.NewProgram(
		app,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	_, err := p.Run()
	return err
}
