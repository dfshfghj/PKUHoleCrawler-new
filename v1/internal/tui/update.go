package tui

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	"treehole/internal/client"
	"treehole/internal/config"
	"treehole/internal/crawler"
	"treehole/internal/db"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pquerna/otp/totp"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case TickMsg:
		return m, tickCmd()

	case LoginMsg:
		if msg.Error != nil {
			m.LastError = msg.Error.Error()
			m.LoggedIn = false
		} else {
			m.LoggedIn = true
			m.LoginUser = msg.Username
		}
		return m, nil

	case LoadStatsMsg:
		if msg.Error == nil {
			m.TotalPosts = msg.PostCount
			m.TotalComments = msg.CommentCount
		}
		return m, nil

	case CrawlMsg:
		if msg.Error != nil {
			log.Printf("[Crawler] 爬虫错误: 第 %d 页, %v", msg.Page, msg.Error)
			m.CrawlerState = CrawlerError
			m.HomeLastError = msg.Error.Error()
		} else {
			m.LastCrawlPage = msg.Page
			m.LastCrawlTime = msg.Duration
			m.TotalPosts = msg.PostsCount
			m.TotalComments = msg.CommentsCount
			if m.CrawlerState == CrawlerRunning {
				if m.CrawlMode == CrawlMonitor {
					return m, tea.Batch(
						crawlMonitorCmd(m.Client, m.Database, m.MonitorPages),
						func() tea.Msg {
							pc, _ := m.Database.GetPostCount()
							cc, _ := m.Database.GetCommentCount()
							return LoadStatsMsg{PostCount: pc, CommentCount: cc}
						},
					)
				}
				return m, tea.Batch(
					crawlPageCmd(m.Client, m.Database, msg.Page+1, 3),
					func() tea.Msg {
						pc, _ := m.Database.GetPostCount()
						cc, _ := m.Database.GetCommentCount()
						return LoadStatsMsg{PostCount: pc, CommentCount: cc}
					},
				)
			} else {
				log.Printf("[Crawler] 爬虫已停止，最终抓取到第 %d 页", msg.Page)
			}
		}
		return m, nil

	case LoadPostsMsg:
		m.PostListLoading = false
		if msg.Error != nil {
			log.Printf("[Posts] 加载帖子列表失败: %v", msg.Error)
			m.PostListError = msg.Error.Error()
		} else {
			log.Printf("[Posts] 加载 %d 条帖子, 总计 %d 条", len(msg.Posts), msg.Total)
			if !m.SearchActive {
				m.PostList = append(m.PostList, msg.Posts...)
			}
			m.PostListTotal = msg.Total
			m.postContent = ""
		}
		return m, nil

	case LoadCommentsMsg:
		if msg.Error != nil {
			log.Printf("[Posts] 加载评论失败: %v", msg.Error)
			m.PostListError = msg.Error.Error()
		} else {
			log.Printf("[Posts] 加载 %d 条评论", len(msg.Comments))
			m.CommentList = msg.Comments
		}
		return m, nil

	case SearchPostsMsg:
		m.PostListLoading = false
		if msg.Error != nil {
			m.PostListError = msg.Error.Error()
		} else {
			if m.SearchActive {
				m.PostList = append(m.PostList, msg.Posts...)
			} else {
				m.PostList = msg.Posts
				m.SelectedPostIdx = 0
				m.PostViewport.GotoTop()
			}
			m.PostListTotal = msg.Total
			m.SearchActive = true
			m.Searching = false
			m.postContent = ""
		}
		return m, nil

	case LoadLogsMsg:
		m.LogLoading = false
		if msg.Error != nil {
			m.LastError = msg.Error.Error()
		} else {
			m.LogLines = msg.Lines
		}
		return m, nil

	case LoadConfigMsg:
		if msg.Error == nil && msg.Config != nil {
			m.Config = msg.Config
			m.ConfigUsername = msg.Config.Username
			m.ConfigPassword = msg.Config.Password
			m.ConfigSecretKey = msg.Config.SecretKey
			m.ConfigFieldIdx = 0
		}
		return m, nil

	case SaveConfigMsg:
		m.ConfigSaving = false
		if msg.Error != nil {
			m.LastError = msg.Error.Error()
		} else {
			m.ConfigSaveOK = true
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

	if msg.String() == "q" && m.Dialog == DialogNone && !m.Searching && m.ConfigFieldIdx < 3 {
		return m, tea.Quit
	}

	// Open dialogs
	if m.Dialog == DialogNone && !m.Searching && !m.ShowPostDetail && m.ConfigFieldIdx < 3 {
		if msg.String() == "c" {
			m.Dialog = DialogConfig
			m.ConfigFieldIdx = 0
			m.ConfigSaveOK = false
			return m, loadConfigCmd()
		}
		if msg.String() == "l" {
			m.Dialog = DialogLogs
			m.LogLoading = true
			return m, loadLogsCmd()
		}
		if msg.String() == "h" {
			m.Dialog = DialogHelp
			return m, nil
		}
	}

	if msg.String() == "tab" && m.Dialog == DialogNone && !m.Searching && !m.ShowPostDetail {
		m.TabCursor = (m.TabCursor + 1) % 2
		m.Page = Page(m.TabCursor)
		if m.Page == PagePosts && len(m.PostList) == 0 {
			m.PostListLoading = true
			return m, loadPostsCmd(m.Database, 0, m.PostPerPage)
		}
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
	switch msg.String() {
	case "m":
		if m.CrawlerState == CrawlerStopped {
			if m.CrawlMode == CrawlSequential {
				m.CrawlMode = CrawlMonitor
			} else {
				m.CrawlMode = CrawlSequential
			}
		}
	case "left":
		if m.HomeButtonIdx > 0 {
			m.HomeButtonIdx--
		}
	case "right":
		if m.HomeButtonIdx < 2 {
			m.HomeButtonIdx++
		}
	case "enter":
		if m.HomeButtonIdx == 0 && m.CrawlerState == CrawlerStopped {
			m.CrawlerState = CrawlerRunning
			m.CrawlerStart = time.Now()
			m.HomeLastError = ""
			log.Printf("[Crawler] 爬虫已启动, 模式: %v", m.CrawlMode)
			if m.CrawlMode == CrawlMonitor {
				return m, tea.Batch(
					crawlMonitorCmd(m.Client, m.Database, m.MonitorPages),
					func() tea.Msg {
						pc, _ := m.Database.GetPostCount()
						cc, _ := m.Database.GetCommentCount()
						return LoadStatsMsg{PostCount: pc, CommentCount: cc}
					},
				)
			}
			return m, tea.Batch(
				crawlPageCmd(m.Client, m.Database, 1, 3),
				func() tea.Msg {
					pc, _ := m.Database.GetPostCount()
					cc, _ := m.Database.GetCommentCount()
					return LoadStatsMsg{PostCount: pc, CommentCount: cc}
				},
			)
		} else if m.HomeButtonIdx == 1 && m.CrawlerState == CrawlerRunning {
			m.CrawlerState = CrawlerStopped
			log.Printf("[Crawler] 爬虫已手动停止")
		} else if m.HomeButtonIdx == 2 {
			if m.CrawlMode == CrawlSequential {
				m.CrawlMode = CrawlMonitor
			} else {
				m.CrawlMode = CrawlSequential
			}
		}
	}
	return m, nil
}

func (m Model) handlePostsKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	if m.Searching {
		switch msg.Type {
		case tea.KeyEscape:
			m.Searching = false
			m.SearchInput = ""
			m.SearchActive = false
			m.postContent = ""
			return m, nil
		case tea.KeyEnter:
			if m.SearchInput != "" {
				m.PostListLoading = true
				return m, searchPostsCmd(m.Database, m.SearchInput, 0, m.PostPerPage)
			}
			return m, nil
		case tea.KeyBackspace:
			if len(m.SearchInput) > 0 {
				m.SearchInput = m.SearchInput[:len(m.SearchInput)-1]
			}
			return m, nil
		default:
			if msg.Type == tea.KeyRunes {
				m.SearchInput += msg.String()
			}
			return m, nil
		}
	}

	if m.ShowPostDetail {
		switch msg.String() {
		case "esc":
			m.ShowPostDetail = false
			m.CurrentPost = nil
			m.CommentList = nil
			return m, nil
		case "up":
			m.CommentViewport.ScrollUp(1)
		case "down":
			m.CommentViewport.ScrollDown(1)
		case "pgup":
			m.CommentViewport.PageUp()
		case "pgdown":
			m.CommentViewport.PageDown()
		}
		return m, nil
	}

	switch msg.String() {
	case "esc":
		if m.SearchActive {
			m.SearchActive = false
			m.SearchInput = ""
			m.postContent = ""
			m.PostListLoading = true
			m.PostList = nil
			m.PostListTotal = 0
			m.SelectedPostIdx = 0
			m.PostViewport.GotoTop()
			return m, loadPostsCmd(m.Database, 0, m.PostPerPage)
		}
	case "r":
		if !m.SearchActive {
			m.PostListLoading = true
			m.PostList = nil
			m.PostListTotal = 0
			m.SelectedPostIdx = 0
			m.postContent = ""
			m.PostViewport.GotoTop()
			return m, loadPostsCmd(m.Database, 0, m.PostPerPage)
		}
		return m, nil
	case "/":
		m.Searching = true
		m.SearchInput = ""
		return m, nil
	case "up":
		if m.SelectedPostIdx > 0 {
			m.SelectedPostIdx--
			m.scrollToSelectedPost()
		} else {
			m.PostViewport.ScrollUp(1)
		}
	case "down":
		if m.SelectedPostIdx < len(m.PostList)-1 {
			m.SelectedPostIdx++
			m.scrollToSelectedPost()
		} else {
			m.PostViewport.ScrollDown(1)
			if m.PostViewport.AtBottom() && m.PostListTotal > len(m.PostList) && !m.PostListLoading {
				m.PostListLoading = true
				offset := len(m.PostList)
				if m.SearchActive {
					return m, searchPostsCmd(m.Database, m.SearchInput, offset, m.PostPerPage)
				}
				return m, loadPostsCmd(m.Database, offset, m.PostPerPage)
			}
		}
	case "enter":
		if len(m.PostList) > 0 && m.SelectedPostIdx < len(m.PostList) {
			post := m.PostList[m.SelectedPostIdx]
			m.ShowPostDetail = true
			m.CurrentPost = &post
			return m, loadCommentsCmd(m.Database, post.Pid)
		}
	case "pgup":
		m.PostViewport.PageUp()
		m.adjustSelectedToViewport()
	case "pgdown":
		m.PostViewport.PageDown()
		m.adjustSelectedToViewport()
		if m.PostViewport.AtBottom() && m.PostListTotal > len(m.PostList) && !m.PostListLoading {
			m.PostListLoading = true
			offset := len(m.PostList)
			if m.SearchActive {
				return m, searchPostsCmd(m.Database, m.SearchInput, offset, m.PostPerPage)
			}
			return m, loadPostsCmd(m.Database, offset, m.PostPerPage)
		}
	}
	return m, nil
}

