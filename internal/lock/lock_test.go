package lock

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileLock(t *testing.T) {
	t.Run("successful lock and unlock", func(t *testing.T) {
		tmpDir := t.TempDir()
		lock := New(filepath.Join(tmpDir, ".devslot.lock"))

		err := lock.Acquire()
		if err != nil {
			t.Errorf("Lock() error = %v, want nil", err)
		}

		// Check that lock file exists
		lockPath := filepath.Join(tmpDir, ".devslot.lock")
		if _, err := os.Stat(lockPath); os.IsNotExist(err) {
			t.Error("Lock file was not created")
		}

		err = lock.Release()
		if err != nil {
			t.Errorf("Unlock() error = %v, want nil", err)
		}
	})

	t.Run("prevents concurrent locks", func(t *testing.T) {
		tmpDir := t.TempDir()
		lockPath := filepath.Join(tmpDir, ".devslot.lock")
		lock1 := New(lockPath)
		lock2 := New(lockPath)

		err := lock1.Acquire()
		if err != nil {
			t.Errorf("First Lock() error = %v, want nil", err)
		}
		defer func() { _ = lock1.Release() }()

		err = lock2.Acquire()
		if err == nil {
			t.Error("Second Lock() expected error, got nil")
		}
		if !contains(err.Error(), "another devslot process is already running") {
			t.Errorf("Second Lock() error = %v, want error containing 'another devslot process is already running'", err)
		}
	})

	t.Run("multiple unlock calls are safe", func(t *testing.T) {
		tmpDir := t.TempDir()
		lock := New(filepath.Join(tmpDir, ".devslot.lock"))

		err := lock.Acquire()
		if err != nil {
			t.Errorf("Lock() error = %v, want nil", err)
		}

		err = lock.Release()
		if err != nil {
			t.Errorf("First Unlock() error = %v, want nil", err)
		}

		// Second unlock should not error
		err = lock.Release()
		if err != nil {
			t.Errorf("Second Unlock() error = %v, want nil", err)
		}
	})

	t.Run("lock file contains PID", func(t *testing.T) {
		tmpDir := t.TempDir()
		lock := New(filepath.Join(tmpDir, ".devslot.lock"))

		err := lock.Acquire()
		if err != nil {
			t.Errorf("Lock() error = %v, want nil", err)
		}
		defer func() { _ = lock.Release() }()

		lockPath := filepath.Join(tmpDir, ".devslot.lock")
		content, err := os.ReadFile(lockPath)
		if err != nil {
			t.Errorf("Failed to read lock file: %v", err)
		}

		if !contains(string(content), "PID:") {
			t.Error("Lock file does not contain PID information")
		}
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && strings.Contains(s, substr)
}
