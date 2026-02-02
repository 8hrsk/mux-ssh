package config

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	input := `
# Comment
prod-db {
    host: 192.168.1.10
    user: root
    port: 22
}

dev-web {
    host: dev.example.com
    user: admin
    identity: ~/.ssh/id_ed25519
}
`
	r := strings.NewReader(input)
	configs, err := Parse(r)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(configs) != 2 {
		t.Errorf("expected 2 configs, got %d", len(configs))
	}

	// Check prod-db
	if configs[0].Alias != "prod-db" || configs[0].Host != "192.168.1.10" || configs[0].User != "root" || configs[0].Port != "22" {
		t.Errorf("prod-db parsed incorrectly: %+v", configs[0])
	}

	// Check dev-web
	if configs[1].Alias != "dev-web" || configs[1].Host != "dev.example.com" || configs[1].User != "admin" || configs[1].IdentityFile != "~/.ssh/id_ed25519" {
		t.Errorf("dev-web parsed incorrectly: %+v", configs[1])
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"Nested blocks", "a { b { } }"},
		{"Unexpected close", "}"},
		{"Missing alias", "{ host: x }"},
		{"Bad pair", "a { host }"},
		{"Unknown key", "a { foo: bar }"},
		{"Unclosed block", "a { host: x"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			_, err := Parse(r)
			if err == nil {
				t.Errorf("expected error for %s, got nil", tt.name)
			}
		})
	}
}
