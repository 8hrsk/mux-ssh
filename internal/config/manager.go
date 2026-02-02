package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"io"
)

const (
	DirName     = ".ssh-ogm"
	ConfigName  = "config"
	ProxiesName = "proxies.conf"
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

// GetConfigPath returns the absolute path to the server config file
func (m *Manager) GetConfigPath() string {
	return filepath.Join(m.HomeDir, DirName, ConfigName)
}

// GetProxiesPath returns the absolute path to the proxies config file
func (m *Manager) GetProxiesPath() string {
	return filepath.Join(m.HomeDir, DirName, ProxiesName)
}

// Headers for documentation
const ServerConfigHeader = `# SSH OGM Server Configuration
# Syntax: Alias { host: ... user: ... }
# Example:
# myserver {
#    host: 1.2.3.4
#    user: root
#    port: 22
#    proxy: myproxy # Optional
# }

`

const ProxyConfigHeader = `# SSH OGM Proxy Configuration
# Syntax: Alias { host: ... port: ... type: ... }
# Types: socks5, http
# Example:
# myproxy {
#    host: proxy.example.com
#    port: 1080
#    type: socks5
#    user: user # Optional
#    password: pass # Optional
# }

`

// Initialize ensures the config directory and files exist with documentation.
func (m *Manager) Initialize() (bool, error) {
	configDir := filepath.Join(m.HomeDir, DirName)
	
	// Check/Create directory
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0700); err != nil {
			return false, fmt.Errorf("failed to create config directory: %w", err)
		}
	}

	// Ensure files exist and have headers
	firstRun, err := m.ensureFile(ConfigName, ServerConfigHeader)
	if err != nil {
		return false, err
	}

	_, err = m.ensureFile(ProxiesName, ProxyConfigHeader)
	if err != nil {
		return false, err
	}

	return firstRun, nil
}

// ensureFile checks if a file exists. If not, creates it with header.
// If it exists, checks if header is present (simple check) and prepends if missing.
// Returns true if created new.
func (m *Manager) ensureFile(name, header string) (bool, error) {
	path := filepath.Join(m.HomeDir, DirName, name)
	created := false

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return false, fmt.Errorf("failed to open %s: %w", name, err)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return false, err
	}

	if stat.Size() == 0 {
		if _, err := f.WriteString(header); err != nil {
			return false, fmt.Errorf("failed to write header to %s: %w", name, err)
		}
		return true, nil
	}

	// Simple check: Read first few bytes to see if it starts with "# SSH OGM"
	// For robust "prepending", we would need to read the whole file and rewrite it.
	// Given the user constraint "If documentation gets distorted... add it at the top",
	// a simple read-check-prepend is safer.

	buf := make([]byte, len(header))
	n, _ := f.ReadAt(buf, 0)
	currentHeader := string(buf[:n])

	// If the file content doesn't start with the expected header (or at least the first line)
	// We should prepend. comparing full header might be brittle if they edit it slightly.
	// Let's check first line.
	expectedFirstLine := strings.Split(header, "\n")[0]
	actualFirstLine := strings.Split(currentHeader, "\n")[0]

	if actualFirstLine != expectedFirstLine {
		// Prepend logic
		content, err := io.ReadAll(f) // Read from current offset (which is 0 because of OpenFile?)
		// wait, OpenFile doesn't reset offset? O_RDWR puts it at 0.
		// logic: ReadAt didn't move offset.
		// Let's confirm: ReadAll reads from current position.
		// To be safe, Seek to 0.
		f.Seek(0, 0)
		content, err = io.ReadAll(f)
		if err != nil {
			return false, fmt.Errorf("failed to read content of %s: %w", name, err)
		}
		
		f.Seek(0, 0)
		f.Truncate(0)
		if _, err := f.WriteString(header + string(content)); err != nil {
			return false, fmt.Errorf("failed to prepend header to %s: %w", name, err)
		}
	}

	return created, nil
}

// AppendTemplate adds a new template block to the specified file
func (m *Manager) AppendTemplate(filename, alias string, isProxy bool) error {
	path := filepath.Join(m.HomeDir, DirName, filename)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	var tmpl string
	if isProxy {
		tmpl = fmt.Sprintf("\n%s {\n    host: proxy.example.com\n    port: 1080\n    type: socks5\n}\n", alias)
	} else {
		tmpl = fmt.Sprintf("\n%s {\n    host: 1.2.3.4\n    user: root\n    port: 22\n}\n", alias)
	}

	if _, err := f.WriteString(tmpl); err != nil {
		return err
	}
	return nil
}
