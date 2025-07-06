package errors

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestUserError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		message    string
		suggestion string
		want       string
	}{
		{
			name:       "with suggestion",
			err:        errors.New("original error"),
			message:    "operation failed",
			suggestion: "Try this to fix it",
			want:       "operation failed: original error\nTry this to fix it",
		},
		{
			name:       "without suggestion",
			err:        errors.New("original error"),
			message:    "operation failed",
			suggestion: "",
			want:       "operation failed: original error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := WithSuggestion(tt.err, tt.message, tt.suggestion)
			if got := err.Error(); got != tt.want {
				t.Errorf("UserError.Error() = %q, want %q", got, tt.want)
			}

			// Test Unwrap
			var userErr *UserError
			if errors.As(err, &userErr) {
				if userErr.Unwrap() != tt.err {
					t.Errorf("UserError.Unwrap() = %v, want %v", userErr.Unwrap(), tt.err)
				}
			}
		})
	}
}

func TestHelperFunctions(t *testing.T) {
	tests := []struct {
		name        string
		errFunc     func() error
		wantMessage string
		wantSuggest string
	}{
		{
			name:        "NotInProject",
			errFunc:     func() error { return NotInProject(errors.New("not found")) },
			wantMessage: "not in a devslot project",
			wantSuggest: "Run 'devslot boilerplate .' to create a new project",
		},
		{
			name:        "SlotAlreadyExists",
			errFunc:     func() error { return SlotAlreadyExists("test-slot") },
			wantMessage: "slot test-slot already exists",
			wantSuggest: "Use a different name or destroy the existing slot with 'devslot destroy test-slot'",
		},
		{
			name:        "SlotNotFound",
			errFunc:     func() error { return SlotNotFound("test-slot") },
			wantMessage: "slot test-slot does not exist",
			wantSuggest: "Run 'devslot list' to see available slots",
		},
		{
			name:        "LockFailed",
			errFunc:     func() error { return LockFailed(errors.New("lock error")) },
			wantMessage: "failed to acquire lock",
			wantSuggest: "Another devslot command may be running",
		},
		{
			name:        "CloneFailed",
			errFunc:     func() error { return CloneFailed("my-repo", errors.New("network error")) },
			wantMessage: "failed to clone my-repo",
			wantSuggest: "Check your network connection and repository URL",
		},
		{
			name:        "FetchFailed",
			errFunc:     func() error { return FetchFailed(errors.New("fetch error")) },
			wantMessage: "failed to fetch latest changes",
			wantSuggest: "Check your network connection and repository access",
		},
		{
			name:        "HookNotExecutable",
			errFunc:     func() error { return HookNotExecutable("post-create") },
			wantMessage: "hook post-create is not executable",
			wantSuggest: "Run 'chmod +x hooks/post-create' to fix",
		},
		{
			name:        "HookFailed",
			errFunc:     func() error { return HookFailed("post-create", errors.New("exit 1")) },
			wantMessage: "hook post-create failed",
			wantSuggest: "Check the hook script at hooks/post-create for errors",
		},
		{
			name:        "WorktreeFailed",
			errFunc:     func() error { return WorktreeFailed("my-repo", errors.New("branch error")) },
			wantMessage: "failed to create worktree for my-repo",
			wantSuggest: "Ensure the branch exists or try 'devslot init' to update repositories",
		},
		{
			name:        "ConfigNotFound",
			errFunc:     func() error { return ConfigNotFound() },
			wantMessage: "devslot.yaml not found in any parent directory",
			wantSuggest: "Run 'devslot boilerplate .' to create a new project",
		},
		{
			name:        "YAMLParseFailed",
			errFunc:     func() error { return YAMLParseFailed(errors.New("yaml error")) },
			wantMessage: "failed to parse YAML",
			wantSuggest: "Check the devslot.yaml syntax",
		},
		{
			name:        "UnsupportedVersion",
			errFunc:     func() error { return UnsupportedVersion(2) },
			wantMessage: "unsupported config version: 2",
			wantSuggest: "Only version 1 is supported",
		},
		{
			name:        "NoBranchesFound",
			errFunc:     func() error { return NoBranchesFound() },
			wantMessage: "no branches found in repository",
			wantSuggest: "The repository may be empty or corrupted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.errFunc()
			errStr := err.Error()

			if !strings.Contains(errStr, tt.wantMessage) {
				t.Errorf("Error message %q does not contain %q", errStr, tt.wantMessage)
			}

			if !strings.Contains(errStr, tt.wantSuggest) {
				t.Errorf("Error message %q does not contain suggestion %q", errStr, tt.wantSuggest)
			}
		})
	}
}

func TestErrorFormatting(t *testing.T) {
	// Test that the error message is properly formatted
	err := WithSuggestion(
		fmt.Errorf("network timeout"),
		"failed to connect",
		"Check your internet connection",
	)

	expected := "failed to connect: network timeout\nCheck your internet connection"
	if err.Error() != expected {
		t.Errorf("Error format incorrect:\ngot:  %q\nwant: %q", err.Error(), expected)
	}
}