package main

import "fmt"

func (m *TerminalModel) View() string {
	if m.Message == "" {
		return fmt.Sprintf("%s\n%s", m.Viewport.View(), m.TextInput.View())
	}

	if m.SpStatus == 0 {
		return fmt.Sprintf("%s\n%s", m.Viewport.View(), m.TextInput.View())
	}
	return fmt.Sprintf("%s\n%s", m.Viewport.View(), m.Spinner.View())
}
