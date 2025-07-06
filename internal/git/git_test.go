package git

import (
	"os"
	"testing"
)

func TestSanitizeBranchComponent(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"yammer", "yammer"},
		{"User.Name", "user-name"},
		{"user@example.com", "user-example-com"},
		{"user name", "user-name"},
		{"user__name", "user__name"},
		{"User-Name", "user-name"},
		{"山田太郎", "user"},
		{"", "user"},
		{"---", "user"},
		{"user.name+tag", "user-name-tag"},
		{"user---name", "user-name"},
		{"_user_", "user"},
		{"123user", "123user"},
		{"user123", "user123"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := SanitizeBranchComponent(tt.input)
			if got != tt.expected {
				t.Errorf("SanitizeBranchComponent(%q) = %q, want %q",
					tt.input, got, tt.expected)
			}
		})
	}
}

func TestGetBranchPrefix(t *testing.T) {
	// Save and restore environment
	origPrefix := os.Getenv("DEVSLOT_BRANCH_PREFIX")
	defer func() {
		if origPrefix != "" {
			os.Setenv("DEVSLOT_BRANCH_PREFIX", origPrefix)
		} else {
			os.Unsetenv("DEVSLOT_BRANCH_PREFIX")
		}
	}()

	// Test environment variable
	os.Setenv("DEVSLOT_BRANCH_PREFIX", "feature/")
	prefix := GetBranchPrefix()
	if prefix != "feature/" {
		t.Errorf("GetBranchPrefix() with env = %q, want %q", prefix, "feature/")
	}

	// Test fallback (unset env var)
	os.Unsetenv("DEVSLOT_BRANCH_PREFIX")
	prefix = GetBranchPrefix()
	// Should either use git email or fallback to "devslot/user/"
	if prefix == "" {
		t.Error("GetBranchPrefix() returned empty string")
	}
}
