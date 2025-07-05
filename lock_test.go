package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileLock(t *testing.T) {
	t.Run("successful lock and unlock", func(t *testing.T) {
		tmpDir := t.TempDir()
		lock := NewFileLock(tmpDir)

		err := lock.Lock()
		if err != nil {
			t.Errorf("Lock() error = %v, want nil", err)
		}

		// Check that lock file exists
		lockPath := filepath.Join(tmpDir, ".devslot.lock")
		if _, err := os.Stat(lockPath); os.IsNotExist(err) {
			t.Error("Lock file was not created")
		}

		err = lock.Unlock()
		if err != nil {
			t.Errorf("Unlock() error = %v, want nil", err)
		}
	})

	t.Run("prevents concurrent locks", func(t *testing.T) {
		tmpDir := t.TempDir()
		lock1 := NewFileLock(tmpDir)
		lock2 := NewFileLock(tmpDir)

		err := lock1.Lock()
		if err != nil {
			t.Errorf("First Lock() error = %v, want nil", err)
		}
		defer func() { _ = lock1.Unlock() }()

		err = lock2.Lock()
		if err == nil {
			t.Error("Second Lock() expected error, got nil")
		}
		if !contains(err.Error(), "another devslot process is already running") {
			t.Errorf("Second Lock() error = %v, want error containing 'another devslot process is already running'", err)
		}
	})

	t.Run("multiple unlock calls are safe", func(t *testing.T) {
		tmpDir := t.TempDir()
		lock := NewFileLock(tmpDir)

		err := lock.Lock()
		if err != nil {
			t.Errorf("Lock() error = %v, want nil", err)
		}

		err = lock.Unlock()
		if err != nil {
			t.Errorf("First Unlock() error = %v, want nil", err)
		}

		// Second unlock should not error
		err = lock.Unlock()
		if err != nil {
			t.Errorf("Second Unlock() error = %v, want nil", err)
		}
	})

	t.Run("lock file contains PID", func(t *testing.T) {
		tmpDir := t.TempDir()
		lock := NewFileLock(tmpDir)

		err := lock.Lock()
		if err != nil {
			t.Errorf("Lock() error = %v, want nil", err)
		}
		defer func() { _ = lock.Unlock() }()

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
