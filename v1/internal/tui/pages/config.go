package pages

import (
	"strings"

	"treehole/internal/config"

	"github.com/charmbracelet/lipgloss"
)

type ConfigModel struct {
	Width       int
	Height      int
	Config      *config.Config
	Username    string
	Password    string
	SecretKey   string
	FieldIdx    int
	Saving      bool
	Saved       bool
	SaveSuccess bool
	LastError   string
}

func NewConfigModel(cfg *config.Config) ConfigModel {
	return ConfigModel{
		Config:    cfg,
		Username:  cfg.Username,
		Password:  cfg.Password,
		SecretKey: cfg.SecretKey,
	}
}

func (m ConfigModel) View() string {
	var b strings.Builder

	b.WriteString(pTitleStyle.Render("配置管理"))
	b.WriteString("\n\n")

	b.WriteString(pSubtitleStyle.Render("config.json"))
	b.WriteString("\n\n")

	fields := []struct {
		label string
		value string
	}{
		{"用户名:", maskField(m.Username, false)},
		{"密码:", maskField(m.Password, true)},
		{"SecretKey:", maskField(m.SecretKey, true)},
	}

	for i, f := range fields {
		labelStyle := pFormLabelStyle
		inputStyle := pFormInput

		if m.FieldIdx == i {
			labelStyle = labelStyle.Foreground(lipgloss.Color("#7D56F4"))
			inputStyle = pFormInputFocused
		}

		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top,
			labelStyle.Render(f.label),
			inputStyle.Render(f.value),
		))
		b.WriteString("\n")
	}

	// Save button
	saveBtn := "保存配置"
	if m.FieldIdx == 3 {
		b.WriteString("\n")
		b.WriteString(pFormSaveActive.Render(saveBtn))
	} else {
		b.WriteString("\n")
		b.WriteString(pFormSaveBtn.Render(saveBtn))
	}

	if m.Saving {
		b.WriteString("\n")
		b.WriteString(pLoadingStyle.Render("保存中..."))
	}

	if m.SaveSuccess {
		b.WriteString("\n")
		b.WriteString(pStatusRunningStyle.Render("配置已保存!"))
	}

	if m.LastError != "" {
		b.WriteString("\n")
		b.WriteString(pErrorStyle.Render("错误: " + m.LastError))
	}

	b.WriteString("\n\n")
	b.WriteString(pHelpStyle.Render("Tab: 切换页面 | ↑↓: 选择字段 | Enter: 编辑/保存 | Esc: 取消编辑 | q: 退出"))

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

func (m ConfigModel) ToConfig() *config.Config {
	return &config.Config{
		Username:  m.Username,
		Password:  m.Password,
		SecretKey: m.SecretKey,
	}
}
