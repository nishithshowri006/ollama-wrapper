package chat

import "fmt"

func (m *TerminalModel) View() string {
	if m.SpStatus == 0 {
		return fmt.Sprintf("%s\n%s", m.Viewport.View(), m.InputView.View())
	}
	return fmt.Sprintf("%s\n%s", m.Viewport.View(), m.Spinner.View())
}
