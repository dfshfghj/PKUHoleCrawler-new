package tui

import (
	"fmt"
	"strings"

	"treehole/internal/models"

	tea "github.com/charmbracelet/bubbletea"
)

type TagsDialogModel struct {
	tags      []models.Tag
	selected  int
	errorText string
}

func NewTagsDialog() TagsDialogModel {
	return TagsDialogModel{tags: []models.Tag{}}
}

func (m TagsDialogModel) initialized() bool {
	return m.tags != nil
}

func (m *TagsDialogModel) SetTags(tags []models.Tag) {
	m.tags = tags
	m.selected = 0
	m.errorText = ""
}

func (m *TagsDialogModel) SetError(err error) {
	if err == nil {
		m.errorText = ""
		return
	}
	m.errorText = err.Error()
}

func (m *TagsDialogModel) Update(msg tea.KeyMsg) {
	switch msg.String() {
	case "up", "k":
		if m.selected > 0 {
			m.selected--
		}
	case "down", "j":
		if m.selected < len(m.tags)-1 {
			m.selected++
		}
	}
}

func (m TagsDialogModel) SelectedTag() *models.Tag {
	if len(m.tags) == 0 || m.selected < 0 || m.selected >= len(m.tags) {
		return nil
	}
	tag := m.tags[m.selected]
	return &tag
}

func (m TagsDialogModel) View(width int) string {
	var b strings.Builder
	b.WriteString(vDialogTitleStyle.Render("标签筛选"))
	b.WriteString("\n\n")
	if m.errorText != "" {
		b.WriteString(vErrorStyle.Render(m.errorText))
		b.WriteString("\n\n")
	}
	if len(m.tags) == 0 {
		b.WriteString(vEmptyStyle.Render("暂无标签"))
	} else {
		for i, tag := range m.tags {
			prefix := "  "
			if i == m.selected {
				prefix = "→ "
			}
			name := tag.Label
			if name == "" {
				name = tag.Name
			}
			b.WriteString(fmt.Sprintf("%s%s (#%d)\n", prefix, name, tag.ID))
		}
	}
	b.WriteString("\n")
	b.WriteString(vDialogHelpStyle.Render("↑↓: 选择 | Enter: 应用 | c: 清除 | Esc: 关闭"))
	return b.String()
}
