package ssh

import (
	"crypto/rand"
	"crypto/rsa"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ssh"
)

func TestCheckConnection(t *testing.T) {
	// Start a mock SSH server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	defer listener.Close()

	// Get the random port
	addr := listener.Addr().String()
	_, port, _ := net.SplitHostPort(addr)

	// Accept connection in goroutine to simulate server
	go func() {
		config := &ssh.ServerConfig{
			NoClientAuth: true,
		}
		// We need a host key
		key, err := generateKey()
		if err != nil {
			return
		}
		config.AddHostKey(key)

		conn, _, _, err := ssh.NewServerConn(mustAccept(listener), config)
		if err == nil {
			conn.Close()
		}
	}()

	t.Run("Online Server", func(t *testing.T) {
		// Attempt check
		status := CheckConnection("127.0.0.1", port)
		assert.Equal(t, StatusOnline, status)
	})

	t.Run("Offline Server", func(t *testing.T) {
		// Random unused port (hopefully)
		status := CheckConnection("127.0.0.1", "54321")
		// Note: 54321 might be used, but unlikely to speak SSH.
		// If it's closed, it returns StatusOffline.
		// If it's open but not SSH, CheckConnection falls back to TCP (Online) or Ping (Online).
		// We assume 54321 is closed or unreachable.
		// Better: bind a port then close it to ensure it's closed?
		// Or just use a non-routable IP?
		status = CheckConnection("192.0.2.1", "22") // TEST-NET-1, reserved for docs/examples, usually unreachable
		assert.Equal(t, StatusOffline, status)
	})
}

func mustAccept(l net.Listener) net.Conn {
	c, err := l.Accept()
	if err != nil {
		panic(err)
	}
	return c
}

func generateKey() (ssh.Signer, error) {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return nil, err
	}
	return ssh.NewSignerFromKey(key)
}

// Rewriting test to use simpler TCP mock that mimics SSH roughly or just accepts connection
// Since CheckConnection logic says:
// 1. SSH Dial -> Success OR specific errors ("unable to authenticate", "handshake failed") -> ONLINE
// 2. TCP Dial -> Success -> ONLINE
// So a simple TCP listener is enough to pass the "TCP Fallback" or even "SSH Handshake" if validation is loose.
// But `CheckConnection` *prioritizes* SSH.
// Let's just listen on TCP. That should trigger the SSH Dial to fail (or succeed if we speak SSH)
// OR the TCP Dial fallback to succeed.
// Either way, it returns StatusOnline.
func TestCheckConnection_SimpleTCP(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	defer listener.Close()

	_, port, _ := net.SplitHostPort(listener.Addr().String())

	// Accept and just close (or hang)
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	status := CheckConnection("127.0.0.1", port)
	assert.Equal(t, StatusOnline, status)
}
