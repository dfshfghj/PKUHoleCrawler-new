package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var shanghaiLocation = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.Local
	}
	return loc
}()

func (m Model) View() string {
	m.ensureDialogModels()

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
	tabBarWidth := w - tabBarStyle.GetHorizontalFrameSize()
	if tabBarWidth < 1 {
		tabBarWidth = 1
	}
	tabBar := tabBarStyle.Width(tabBarWidth).Render(lipgloss.JoinHorizontal(lipgloss.Left, tabItems...))

	// Footer - fixed bottom
	footerText := fmt.Sprintf("TreeHole TUI v1.0 | h: 帮助 | %s", time.Now().In(shanghaiLocation).Format("15:04:05"))
	footer := lipgloss.NewStyle().
		Width(w).
		Align(lipgloss.Right).
		Foreground(colorMuted).
		Background(colorBg).
		Render(footerText)

	chromeHeight := lipgloss.Height(tabBar) + lipgloss.Height(footer)

	// Content area - fixed height between tab bar and footer
	contentHeight := h - chromeHeight
	if contentHeight < 1 {
		contentHeight = 1
	}
	content := m.renderContent(contentHeight)
	contentBlock := lipgloss.Place(
		w,
		contentHeight,
		lipgloss.Left,
		lipgloss.Top,
		content,
		lipgloss.WithWhitespaceBackground(colorBg),
	)

	body := lipgloss.JoinVertical(lipgloss.Left, tabBar, contentBlock, footer)

	// Dialog overlay
	if m.Dialog != DialogNone {
		dialog := m.renderDialog()
		body = lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, dialog,
			lipgloss.WithWhitespaceBackground(colorBg))
	}

	rendered := lipgloss.Place(
		w,
		h,
		lipgloss.Left,
		lipgloss.Top,
		body,
		lipgloss.WithWhitespaceBackground(colorBg),
	)
	rendered = baseStyle.Render(rendered)
	if m.Capture != nil {
		m.Capture.RecordFrame(rendered)
	}

	return rendered
}

func (m Model) renderContent(contentHeight int) string {
	switch m.Page {
	case PageHome:
		return m.renderHome(contentHeight)
	case PagePosts:
		return m.renderPosts(contentHeight)
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

func (m Model) renderHome(contentHeight int) string {
	return m.Home.View(m.Width, contentHeight)
}

func (m Model) renderPosts(contentHeight int) string {
	return m.Posts.View(m.Width, contentHeight)
}

func (m Model) renderConfigDialog() string {
	return m.renderDialogCard(m.ConfigDialog.View(m.Width))
}

func (m Model) renderLogsDialog() string {
	return m.renderDialogCard(m.LogsDialog.View(m.Width))
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

	return m.renderDialogCard(b.String())
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (m Model) renderDialogCard(content string) string {
	width := minInt(70, maxInt(40, m.Width-8))
	return dialogCard.Width(width).Render(content)
}
