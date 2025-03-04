package chat

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	basestyle         = lipgloss.NewStyle()
	textinputstyle    = basestyle.BorderForeground(lipgloss.Color("0"))
	spinnerstyle      = basestyle.Foreground(lipgloss.Color("5"))
	assistantstyle    = basestyle.Foreground(lipgloss.Color("1"))
	userstyle         = basestyle.Foreground(lipgloss.Color("2"))
	viewportStyle     = basestyle.Border(lipgloss.RoundedBorder(), true).Padding(2).Margin(1)
	textareaBlurStyle = basestyle.Foreground(lipgloss.Color("4")).Border(lipgloss.RoundedBorder(), true)
	textareaStyle     = basestyle.Foreground(lipgloss.Color("3")).Border(lipgloss.RoundedBorder(), true)
)
