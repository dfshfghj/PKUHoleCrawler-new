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

	// Tab bar - full width
	tabs := []string{"首页", "帖子", "配置", "日志"}
	var tabItems []string
	for i, t := range tabs {
		if i == m.TabCursor {
			tabItems = append(tabItems, tabItemActiveStyle.Render(t))
		} else {
			tabItems = append(tabItems, tabItemStyle.Render(t))
		}
	}
	tabBar := tabBarStyle.Width(w).Render(lipgloss.JoinHorizontal(lipgloss.Left, tabItems...))

	// Content - full width with padding from contentStyle
	content := contentStyle.Width(w).Render(m.renderContent())

	// Footer - full width, text right-aligned
	footerText := fmt.Sprintf("Tab: 切换页面 | q: 退出 | %s", time.Now().Format("15:04:05"))
	footerWidth := lipgloss.Width(footerText)
	leftPad := w - footerWidth
	if leftPad < 0 {
		leftPad = 0
	}
	footer := strings.Repeat(" ", leftPad) + footerText

	body := tabBar + "\n" + content + "\n" + footer

	return baseStyle.Width(w).Height(h).Render(body)
}

func (m Model) renderContent() string {
	switch m.Page {
	case PageHome:
		return m.renderHome()
	case PagePosts:
		return m.renderPosts()
	case PageConfig:
		return m.renderConfig()
	case PageLogs:
		return m.renderLogs()
	default:
		return "Unknown page"
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
	b.WriteString("\n\n")

	searchLabel := "按 / 搜索"
	if m.Searching {
		searchLabel = "输入关键词 (Enter搜索, Esc取消): " + m.SearchInput
		b.WriteString(vSearchInputFocused.Render(searchLabel))
	} else {
		b.WriteString(vSearchInput.Render(searchLabel))
	}
	b.WriteString("\n")

	if m.PostListLoading {
		b.WriteString("\n")
		b.WriteString(vLoadingStyle.Render("加载中..."))
		return b.String()
	}

	if m.PostListError != "" {
		b.WriteString("\n")
		b.WriteString(vErrorStyle.Render("错误: " + m.PostListError))
	}

	if len(m.PostList) == 0 {
		b.WriteString("\n")
		b.WriteString(vEmptyStyle.Render("暂无数据"))
	} else {
		for i, post := range m.PostList {
			style := vListItemStyle
			if i == m.PostListCursor {
				style = vListItemSelectedStyle
			}

			text := post.Text
			if len(text) > 80 {
				text = text[:80] + "..."
			}

			ts := time.Unix(int64(post.Timestamp), 0).Format("2006-01-02 15:04")

			line := fmt.Sprintf("#%-6d %-10s %s", post.Pid, ts, text)
			if post.Anonymous == 0 {
				line = fmt.Sprintf("#%-6d %-10s [实名] %s", post.Pid, ts, text)
			}

			b.WriteString(style.Render(line))
			b.WriteString("\n")
		}
	}

	if m.PostListTotal > 0 {
		totalPages := (m.PostListTotal + m.PostListPerPage - 1) / m.PostListPerPage
		b.WriteString("\n")
		b.WriteString(vPaginationStyle.Render(
			fmt.Sprintf("第 %d/%d 页 (共 %d 条) | ←→翻页 | ↑↓选择 | Enter查看 | /搜索",
				m.PostListPage, totalPages, m.PostListTotal),
		))
	}

	return b.String()
}

func (m Model) renderPostDetail() string {
	var b strings.Builder

	if m.CurrentPost == nil {
		return "无帖子数据"
	}

	ts := time.Unix(int64(m.CurrentPost.Timestamp), 0).Format("2006-01-02 15:04")
	b.WriteString(vPostPidStyle.Render(fmt.Sprintf("#%d", m.CurrentPost.Pid)))
	b.WriteString("  ")
	b.WriteString(vPostMetaStyle.Render(ts))
	b.WriteString(fmt.Sprintf("  回复: %d  点赞: %d", m.CurrentPost.Reply, m.CurrentPost.Likenum))
	if m.CurrentPost.Tag != "" {
		b.WriteString("  标签: " + m.CurrentPost.Tag)
	}
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
		for i, c := range m.CommentList {
			style := vListItemStyle
			if i == m.CommentCursor {
				style = vListItemSelectedStyle
			}

			cName := c.Name
			if cName == "" {
				cName = "匿名"
			}
			cTs := time.Unix(int64(c.Timestamp), 0).Format("15:04")
			cText := c.Text
			if len(cText) > 100 {
				cText = cText[:100] + "..."
			}

			line := fmt.Sprintf("%s %s: %s", cTs, cName, cText)
			b.WriteString(style.Render(line))
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (m Model) renderConfig() string {
	var b strings.Builder

	b.WriteString(vTitleStyle.Render("配置管理"))
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
			labelStyle = labelStyle.Foreground(lipgloss.Color("#FF69B4"))
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

	return b.String()
}

func (m Model) renderLogs() string {
	var b strings.Builder

	b.WriteString(vTitleStyle.Render("运行日志"))
	b.WriteString("\n\n")

	if m.LogLoading {
		b.WriteString(vLoadingStyle.Render("加载日志中..."))
		return b.String()
	}

	if m.LastError != "" {
		b.WriteString(vErrorStyle.Render("错误: " + m.LastError))
		b.WriteString("\n")
	}

	if len(m.LogLines) == 0 {
		b.WriteString(vEmptyStyle.Render("暂无日志"))
	} else {
		end := m.LogOffset + 20
		if end > len(m.LogLines) {
			end = len(m.LogLines)
		}

		for i := m.LogOffset; i < end; i++ {
			line := m.LogLines[i]
			if len(line) > m.Width-6 {
				line = line[:m.Width-6]
			}
			b.WriteString(vLogLineStyle.Render(line))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	totalLines := len(m.LogLines)
	b.WriteString(vPaginationStyle.Render(
		fmt.Sprintf("日志: %d 行 | 当前: %d-%d | ↑↓/PgUp/PgDn滚动 | r: 刷新",
			totalLines, m.LogOffset+1, minInt(m.LogOffset+20, totalLines)),
	))

	return b.String()
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
