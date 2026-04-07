package tui

import (
	"time"

	"treehole/internal/client"
	"treehole/internal/config"
	"treehole/internal/db"
	"treehole/internal/models"

	"github.com/charmbracelet/bubbles/viewport"
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

type CrawlMsg struct {
	PostsCount    int
	CommentsCount int
	Page          int
	Duration      time.Duration
	Error         error
}

type TickMsg time.Time

type LoginMsg struct {
	Username string
	Error    error
}

type LoadPostsMsg struct {
	Posts []models.Post
	Total int
	Page  int
	Error error
}

type LoadCommentsMsg struct {
	Comments []models.Comment
	Error    error
}

type SearchPostsMsg struct {
	Posts []models.Post
	Total int
	Page  int
	Error error
}

type LoadStatsMsg struct {
	PostCount    int
	CommentCount int
	Error        error
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

	LoggedIn     bool
	LoginUser    string
	CrawlerState CrawlerState
	CrawlerStart time.Time
	CrawlMode    CrawlMode
	MonitorPages int
	Database     *db.Database
	Client       *client.Client
	Config       *config.Config

	TotalPosts    int
	TotalComments int
	LastCrawlPage int
	LastCrawlTime time.Duration

	HomeButtonIdx int
	HomeLastError string

	PostList        []models.Post
	PostListTotal   int
	PostListLoading bool
	PostListError   string
	PostPerPage     int
	PostViewport    *viewport.Model
	postContent     string
	SelectedPostIdx int

	ShowPostDetail  bool
	CurrentPost     *models.Post
	CommentList     []models.Comment
	CommentViewport *viewport.Model
	commentContent  string

	Searching      bool
	SearchInput    string
	SearchActive   bool
	SearchResults  []models.Post
	SearchTotal    int
	SearchViewport *viewport.Model

	ConfigUsername  string
	ConfigPassword  string
	ConfigSecretKey string
	ConfigFieldIdx  int
	ConfigSaving    bool
	ConfigSaveOK    bool

	LogLines   []string
	LogOffset  int
	LogLoading bool

	LastError string
}

func NewModel(database *db.Database, client *client.Client, cfg *config.Config) Model {
	pv := viewport.New(0, 0)
	cv := viewport.New(0, 0)
	return Model{
		Page:            PagePosts,
		TabCursor:       1,
		Dialog:          DialogNone,
		Database:        database,
		Client:          client,
		Config:          cfg,
		CrawlerState:    CrawlerStopped,
		CrawlMode:       CrawlSequential,
		MonitorPages:    3,
		PostPerPage:     20,
		PostViewport:    &pv,
		CommentViewport: &cv,
		ConfigUsername:  cfg.Username,
		ConfigPassword:  cfg.Password,
		ConfigSecretKey: cfg.SecretKey,
		LogOffset:       0,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			return LoginMsg{Username: m.Config.Username}
		},
		func() tea.Msg {
			pc, _ := m.Database.GetPostCount()
			cc, _ := m.Database.GetCommentCount()
			return LoadStatsMsg{PostCount: pc, CommentCount: cc}
		},
		func() tea.Msg {
			posts, err := m.Database.GetPosts(0, m.PostPerPage)
			if err != nil {
				return LoadPostsMsg{Error: err}
			}
			total, _ := m.Database.GetPostCount()
			return LoadPostsMsg{Posts: posts, Total: total, Page: 1}
		},
		tickCmd(),
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}
