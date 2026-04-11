package tui

import (
	"time"

	"treehole/internal/client"
	"treehole/internal/config"
	"treehole/internal/db"
	"treehole/internal/models"

	tea "github.com/charmbracelet/bubbletea"
)

type Page int

const (
	PageHome Page = iota
	PagePosts
)

type DialogType int

const (
	DialogNone DialogType = iota
	DialogConfig
	DialogLogs
	DialogHelp
)

type CrawlerState int

const (
	CrawlerStopped CrawlerState = iota
	CrawlerRunning
	CrawlerError
)

type HomeFocus int

const (
	HomeFocusStart HomeFocus = iota
	HomeFocusStop
	HomeFocusMode
)

type PostsMode int

const (
	PostsModeList PostsMode = iota
	PostsModeSearchInput
	PostsModeSearchResults
	PostsModeDetail
)

type DetailFocus int

const (
	DetailFocusPost DetailFocus = iota
	DetailFocusComments
)

type CrawlMsg struct {
	Page     int
	Duration time.Duration
	Error    error
}

type TickMsg time.Time

type LoginMsg struct {
	Username string
	Error    error
}

type LoadPostsMsg struct {
	Posts   []models.Post
	Cursor  int
	HasMore bool
	Error   error
}

type LoadCommentsMsg struct {
	Comments []models.Comment
	Cursor   int32
	HasMore  bool
	SortAsc  bool
	Error    error
}

type SearchPostsMsg struct {
	Posts   []models.Post
	Cursor  int
	HasMore bool
	Error   error
}

type LoadPostDetailMsg struct {
	Post     *models.Post
	Comments []models.Comment
	HasMore  bool
	SortAsc  bool
	Error    error
}

type LoadLogsMsg struct {
	Lines []string
	Error error
}

type LoadConfigMsg struct {
	Config *config.Config
	Error  error
}

type SaveConfigMsg struct {
	Error error
}

type CrawlMode int

const (
	CrawlSequential CrawlMode = iota
	CrawlMonitor
)

type Model struct {
	Page      Page
	Width     int
	Height    int
	TabCursor int

	Dialog DialogType

	Home     HomePageModel
	Database *db.Database
	Client   *client.Client
	Config   *config.Config

	Posts PostsPageModel

	ConfigDialog ConfigDialogModel
	LogsDialog   LogsDialogModel

	LastError string
	Capture   *CaptureSink
}

func NewModel(database *db.Database, client *client.Client, cfg *config.Config) Model {
	applyTheme("")

	return Model{
		Page:         PagePosts,
		TabCursor:    1,
		Dialog:       DialogNone,
		Home:         NewHomePageModel(),
		Database:     database,
		Client:       client,
		Config:       cfg,
		Posts:        NewPostsPageModel(),
		ConfigDialog: NewConfigDialog(cfg),
		LogsDialog:   NewLogsDialog(),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			return LoginMsg{Username: m.Config.Username}
		},
		loadPostsCmd(m.Database, 0, m.Posts.PostPerPage),
		tickCmd(),
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func (m *Model) ensureDialogModels() {
	if !m.ConfigDialog.initialized() {
		m.ConfigDialog = NewConfigDialog(m.Config)
	}
	if !m.LogsDialog.initialized() {
		m.LogsDialog = NewLogsDialog()
	}
	m.Posts.ensureInitialized()
}

func (m Model) calcPostViewportHeight() int {
	return m.Posts.calcPostViewportHeight(m.Height)
}