func (m *Model) scrollToSelectedPost() {
	if len(m.PostList) == 0 {
		return
	}
	targetLine := 0
	for i := 0; i < m.SelectedPostIdx && i < len(m.PostList); i++ {
		targetLine += 1 + strings.Count(m.PostList[i].Text, "\n") + 2
	}
	if targetLine < m.PostViewport.YOffset {
		m.PostViewport.SetYOffset(targetLine)
	} else {
		visibleLines := m.PostViewport.VisibleLineCount()
		if targetLine >= m.PostViewport.YOffset+visibleLines {
			m.PostViewport.SetYOffset(targetLine - visibleLines + 3)
		}
	}
}

func (m *Model) adjustSelectedToViewport() {
	if len(m.PostList) == 0 {
		return
	}
	yOffset := m.PostViewport.YOffset
	visibleLines := m.PostViewport.VisibleLineCount()
	lineIdx := 0
	for i := 0; i < len(m.PostList); i++ {
		postLines := 1 + strings.Count(m.PostList[i].Text, "\n") + 1
		if lineIdx+postLines > yOffset && lineIdx < yOffset+visibleLines {
			m.SelectedPostIdx = i
			return
		}
		lineIdx += postLines
	}
}

func (m Model) handleConfigKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	if m.ConfigFieldIdx < 3 {
		switch msg.Type {
		case tea.KeyEscape:
			m.Dialog = DialogNone
			return m, nil
		case tea.KeyEnter:
			m.ConfigFieldIdx = 3
			return m, nil
		case tea.KeyBackspace:
			switch m.ConfigFieldIdx {
			case 0:
				if len(m.ConfigUsername) > 0 {
					m.ConfigUsername = m.ConfigUsername[:len(m.ConfigUsername)-1]
				}
			case 1:
				if len(m.ConfigPassword) > 0 {
					m.ConfigPassword = m.ConfigPassword[:len(m.ConfigPassword)-1]
				}
			case 2:
				if len(m.ConfigSecretKey) > 0 {
					m.ConfigSecretKey = m.ConfigSecretKey[:len(m.ConfigSecretKey)-1]
				}
			}
			return m, nil
		case tea.KeyUp:
			if m.ConfigFieldIdx > 0 {
				m.ConfigFieldIdx--
			}
			return m, nil
		case tea.KeyDown:
			if m.ConfigFieldIdx < 2 {
				m.ConfigFieldIdx++
			}
			return m, nil
		default:
			if msg.Type == tea.KeyRunes {
				switch m.ConfigFieldIdx {
				case 0:
					m.ConfigUsername += msg.String()
				case 1:
					m.ConfigPassword += msg.String()
				case 2:
					m.ConfigSecretKey += msg.String()
				}
			}
			return m, nil
		}
	}

	if msg.String() == "enter" && m.ConfigFieldIdx == 3 {
		m.ConfigSaving = true
		newCfg := &config.Config{
			Username:  m.ConfigUsername,
			Password:  m.ConfigPassword,
			SecretKey: m.ConfigSecretKey,
		}
		return m, saveConfigCmd(newCfg)
	}

	switch msg.String() {
	case "up":
		if m.ConfigFieldIdx > 0 {
			m.ConfigFieldIdx--
		}
	case "down":
		if m.ConfigFieldIdx < 3 {
			m.ConfigFieldIdx++
		}
	case "enter":
		if m.ConfigFieldIdx < 3 {
			return m, nil
		}
	}
	return m, nil
}

