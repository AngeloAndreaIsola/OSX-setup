package verifier

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"setupper/internal/manifest"
)

var (
	titleStyle     = lipgloss.NewStyle().Bold(true)
	sectionStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).MarginTop(1)
	successStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	warningStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	errorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	dimStyle       = lipgloss.NewStyle().Faint(true)
)

type verifyMsg struct {
	results []CheckResult
	obs     *manifest.ObservedManifest
	err     error
}

type model struct {
	verifier *Verifier
	des      *manifest.DesiredManifest
	deep     bool
	spinner  spinner.Model
	results  []CheckResult
	obs      *manifest.ObservedManifest
	done     bool
	err      error
}

func initialModel(v *Verifier, des *manifest.DesiredManifest, deep bool) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	
	return model{
		verifier: v,
		des:      des,
		deep:     deep,
		spinner:  s,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.runVerify())
}

func (m model) runVerify() tea.Cmd {
	return func() tea.Msg {
		results, obs, err := m.verifier.Verify(context.Background(), m.des, m.deep)
		return verifyMsg{results: results, obs: obs, err: err}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
		
	case verifyMsg:
		m.err = msg.err
		m.results = msg.results
		m.obs = msg.obs
		m.done = true
		return m, tea.Quit
		
	case spinner.TickMsg:
		if m.done {
			return m, nil
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	
	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("❌ Verification failed: %v\n", m.err))
	}
	
	var b strings.Builder
	
	if !m.done {
		msg := "Verifying system state..."
		if m.deep {
			msg = "Running deep verification..."
		}
		b.WriteString(fmt.Sprintf("%s %s\n", m.spinner.View(), msg))
		return b.String()
	}
	
	mode := "fast mode"
	if m.deep {
		mode = "deep mode"
	}
	b.WriteString(titleStyle.Render("Verification Results") + " " + dimStyle.Render(fmt.Sprintf("[%s]", mode)) + "\n")
	
	// Group results by type
	grouped := make(map[string][]CheckResult)
	for _, res := range m.results {
		grouped[res.Resource.Type] = append(grouped[res.Resource.Type], res)
	}
	
	// Sort types
	var types []string
	for t := range grouped {
		types = append(types, t)
	}
	sort.Strings(types)
	
	for _, t := range types {
		headerName := t
		switch t {
		case "brew":
			headerName = "Homebrew Formulas"
		case "cask":
			headerName = "Homebrew Casks"
		case "mas":
			headerName = "App Store Apps"
		default:
			if len(t) > 0 {
				headerName = strings.ToUpper(t[:1]) + t[1:]
			}
		}
		b.WriteString(sectionStyle.Render(fmt.Sprintf("── %s ──", headerName)) + "\n")
		
		for _, res := range grouped[t] {
			icon := "❓"
			switch res.Status {
			case StatusPassed:
				icon = successStyle.Render("✅")
			case StatusWarning:
				icon = warningStyle.Render("⚠️")
			case StatusFailed:
				icon = errorStyle.Render("❌")
			}
			b.WriteString(fmt.Sprintf("%s %s:%s %s\n", icon, res.Resource.Type, res.Resource.Name, res.Message))
		}
	}
	
	passed, warnings, failed := Summarize(m.results)
	summary := fmt.Sprintf("%d passed, %d warnings, %d failed", passed, warnings, failed)
	
	var summaryStyled string
	if failed > 0 {
		summaryStyled = errorStyle.Render(summary)
	} else if warnings > 0 {
		summaryStyled = warningStyle.Render(summary)
	} else {
		summaryStyled = successStyle.Render(summary)
	}
	
	b.WriteString("\n" + summaryStyled + "\n")
	
	return b.String()
}

func RunVerifyTUI(v *Verifier, des *manifest.DesiredManifest, deep bool) ([]CheckResult, *manifest.ObservedManifest, error) {
	p := tea.NewProgram(initialModel(v, des, deep))
	m, err := p.Run()
	if err != nil {
		return nil, nil, err
	}
	
	mod := m.(model)
	return mod.results, mod.obs, mod.err
}
