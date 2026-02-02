package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	DirName    = ".ssh-ogm"
	ConfigName = "config"
)

// Manager handles configuration file operations
type Manager struct {
	HomeDir string
}

// NewManager creates a new configuration manager
func NewManager() (*Manager, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home dir: %w", err)
	}
	return &Manager{HomeDir: home}, nil
}

// GetConfigPath returns the absolute path to the config file
func (m *Manager) GetConfigPath() string {
	return filepath.Join(m.HomeDir, DirName, ConfigName)
}

// Initialize ensures the config directory exists. Returns true if created for the first time.
func (m *Manager) Initialize() (bool, error) {
	configDir := filepath.Join(m.HomeDir, DirName)
	
	// Check if directory exists
	_, err := os.Stat(configDir)
	if os.IsNotExist(err) {
		// Create directory
		if err := os.MkdirAll(configDir, 0700); err != nil {
			return false, fmt.Errorf("failed to create config directory: %w", err)
		}
		
		// Create empty config file
		configPath := filepath.Join(configDir, ConfigName)
		f, err := os.Create(configPath)
		if err != nil {
			return false, fmt.Errorf("failed to create config file: %w", err)
		}
		defer f.Close()
		
		// Write a template or empty comment
		_, err = f.WriteString("# SSH OGM Configuration\n# Syntax: Alias { host: ... user: ... }\n\n")
		if err != nil {
             return false, fmt.Errorf("failed to write initial config: %w", err)
        }

		return true, nil // First run
	} else if err != nil {
		return false, fmt.Errorf("failed to stat config dir: %w", err)
	}

	return false, nil // Already exists
}