func (m Model) handleLogsKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "up":
		if m.LogOffset > 0 {
			m.LogOffset--
		}
	case "down":
		if m.LogOffset < len(m.LogLines)-1 {
			m.LogOffset++
		}
	case "pgup":
		m.LogOffset -= 20
		if m.LogOffset < 0 {
			m.LogOffset = 0
		}
	case "pgdown":
		m.LogOffset += 20
		if m.LogOffset >= len(m.LogLines) {
			m.LogOffset = len(m.LogLines) - 1
		}
	case "r":
		m.LogLoading = true
		return m, loadLogsCmd()
	}
	return m, nil
}

func crawlPageCmd(c *client.Client, database *db.Database, page int, endPage int) tea.Cmd {
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

		pc, _ := database.GetPostCount()
		cc, _ := database.GetCommentCount()
		log.Printf("[Crawler] 数据库总计: %d 条帖子, %d 条评论", pc, cc)

		return CrawlMsg{
			PostsCount:    pc,
			CommentsCount: cc,
			Page:          page,
			Duration:      duration,
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
		pc, _ := database.GetPostCount()
		cc, _ := database.GetCommentCount()

		return CrawlMsg{
			PostsCount:    pc,
			CommentsCount: cc,
			Page:          monitorPages,
			Duration:      duration,
		}
	}
}

func loadPostsCmd(database *db.Database, offset, limit int) tea.Cmd {
	return func() tea.Msg {
		posts, err := database.GetPosts(offset, limit)
		if err != nil {
			return LoadPostsMsg{Error: err}
		}
		total, _ := database.GetPostCount()
		page := (offset / limit) + 1
		return LoadPostsMsg{Posts: posts, Total: total, Page: page}
	}
}

func loadCommentsCmd(database *db.Database, pid int32) tea.Cmd {
	return func() tea.Msg {
		comments, err := database.GetCommentsByPid(pid)
		if err != nil {
			return LoadCommentsMsg{Error: err}
		}
		return LoadCommentsMsg{Comments: comments}
	}
}

func searchPostsCmd(database *db.Database, keyword string, offset, limit int) tea.Cmd {
	return func() tea.Msg {
		posts, err := database.SearchPosts(keyword, offset, limit)
		if err != nil {
			return SearchPostsMsg{Error: err}
		}
		total, _ := database.SearchPostsCount(keyword)
		page := (offset / limit) + 1
		return SearchPostsMsg{Posts: posts, Total: total, Page: page}
	}
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
