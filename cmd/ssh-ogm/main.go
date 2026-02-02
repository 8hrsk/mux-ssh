package main

import (
	"fmt"
	"os"
	"ssh-ogm/internal/config"
	"ssh-ogm/internal/ssh"
	"ssh-ogm/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Initialize Config Manager
	mgr, err := config.NewManager()
	if err != nil {
		fmt.Printf("Error initializing config manager: %v\n", err)
		os.Exit(1)
	}

	// Check/Create Config
	isFirstRun, err := mgr.Initialize()
	if err != nil {
		fmt.Printf("Error creating config: %v\n", err)
		os.Exit(1)
	}

	if isFirstRun {
		// Run First Run TUI
		p := tea.NewProgram(tui.NewFirstRunModel(mgr.GetConfigPath()))
		m, err := p.Run()
		if err != nil {
			fmt.Printf("Alas, there's been an error: %v", err)
			os.Exit(1)
		}

		// Handle user choice
		if model, ok := m.(tui.FirstRunModel); ok && model.Chosen {
			err := config.OpenEditor(mgr.GetConfigPath(), model.Result())
			if err != nil {
				fmt.Printf("Error opening editor: %v\n", err)
			}
			fmt.Println("Configuration file opened. Please restart SSH OGM after editing.")
			os.Exit(0)
		} else if model.Quitting {
			fmt.Println("Setup skipped.")
			os.Exit(0)
		}
	}

	// TODO: Normal Run - Load Config and Start Dashboard
	fmt.Println("Loading configuration...")
	f, err := os.Open(mgr.GetConfigPath())
	if err != nil {
		fmt.Printf("Error opening config: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	configs, err := config.Parse(f)
	if err != nil {
		fmt.Printf("Error parsing config: %v\n", err)
		os.Exit(1)
	}

	// Parse Proxy Config
	fmt.Println("Loading proxy configuration...")
	pf, err := os.Open(mgr.GetProxiesPath())
	if err != nil {
		fmt.Printf("Error opening proxies config: %v\n", err)
		os.Exit(1)
	}
	defer pf.Close()

	proxies, err := config.Parse(pf)
	if err != nil {
		fmt.Printf("Error parsing proxies config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Loaded %d servers and %d proxies.\n", len(configs), len(proxies))

	// Start Dashboard
	p := tea.NewProgram(tui.NewDashboardModel(configs, proxies, mgr))
	m, err := p.Run()
	if err != nil {
		fmt.Printf("Error running dashboard: %v\n", err)
		os.Exit(1)
	}

	if dashboard, ok := m.(tui.DashboardModel); ok && dashboard.Selected != nil {
		fmt.Printf("Connecting to %s...\n", dashboard.Selected.Alias)
		
		// Find Proxy
		var proxyCfg *config.HostConfig
		if dashboard.Selected.Proxy != "" {
			for _, p := range proxies {
				if p.Alias == dashboard.Selected.Proxy {
					proxyCfg = &p
					break
				}
			}
			if proxyCfg == nil {
				fmt.Printf("Warning: Proxy '%s' not found in proxies.conf\n", dashboard.Selected.Proxy)
				// Proceed without proxy or exit? 
				// Better to error out as the user intended a proxy.
				fmt.Println("Aborting connection.")
				os.Exit(1)
			}
			fmt.Printf("Using proxy: %s (%s:%s)\n", proxyCfg.Alias, proxyCfg.Host, proxyCfg.Port)
		}

		err := ssh.Connect(*dashboard.Selected, proxyCfg)
		if err != nil {
			fmt.Printf("Error connecting: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("Exiting.")
	}
}
