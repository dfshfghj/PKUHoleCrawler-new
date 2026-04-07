package pages

import "github.com/charmbracelet/lipgloss"

var (
	pTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true).
			Padding(0, 1)

	pSubtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#B0B0B0")).
			Bold(true).
			Padding(0, 1)

	pStatusCard = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#4A4A4A")).
			Padding(1, 2).
			Background(lipgloss.Color("#1A1A1A"))

	pStatsCard = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#4A4A4A")).
			Padding(1, 2).
			Background(lipgloss.Color("#1A1A1A"))

	pStatLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	pStatValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)

	pStatusRunningStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true)

	pStatusStoppedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888888")).
				Bold(true)

	pButtonDefault = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Background(lipgloss.Color("#4A4A4A")).
			Padding(0, 2).
			Margin(0, 1).
			Bold(true)

	pButtonActive = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#FF69B4")).
			Padding(0, 2).
			Margin(0, 1).
			Bold(true)

	pButtonActiveDanger = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#000000")).
				Background(lipgloss.Color("#CC5588")).
				Padding(0, 2).
				Margin(0, 1).
				Bold(true)

	pButtonDisabled = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4A4A4A")).
			Background(lipgloss.Color("#1A1A1A")).
			Padding(0, 2).
			Margin(0, 1)

	pErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Padding(0, 1)

	pHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Padding(1, 0)

	pListItemStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(lipgloss.Color("#B0B0B0"))

	pListItemSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF69B4")).
				Background(lipgloss.Color("#2A1520")).
				Padding(0, 1).
				Bold(true)

	pSearchInput = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#4A4A4A")).
			Padding(0, 1).
			Width(60).
			Foreground(lipgloss.Color("#888888"))

	pSearchInputFocused = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("#FF69B4")).
				Padding(0, 1).
				Width(60).
				Foreground(lipgloss.Color("#FFFFFF"))

	pLoadingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Padding(1, 0)

	pEmptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4A4A4A")).
			Padding(1, 0)

	pPaginationStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888888")).
				Padding(1, 0)

	pPostPidStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#87CEEB")).
			Bold(true)

	pPostTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#B0B0B0")).
			Padding(0, 0, 0, 2)

	pPostMetaStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	pPostTimeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#90EE90"))

	pPostReplyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFA500"))

	pPostLikeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF69B4"))

	pDividerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4A4A4A"))

	pFormLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Width(12).
			Align(lipgloss.Right).
			Padding(0, 1)

	pFormInput = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#4A4A4A")).
			Padding(0, 1).
			Width(40)

	pFormInputFocused = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("#FF69B4")).
				Padding(0, 1).
				Width(40)

	pFormSaveBtn = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Background(lipgloss.Color("#4A4A4A")).
			Padding(0, 2).
			Margin(1, 0, 0, 13).
			Bold(true)

	pFormSaveActive = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#FF69B4")).
			Padding(0, 2).
			Margin(1, 0, 0, 13).
			Bold(true)

	pLogStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#B0B0B0")).
			Background(lipgloss.Color("#1A1A1A")).
			Padding(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#4A4A4A"))

	pLogLineStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#B0B0B0"))
)
