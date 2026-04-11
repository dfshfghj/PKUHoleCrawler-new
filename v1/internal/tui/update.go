package tui

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"time"

	"treehole/internal/client"
	"treehole/internal/config"
	"treehole/internal/crawler"
	"treehole/internal/db"
	"treehole/internal/models"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pquerna/otp/totp"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.ensureDialogModels()

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.syncPostsPage()
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case TickMsg:
		return m, tickCmd()

	case LoginMsg:
		if msg.Error != nil {
			m.LastError = msg.Error.Error()
			m.Home.LoggedIn = false
		} else {
			m.Home.LoggedIn = true
			m.Home.LoginUser = msg.Username
		}
		return m, nil

	case CrawlMsg:
		if msg.Error != nil {
			log.Printf("[Crawler] 爬虫错误: 第 %d 页, %v", msg.Page, msg.Error)
			m.Home.CrawlerState = CrawlerError
			m.Home.HomeLastError = msg.Error.Error()
		} else {
			m.Home.LastCrawlPage = msg.Page
			m.Home.LastCrawlTime = msg.Duration
			if m.Home.CrawlerState == CrawlerRunning {
				if m.Home.CrawlMode == CrawlMonitor {
					return m, crawlMonitorCmd(m.Client, m.Database, m.Home.MonitorPages)
				}
				return m, crawlPageCmd(m.Client, m.Database, msg.Page+1)
			} else {
				log.Printf("[Crawler] 爬虫已停止，最终抓取到第 %d 页", msg.Page)
			}
		}
		return m, nil

	case LoadPostsMsg:
		m.Posts.PostListLoading = false
		if msg.Error != nil {
			log.Printf("[Posts] 加载帖子列表失败: %v", msg.Error)
			m.Posts.PostListError = msg.Error.Error()
		} else {
			log.Printf("[Posts] 加载 %d 条帖子", len(msg.Posts))
			m.Posts.PostListError = ""
			if !m.Posts.SearchActive {
				if msg.Cursor == 0 {
					m.Posts.PostList = msg.Posts
					m.Posts.SelectedPostIdx = 0
					m.Posts.CursorLine = 0
					m.Posts.PostViewport.GotoTop()
				} else {
					m.Posts.PostList = append(m.Posts.PostList, msg.Posts...)
				}
				m.Posts.PostsMode = PostsModeList
			}
			m.Posts.PostListTotal = len(m.Posts.PostList)
			m.Posts.PostListCursor = nextPostCursor(msg.Posts)
			m.Posts.PostListHasMore = msg.HasMore
		}
		m.syncPostsPage()
		return m, nil

	case LoadCommentsMsg:
		m.Posts.CommentListLoading = false
		if msg.Error != nil {
			log.Printf("[Posts] 加载评论失败: %v", msg.Error)
			m.Posts.CommentListError = msg.Error.Error()
		} else {
			log.Printf("[Posts] 加载 %d 条评论", len(msg.Comments))
			m.Posts.CommentListError = ""
			if msg.Cursor == 0 {
				m.Posts.CommentList = msg.Comments
				m.Posts.CommentViewport.GotoTop()
			} else {
				m.Posts.CommentList = append(m.Posts.CommentList, msg.Comments...)
			}
			m.Posts.CommentListCursor = nextCommentCursor(msg.Comments)
			m.Posts.CommentListHasMore = msg.HasMore
			m.Posts.CommentSortAsc = msg.SortAsc
		}
		m.syncPostsPage()
		return m, nil

	case SearchPostsMsg:
		m.Posts.PostListLoading = false
		if msg.Error != nil {
			m.Posts.PostListError = msg.Error.Error()
		} else {
			log.Printf("[Posts] 搜索加载 %d 条帖子", len(msg.Posts))
			m.Posts.PostListError = ""
			if msg.Cursor == 0 {
				m.Posts.PostList = msg.Posts
				m.Posts.SelectedPostIdx = 0
				m.Posts.CursorLine = 0
				m.Posts.PostViewport.GotoTop()
			} else if m.Posts.SearchActive {
				m.Posts.PostList = append(m.Posts.PostList, msg.Posts...)
			}
			m.Posts.PostListTotal = len(m.Posts.PostList)
			m.Posts.PostListCursor = nextPostCursor(msg.Posts)
			m.Posts.PostListHasMore = msg.HasMore
			m.Posts.SearchActive = true
			m.Posts.Searching = false
			m.Posts.PostsMode = PostsModeSearchResults
		}
		m.syncPostsPage()
		return m, nil

	case LoadPostDetailMsg:
		m.Posts.CommentListLoading = false
		if msg.Error != nil {
			m.Posts.CommentListError = msg.Error.Error()
		} else {
			m.Posts.CommentListError = ""
			m.Posts.CurrentPost = msg.Post
			m.Posts.CommentList = msg.Comments
			m.Posts.CommentListCursor = nextCommentCursor(msg.Comments)
			m.Posts.CommentListHasMore = msg.HasMore
			m.Posts.CommentSortAsc = msg.SortAsc
			m.Posts.commentContent = ""
			m.Posts.postBodyContent = ""
			m.Posts.PostBodyViewport.GotoTop()
			m.Posts.CommentViewport.GotoTop()
		}
		m.syncPostsPage()
		return m, nil

	case LoadLogsMsg:
		if msg.Error != nil {
			m.LogsDialog.SetError(msg.Error)
		} else {
			m.LogsDialog.SetLines(msg.Lines)
		}
		return m, nil

	case LoadConfigMsg:
		if msg.Error == nil && msg.Config != nil {
			m.Config = msg.Config
			m.ConfigDialog.SetConfig(msg.Config)
		}
		return m, nil

	case SaveConfigMsg:
		if msg.Error != nil {
			m.LastError = msg.Error.Error()
			m.ConfigDialog.SetSaveResult(msg.Error)
		} else {
			m.ConfigDialog.SetSaveResult(nil)
		}
		return m, nil
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	// Dialog close
	if msg.String() == "esc" && m.Dialog != DialogNone {
		m.Dialog = DialogNone
		return m, nil
	}

	if msg.String() == "q" && m.Dialog == DialogNone && !m.Posts.Searching {
		return m, tea.Quit
	}

	// Open dialogs
	if m.Dialog == DialogNone && !m.Posts.Searching && !m.Posts.ShowPostDetail {
		if msg.String() == "c" {
			m.Dialog = DialogConfig
			m.ConfigDialog = NewConfigDialog(m.Config)
			return m, loadConfigCmd()
		}
		if msg.String() == "l" {
			m.Dialog = DialogLogs
			m.LogsDialog.SetLoading(true)
			return m, loadLogsCmd()
		}
		if msg.String() == "h" {
			m.Dialog = DialogHelp
			return m, nil
		}
	}

	if msg.String() == "tab" && m.Dialog == DialogNone && !m.Posts.Searching && !m.Posts.ShowPostDetail {
		m.TabCursor = (m.TabCursor + 1) % 2
		m.Page = Page(m.TabCursor)
		if m.Page == PagePosts && len(m.Posts.PostList) == 0 {
			m.Posts.PostListLoading = true
			return m, loadPostsCmd(m.Database, 0, m.Posts.PostPerPage)
		}
		m.syncPostsPage()
		return m, nil
	}

	// Route to page handlers when no dialog
	if m.Dialog == DialogNone {
		switch m.Page {
		case PageHome:
			return m.handleHomeKey(msg)
		case PagePosts:
			return m.handlePostsKey(msg)
		}
	} else {
		// Dialog handlers
		switch m.Dialog {
		case DialogConfig:
			return m.handleConfigKey(msg)
		case DialogLogs:
			return m.handleLogsKey(msg)
		case DialogHelp:
			return m, nil
		}
	}

	return m, nil
}

