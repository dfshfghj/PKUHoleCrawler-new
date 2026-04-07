package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	w := m.Width
	if w < 1 {
		w = 80
	}
	h := m.Height
	if h < 1 {
		h = 24
	}

	// Tab bar - fixed top
	tabs := []string{"首页", "帖子"}
	var tabItems []string
	for i, t := range tabs {
		if i == m.TabCursor {
			tabItems = append(tabItems, tabItemActiveStyle.Render(t))
		} else {
			tabItems = append(tabItems, tabItemStyle.Render(t))
		}
	}
	tabBar := tabBarStyle.Width(w).Render(lipgloss.JoinHorizontal(lipgloss.Left, tabItems...))

	// Footer - fixed bottom
	loc, _ := time.LoadLocation("Asia/Shanghai")
	footerText := fmt.Sprintf("PKUHole Crawler v1.0 | h: 帮助 | %s", time.Now().In(loc).Format("15:04:05"))
	footerWidth := lipgloss.Width(footerText)
	leftPad := w - footerWidth
	if leftPad < 0 {
		leftPad = 0
	}
	footer := strings.Repeat(" ", leftPad) + footerText

	// Content area - fixed height between tab bar and footer
	contentHeight := h - 3 // tab bar(1) + separator(1) + footer(1)
	if contentHeight < 1 {
		contentHeight = 1
	}
	content := m.renderContent()
	contentBlock := lipgloss.NewStyle().
		Height(contentHeight).
		Width(w).
		Background(colorBlack).
		Render(content)

	body := tabBar + "\n" + contentBlock + "\n" + footer

	// Dialog overlay
	if m.Dialog != DialogNone {
		dialog := m.renderDialog()
		body = lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, dialog,
			lipgloss.WithWhitespaceBackground(colorBlack))
	}

	return baseStyle.Width(w).Height(h).Render(body)
}

func (m Model) renderContent() string {
	switch m.Page {
	case PageHome:
		return m.renderHome()
	case PagePosts:
		return m.renderPosts()
	default:
		return "Unknown page"
	}
}

func (m Model) renderDialog() string {
	switch m.Dialog {
	case DialogConfig:
		return m.renderConfigDialog()
	case DialogLogs:
		return m.renderLogsDialog()
	case DialogHelp:
		return m.renderHelpDialog()
	default:
		return ""
	}
}

func (m Model) renderHome() string {
	var b strings.Builder

	b.WriteString(vTitleStyle.Render("PKUHole Crawler"))
	b.WriteString("\n\n")

	if m.LoggedIn {
		b.WriteString(vStatusCard.Render(
			lipgloss.JoinHorizontal(lipgloss.Top,
				vStatLabelStyle.Render("登录状态: "),
				vStatusRunningStyle.Render("已登录"),
				vStatLabelStyle.Render("  用户: "),
				vStatValueStyle.Render(m.LoginUser),
			),
		))
	} else {
		b.WriteString(vStatusCard.Render(
			vStatusStoppedStyle.Render("未登录"),
		))
	}

	var crawlerStatus string
	switch m.CrawlerState {
	case CrawlerRunning:
		crawlerStatus = vStatusRunningStyle.Render("运行中")
	case CrawlerStopped:
		crawlerStatus = vStatusStoppedStyle.Render("已停止")
	case CrawlerError:
		crawlerStatus = vStatusStoppedStyle.Render("错误")
	}

	elapsed := "0s"
	if m.CrawlerState == CrawlerRunning && !m.CrawlerStart.IsZero() {
		elapsed = time.Since(m.CrawlerStart).Round(time.Second).String()
	}

	b.WriteString("\n")
	b.WriteString(vStatusCard.Render(
		lipgloss.JoinHorizontal(lipgloss.Top,
			vStatLabelStyle.Render("爬虫状态: "),
			crawlerStatus,
			vStatLabelStyle.Render("  运行时长: "),
			vStatValueStyle.Render(elapsed),
		),
	))

	b.WriteString("\n")
	b.WriteString(vStatsCard.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.JoinHorizontal(lipgloss.Top,
				vStatLabelStyle.Render("帖子总数: "),
				vStatValueStyle.Render(fmt.Sprintf("%d", m.TotalPosts)),
				vStatLabelStyle.Render("  评论总数: "),
				vStatValueStyle.Render(fmt.Sprintf("%d", m.TotalComments)),
			),
			lipgloss.JoinHorizontal(lipgloss.Top,
				vStatLabelStyle.Render("上次爬取: "),
				vStatValueStyle.Render(fmt.Sprintf("第%d页", m.LastCrawlPage)),
				vStatLabelStyle.Render("  耗时: "),
				vStatValueStyle.Render(m.LastCrawlTime.Round(time.Millisecond).String()),
			),
		),
	))

	// Crawl mode selection
	modeLabel := "顺序爬取"
	if m.CrawlMode == CrawlMonitor {
		modeLabel = fmt.Sprintf("监控模式(前%d页)", m.MonitorPages)
	}
	modeStyle := vStatLabelStyle
	if m.HomeButtonIdx == 2 {
		modeStyle = modeStyle.Foreground(colorPink)
	}
	b.WriteString("\n")
	b.WriteString(vStatusCard.Render(
		lipgloss.JoinHorizontal(lipgloss.Top,
			modeStyle.Render("爬取模式: "),
			vStatValueStyle.Render(modeLabel),
			vStatLabelStyle.Render("  m: 切换"),
		),
	))

	b.WriteString("\n")
	buttons := []string{"启动爬虫", "停止爬虫"}
	var btns []string
	for i, label := range buttons {
		disabled := (i == 0 && m.CrawlerState == CrawlerRunning) ||
			(i == 1 && m.CrawlerState != CrawlerRunning)
		if disabled {
			btns = append(btns, vButtonDisabled.Render(label))
		} else if m.HomeButtonIdx == i {
			if i == 1 {
				btns = append(btns, vButtonActiveDanger.Render(label))
			} else {
				btns = append(btns, vButtonActive.Render(label))
			}
		} else {
			btns = append(btns, vButtonDefault.Render(label))
		}
	}
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Center, btns...))

	if m.HomeLastError != "" {
		b.WriteString("\n\n")
		b.WriteString(vErrorStyle.Render("错误: " + m.HomeLastError))
	}

	return b.String()
}

