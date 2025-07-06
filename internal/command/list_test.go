package command

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yammerjp/devslot/internal/testutil"
)

func TestListCmd_Run(t *testing.T) {
	tests := []struct {
		name         string
		setupFunc    func(t *testing.T, projectRoot string)
		wantContains []string
		wantErr      bool
	}{
		{
			name: "no slots",
			setupFunc: func(t *testing.T, projectRoot string) {
				// slots directory exists but is empty
			},
			wantContains: []string{
				"No slots found",
				"Create a new slot with 'devslot create <slot-name>'",
			},
			wantErr: false,
		},
		{
			name: "multiple slots",
			setupFunc: func(t *testing.T, projectRoot string) {
				slotsDir := filepath.Join(projectRoot, "slots")
				os.MkdirAll(filepath.Join(slotsDir, "dev"), 0755)
				os.MkdirAll(filepath.Join(slotsDir, "staging"), 0755)
				os.MkdirAll(filepath.Join(slotsDir, "feature-x"), 0755)
			},
			wantContains: []string{
				"Available slots:",
				"- dev",
				"- staging",
				"- feature-x",
			},
			wantErr: false,
		},
		{
			name: "slots directory with files",
			setupFunc: func(t *testing.T, projectRoot string) {
				slotsDir := filepath.Join(projectRoot, "slots")
				os.MkdirAll(filepath.Join(slotsDir, "valid-slot"), 0755)
				// Create a file (should be ignored)
				os.WriteFile(filepath.Join(slotsDir, "README.md"), []byte("test"), 0644)
			},
			wantContains: []string{
				"Available slots:",
				"- valid-slot",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectRoot := testutil.TempDir(t)
			testutil.CreateProjectStructure(t, projectRoot)

			cleanup := testutil.Chdir(t, projectRoot)
			defer cleanup()

			if tt.setupFunc != nil {
				tt.setupFunc(t, projectRoot)
			}

			var buf bytes.Buffer
			cmd := &ListCmd{}
			ctx := &Context{Writer: &buf, Logger: nil}

			err := cmd.Run(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListCmd.Run() error = %v, wantErr %v", err, tt.wantErr)
			}

			output := buf.String()
			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("ListCmd.Run() output = %v, want to contain %v", output, want)
				}
			}
		})
	}
}

func TestListCmd_NotInProject(t *testing.T) {
	// Create a temporary directory without devslot.yaml
	tmpDir := testutil.TempDir(t)
	cleanup := testutil.Chdir(t, tmpDir)
	defer cleanup()

	var buf bytes.Buffer
	cmd := &ListCmd{}
	ctx := &Context{Writer: &buf, Logger: nil}

	err := cmd.Run(ctx)
	if err == nil {
		t.Error("expected error when not in devslot project")
	}
	if !strings.Contains(err.Error(), "not in a devslot project") {
		t.Errorf("expected 'not in a devslot project' error, got: %v", err)
	}
}