package pages

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

type HomeModel struct {
	LoggedIn      bool
	LoginUser     string
	CrawlerState  CrawlState
	CrawlerStart  time.Time
	TotalPosts    int
	TotalComments int
	LastCrawlPage int
	LastCrawlTime time.Duration
	Width         int
	Height        int
	ButtonIdx     int
	LastError     string
}

func NewHomeModel() HomeModel {
	return HomeModel{
		ButtonIdx: 0,
	}
}

func (m HomeModel) View() string {
	var b strings.Builder

	b.WriteString(pTitleStyle.Render("PKUHole Crawler"))
	b.WriteString("\n\n")

	if m.LoggedIn {
		b.WriteString(pStatusCard.Render(
			lipgloss.JoinHorizontal(lipgloss.Top,
				pStatLabelStyle.Render("登录状态: "),
				pStatusRunningStyle.Render("已登录"),
				pStatLabelStyle.Render("  用户: "),
				pStatValueStyle.Render(m.LoginUser),
			),
		))
	} else {
		b.WriteString(pStatusCard.Render(
			pStatusStoppedStyle.Render("未登录"),
		))
	}

	var crawlerStatus string
	switch m.CrawlerState {
	case CrawlRunning:
		crawlerStatus = pStatusRunningStyle.Render("运行中")
	case CrawlStopped:
		crawlerStatus = pStatusStoppedStyle.Render("已停止")
	case CrawlError:
		crawlerStatus = pStatusStoppedStyle.Render("错误")
	}

	elapsed := "0s"
	if m.CrawlerState == CrawlRunning && !m.CrawlerStart.IsZero() {
		elapsed = time.Since(m.CrawlerStart).Round(time.Second).String()
	}

	b.WriteString("\n")
	b.WriteString(pStatusCard.Render(
		lipgloss.JoinHorizontal(lipgloss.Top,
			pStatLabelStyle.Render("爬虫状态: "),
			crawlerStatus,
			pStatLabelStyle.Render("  运行时长: "),
			pStatValueStyle.Render(elapsed),
		),
	))

	b.WriteString("\n")
	b.WriteString(pStatsCard.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.JoinHorizontal(lipgloss.Top,
				pStatLabelStyle.Render("帖子总数: "),
				pStatValueStyle.Render(fmt.Sprintf("%d", m.TotalPosts)),
				pStatLabelStyle.Render("  评论总数: "),
				pStatValueStyle.Render(fmt.Sprintf("%d", m.TotalComments)),
			),
			lipgloss.JoinHorizontal(lipgloss.Top,
				pStatLabelStyle.Render("上次爬取: "),
				pStatValueStyle.Render(fmt.Sprintf("第%d页", m.LastCrawlPage)),
				pStatLabelStyle.Render("  耗时: "),
				pStatValueStyle.Render(m.LastCrawlTime.Round(time.Millisecond).String()),
			),
		),
	))

	b.WriteString("\n")
	buttons := []string{"启动爬虫", "停止爬虫"}

	var btns []string
	for i, label := range buttons {
		disabled := (i == 0 && m.CrawlerState == CrawlRunning) ||
			(i == 1 && m.CrawlerState != CrawlRunning)
		if disabled {
			btns = append(btns, pButtonDisabled.Render(label))
		} else if m.ButtonIdx == i {
			if i == 1 {
				btns = append(btns, pButtonActiveDanger.Render(label))
			} else {
				btns = append(btns, pButtonActive.Render(label))
			}
		} else {
			btns = append(btns, pButtonDefault.Render(label))
		}
	}
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Center, btns...))

	if m.LastError != "" {
		b.WriteString("\n\n")
		b.WriteString(pErrorStyle.Render("错误: " + m.LastError))
	}

	b.WriteString("\n\n")
	b.WriteString(pHelpStyle.Render("Tab: 切换页面 | ←→: 选择按钮 | Enter: 确认 | q: 退出"))

	return b.String()
}