func (m Model) renderPosts() string {
	if m.ShowPostDetail {
		return m.renderPostDetail()
	}

	var b strings.Builder

	if m.SearchActive {
		b.WriteString(vTitleStyle.Render(fmt.Sprintf("搜索结果: %s", m.SearchInput)))
	} else {
		b.WriteString(vTitleStyle.Render("帖子列表"))
	}
	b.WriteString("\n")

	searchLabel := "按 / 搜索"
	if m.Searching {
		searchLabel = "输入关键词 (Enter搜索, Esc取消): " + m.SearchInput
		b.WriteString(vSearchInputFocused.Render(searchLabel))
	} else {
		b.WriteString(vSearchInput.Render(searchLabel))
	}
	b.WriteString("\n")

	if m.PostListLoading {
		b.WriteString(vLoadingStyle.Render("加载中..."))
		return b.String()
	}

	if m.PostListError != "" {
		b.WriteString(vErrorStyle.Render("错误: " + m.PostListError))
		b.WriteString("\n")
	}

	if len(m.PostList) == 0 {
		b.WriteString(vEmptyStyle.Render("暂无数据"))
	} else {
		contentWidth := m.Width - 8
		if contentWidth < 20 {
			contentWidth = 20
		}

		selStyle := lipgloss.NewStyle().
			Foreground(colorPink).
			Bold(true).
			Padding(0, 0, 0, 1).
			Background(colorBlack).
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(colorPink).
			Render

		var content strings.Builder
		for i, post := range m.PostList {
			if i > 0 {
				content.WriteString("\n")
			}

			loc, _ := time.LoadLocation("Asia/Shanghai")
			ts := time.Unix(int64(post.Timestamp), 0).In(loc).Format("2006-01-02 15:04")
			replyStr := vPostReplyStyle.Render(fmt.Sprintf("回复:%d", post.Reply))
			likeStr := vPostLikeStyle.Render(fmt.Sprintf("赞:%d", post.Likenum))
			meta := replyStr + " " + likeStr
			pidStr := vPostPidStyle.Render(fmt.Sprintf("#%-6d", post.Pid))
			tsStr := vPostTimeStyle.Render(ts)
			header := pidStr + " " + tsStr + "  " + meta
			if !post.Anonymous {
				header = pidStr + " [实名] " + tsStr + "  " + meta
			}

			if i == m.SelectedPostIdx {
				content.WriteString(selStyle(header) + "\n")
				text := strings.ReplaceAll(post.Text, "\r\n", "\n")
				for _, line := range strings.Split(text, "\n") {
					content.WriteString(selStyle("  "+line) + "\n")
				}
			} else {
				content.WriteString(header + "\n")
				text := strings.ReplaceAll(post.Text, "\r\n", "\n")
				for _, line := range strings.Split(text, "\n") {
					content.WriteString("  " + line + "\n")
				}
			}
		}

		newContent := content.String()
		if m.postContent != newContent || m.PostViewport.Width != contentWidth || m.PostViewport.Height != m.calcPostViewportHeight() {
			m.PostViewport.Width = contentWidth
			m.PostViewport.Height = m.calcPostViewportHeight()
			m.PostViewport.SetContent(newContent)
			m.postContent = newContent
		}
		b.WriteString(m.PostViewport.View())

		b.WriteString("\n")
		b.WriteString(vPaginationStyle.Render(
			fmt.Sprintf("↑↓: 选择 | Enter: 查看 | /: 搜索 | r: 刷新 | PgUp/PgDn: 快滚 | 已加载 %d/%d",
				len(m.PostList), m.PostListTotal),
		))
	}

	return b.String()
}

