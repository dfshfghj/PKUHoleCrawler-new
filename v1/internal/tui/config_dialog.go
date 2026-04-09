package tui

import (
	"strings"

	"treehole/internal/config"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const configSaveButtonIndex = 3

type ConfigDialogModel struct {
	inputs  []textinput.Model
	focus   int
	saving  bool
	saveOK  bool
	lastErr string
}

func NewConfigDialog(cfg *config.Config) ConfigDialogModel {
	inputs := make([]textinput.Model, 3)
	placeholders := []string{"用户名", "密码", "SecretKey"}
	values := []string{"", "", ""}
	if cfg != nil {
		values[0] = cfg.Username
		values[1] = cfg.Password
		values[2] = cfg.SecretKey
	}

	for i := range inputs {
		inputs[i] = textinput.New()
		inputs[i].Prompt = ""
		inputs[i].Placeholder = placeholders[i]
		inputs[i].SetValue(values[i])
		inputs[i].Width = 40
	}
	inputs[1].EchoMode = textinput.EchoPassword
	inputs[1].EchoCharacter = '*'
	inputs[2].EchoMode = textinput.EchoPassword
	inputs[2].EchoCharacter = '*'

	m := ConfigDialogModel{inputs: inputs}
	m.setFocus(0)
	return m
}

func (m ConfigDialogModel) initialized() bool {
	return len(m.inputs) == 3
}

func (m *ConfigDialogModel) setFocus(idx int) {
	if idx < 0 {
		idx = 0
	}
	if idx > configSaveButtonIndex {
		idx = configSaveButtonIndex
	}
	m.focus = idx
	for i := range m.inputs {
		if i == idx {
			m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
}

func (m *ConfigDialogModel) SetConfig(cfg *config.Config) {
	if !m.initialized() {
		*m = NewConfigDialog(cfg)
		return
	}
	if cfg == nil {
		return
	}
	m.inputs[0].SetValue(cfg.Username)
	m.inputs[1].SetValue(cfg.Password)
	m.inputs[2].SetValue(cfg.SecretKey)
	m.saveOK = false
	m.lastErr = ""
	m.setFocus(0)
}

func (m *ConfigDialogModel) SetSaving(saving bool) {
	m.saving = saving
}

func (m *ConfigDialogModel) SetSaveResult(err error) {
	m.saving = false
	if err != nil {
		m.saveOK = false
		m.lastErr = err.Error()
		return
	}
	m.saveOK = true
	m.lastErr = ""
}

func (m *ConfigDialogModel) FocusIndex() int {
	return m.focus
}

func (m *ConfigDialogModel) Username() string {
	if !m.initialized() {
		return ""
	}
	return m.inputs[0].Value()
}

func (m *ConfigDialogModel) Password() string {
	if !m.initialized() {
		return ""
	}
	return m.inputs[1].Value()
}

func (m *ConfigDialogModel) SecretKey() string {
	if !m.initialized() {
		return ""
	}
	return m.inputs[2].Value()
}

func (m *ConfigDialogModel) ToConfig() *config.Config {
	return &config.Config{
		Username:  m.Username(),
		Password:  m.Password(),
		SecretKey: m.SecretKey(),
	}
}

func (m *ConfigDialogModel) Update(msg tea.KeyMsg) tea.Cmd {
	switch msg.Type {
	case tea.KeyEscape:
		return nil
	case tea.KeyUp:
		if m.focus > 0 {
			m.setFocus(m.focus - 1)
		}
		return nil
	case tea.KeyDown:
		if m.focus < configSaveButtonIndex {
			m.setFocus(m.focus + 1)
		}
		return nil
	case tea.KeyEnter:
		if m.focus < configSaveButtonIndex {
			m.setFocus(configSaveButtonIndex)
		}
		return nil
	}

	if m.focus < configSaveButtonIndex {
		var cmd tea.Cmd
		m.inputs[m.focus], cmd = m.inputs[m.focus].Update(msg)
		return cmd
	}
	return nil
}

func (m ConfigDialogModel) View(width int) string {
	var b strings.Builder

	b.WriteString(vDialogTitleStyle.Render("配置管理"))
	b.WriteString("\n\n")
	b.WriteString(vSubtitleStyle.Render("config.json"))
	b.WriteString("\n\n")

	fields := []struct {
		label string
		input textinput.Model
	}{
		{"用户名:", m.inputs[0]},
		{"密码:", m.inputs[1]},
		{"SecretKey:", m.inputs[2]},
	}

	inputWidth := minInt(40, maxInt(24, width-28))
	for i, f := range fields {
		labelStyle := vFormLabelStyle
		inputStyle := vFormInput.Width(inputWidth)
		if m.focus == i {
			labelStyle = labelStyle.Foreground(colorAccent)
			inputStyle = vFormInputFocused.Width(inputWidth)
		}
		f.input.Width = inputWidth

		b.WriteString(lipgloss.JoinHorizontal(
			lipgloss.Top,
			labelStyle.Render(f.label),
			inputStyle.Render(f.input.View()),
		))
		b.WriteString("\n")
	}

	saveBtn := "保存配置"
	if m.focus == configSaveButtonIndex {
		b.WriteString("\n")
		b.WriteString(vFormSaveActive.Render(saveBtn))
	} else {
		b.WriteString("\n")
		b.WriteString(vFormSaveBtn.Render(saveBtn))
	}

	if m.saving {
		b.WriteString("\n")
		b.WriteString(vLoadingStyle.Render("保存中..."))
	}
	if m.saveOK {
		b.WriteString("\n")
		b.WriteString(vStatusRunningStyle.Render("配置已保存!"))
	}
	if m.lastErr != "" {
		b.WriteString("\n")
		b.WriteString(vErrorStyle.Render("错误: " + m.lastErr))
	}

	b.WriteString("\n\n")
	b.WriteString(vDialogHelpStyle.Render("↑↓: 选择 | Enter: 前往保存/保存 | Esc: 关闭"))
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
