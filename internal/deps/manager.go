package deps

import (
	"fmt"
	"os/exec"
	"runtime"
)

// IsNetcatAvailable checks if nc, ncat, or netcat is in the PATH.
func IsNetcatAvailable() bool {
	cmds := []string{"nc", "ncat", "netcat"}
	for _, cmd := range cmds {
		if _, err := exec.LookPath(cmd); err == nil {
			return true
		}
	}
	return false
}

// InstallNetcat attempts to install netcat based on the OS.
func InstallNetcat() error {
	switch runtime.GOOS {
	case "windows":
		return installWindows()
	case "darwin":
		return installMac()
	case "linux":
		return installLinux()
	default:
		return fmt.Errorf("automatic installation not supported on %s", runtime.GOOS)
	}
}

func installWindows() error {
	// Try Winget first (Standard on Windows 10/11)
	if _, err := exec.LookPath("winget"); err == nil {
		// Insecure.Nmap includes ncat
		cmd := exec.Command("winget", "install", "Insecure.Nmap", "--accept-source-agreements", "--accept-package-agreements")
		return cmd.Run()
	}

	// Try Scoop
	if _, err := exec.LookPath("scoop"); err == nil {
		cmd := exec.Command("scoop", "install", "ncat")
		return cmd.Run()
	}

	// Fallback: Open Browser
	return openBrowser("https://nmap.org/download.html")
}

func installMac() error {
	if _, err := exec.LookPath("brew"); err == nil {
		cmd := exec.Command("brew", "install", "netcat")
		return cmd.Run()
	}
	return fmt.Errorf("homebrew not found")
}

func installLinux() error {
	// Debian/Ubuntu
	if _, err := exec.LookPath("apt-get"); err == nil {
		// sudo is likely required, which might prompt password in terminal.
		// TUI might block this or it might work if we attach Stdin.
		// Simple attempt:
		cmd := exec.Command("sudo", "apt-get", "install", "-y", "netcat")
		// We can't easily attach stdin in bubbletea's exec unless we suspend.
		// For now, let's look for user-level install or error out.
		// Actually, `pkexec` or forcing a terminal prompt might be needed.
		// Let's rely on standard command execution. If it fails, user does it manually.
		return cmd.Run()
	}
	// Fedora/RHEL
	if _, err := exec.LookPath("dnf"); err == nil {
		cmd := exec.Command("sudo", "dnf", "install", "-y", "nmap-ncat")
		return cmd.Run()
	}
	// Arch
	if _, err := exec.LookPath("pacman"); err == nil {
		cmd := exec.Command("sudo", "pacman", "-S", "--noconfirm", "gnu-netcat")
		return cmd.Run()
	}
	
	return fmt.Errorf("package manager not found or supported")
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Run()
}
