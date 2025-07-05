package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "version command",
			args: []string{"version"},
			want: "devslot version dev\n",
		},
		{
			name: "version flag",
			args: []string{"-v"},
			want: "devslot version dev",
		},
		{
			name: "version long flag",
			args: []string{"--version"},
			want: "devslot version dev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			app := NewApp(&buf)
			err := app.Run(tt.args)
			// Version flag returns an error like help flag
			if tt.name != "version command" && err == nil {
				t.Errorf("expected error for version flag")
			}

			output := buf.String()
			if !strings.Contains(output, tt.want) {
				t.Errorf("expected output %q, got %q", tt.want, output)
			}
		})
	}
}
