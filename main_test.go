package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestHelp(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "help flag",
			args: []string{"--help"},
			want: []string{
				"Usage: devslot <command>",
				"Development environment manager for multi-repo worktrees",
				"Commands:",
				"boilerplate",
				"init",
				"create",
				"destroy",
				"reload",
				"list",
				"doctor",
				"version",
				"Flags:",
				"-h, --help",
				"-v, --version",
			},
		},
		{
			name: "help shorthand",
			args: []string{"-h"},
			want: []string{
				"Usage: devslot <command>",
				"Development environment manager for multi-repo worktrees",
			},
		},
		{
			name: "no arguments shows help",
			args: []string{},
			want: []string{
				"Usage: devslot <command>",
				"Development environment manager for multi-repo worktrees",
				"Commands:",
				"Flags:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			app := NewApp(&buf)
			err := app.Run(tt.args)
			// Help flag returns an error but that's expected
			if err != nil && !strings.Contains(err.Error(), "expected one of") {
				t.Errorf("unexpected error: %v", err)
			}

			output := buf.String()
			for _, want := range tt.want {
				if !strings.Contains(output, want) {
					t.Errorf("output missing expected text: %q\nGot:\n%s", want, output)
				}
			}
		})
	}
}