func (m Model) renderPostDetail() string {
	var b strings.Builder

	if m.CurrentPost == nil {
		return "无帖子数据"
	}
	loc, _ := time.LoadLocation("Asia/Shanghai")
	ts := time.Unix(int64(m.CurrentPost.Timestamp), 0).In(loc).Format("2006-01-02 15:04")
	b.WriteString(vPostPidStyle.Render(fmt.Sprintf("#%d", m.CurrentPost.Pid)))
	b.WriteString("  ")
	b.WriteString(vPostTimeStyle.Render(ts))
	b.WriteString("  ")
	b.WriteString(vPostReplyStyle.Render(fmt.Sprintf("回复: %d", m.CurrentPost.Reply)))
	b.WriteString("  ")
	b.WriteString(vPostLikeStyle.Render(fmt.Sprintf("点赞: %d", m.CurrentPost.Likenum)))
	b.WriteString("\n\n")

	b.WriteString(vPostTextStyle.Render(m.CurrentPost.Text))
	b.WriteString("\n\n")

	b.WriteString(vDividerStyle.Render(strings.Repeat("─", 60)))
	b.WriteString("\n\n")

	b.WriteString(vSubtitleStyle.Render(fmt.Sprintf("评论 (%d):", len(m.CommentList))))
	b.WriteString("\n\n")

	if len(m.CommentList) == 0 {
		b.WriteString(vEmptyStyle.Render("暂无评论"))
	} else {
		var content strings.Builder
		for i, c := range m.CommentList {
			if i > 0 {
				content.WriteString("\n")
			}

			cName := c.NameTag
			if cName == "" {
				cName = "匿名"
			}
			loc, _ := time.LoadLocation("Asia/Shanghai")
			cTs := time.Unix(int64(c.Timestamp), 0).In(loc).Format("15:04")
			cText := c.Text

			content.WriteString(fmt.Sprintf("%s %s:\n  %s", cTs, cName, cText))
		}

		contentWidth := m.Width - 8
		if contentWidth < 20 {
			contentWidth = 20
		}
		newContent := content.String()
		vpHeight := m.calcPostViewportHeight() - 6
		if m.commentContent != newContent || m.CommentViewport.Width != contentWidth || m.CommentViewport.Height != vpHeight {
			m.CommentViewport.Width = contentWidth
			m.CommentViewport.Height = vpHeight
			m.CommentViewport.SetContent(newContent)
			m.commentContent = newContent
		}
		b.WriteString(m.CommentViewport.View())
	}

	b.WriteString("\n")
	b.WriteString(vPaginationStyle.Render("Esc: 返回列表 | ↑↓: 滚动评论"))

	return b.String()
}

func (m Model) calcPostViewportHeight() int {
	headerLines := 3
	paddingLines := 2
	paginationLines := 1
	avail := m.Height - 3 - headerLines - paddingLines - paginationLines
	if avail < 3 {
		avail = 3
	}
	return avail
}

