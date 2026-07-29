package planner

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"setupper/internal/manifest"
)

type item struct {
	resource manifest.Resource
	action   Action
	selected bool
}

type model struct {
	cursor   int
	items    []item
	quitting bool
	done     bool
}

func initialModel(missing []manifest.Resource, unmanaged []manifest.Resource) model {
	var items []item
	for _, r := range missing {
		items = append(items, item{resource: r, action: ActionInstall, selected: true})
	}
	for _, r := range unmanaged {
		items = append(items, item{resource: r, action: ActionRemove, selected: false}) // defaults to false (don't remove)
	}
	return model{items: items}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}

		case " ":
			if len(m.items) > 0 {
				m.items[m.cursor].selected = !m.items[m.cursor].selected
			}
			
		case "enter":
			// if space also triggered enter, we don't want it to skip. But standard tea allows multiple keys.
			// Space is toggle, Enter is confirm.
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return "Operation cancelled.\n"
	}
	if m.done {
		return "Plan confirmed!\n"
	}
	if len(m.items) == 0 {
		return "Nothing to do! Your system matches the desired state.\n(Press q to quit)\n"
	}

	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Review Execution Plan\n\n"))
	b.WriteString("Use ↑/↓ to navigate, Space to toggle, Enter to confirm.\n\n")

	for i, it := range m.items {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		
		check := " "
		if it.selected {
			check = "x"
		}
		
		color := lipgloss.Color("2") // Green
		if it.action == ActionRemove {
			color = lipgloss.Color("1") // Red
		}
		style := lipgloss.NewStyle().Foreground(color)

		label := fmt.Sprintf("[%s] %s %s (%s)", check, it.action, it.resource.Name, it.resource.Type)
		b.WriteString(fmt.Sprintf("%s %s\n", cursor, style.Render(label)))
	}

	return b.String()
}

// RunInteractiveChecklist runs the Bubble Tea checklist and returns selected resources.
func RunInteractiveChecklist(missing, unmanaged []manifest.Resource) ([]manifest.Resource, []manifest.Resource, error) {
	if len(missing) == 0 && len(unmanaged) == 0 {
		return nil, nil, nil
	}

	p := tea.NewProgram(initialModel(missing, unmanaged))
	m, err := p.Run()
	if err != nil {
		return nil, nil, err
	}
	
	mod := m.(model)
	if mod.quitting {
		return nil, nil, fmt.Errorf("cancelled by user")
	}

	var installs []manifest.Resource
	var removes []manifest.Resource
	for _, it := range mod.items {
		if it.selected {
			if it.action == ActionInstall {
				installs = append(installs, it.resource)
			} else if it.action == ActionRemove {
				removes = append(removes, it.resource)
			}
		}
	}

	return installs, removes, nil
}
