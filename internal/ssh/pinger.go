package ssh

import (
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

type ServerStatus int

const (
	StatusChecking ServerStatus = iota
	StatusOnline
	StatusOffline
)

// ServerHealth holds the status of a server
type ServerHealth struct {
	Alias  string
	Status ServerStatus
	Error  error
}

// CheckConnection attempts to check server availability via:
// 1. SSH Handshake (Gold standard)
// 2. TCP Connect (Fallback if SSH auth fails fast)
// 3. ICMP Ping (Fallback if port is closed/filtered)
func CheckConnection(host, port string) ServerStatus {
	// 1. Try SSH Handshake
	// We use "none" auth method. If server is up, it will reject us with "unable to authenticate",
	// which counts as GREEN (Online).
	sshConfig := &ssh.ClientConfig{
		User:            "test", // User doesn't matter for handshake check
		Auth:            []ssh.AuthMethod{ssh.Password("test")}, 
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // We just want to check reachability
		Timeout:         4 * time.Second,
	}

	target := fmt.Sprintf("%s:%s", host, cmdPort(port))
	conn, err := ssh.Dial("tcp", target, sshConfig)
	if err == nil {
		conn.Close()
		return StatusOnline
	}

	// Analyze SSH error
	// If the error implies we reached the server but failed auth, it's ONLINE.
	// Common errors: "ssh: handshake failed", "unable to authenticate"
	errMsg := err.Error()
	if strings.Contains(errMsg, "unable to authenticate") || 
	   strings.Contains(errMsg, "handshake failed") ||
	   strings.Contains(errMsg, "no common algorithm") {
		return StatusOnline
	}

	// 2. Fallback: Simple TCP Dial (in case SSH Dial logic was too strict)
	timeout := 2 * time.Second
	tcpConn, err := net.DialTimeout("tcp", target, timeout)
	if err == nil {
		tcpConn.Close()
		return StatusOnline
	}

	// 3. Fallback: ICMP Ping
	if checkPing(host) {
		return StatusOnline
	}

	return StatusOffline
}

func cmdPort(p string) string {
	if p == "" {
		return "22"
	}
	return p
}

func checkPing(host string) bool {
	var cmd *exec.Cmd
	
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("ping", "-n", "1", "-w", "1000", host) // -n 1 count, -w 1000ms timeout
	default:
		// Linux/Mac: -c 1 count, -W 2 timeout (seconds)
		// Note: macOS ping -W is in milliseconds, Linux is seconds? 
		// Actually macOS: -W wait time in ms. Linux: -W timeout in seconds. 
		// To be safe for both, let's use -W 1000 (valid for mac) or -W 1 (valid for linux)?
		// Linux 'ping' often supports -W in seconds. macOS supports -t timeout.
		// Let's use a simpler heuristic or just -c 1.
		cmd = exec.Command("ping", "-c", "1", host) 
	}
	
	err := cmd.Run()
	return err == nil
}
