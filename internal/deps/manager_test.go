package deps

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsNetcatAvailable(t *testing.T) {
	// Backup original lookPath
	origLookPath := lookPath
	defer func() { lookPath = origLookPath }()

	tests := []struct {
		name     string
		mockFunc func(file string) (string, error)
		want     bool
	}{
		{
			name: "nc available",
			mockFunc: func(file string) (string, error) {
				if file == "nc" {
					return "/usr/bin/nc", nil
				}
				return "", fmt.Errorf("not found")
			},
			want: true,
		},
		{
			name: "ncat available",
			mockFunc: func(file string) (string, error) {
				if file == "ncat" {
					return "/usr/bin/ncat", nil
				}
				return "", fmt.Errorf("not found")
			},
			want: true,
		},
		{
			name: "none available",
			mockFunc: func(file string) (string, error) {
				return "", fmt.Errorf("not found")
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lookPath = tt.mockFunc
			got := IsNetcatAvailable()
			assert.Equal(t, tt.want, got)
		})
	}
}
