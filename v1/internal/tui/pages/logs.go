package pages

import (
	"fmt"
	"strings"
)

type LogsModel struct {
	Width     int
	Height    int
	LogLines  []string
	Offset    int
	Loading   bool
	Viewport  int
	LastError string
}

func NewLogsModel() LogsModel {
	return LogsModel{
		Viewport: 20,
	}
}

func (m LogsModel) View() string {
	var b strings.Builder

	b.WriteString(pTitleStyle.Render("运行日志"))
	b.WriteString("\n\n")

	if m.Loading {
		b.WriteString(pLoadingStyle.Render("加载日志中..."))
		return b.String()
	}

	if m.LastError != "" {
		b.WriteString(pErrorStyle.Render("错误: " + m.LastError))
		b.WriteString("\n")
	}

	if len(m.LogLines) == 0 {
		b.WriteString(pEmptyStyle.Render("暂无日志"))
	} else {
		end := m.Offset + m.Viewport
		if end > len(m.LogLines) {
			end = len(m.LogLines)
		}

		for i := m.Offset; i < end; i++ {
			line := m.LogLines[i]
			if len(line) > m.Width-6 {
				line = line[:m.Width-6]
			}
			b.WriteString(pLogLineStyle.Render(line))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	totalLines := len(m.LogLines)
	b.WriteString(pPaginationStyle.Render(
		fmt.Sprintf("日志: %d 行 | 当前: %d-%d | ↑↓/PgUp/PgDn滚动 | r: 刷新",
			totalLines, m.Offset+1, min(m.Offset+m.Viewport, totalLines)),
	))

	b.WriteString("\n")
	b.WriteString(pHelpStyle.Render("Tab: 切换页面 | q: 退出"))

	return b.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
