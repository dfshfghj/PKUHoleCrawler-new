package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Black/white theme with pink accent
	colorBlack     = lipgloss.Color("#000000")
	colorDarkGray  = lipgloss.Color("#1A1A1A")
	colorMidGray   = lipgloss.Color("#4A4A4A")
	colorGray      = lipgloss.Color("#888888")
	colorLightGray = lipgloss.Color("#B0B0B0")
	colorWhite     = lipgloss.Color("#FFFFFF")
	colorPink      = lipgloss.Color("#FF69B4")
	colorPinkDark  = lipgloss.Color("#CC5588")
	colorPinkBg    = lipgloss.Color("#2A1520")

	baseStyle = lipgloss.NewStyle().
			Foreground(colorWhite).
			Background(colorBlack)

	tabBarStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(colorMidGray).
			BorderBottom(true).
			Padding(0, 1)

	tabItemStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(colorGray)

	tabItemActiveStyle = lipgloss.NewStyle().
				Padding(0, 2).
				Foreground(colorPink).
				Bold(true).
				Background(colorPinkBg)

	contentStyle = lipgloss.NewStyle().
			Padding(1, 2)

	footerStyle = lipgloss.NewStyle().
			Foreground(colorGray).
			Border(lipgloss.NormalBorder()).
			BorderForeground(colorMidGray).
			BorderTop(true).
			Padding(0, 1)

	vTitleStyle = lipgloss.NewStyle().
			Foreground(colorWhite).
			Bold(true).
			Padding(0, 1)

	vSubtitleStyle = lipgloss.NewStyle().
			Foreground(colorLightGray).
			Bold(true).
			Padding(0, 1)

	vStatusCard = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorMidGray).
			Padding(1, 2).
			Background(colorDarkGray)

	vStatsCard = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorMidGray).
			Padding(1, 2).
			Background(colorDarkGray)

	vStatLabelStyle = lipgloss.NewStyle().
			Foreground(colorGray)

	vStatValueStyle = lipgloss.NewStyle().
			Foreground(colorWhite).
			Bold(true)

	vStatusRunningStyle = lipgloss.NewStyle().
				Foreground(colorWhite).
				Bold(true)

	vStatusStoppedStyle = lipgloss.NewStyle().
				Foreground(colorGray).
				Bold(true)

	vButtonDefault = lipgloss.NewStyle().
			Foreground(colorGray).
			Background(colorMidGray).
			Padding(0, 2).
			Margin(0, 1).
			Bold(true)

	vButtonActive = lipgloss.NewStyle().
			Foreground(colorBlack).
			Background(colorPink).
			Padding(0, 2).
			Margin(0, 1).
			Bold(true)

	vButtonActiveDanger = lipgloss.NewStyle().
				Foreground(colorBlack).
				Background(colorPinkDark).
				Padding(0, 2).
				Margin(0, 1).
				Bold(true)

	vButtonDisabled = lipgloss.NewStyle().
			Foreground(colorMidGray).
			Background(colorDarkGray).
			Padding(0, 2).
			Margin(0, 1)

	vErrorStyle = lipgloss.NewStyle().
			Foreground(colorGray).
			Padding(0, 1)

	vHelpStyle = lipgloss.NewStyle().
			Foreground(colorGray).
			Padding(1, 0)

	vListItemStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(colorLightGray)

	vListItemSelectedStyle = lipgloss.NewStyle().
				Foreground(colorPink).
				Background(colorPinkBg).
				Padding(0, 1).
				Bold(true)

	vSearchInput = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(colorMidGray).
			Padding(0, 1).
			Width(60).
			Foreground(colorGray)

	vSearchInputFocused = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(colorPink).
				Padding(0, 1).
				Width(60).
				Foreground(colorWhite)

	vLoadingStyle = lipgloss.NewStyle().
			Foreground(colorGray).
			Padding(1, 0)

	vEmptyStyle = lipgloss.NewStyle().
			Foreground(colorMidGray).
			Padding(1, 0)

	vPaginationStyle = lipgloss.NewStyle().
				Foreground(colorGray).
				Padding(1, 0)

	vPostPidStyle = lipgloss.NewStyle().
			Foreground(colorWhite).
			Bold(true)

	vPostTextStyle = lipgloss.NewStyle().
			Foreground(colorLightGray).
			Padding(0, 0, 0, 2)

	vPostMetaStyle = lipgloss.NewStyle().
			Foreground(colorGray)

	vDividerStyle = lipgloss.NewStyle().
			Foreground(colorMidGray)

	vFormLabelStyle = lipgloss.NewStyle().
			Foreground(colorGray).
			Width(12).
			Align(lipgloss.Right).
			Padding(0, 1)

	vFormInput = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(colorMidGray).
			Padding(0, 1).
			Width(40).
			Foreground(colorLightGray)

	vFormInputFocused = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(colorPink).
				Padding(0, 1).
				Width(40).
				Foreground(colorWhite)

	vFormSaveBtn = lipgloss.NewStyle().
			Foreground(colorGray).
			Background(colorMidGray).
			Padding(0, 2).
			Margin(1, 0, 0, 13).
			Bold(true)

	vFormSaveActive = lipgloss.NewStyle().
			Foreground(colorBlack).
			Background(colorPink).
			Padding(0, 2).
			Margin(1, 0, 0, 13).
			Bold(true)

	vLogLineStyle = lipgloss.NewStyle().
			Foreground(colorLightGray)
)