func (m Model) handleHomeKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	action := m.Home.Update(msg)
	switch action {
	case HomeActionStartCrawler:
		log.Printf("[Crawler] 爬虫已启动, 模式: %v", m.Home.CrawlMode)
		if m.Home.CrawlMode == CrawlMonitor {
			return m, crawlMonitorCmd(m.Client, m.Database, m.Home.MonitorPages)
		}
		return m, crawlPageCmd(m.Client, m.Database, 1)
	case HomeActionStopCrawler:
		log.Printf("[Crawler] 爬虫已手动停止")
	}
	return m, nil
}

func (m Model) handlePostsKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	if m.Posts.Searching {
		switch msg.Type {
		case tea.KeyEscape:
			return m.cancelSearchInput()
		case tea.KeyEnter:
			if m.Posts.SearchInput != "" {
				m.Posts.PostListLoading = true
				m.Posts.PostsMode = PostsModeSearchInput
				return m, searchPostsCmd(m.Database, m.Posts.SearchInput, 0, m.Posts.PostPerPage)
			}
			return m, nil
		case tea.KeyBackspace:
			if len(m.Posts.SearchInput) > 0 {
				m.Posts.SearchInput = m.Posts.SearchInput[:len(m.Posts.SearchInput)-1]
			}
			m.syncPostsPage()
			return m, nil
		default:
			if msg.Type == tea.KeyRunes {
				m.Posts.SearchInput += msg.String()
			}
			m.syncPostsPage()
			return m, nil
		}
	}

	if m.Posts.ShowPostDetail {
		switch msg.String() {
		case "esc":
			m.Posts.ShowPostDetail = false
			m.Posts.CurrentPost = nil
			m.Posts.postBodyContent = ""
			m.Posts.resetComments()
			m.Posts.commentContent = ""
			m.Posts.PostBodyViewport.GotoTop()
			m.Posts.DetailFocus = DetailFocusComments
			if m.Posts.SearchActive {
				m.Posts.PostsMode = PostsModeSearchResults
			} else {
				m.Posts.PostsMode = PostsModeList
			}
			m.syncPostsPage()
			return m, nil
		case "tab":
			if m.Posts.DetailFocus == DetailFocusPost {
				m.Posts.DetailFocus = DetailFocusComments
			} else {
				m.Posts.DetailFocus = DetailFocusPost
			}
		case "s":
			if m.Posts.CurrentPost != nil {
				nextSortAsc := !m.Posts.CommentSortAsc
				m.Posts.resetComments()
				m.Posts.CommentListLoading = true
				return m, loadCommentsCmd(m.Database, m.Posts.CurrentPost.Pid, nextSortAsc, 0)
			}
		case "r":
			if m.Posts.CurrentPost != nil {
				m.Posts.CommentListLoading = true
				m.Posts.CommentListError = ""
				return m, loadPostDetailCmd(m.Database, m.Posts.CurrentPost.Pid, m.Posts.CommentSortAsc)
			}
		case "up":
			if m.Posts.DetailFocus == DetailFocusPost {
				m.Posts.PostBodyViewport.ScrollUp(1)
			} else {
				m.Posts.CommentViewport.ScrollUp(1)
			}
		case "down":
			if m.Posts.DetailFocus == DetailFocusPost {
				m.Posts.PostBodyViewport.ScrollDown(1)
			} else {
				m.Posts.CommentViewport.ScrollDown(1)
				if m.Posts.CurrentPost != nil && m.Posts.shouldPrefetchCommentsMore() {
					m.Posts.CommentListLoading = true
					return m, loadCommentsCmd(m.Database, m.Posts.CurrentPost.Pid, m.Posts.CommentSortAsc, m.Posts.CommentListCursor)
				}
			}
		case "pgup":
			if m.Posts.DetailFocus == DetailFocusPost {
				m.Posts.PostBodyViewport.PageUp()
			} else {
				m.Posts.CommentViewport.PageUp()
			}
		case "pgdown":
			if m.Posts.DetailFocus == DetailFocusPost {
				m.Posts.PostBodyViewport.PageDown()
			} else {
				m.Posts.CommentViewport.PageDown()
				if m.Posts.CurrentPost != nil && m.Posts.shouldPrefetchCommentsMore() {
					m.Posts.CommentListLoading = true
					return m, loadCommentsCmd(m.Database, m.Posts.CurrentPost.Pid, m.Posts.CommentSortAsc, m.Posts.CommentListCursor)
				}
			}
		}
		m.syncPostsPage()
		return m, nil
	}

	switch msg.String() {
	case "esc":
		if m.Posts.SearchActive {
			return m.clearSearchResults()
		}
	case "r":
		if !m.Posts.SearchActive {
			m.Posts.PostListLoading = true
			m.Posts.resetList()
			return m, loadPostsCmd(m.Database, 0, m.Posts.PostPerPage)
		}
		return m, nil
	case "/":
		m.Posts.Searching = true
		m.Posts.PostsMode = PostsModeSearchInput
		m.Posts.SearchInput = ""
		return m, nil
	case "up":
		m.Posts.moveCursor(-1)
	case "down":
		m.Posts.moveCursor(1)
		if m.Posts.shouldPrefetchMore() {
			m.Posts.PostListLoading = true
			if m.Posts.SearchActive {
				return m, searchPostsCmd(m.Database, m.Posts.SearchInput, m.Posts.PostListCursor, m.Posts.PostPerPage)
			}
			return m, loadPostsCmd(m.Database, m.Posts.PostListCursor, m.Posts.PostPerPage)
		}
	case "enter":
		if len(m.Posts.PostList) > 0 && m.Posts.SelectedPostIdx < len(m.Posts.PostList) {
			post := m.Posts.PostList[m.Posts.SelectedPostIdx]
			m.Posts.ShowPostDetail = true
			m.Posts.PostsMode = PostsModeDetail
			m.Posts.CurrentPost = &post
			m.Posts.resetComments()
			m.Posts.CommentListLoading = true
			m.Posts.PostBodyViewport.GotoTop()
			m.Posts.DetailFocus = DetailFocusComments
			m.syncPostsPage()
			return m, loadCommentsCmd(m.Database, post.Pid, true, 0)
		}
	case "pgup":
		m.Posts.pageMove(-1)
	case "pgdown":
		m.Posts.pageMove(1)
		if m.Posts.shouldPrefetchMore() &&
			m.Posts.PostListHasMore &&
			!m.Posts.PostListLoading {
			m.Posts.PostListLoading = true
			if m.Posts.SearchActive {
				return m, searchPostsCmd(m.Database, m.Posts.SearchInput, m.Posts.PostListCursor, m.Posts.PostPerPage)
			}
			return m, loadPostsCmd(m.Database, m.Posts.PostListCursor, m.Posts.PostPerPage)
		}
	}
	m.syncPostsPage()
	return m, nil
}

