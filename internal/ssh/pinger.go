package ssh

import (
	"context"
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

func CheckConnection(host, port string) ServerStatus {
	sshConfig := &ssh.ClientConfig{
		User:            "test", // User doesn't matter for handshake check
		Auth:            []ssh.AuthMethod{ssh.Password("test")},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // We just want to check reachability
		Timeout:         4 * time.Second,
	}

	target := net.JoinHostPort(host, cmdPort(port))
	conn, err := ssh.Dial("tcp", target, sshConfig)
	if err == nil {
		conn.Close()
		return StatusOnline
	}

	errMsg := err.Error()
	if strings.Contains(errMsg, "unable to authenticate") ||
		strings.Contains(errMsg, "handshake failed") ||
		strings.Contains(errMsg, "no common algorithm") {
		return StatusOnline
	}

	timeout := 2 * time.Second
	tcpConn, err := net.DialTimeout("tcp", target, timeout)
	if err == nil {
		tcpConn.Close()
		return StatusOnline
	}

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
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.CommandContext(ctx, "ping", "-n", "1", "-w", "1000", host)
	default:
		cmd = exec.CommandContext(ctx, "ping", "-c", "1", host)
	}

	err := cmd.Run()
	return err == nil
}
