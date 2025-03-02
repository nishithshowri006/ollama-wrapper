package main

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	basestyle      = lipgloss.NewStyle()
	textinputstyle = basestyle.BorderForeground(lipgloss.Color("0"))
	spinnerstyle   = basestyle.Foreground(lipgloss.Color("5"))
	assistantstyle = basestyle.Foreground(lipgloss.Color("1"))
	userstyle      = basestyle.Foreground(lipgloss.Color("2"))
	viewportStyle  = basestyle.Border(lipgloss.RoundedBorder(), true)
)
