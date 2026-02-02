package ssh

import (
	"fmt"
	"net"
	"time"
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

// CheckConnection attempts to dial the address to check availability
func CheckConnection(host, port string) ServerStatus {
	timeout := 2 * time.Second
	target := fmt.Sprintf("%s:%s", host, port)
	
	conn, err := net.DialTimeout("tcp", target, timeout)
	if err != nil {
		return StatusOffline
	}
	conn.Close()
	return StatusOnline
}
