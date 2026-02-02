package tui

import (
	"fmt"
	"ssh-ogm/internal/config"
	"ssh-ogm/internal/ssh"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ViewState defines which tab is active
type ViewState int

const (
	ViewServers ViewState = iota
	ViewProxies
)

type DashboardModel struct {
	ConfigManager *config.Manager
	
	Configs    []config.HostConfig
	Proxies    []config.HostConfig
	
	ServerStatuses   map[string]ssh.ServerStatus
	ProxyStatuses    map[string]ssh.ServerStatus
	
	Cursor     int
	Results    map[string]ssh.ServerStatus // Temporary holding for batch updates? No, direct map update is fine.

	ActiveView ViewState
	Selected   *config.HostConfig
	Quitting   bool
	WindowSize tea.WindowSizeMsg

	// For feedback
	Message string
}

type PingResultMsg ssh.ServerHealth

// NewDashboardModel initializes the dashboard with servers and proxies
func NewDashboardModel(configs, proxies []config.HostConfig, mgr *config.Manager) DashboardModel {
	sStatuses := make(map[string]ssh.ServerStatus)
	for _, c := range configs {
		sStatuses[c.Alias] = ssh.StatusChecking
	}
	
	pStatuses := make(map[string]ssh.ServerStatus)
	for _, p := range proxies {
		pStatuses[p.Alias] = ssh.StatusChecking
	}

	return DashboardModel{
		ConfigManager:  mgr,
		Configs:        configs,
		Proxies:        proxies,
		ServerStatuses: sStatuses,
		ProxyStatuses:  pStatuses,
		Cursor:         0,
		ActiveView:     ViewServers,
	}
}

// checkServerCmd creates a command to check a single host
func checkHostCmd(c config.HostConfig) tea.Cmd {
	return func() tea.Msg {
		status := ssh.CheckConnection(c.Host, c.Port)
		return PingResultMsg{
			Alias:  c.Alias,
			Status: status,
		}
	}
}

// checkBatch triggers checks for a list of configs
func checkBatch(configs []config.HostConfig) tea.Cmd {
	var cmds []tea.Cmd
	for _, c := range configs {
		cmds = append(cmds, checkHostCmd(c))
	}
	return tea.Batch(cmds...)
}

func (m DashboardModel) Init() tea.Cmd {
	return tea.Batch(
		checkBatch(m.Configs),
		checkBatch(m.Proxies),
	)
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
			// Max cursor depends on active view
			maxLen := len(m.Configs)
			if m.ActiveView == ViewProxies {
				maxLen = len(m.Proxies)
			}
			if m.Cursor < maxLen-1 {
				m.Cursor++
			}

		case "left", "h", "right", "l", "tab":
			// Toggle View
			if m.ActiveView == ViewServers {
				m.ActiveView = ViewProxies
			} else {
				m.ActiveView = ViewServers
			}
			m.Cursor = 0
			m.Message = ""

		case "enter":
			if m.ActiveView == ViewServers && len(m.Configs) > 0 {
				selected := m.Configs[m.Cursor]
				m.Selected = &selected
				return m, tea.Quit
			}
			// Proxies are not "connectable" directly in the main flow, 
			// though user said "navigate... select...". 
			// Usually we don't SSH *into* a proxy to do work, we use it. 
			// For now, let's say Enter does nothing on proxies or maybe shows details?
			// User request: "When user selects servers - they see... what proxy is used".
			// Proxy page: "List of all proxies... and their status".
			// Doesn't explicitly say "Connect to proxy".
			
		case "r":
			// Reload: Set all current view items to Checking (Blue) and re-trigger
			if m.ActiveView == ViewServers {
				for k := range m.ServerStatuses {
					m.ServerStatuses[k] = ssh.StatusChecking
				}
				return m, checkBatch(m.Configs)
			} else {
				for k := range m.ProxyStatuses {
					m.ProxyStatuses[k] = ssh.StatusChecking
				}
				return m, checkBatch(m.Proxies)
			}

		case "a":
			// Add Template
			var err error
			if m.ActiveView == ViewServers {
				err = m.ConfigManager.AppendTemplate(config.ConfigName, "new_server", false)
				if err == nil {
					// Open Editor
					config.OpenEditor(m.ConfigManager.GetConfigPath(), config.EditorSystem) 
					// Ideally we pause the TUI or ask user to restart?
					m.Message = "Template added! Restart CLI to apply."
					return m, tea.Quit // User usually needs to edit it now
				}
			} else {
				err = m.ConfigManager.AppendTemplate(config.ProxiesName, "new_proxy", true)
				if err == nil {
					config.OpenEditor(m.ConfigManager.GetProxiesPath(), config.EditorSystem)
					m.Message = "Proxy template added! Restart CLI to apply."
					return m, tea.Quit
				}
			}
			if err != nil {
				m.Message = fmt.Sprintf("Error adding template: %v", err)
			}
		}

	case PingResultMsg:
		// Update status map
		if _, ok := m.ServerStatuses[msg.Alias]; ok {
			m.ServerStatuses[msg.Alias] = msg.Status
		} else if _, ok := m.ProxyStatuses[msg.Alias]; ok {
			m.ProxyStatuses[msg.Alias] = msg.Status
		}

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
		if m.Message != "" {
			return m.Message + "\n"
		}
		return "Goodbye!\n"
	}

	// Header / Tabs
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render("SSH OGM Dashboard")
	
	tabServer := "Servers"
	tabProxy := "Proxies"
	
	activeTabStyle := lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, true, false).BorderForeground(lipgloss.Color("205")).Foreground(lipgloss.Color("205")).Bold(true).Padding(0, 1)
	inactiveTabStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Padding(0, 1)

	if m.ActiveView == ViewServers {
		tabServer = activeTabStyle.Render(tabServer)
		tabProxy = inactiveTabStyle.Render(tabProxy)
	} else {
		tabServer = inactiveTabStyle.Render(tabServer)
		tabProxy = activeTabStyle.Render(tabProxy)
	}

	tabs := lipgloss.JoinHorizontal(lipgloss.Top, tabServer, tabProxy)
	header := fmt.Sprintf("%s\n\n%s\n", title, tabs)
	s := header
	
	// Content
	list := m.Configs
	statuses := m.ServerStatuses
	if m.ActiveView == ViewProxies {
		list = m.Proxies
		statuses = m.ProxyStatuses
	}

	if len(list) == 0 {
		s += "\n  No items found. Press 'a' to add a template.\n"
	}

	for i, c := range list {
		cursor := "  "
		if m.Cursor == i {
			cursor = "> "
		}

		// Status Dot
		statusDot := "●"
		statusStyle := lipgloss.NewStyle()
		
		stat, _ := statuses[c.Alias]
		switch stat {
		case ssh.StatusChecking:
			statusStyle = statusStyle.Foreground(lipgloss.Color("33")) // Blue
		case ssh.StatusOnline:
			statusStyle = statusStyle.Foreground(lipgloss.Color("46")) // Green
		case ssh.StatusOffline:
			statusStyle = statusStyle.Foreground(lipgloss.Color("196")) // Red
		}
		dot := statusStyle.Render(statusDot)
		
		// Row Render
		var details string
		if m.ActiveView == ViewServers {
			details = fmt.Sprintf("%s (%s@%s)", c.Alias, c.User, c.Host)
			if c.Proxy != "" {
				details += lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(fmt.Sprintf(" via %s", c.Proxy))
			}
		} else {
			details = fmt.Sprintf("%s (%s:%s %s)", c.Alias, c.Host, c.Port, c.Type)
		}

		row := fmt.Sprintf("%s %s %s", cursor, dot, details)
		
		if m.Cursor == i {
			s += lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86")).Render(row) + "\n"
		} else {
			s += row + "\n"
		}
	}

	s += "\n(q: quit, r: reload, a: add, tab: switch view)\n"
	if m.Message != "" {
		s += lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render(m.Message) + "\n"
	}

	return s
}
