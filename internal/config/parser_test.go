package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []HostConfig
		wantErr bool
	}{
		{
			name: "valid server config",
			input: `
server1 {
    host: 192.168.1.1
    user: root
    port: 22
    identity: ~/.ssh/id_rsa
}
`,
			want: []HostConfig{
				{
					Alias:        "server1",
					Host:         "192.168.1.1",
					User:         "root",
					Port:         "22",
					IdentityFile: "~/.ssh/id_rsa",
				},
			},
			wantErr: false,
		},
		{
			name: "valid proxy config",
			input: `
proxy1 {
    host: proxy.example.com
    port: 1080
    type: socks5
}
`,
			want: []HostConfig{
				{
					Alias: "proxy1",
					Host:  "proxy.example.com",
					Port:  "1080",
					Type:  "socks5",
				},
			},
			wantErr: false,
		},
		{
			name: "server with proxy",
			input: `
server2 {
    host: 10.0.0.1
    proxy: proxy1
}
`,
			want: []HostConfig{
				{
					Alias: "server2",
					Host:  "10.0.0.1",
					Proxy: "proxy1",
				},
			},
			wantErr: false,
		},
		{
			name: "multiple entries",
			input: `
s1 {
    host: 1.1.1.1
}
s2 {
    host: 2.2.2.2
}
`,
			want: []HostConfig{
				{Alias: "s1", Host: "1.1.1.1"},
				{Alias: "s2", Host: "2.2.2.2"},
			},
			wantErr: false,
		},
		{
			name:    "invalid syntax - missing brace",
			input:   `s1 { host: 1.1.1.1`,
			want:    nil,
			wantErr: false,
		},
		{
			name: "invalid syntax - key without value",
			input: `
s1 {
    host
}
`,
			want: []HostConfig{
				{Alias: "s1"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			got, err := Parse(r)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