func (m Model) cancelSearchInput() (Model, tea.Cmd) {
	m.Posts.Searching = false
	m.Posts.SearchInput = ""
	if m.Posts.SearchActive {
		m.Posts.PostsMode = PostsModeSearchResults
	} else {
		m.Posts.PostsMode = PostsModeList
	}
	m.syncPostsPage()
	return m, nil
}

func (m Model) clearSearchResults() (Model, tea.Cmd) {
	m.Posts.SearchActive = false
	m.Posts.Searching = false
	m.Posts.SearchInput = ""
	m.Posts.PostListLoading = true
	m.Posts.resetList()
	m.syncPostsPage()
	return m, loadPostsCmd(m.Database, 0, m.Posts.PostPerPage)
}

func (m *Model) syncPostsPage() {
	m.Posts.syncViewports(m.Width, m.contentAreaHeightForSize(m.Width, m.Height))
}

func (m Model) handleConfigKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	if msg.Type == tea.KeyEscape {
		m.Dialog = DialogNone
		return m, nil
	}
	if msg.Type == tea.KeyEnter && m.ConfigDialog.FocusIndex() == configSaveButtonIndex {
		m.ConfigDialog.SetSaving(true)
		return m, saveConfigCmd(m.ConfigDialog.ToConfig())
	}
	cmd := m.ConfigDialog.Update(msg)
	return m, cmd
}

