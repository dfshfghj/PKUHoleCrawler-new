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
	PageConfig
	PageLogs
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

type Model struct {
	Page      Page
	Width     int
	Height    int
	TabCursor int

	LoggedIn     bool
	LoginUser    string
	CrawlerState CrawlerState
	CrawlerStart time.Time
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
	PostListPage    int
	PostListPerPage int
	PostListCursor  int
	PostListLoading bool
	PostListError   string

	ShowPostDetail bool
	CurrentPost    *models.Post
	CommentList    []models.Comment
	CommentCursor  int

	Searching    bool
	SearchInput  string
	SearchActive bool

	ConfigUsername  string
	ConfigPassword  string
	ConfigSecretKey string
	ConfigFieldIdx  int
	ConfigSaving    bool
	ConfigSaved     bool
	ConfigSaveOK    bool

	LogLines   []string
	LogOffset  int
	LogLoading bool

	LastError string
}

func NewModel(database *db.Database, client *client.Client, cfg *config.Config) Model {
	return Model{
		Page:            PageHome,
		TabCursor:       0,
		Database:        database,
		Client:          client,
		Config:          cfg,
		CrawlerState:    CrawlerStopped,
		PostListPerPage: 10,
		PostListPage:    1,
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
		tickCmd(),
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}