func (m Model) renderConfigDialog() string {
	var b strings.Builder

	b.WriteString(vDialogTitleStyle.Render("配置管理"))
	b.WriteString("\n\n")

	b.WriteString(vSubtitleStyle.Render("config.json"))
	b.WriteString("\n\n")

	fields := []struct {
		label string
		value string
	}{
		{"用户名:", maskField(m.ConfigUsername, false)},
		{"密码:", maskField(m.ConfigPassword, true)},
		{"SecretKey:", maskField(m.ConfigSecretKey, true)},
	}

	for i, f := range fields {
		labelStyle := vFormLabelStyle
		inputStyle := vFormInput

		if m.ConfigFieldIdx == i {
			labelStyle = labelStyle.Foreground(colorPink)
			inputStyle = vFormInputFocused
		}

		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top,
			labelStyle.Render(f.label),
			inputStyle.Render(f.value),
		))
		b.WriteString("\n")
	}

	saveBtn := "保存配置"
	if m.ConfigFieldIdx == 3 {
		b.WriteString("\n")
		b.WriteString(vFormSaveActive.Render(saveBtn))
	} else {
		b.WriteString("\n")
		b.WriteString(vFormSaveBtn.Render(saveBtn))
	}

	if m.ConfigSaving {
		b.WriteString("\n")
		b.WriteString(vLoadingStyle.Render("保存中..."))
	}

	if m.ConfigSaveOK {
		b.WriteString("\n")
		b.WriteString(vStatusRunningStyle.Render("配置已保存!"))
	}

	if m.LastError != "" {
		b.WriteString("\n")
		b.WriteString(vErrorStyle.Render("错误: " + m.LastError))
	}

	b.WriteString("\n\n")
	b.WriteString(vDialogHelpStyle.Render("↑↓: 选择 | Enter: 编辑/保存 | Esc: 关闭"))

	return dialogCard.Render(b.String())
}

func (m Model) renderLogsDialog() string {
	var b strings.Builder

	b.WriteString(vDialogTitleStyle.Render("运行日志"))
	b.WriteString("\n\n")

	if m.LogLoading {
		b.WriteString(vLoadingStyle.Render("加载日志中..."))
	} else if len(m.LogLines) == 0 {
		b.WriteString(vEmptyStyle.Render("暂无日志"))
	} else {
		end := m.LogOffset + 15
		if end > len(m.LogLines) {
			end = len(m.LogLines)
		}

		for i := m.LogOffset; i < end; i++ {
			line := m.LogLines[i]
			if len(line) > 70 {
				line = line[:70]
			}
			b.WriteString(vLogLineStyle.Render(line))
			b.WriteString("\n")
		}

		b.WriteString("\n")
		totalLines := len(m.LogLines)
		b.WriteString(vPaginationStyle.Render(
			fmt.Sprintf("日志: %d 行 | 当前: %d-%d | ↑↓/PgUp/PgDn滚动 | r: 刷新",
				totalLines, m.LogOffset+1, minInt(m.LogOffset+15, totalLines)),
		))
	}

	b.WriteString("\n")
	b.WriteString(vDialogHelpStyle.Render("Esc: 关闭"))

	return dialogCard.Render(b.String())
}

func (m Model) renderHelpDialog() string {
	var b strings.Builder

	b.WriteString(vDialogTitleStyle.Render("快捷键帮助"))
	b.WriteString("\n\n")

	helpItems := []struct {
		key  string
		desc string
	}{
		{"h", "打开/关闭此帮助菜单"},
		{"c", "打开配置管理对话框"},
		{"l", "打开运行日志查看器"},
		{"Tab", "在首页和帖子列表之间切换"},
		{"q", "退出程序"},
		{"", ""},
		{"m", "切换爬取模式（顺序/监控）"},
		{"←→", "选择启动/停止爬虫按钮"},
		{"Enter", "执行选中的操作"},
		{"", ""},
		{"/", "搜索帖子"},
		{"r", "刷新帖子列表"},
		{"↑↓", "选择帖子 / 滚动评论"},
		{"PgUp/PgDn", "快速滚动"},
	}

	for _, item := range helpItems {
		if item.key == "" {
			b.WriteString("\n")
			continue
		}
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top,
			vStatValueStyle.Width(12).Render(item.key),
			vStatLabelStyle.Render(item.desc),
		))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(vDialogHelpStyle.Render("Esc: 关闭"))

	return dialogCard.Render(b.String())
}

func maskField(s string, mask bool) string {
	if s == "" {
		return "(空)"
	}
	if mask {
		return strings.Repeat("*", len(s))
	}
	return s
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
