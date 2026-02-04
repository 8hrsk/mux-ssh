package ssh

import (
	"ssh-ogm/internal/config"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConstructSSHArgs(t *testing.T) {
	tests := []struct {
		name     string
		cfg      config.HostConfig
		proxyCfg *config.HostConfig
		want     []string
	}{
		{
			name: "simple host",
			cfg: config.HostConfig{
				Alias: "test",
				Host:  "1.2.3.4",
				User:  "root",
				Port:  "22",
			},
			want: []string{"-p", "22", "root@1.2.3.4"},
		},
		{
			name: "host with identity",
			cfg: config.HostConfig{
				Alias:        "test",
				Host:         "example.com",
				User:         "user",
				Port:         "2222",
				IdentityFile: "~/.ssh/id_rsa",
			},
			want: []string{"-p", "2222", "-i", "~/.ssh/id_rsa", "user@example.com"},
		},
		{
			name: "host with socks5 proxy",
			cfg: config.HostConfig{
				Alias: "test",
				Host:  "10.0.0.1",
				User:  "admin",
				Port:  "22",
			},
			proxyCfg: &config.HostConfig{
				Host: "proxy.local",
				Port: "1080",
				Type: "socks5",
			},
			want: []string{
				"-p", "22",
				"-o", "ProxyCommand=nc -x proxy.local:1080 %h %p",
				"admin@10.0.0.1",
			},
		},
		{
			name: "host with http proxy",
			cfg: config.HostConfig{
				Alias: "test",
				Host:  "10.0.0.1",
				User:  "root",
				Port:  "22",
			},
			proxyCfg: &config.HostConfig{
				Host: "gateway.local",
				Port: "8080",
				Type: "http",
			},
			want: []string{
				"-p", "22",
				"-o", "ProxyCommand=nc -X connect -x gateway.local:8080 %h %p",
				"root@10.0.0.1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConstructSSHArgs(tt.cfg, tt.proxyCfg)
			assert.Equal(t, tt.want, got)
		})
	}
}
