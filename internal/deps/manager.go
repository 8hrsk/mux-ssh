package deps

import (
	"fmt"
	"os/exec"
	"runtime"
)

// var for mocking in tests
var lookPath = exec.LookPath

// IsNetcatAvailable checks if nc, ncat, or netcat is in the PATH.
func IsNetcatAvailable() bool {
	cmds := []string{"nc", "ncat", "netcat"}
	for _, cmd := range cmds {
		if _, err := lookPath(cmd); err == nil {
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
	if _, err := exec.LookPath("winget"); err == nil {
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
	if _, err := exec.LookPath("apt-get"); err == nil {
		cmd := exec.Command("sudo", "apt-get", "install", "-y", "netcat")
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