func (m Model) handleLogsKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	if msg.Type == tea.KeyEscape {
		m.Dialog = DialogNone
		return m, nil
	}
	cmd := m.LogsDialog.Update(msg)
	return m, cmd
}

func crawlPageCmd(c *client.Client, database *db.Database, page int) tea.Cmd {
	return func() tea.Msg {
		log.Printf("[Crawler] 开始抓取第 %d 页", page)
		startTime := time.Now()
		result, err := crawler.FetchAndSave(c, database, page, false, 200, 200, false, false)
		duration := time.Since(startTime)

		if err != nil {
			log.Printf("[Crawler] 第 %d 页抓取失败: %v (耗时 %v)", page, err, duration)
			return CrawlMsg{Error: err, Page: page}
		}

		log.Printf("[Crawler] 第 %d 页抓取完成: %d 条帖子, %d 条评论 (耗时 %v)", page, result.PostCount, result.CommentCount, duration)

		return CrawlMsg{
			Page:     page,
			Duration: duration,
		}
	}
}

func crawlMonitorCmd(c *client.Client, database *db.Database, monitorPages int) tea.Cmd {
	return func() tea.Msg {
		startTime := time.Now()
		totalPosts := 0
		totalComments := 0

		for page := 1; page <= monitorPages; page++ {
			result, err := crawler.FetchAndSave(c, database, page, false, 200, 200, false, false)
			if err != nil {
				log.Printf("[Crawler] 监控模式第 %d 页抓取失败: %v", page, err)
				continue
			}
			totalPosts += result.PostCount
			totalComments += result.CommentCount

			log.Printf("[Crawler] 监控第 %d 页完成: +%d帖子 +%d评论", page, result.PostCount, result.CommentCount)
		}

		duration := time.Since(startTime)

		return CrawlMsg{
			Page:     monitorPages,
			Duration: duration,
		}
	}
}

