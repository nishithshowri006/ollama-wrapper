package chat

import (
	"fmt"
)

func (m Model) View() string {
	switch m.WhichView {
	case ListView:
		return m.ListModel.View()
	case ChatView:
		if m.SpStatus == 0 {
			return fmt.Sprintf("%s\n%s", m.ViewportModel.View(), m.InputView.View())
		}
		return fmt.Sprintf("%s\n%s", m.ViewportModel.View(), m.Spinner.View())
	default:
		return ""
	}
}
