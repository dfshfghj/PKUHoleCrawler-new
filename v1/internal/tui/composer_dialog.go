package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type ComposerMode int

const (
	ComposerModePost ComposerMode = iota
	ComposerModeComment
)

type ComposerDialogModel struct {
	input       textinput.Model
	mode        ComposerMode
	title       string
	description string
	errorText   string
}

func NewComposerDialog() ComposerDialogModel {
	input := textinput.New()
	input.Placeholder = "输入内容"
	input.Focus()
	input.CharLimit = 2000
	input.Width = 50
	return ComposerDialogModel{input: input, title: "发布内容"}
}

func (m ComposerDialogModel) initialized() bool {
	return m.input.Width > 0
}

func (m *ComposerDialogModel) Configure(mode ComposerMode) {
	m.mode = mode
	m.errorText = ""
	m.input.SetValue("")
	m.input.Focus()
	if mode == ComposerModeComment {
		m.title = "发布评论"
		m.description = "输入评论内容（单行）"
	} else {
		m.title = "发布帖子"
		m.description = "输入帖子内容（单行）"
	}
}

func (m *ComposerDialogModel) SetError(err error) {
	if err == nil {
		m.errorText = ""
		return
	}
	m.errorText = err.Error()
}

func (m *ComposerDialogModel) Update(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return cmd
}

func (m ComposerDialogModel) Value() string {
	return strings.TrimSpace(m.input.Value())
}

func (m ComposerDialogModel) Mode() ComposerMode {
	return m.mode
}

func (m ComposerDialogModel) View(width int) string {
	var b strings.Builder
	b.WriteString(vDialogTitleStyle.Render(m.title))
	b.WriteString("\n\n")
	b.WriteString(m.description)
	b.WriteString("\n\n")
	b.WriteString(m.input.View())
	if m.errorText != "" {
		b.WriteString("\n\n")
		b.WriteString(vErrorStyle.Render(m.errorText))
	}
	b.WriteString("\n\n")
	b.WriteString(vDialogHelpStyle.Render("Enter: 提交 | Esc: 取消"))
	return b.String()
}
