package tui

import (
	"fmt"
	"ssh-ogm/internal/config"
	"ssh-ogm/internal/deps"
	"ssh-ogm/internal/ssh"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ViewState defines which tab is active
type ViewState int

const (
	ViewServers ViewState = iota
	ViewProxies
	ViewEditorPrompt
	ViewInstallPrompt
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

	// Editor Prompt State
	PromptChoice int
	EditorTarget string

	// Installation State
	Installing    bool
	InstallError  error
	Spinner       spinner.Model

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

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return DashboardModel{
		ConfigManager:  mgr,
		Configs:        configs,
		Proxies:        proxies,
		ServerStatuses: sStatuses,
		ProxyStatuses:  pStatuses,
		Cursor:         0,
		ActiveView:     ViewServers,
		Spinner:        s,
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

// Msg types for installation
type installFinishedMsg struct{ err error }

func installNetcatCmd() tea.Cmd {
	return func() tea.Msg {
		err := deps.InstallNetcat()
		return installFinishedMsg{err}
	}
}

func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle Spinner
	var cmd tea.Cmd
	if m.Installing {
		m.Spinner, cmd = m.Spinner.Update(msg)
		return m, cmd
	}

	// Handle Install Prompt
	if m.ActiveView == ViewInstallPrompt {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "y", "Y":
				m.Installing = true
				return m, tea.Batch(m.Spinner.Tick, installNetcatCmd())
			case "n", "N", "esc", "q":
				m.ActiveView = ViewServers
				m.Message = "Proxy setup cancelled. Netcat is required."
				return m, nil
			}
		case installFinishedMsg:
			m.Installing = false
			if msg.err != nil {
				m.Message = fmt.Sprintf("Installation failed: %v", msg.err)
				m.ActiveView = ViewServers
			} else {
				m.Message = "Netcat installed successfully!"
				// Resume where we left off? 
				// For simplicity, go back to Proxy view or Servers.
				m.ActiveView = ViewProxies
			}
		}
		return m, nil
	}

	// Handle Editor Prompt Inputs
	if m.ActiveView == ViewEditorPrompt {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "up", "k":
				if m.PromptChoice > 0 {
					m.PromptChoice--
				}
			case "down", "j":
				if m.PromptChoice < 1 {
					m.PromptChoice++
				}
			case "enter":
				// Launch Editor
				editorType := config.EditorSystem
				if m.PromptChoice == 1 {
					editorType = config.EditorTerminal
				}
				config.OpenEditor(m.EditorTarget, editorType)
				m.Message = "Configuration edited. Please restart to apply changes."
				return m, tea.Quit
			case "esc", "q":
				// Cancel
				m.ActiveView = ViewServers // Default back to whatever? Or store previous?
				// Simple fallback:
				m.ActiveView = ViewServers
				m.Message = "Edit cancelled."
			}
		}
		return m, nil
	}

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
			// Proxies are not "connectable" directly
			
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
			// Add
			var targetFile string
			if m.ActiveView == ViewServers {
				m.ConfigManager.AppendTemplate(config.ConfigName, "new_server", false)
				targetFile = m.ConfigManager.GetConfigPath()
			} else {
				// Check Dependencies for Proxy
				if !deps.IsNetcatAvailable() {
					m.ActiveView = ViewInstallPrompt
					return m, nil
				}
				m.ConfigManager.AppendTemplate(config.ProxiesName, "new_proxy", true)
				targetFile = m.ConfigManager.GetProxiesPath()
			}
			
			// Setup Prompt
			m.EditorTarget = targetFile
			m.PromptChoice = 0
			m.ActiveView = ViewEditorPrompt
			m.Message = "Template added. Select editor:"

		case "e":
			// Edit
			// Check Dependencies if editing proxies (implied usage)
			if m.ActiveView == ViewProxies && !deps.IsNetcatAvailable() {
				m.ActiveView = ViewInstallPrompt
				return m, nil
			}

			var targetFile string
			if m.ActiveView == ViewServers {
				targetFile = m.ConfigManager.GetConfigPath()
			} else {
				targetFile = m.ConfigManager.GetProxiesPath()
			}

			// Setup Prompt
			m.EditorTarget = targetFile
			m.PromptChoice = 0
			m.ActiveView = ViewEditorPrompt
			m.Message = "Select editor to open config:"
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
		return ""
	}

	// Install Prompt View
	if m.ActiveView == ViewInstallPrompt {
		s := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render("Dependency Missing") + "\n\n"
		
		if m.Installing {
			s += fmt.Sprintf("%s Installing Netcat...\n", m.Spinner.View())
		} else {
			s += "Netcat is required to support Proxy tunneling.\n"
			s += "The system could not find 'nc', 'ncat', or 'netcat' in your PATH.\n\n"
			s += "Do you want to attempt automatic installation? (y/n)\n"
			s += lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("(Windows: Winget/Scoop, Mac: Brew, Linux: Apt/Dnf/Pacman)") + "\n"
		}
		
		if m.Message != "" {
			s += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(m.Message) + "\n"
		}

		return s
	}

	// Editor Prompt View
	if m.ActiveView == ViewEditorPrompt {
		s := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render("Edit Configuration") + "\n\n"
		if m.Message != "" {
			s += m.Message + "\n\n"
		}
		
		cursor := "> "
		noCursor := "  "
		
		// Option 0
		if m.PromptChoice == 0 {
			s += lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Render(cursor + "System Editor (Visual)") + "\n"
		} else {
			s += noCursor + "System Editor (Visual)\n"
		}

		// Option 1
		if m.PromptChoice == 1 {
			s += lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Render(cursor + "Terminal Editor (Vim/Nano)") + "\n"
		} else {
			s += noCursor + "Terminal Editor (Vim/Nano)\n"
		}

		s += "\n(Enter to select, Esc to cancel)\n"
		return s
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

	s += "\n(q: quit, r: reload, a: add, e: edit, tab: switch view)\n"
	if m.Message != "" {
		s += lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render(m.Message) + "\n"
	}

	return s
}