func loadPostsCmd(database *db.Database, cursor, limit int) tea.Cmd {
	return func() tea.Msg {
		posts, err := database.GetPostsCursor(cursor, limit, false)
		if err != nil {
			return LoadPostsMsg{Error: err}
		}
		return LoadPostsMsg{
			Posts:   posts,
			Cursor:  cursor,
			HasMore: len(posts) == limit,
		}
	}
}

func loadCommentsCmd(database *db.Database, pid int32, sortAsc bool, cursor ...int32) tea.Cmd {
	return func() tea.Msg {
		const batchSize = 50
		begin := int32(0)
		if len(cursor) > 0 {
			begin = cursor[0]
		}
		comments, err := database.GetCommentsByPidCursor(pid, begin, batchSize, sortAsc)
		if err != nil {
			return LoadCommentsMsg{Error: err}
		}
		return LoadCommentsMsg{
			Comments: comments,
			Cursor:   begin,
			HasMore:  len(comments) == batchSize,
			SortAsc:  sortAsc,
		}
	}
}

func loadPostDetailCmd(database *db.Database, pid int32, sortAsc bool) tea.Cmd {
	return func() tea.Msg {
		post, err := database.GetPostByPid(pid)
		if err != nil {
			return LoadPostDetailMsg{Error: err}
		}
		const batchSize = 50
		comments, err := database.GetCommentsByPidCursor(pid, 0, batchSize, sortAsc)
		if err != nil {
			return LoadPostDetailMsg{Error: err}
		}
		return LoadPostDetailMsg{
			Post:     post,
			Comments: comments,
			HasMore:  len(comments) == batchSize,
			SortAsc:  sortAsc,
		}
	}
}

func searchPostsCmd(database *db.Database, keyword string, cursor, limit int) tea.Cmd {
	return func() tea.Msg {
		posts, err := database.SearchPostsCursor(keyword, cursor, limit, false)
		if err != nil {
			return SearchPostsMsg{Error: err}
		}
		return SearchPostsMsg{
			Posts:   posts,
			Cursor:  cursor,
			HasMore: len(posts) == limit,
		}
	}
}

func nextPostCursor(posts []models.Post) int {
	if len(posts) == 0 {
		return 0
	}
	return int(posts[len(posts)-1].Pid)
}

func nextCommentCursor(comments []models.Comment) int32 {
	if len(comments) == 0 {
		return 0
	}
	return comments[len(comments)-1].Cid
}

func loadLogsCmd() tea.Cmd {
	return func() tea.Msg {
		file, err := os.Open("crawler.log")
		if err != nil {
			return LoadLogsMsg{Error: err}
		}
		defer file.Close()

		var lines []string
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}

		if len(lines) > 500 {
			lines = lines[len(lines)-500:]
		}

		for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
			lines[i], lines[j] = lines[j], lines[i]
		}

		return LoadLogsMsg{Lines: lines}
	}
}

func loadConfigCmd() tea.Cmd {
	return func() tea.Msg {
		cfg, err := config.LoadConfig()
		if err != nil {
			return LoadConfigMsg{Error: err}
		}
		return LoadConfigMsg{Config: cfg}
	}
}

