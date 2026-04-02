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
			m.PostList = msg.Posts
			m.PostListTotal = msg.Total
			m.PostListPage = msg.Page
			m.PostListCursor = 0
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
			m.PostList = msg.Posts
			m.PostListTotal = msg.Total
			m.PostListPage = msg.Page
			m.PostListCursor = 0
			m.SearchActive = true
			m.Searching = false
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
			m.ConfigSaved = true
			m.ConfigSaveOK = true
		}
		return m, nil
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	if msg.String() == "q" && !m.Searching && m.ConfigFieldIdx < 3 {
		return m, tea.Quit
	}

	if msg.String() == "tab" && !m.Searching && m.ConfigFieldIdx < 3 && !m.ShowPostDetail {
		m.TabCursor = (m.TabCursor + 1) % 4
		m.Page = Page(m.TabCursor)
		if m.Page == PagePosts && len(m.PostList) == 0 {
			m.PostListLoading = true
			return m, loadPostsCmd(m.Database, 0, m.PostListPerPage)
		}
		if m.Page == PageLogs {
			m.LogLoading = true
			return m, loadLogsCmd()
		}
		if m.Page == PageConfig {
			return m, loadConfigCmd()
		}
		return m, nil
	}

	switch m.Page {
	case PageHome:
		return m.handleHomeKey(msg)
	case PagePosts:
		return m.handlePostsKey(msg)
	case PageConfig:
		return m.handleConfigKey(msg)
	case PageLogs:
		return m.handleLogsKey(msg)
	}

	return m, nil
}

func (m Model) handleHomeKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "left":
		if m.HomeButtonIdx > 0 {
			m.HomeButtonIdx--
		}
	case "right":
		if m.HomeButtonIdx < 1 {
			m.HomeButtonIdx++
		}
	case "enter":
		if m.HomeButtonIdx == 0 && m.CrawlerState == CrawlerStopped {
			m.CrawlerState = CrawlerRunning
			m.CrawlerStart = time.Now()
			m.HomeLastError = ""
			log.Printf("[Crawler] 爬虫已启动")
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
			return m, nil
		case tea.KeyEnter:
			if m.SearchInput != "" {
				m.PostListLoading = true
				return m, searchPostsCmd(m.Database, m.SearchInput, 0, m.PostListPerPage)
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
			if m.CommentCursor > 0 {
				m.CommentCursor--
			}
		case "down":
			if m.CommentCursor < len(m.CommentList)-1 {
				m.CommentCursor++
			}
		}
		return m, nil
	}

	switch msg.String() {
	case "/":
		m.Searching = true
		m.SearchInput = ""
		return m, nil
	case "up":
		if m.PostListCursor > 0 {
			m.PostListCursor--
		}
	case "down":
		if m.PostListCursor < len(m.PostList)-1 {
			m.PostListCursor++
		}
	case "enter":
		if len(m.PostList) > 0 && m.PostListCursor < len(m.PostList) {
			post := m.PostList[m.PostListCursor]
			m.ShowPostDetail = true
			m.CurrentPost = &post
			m.CommentCursor = 0
			return m, loadCommentsCmd(m.Database, post.Pid)
		}
	case "left":
		if m.PostListPage > 1 {
			m.PostListPage--
			m.PostListLoading = true
			offset := (m.PostListPage - 1) * m.PostListPerPage
			return m, loadPostsCmd(m.Database, offset, m.PostListPerPage)
		}
	case "right":
		if m.PostListPage*m.PostListPerPage < m.PostListTotal {
			m.PostListPage++
			m.PostListLoading = true
			offset := (m.PostListPage - 1) * m.PostListPerPage
			return m, loadPostsCmd(m.Database, offset, m.PostListPerPage)
		}
	}
	return m, nil
}

func (m Model) handleConfigKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	if m.ConfigFieldIdx < 3 {
		switch msg.Type {
		case tea.KeyEscape:
			m.ConfigUsername = m.Config.Username
			m.ConfigPassword = m.Config.Password
			m.ConfigSecretKey = m.Config.SecretKey
			m.ConfigFieldIdx = 3
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
		result, err := crawler.FetchAndSave(c, database, page)
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

func loadCommentsCmd(database *db.Database, pid int) tea.Cmd {
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
	c, err := client.NewClient()
	if err != nil {
		log.Printf("[Auth] 创建客户端失败: %v", err)
		return nil, nil, err
	}

	log.Printf("[Auth] 尝试使用已有 Cookie 登录...")
	resp, err := c.UnRead()
	if err == nil && resp.StatusCode == 200 {
		c.SaveCookies()
		cfg, _ := config.LoadConfig()
		log.Printf("[Auth] Cookie 登录成功")
		return c, cfg, nil
	}
	if resp != nil {
		resp.Body.Close()
	}
	log.Printf("[Auth] Cookie 登录失败，尝试账号密码登录...")

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Printf("[Auth] 加载配置文件失败: %v", err)
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
