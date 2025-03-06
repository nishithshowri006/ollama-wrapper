package chat

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	basestyle             = lipgloss.NewStyle()
	textinputstyle        = basestyle.BorderForeground(lipgloss.Color("0"))
	spinnerstyle          = basestyle.Foreground(lipgloss.Color("5"))
	assistantstyle        = basestyle.Foreground(lipgloss.Color("1"))
	assistantcontentstyle = assistantstyle.Border(lipgloss.HiddenBorder(), true)
	userstyle             = basestyle.Foreground(lipgloss.Color("2"))
	usercontentstyle      = userstyle.Border(lipgloss.HiddenBorder(), true)
	viewportStyle         = basestyle.MarginLeft(1).MarginRight(1)
	textareaBlurStyle     = basestyle.Foreground(lipgloss.Color("4")).Border(lipgloss.RoundedBorder(), true)
	textareaStyle         = basestyle.Foreground(lipgloss.Color("3")).Border(lipgloss.RoundedBorder(), true)
)
