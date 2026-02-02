package tui

import (
	"fmt"
	"ssh-ogm/internal/config"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type FirstRunModel struct {
	ConfigPath string
	Choice     int
	Quitting   bool
	Chosen     bool
	Err        error
}

func NewFirstRunModel(configPath string) FirstRunModel {
	return FirstRunModel{
		ConfigPath: configPath,
		Choice:     0, // 0 = System, 1 = Terminal
	}
}

func (m FirstRunModel) Init() tea.Cmd {
	return nil
}

func (m FirstRunModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.Quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.Choice > 0 {
				m.Choice--
			}
		case "down", "j":
			if m.Choice < 1 {
				m.Choice++
			}
		case "enter":
			m.Chosen = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m FirstRunModel) View() string {
	if m.Err != nil {
		return fmt.Sprintf("Error: %v\n", m.Err)
	}
	if m.Chosen {
		return "Opening editor...\n"
	}
	if m.Quitting {
		return "Setup skipped.\n"
	}

	var s string
	s += lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render("Welcome to SSH OGM!") + "\n\n"
	s += fmt.Sprintf("Configuration created at: %s\n\n", m.ConfigPath)
	s += "How would you like to edit the configuration?\n\n"

	cursor := "> "
	noCursor := "  "

	// Option 0: System Editor
	if m.Choice == 0 {
		s += lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Render(cursor + "Open in System Editor (Visual)") + "\n"
	} else {
		s += noCursor + "Open in System Editor (Visual)\n"
	}

	// Option 1: Terminal Editor
	if m.Choice == 1 {
		s += lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Render(cursor + "Open in Terminal Editor (Vim/Nano)") + "\n"
	} else {
		s += noCursor + "Open in Terminal Editor (Vim/Nano)\n"
	}

	s += "\n(Use arrow keys to navigate, Enter to select)\n"

	return s
}

// Result returns the selected editor type
func (m FirstRunModel) Result() config.EditorType {
	if m.Choice == 0 {
		return config.EditorSystem
	}
	return config.EditorTerminal
}
