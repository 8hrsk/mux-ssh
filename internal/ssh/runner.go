package ssh

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"ssh-ogm/internal/config"
)

// Connect connects to the host defined in the config, optionally via a proxy
func Connect(cfg config.HostConfig, proxyCfg *config.HostConfig) error {
	// Construct arguments
	args := []string{}
	// Port
	if cfg.Port != "" {
		args = append(args, "-p", cfg.Port)
	}
	// Identity
	if cfg.IdentityFile != "" {
		args = append(args, "-i", cfg.IdentityFile)
	}

	// Proxy Command Logic
	if proxyCfg != nil {
		// Detect nc/netcat
		// We assumes 'nc' is available on mac/linux as discussed.
		// Command: nc -x proxyHost:proxyPort host port (for SOCKS5)
		//          nc -X connect -x proxyHost:proxyPort host port (for HTTP/HTTPS CONNECT if suppported)
		// macOS nc supports -X (proto) -x (proxy).
		// Linux nc (openbsd) supports same. Traditional netcat might not.
		// User mentioned "type(http/socks5)".
		
		// Construct ProxyCommand string
		var proxyCmd string
		proxyHost := proxyCfg.Host
		proxyPort := proxyCfg.Port
		
		switch proxyCfg.Type {
		case "socks5":
			// nc -x proxy:port %h %p
			proxyCmd = fmt.Sprintf("nc -x %s:%s %%h %%p", proxyHost, proxyPort)
		case "http":
			// nc -X connect -x proxy:port %h %p
			proxyCmd = fmt.Sprintf("nc -X connect -x %s:%s %%h %%p", proxyHost, proxyPort)
		default:
			// Default to socks5 if unspecified or use safe default?
			// Let's assume socks5 as it's common for SSH.
			proxyCmd = fmt.Sprintf("nc -x %s:%s %%h %%p", proxyHost, proxyPort)
		}

		args = append(args, "-o", fmt.Sprintf("ProxyCommand=%s", proxyCmd))
	}

	// User@Host
	target := cfg.Host
	if cfg.User != "" {
		target = fmt.Sprintf("%s@%s", cfg.User, cfg.Host)
	}
	args = append(args, target)

	// Try to spawn in a new window based on OS
	var cmd *exec.Cmd
	
	switch runtime.GOOS {
	case "darwin":
        // Fallback to inline for now to ensure reliability, as 'open' is complex with args.
        cmd = exec.Command("ssh", args...)
        cmd.Stdin = os.Stdin
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr

	case "windows":
		// start ssh ...
		// "start" is a cmd shell command.
		// cmd /c start ssh -p ...
		winArgs := append([]string{"/c", "start", "ssh"}, args...)
		cmd = exec.Command("cmd", winArgs...)

	case "linux":
		// try gnome-terminal or x-terminal-emulator
		if path, err := exec.LookPath("gnome-terminal"); err == nil {
			// gnome-terminal -- ssh ...
			tArgs := append([]string{"--", "ssh"}, args...)
			cmd = exec.Command(path, tArgs...)
		} else if path, err := exec.LookPath("x-terminal-emulator"); err == nil {
			// x-terminal-emulator -e ssh ...
			tArgs := append([]string{"-e", "ssh"}, args...)
			cmd = exec.Command(path, tArgs...)

		} else {
			// Fallback inline
			cmd = exec.Command("ssh", args...)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		}
	default:
		// Fallback inline
		cmd = exec.Command("ssh", args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	return cmd.Run()
}
