package listui

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nishithshowri006/ollama-wrapper/internal/ollama"
)

type ListModel struct {
	ListView     list.Model
	LlmModelList []ItemModel
	Client       *ollama.Ollama
}
type ItemModel struct {
	ollama.ModelsMetadata
}

func (i ItemModel) Title() string {
	return i.Name
}

func (i ItemModel) Description() string {
	return fmt.Sprintf("Size: %d\nFamily: %s\nQuantizationLevels: %s\n",
		i.Size, i.Details.Family, i.Details.QuantizationLevel,
	)
}

func (i ItemModel) FilterValue() string {
	return i.Name
}

func InitilizeModel() *ListModel {
	client := ollama.NewClient("", "")
	lv := list.New(nil, list.NewDefaultDelegate(), 0, 0)
	return &ListModel{
		Client:   client,
		ListView: lv,
	}
}

func (m *ListModel) Init() tea.Cmd {
	return nil
}

func (m *ListModel) getModelList() []ollama.ModelsMetadata {
	mlist, err := m.Client.ListModels()
	if err != nil {
		log.Fatal(err)
	}
	return mlist
}

func (m *ListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		id := list.NewDefaultDelegate()

		llmList := m.getModelList()
		items := make([]list.Item, len(llmList))
		id.Styles = list.NewDefaultItemStyles()
		id.Styles.SelectedDesc.Width(msg.Width)
		id.ShowDescription = true
		m.ListView.SetDelegate(id)
		for i := range len(llmList) {
			items[i] = ItemModel{llmList[i]}
		}
		m.ListView.SetItems(items)
		m.ListView.SetHeight(msg.Height)
		m.ListView.SetWidth(msg.Width)
	case tea.KeyMsg:
		switch msg.String() {
		case tea.KeyCtrlC.String(), "q":
			return m, tea.Quit
		}
	}
	m.ListView, cmd = m.ListView.Update(msg)
	return m, cmd
}
func (m *ListModel) View() string {
	return m.ListView.View()
}