func saveConfigCmd(cfg *config.Config) tea.Cmd {
	return func() tea.Msg {
		existing, err := config.LoadConfig()
		if err == nil {
			cfg.Database = existing.Database
			cfg.Cors = existing.Cors
		}
		data, err := json.MarshalIndent(cfg, "", "    ")
		if err != nil {
			return SaveConfigMsg{Error: err}
		}
		err = os.WriteFile("config.json", data, 0644)
		if err != nil {
			return SaveConfigMsg{Error: err}
		}
		return SaveConfigMsg{}
	}
}

func InitClientForTUI() (*client.Client, *config.Config, error) {
	log.Printf("[Auth] 正在初始化客户端...")

	cfg, cfgErr := config.LoadConfig()
	deviceUUID := ""
	if cfgErr == nil && cfg != nil {
		deviceUUID = cfg.DeviceUUID
	}

	c, err := client.NewClient(deviceUUID)
	if err != nil {
		log.Printf("[Auth] 创建客户端失败: %v", err)
		return nil, nil, err
	}

	log.Printf("[Auth] 尝试使用已有 Cookie 登录...")
	resp, err := c.UnRead()
	if err == nil && resp.StatusCode == 200 {
		c.SaveCookies()
		if cfg == nil {
			cfg, _ = config.LoadConfig()
		}
		log.Printf("[Auth] Cookie 登录成功")
		return c, cfg, nil
	}
	if resp != nil {
		resp.Body.Close()
	}
	log.Printf("[Auth] Cookie 登录失败，尝试账号密码登录...")

	if cfg == nil {
		log.Printf("[Auth] 加载配置文件失败: %v", cfgErr)
		return c, nil, nil
	}

	log.Printf("[Auth] 正在执行 OAuth 登录...")
	oauthResult, err := c.OAuthLogin(cfg.Username, cfg.Password)
	if err != nil {
		log.Printf("[Auth] OAuth 登录失败: %v", err)
		return c, cfg, nil
	}

	token, ok := oauthResult["token"].(string)
	if !ok {
		log.Printf("[Auth] 未获取到 OAuth token")
		return c, cfg, nil
	}

	log.Printf("[Auth] 正在执行 SSO 登录...")
	err = c.SSOLogin(token)
	if err != nil {
		log.Printf("[Auth] SSO 登录失败: %v", err)
		return c, cfg, nil
	}

	log.Printf("[Auth] 正在验证登录状态...")
	resp, err = c.UnRead()
	if err != nil {
		log.Printf("[Auth] 验证登录状态失败: %v", err)
		return c, cfg, nil
	}

	var unReadResult map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&unReadResult)
	resp.Body.Close()
	if err != nil {
		log.Printf("[Auth] 解析验证结果失败: %v", err)
		return c, cfg, nil
	}

	if success, ok := unReadResult["success"].(bool); ok && success {
		c.SaveCookies()
		log.Printf("[Auth] 账号密码登录成功")
		return c, cfg, nil
	}

	if message, ok := unReadResult["message"].(string); ok && message == "请进行令牌验证" {
		log.Printf("[Auth] 需要 TOTP 令牌验证...")
		totpToken, _ := totp.GenerateCode(cfg.SecretKey, time.Now())
		resp, err = c.LoginByToken(totpToken)
		if err != nil {
			log.Printf("[Auth] TOTP 登录失败: %v", err)
			return c, cfg, nil
		}
		resp.Body.Close()

		log.Printf("[Auth] 正在验证 TOTP 登录状态...")
		resp, err = c.UnRead()
		if err != nil {
			log.Printf("[Auth] 验证 TOTP 登录状态失败: %v", err)
			return c, cfg, nil
		}

		var finalResult map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&finalResult)
		resp.Body.Close()
		if err != nil {
			log.Printf("[Auth] 解析 TOTP 验证结果失败: %v", err)
			return c, cfg, nil
		}

		if success, ok := finalResult["success"].(bool); ok && success {
			c.SaveCookies()
			log.Printf("[Auth] TOTP 令牌验证成功，登录完成")
			return c, cfg, nil
		}
		log.Printf("[Auth] TOTP 令牌验证失败")
	}

	log.Printf("[Auth] 所有登录方式均失败")
	return c, cfg, nil
}
