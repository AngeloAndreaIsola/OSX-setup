package installer

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"setupper/internal/planner"
	"setupper/internal/runner"
)

type applyMsg struct {
	idx   int
	err   error
}

type model struct {
	plan      *planner.ExecutionPlan
	runner    runner.Runner
	options   ApplyOptions
	spinner   spinner.Model
	results   []Result
	current   int
	done      bool
	quitting  bool
}

func initialModel(p *planner.ExecutionPlan, r runner.Runner, opts ApplyOptions) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	
	results := make([]Result, len(p.Steps))
	for i, step := range p.Steps {
		results[i] = Result{Step: step, Error: nil}
	}

	return model{
		plan:    p,
		runner:  r,
		options: opts,
		spinner: s,
		results: results,
		current: 0,
	}
}

func (m model) Init() tea.Cmd {
	if len(m.plan.Steps) == 0 {
		return tea.Quit
	}
	return tea.Batch(m.spinner.Tick, m.runCurrent())
}

func (m model) runCurrent() tea.Cmd {
	if m.current >= len(m.plan.Steps) {
		return nil
	}
	idx := m.current
	step := m.plan.Steps[idx]
	r := m.runner
	
	return func() tea.Msg {
		err := ExecuteStep(context.Background(), r, step)
		return applyMsg{idx: idx, err: err}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		}
		
	case applyMsg:
		m.results[msg.idx].Error = msg.err
		if msg.err != nil && m.options.FailFast {
			m.done = true
			return m, tea.Quit
		}
		
		m.current++
		if m.current >= len(m.plan.Steps) {
			m.done = true
			return m, tea.Quit
		}
		return m, m.runCurrent()
		
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	
	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return "Apply cancelled.\n"
	}
	
	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Applying Execution Plan\n\n"))
	
	for i, r := range m.results {
		if i < m.current {
			if r.Error != nil {
				b.WriteString(fmt.Sprintf("❌ %s %s: %v\n", r.Step.Action, r.Step.Resource.Name, r.Error))
			} else {
				b.WriteString(fmt.Sprintf("✅ %s %s\n", r.Step.Action, r.Step.Resource.Name))
			}
		} else if i == m.current && !m.done {
			b.WriteString(fmt.Sprintf("%s %s %s...\n", m.spinner.View(), r.Step.Action, r.Step.Resource.Name))
		} else {
			b.WriteString(fmt.Sprintf("   %s %s\n", r.Step.Action, r.Step.Resource.Name))
		}
	}
	
	if m.done {
		b.WriteString("\nApply complete!\n")
	}
	
	return b.String()
}

func RunApplyTUI(plan *planner.ExecutionPlan, r runner.Runner, opts ApplyOptions) ([]Result, error) {
	if len(plan.Steps) == 0 {
		return nil, nil
	}
	
	p := tea.NewProgram(initialModel(plan, r, opts))
	m, err := p.Run()
	if err != nil {
		return nil, err
	}
	
	mod := m.(model)
	return mod.results, nil
}
