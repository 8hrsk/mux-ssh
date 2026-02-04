package config

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type HostConfig struct {
	Alias        string
	Host         string
	User         string
	Port         string
	IdentityFile string

	// Proxy specific
	Proxy    string // Name of the proxy to use (for Servers)
	Password string // (for Proxies)
	Type     string // socks5, http (for Proxies)
}

// Parse reads the configuration from the reader and returns a list of HostConfigs
func Parse(r io.Reader) ([]HostConfig, error) {
	scanner := bufio.NewScanner(r)
	var configs []HostConfig
	var currentConfig *HostConfig

	lineNum := 0
	inBlock := false

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasSuffix(line, "{") {
			if inBlock {
				return nil, fmt.Errorf("line %d: nested blocks or missing closing brace not supported", lineNum)
			}
			alias := strings.TrimSpace(strings.TrimSuffix(line, "{"))
			if alias == "" {
				return nil, fmt.Errorf("line %d: missing alias before '{'", lineNum)
			}
			currentConfig = &HostConfig{Alias: alias}
			inBlock = true
			continue
		}

		if line == "}" {
			if !inBlock {
				return nil, fmt.Errorf("line %d: unexpected closing brace", lineNum)
			}
			if currentConfig != nil {
				configs = append(configs, *currentConfig)
				currentConfig = nil
			}
			inBlock = false
			continue
		}

		if inBlock {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("line %d: expected 'key: value'", lineNum)
			}
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			switch key {
			case "host":
				currentConfig.Host = value
			case "user":
				currentConfig.User = value
			case "port":
				currentConfig.Port = value
			case "identity":
				currentConfig.IdentityFile = value
			case "proxy":
				currentConfig.Proxy = value
			case "password":
				currentConfig.Password = value
			case "type":
				currentConfig.Type = value
			default:
				return nil, fmt.Errorf("line %d: unknown key '%s'", lineNum, key)
			}
			continue
		}

		return nil, fmt.Errorf("line %d: unexpected text outside block: %s", lineNum, line)
	}

	if inBlock {
		return nil, fmt.Errorf("unexpected end of file: missing closing brace")
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return configs, nil
}
