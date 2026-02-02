package tui

import (
	"fmt"
	"ssh-ogm/internal/config"
	"ssh-ogm/internal/ssh"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type DashboardModel struct {
	Configs    []config.HostConfig
	Statuses   map[string]ssh.ServerStatus
	Cursor     int
	Selected   *config.HostConfig
	Quitting   bool
	WindowSize tea.WindowSizeMsg
}

type TickMsg time.Time
type PingResultMsg ssh.ServerHealth

func NewDashboardModel(configs []config.HostConfig) DashboardModel {
	statuses := make(map[string]ssh.ServerStatus)
	for _, c := range configs {
		statuses[c.Alias] = ssh.StatusChecking
	}
	return DashboardModel{
		Configs:  configs,
		Statuses: statuses,
		Cursor:   0,
	}
}

// checkServerCmd creates a command to check a single server
func checkServerCmd(c config.HostConfig) tea.Cmd {
	return func() tea.Msg {
		status := ssh.CheckConnection(c.Host, c.Port)
		return PingResultMsg{
			Alias:  c.Alias,
			Status: status,
		}
	}
}

// checkAllServersCmd triggers checks for all servers
func checkAllServersCmd(configs []config.HostConfig) tea.Cmd {
	var cmds []tea.Cmd
	for _, c := range configs {
		cmds = append(cmds, checkServerCmd(c))
	}
	return tea.Batch(cmds...)
}

func (m DashboardModel) Init() tea.Cmd {
	return checkAllServersCmd(m.Configs)
}

func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.Quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < len(m.Configs)-1 {
				m.Cursor++
			}
		case "enter":
			selected := m.Configs[m.Cursor]
			m.Selected = &selected
			return m, tea.Quit
		case "r":
			// Refresh all
			return m, checkAllServersCmd(m.Configs)
		}

	case PingResultMsg:
		m.Statuses[msg.Alias] = msg.Status

	case tea.WindowSizeMsg:
		m.WindowSize = msg
	}

	return m, nil
}

func (m DashboardModel) View() string {
	if m.Selected != nil {
		return fmt.Sprintf("Connecting to %s...\n", m.Selected.Alias)
	}
	if m.Quitting {
		return "Goodbye!\n"
	}

	s := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render("SSH OGM Dashboard") + "\n"
	s += lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Select a server to connect (q to quit, r to reload)") + "\n\n"

	for i, c := range m.Configs {
		cursor := "  "
		if m.Cursor == i {
			cursor = "> "
		}

		// Status Dot
		statusDot := "●"
		statusStyle := lipgloss.NewStyle()
		
		switch m.Statuses[c.Alias] {
		case ssh.StatusChecking:
			statusStyle = statusStyle.Foreground(lipgloss.Color("33")) // Blue
		case ssh.StatusOnline:
			statusStyle = statusStyle.Foreground(lipgloss.Color("46")) // Green
		case ssh.StatusOffline:
			statusStyle = statusStyle.Foreground(lipgloss.Color("196")) // Red
		}
		
		dot := statusStyle.Render(statusDot)
		
		// Row Render
		row := fmt.Sprintf("%s %s %s (%s@%s)", cursor, dot, c.Alias, c.User, c.Host)
		
		if m.Cursor == i {
			s += lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86")).Render(row) + "\n"
		} else {
			s += row + "\n"
		}
	}

	return s
}
