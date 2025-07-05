package command

import (
	"bytes"
	"testing"
)

func TestVersionCmd_Run(t *testing.T) {
	var buf bytes.Buffer
	cmd := &VersionCmd{}
	ctx := &Context{Writer: &buf}

	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("VersionCmd.Run() error = %v", err)
	}

	output := buf.String()
	if !contains(output, "devslot version") {
		t.Errorf("VersionCmd.Run() output = %v, want to contain 'devslot version'", output)
	}
}