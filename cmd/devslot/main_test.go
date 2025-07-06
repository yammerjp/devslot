package main

import (
	"bytes"
	"testing"
)

func TestApp_Run(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		wantErr        bool
		wantOutputText string
	}{
		{
			name:           "no arguments shows help",
			args:           []string{},
			wantErr:        true, // kong returns error for help
			wantOutputText: "Usage: devslot <command>",
		},
		{
			name:           "help command",
			args:           []string{"--help"},
			wantErr:        true, // kong returns an error for help
			wantOutputText: "Usage: devslot <command>",
		},
		{
			name:           "version command",
			args:           []string{"version"},
			wantErr:        false,
			wantOutputText: "devslot version",
		},
		{
			name:           "unknown command",
			args:           []string{"unknown"},
			wantErr:        true,
			wantOutputText: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			app := NewApp(&buf)

			// Override exit handler for testing
			exitCode := -1
			app.SetExitHandler(func(code int) {
				exitCode = code
			})

			err := app.Run(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("App.Run() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantOutputText != "" {
				output := buf.String()
				if !contains(output, tt.wantOutputText) {
					t.Errorf("App.Run() output = %v, want to contain %v", output, tt.wantOutputText)
				}
			}

			// For help cases, check that exit was called with 0
			if tt.name == "help command" || tt.name == "no arguments shows help" {
				if exitCode != 0 {
					t.Errorf("Expected exit code 0 for help, got %d", exitCode)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && testContainsHelper(s, substr)
}

func testContainsHelper(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
