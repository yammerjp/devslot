package lock

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// FileLock represents a file-based lock
type FileLock struct {
	path string
	file *os.File
}

// New creates a new file lock
func New(lockPath string) *FileLock {
	return &FileLock{
		path: lockPath,
	}
}

// Acquire attempts to acquire the lock
func (l *FileLock) Acquire() error {
	dir := filepath.Dir(l.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create lock directory: %w", err)
	}

	file, err := os.OpenFile(l.path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		if os.IsExist(err) {
			return errors.New("lock is already held")
		}
		return fmt.Errorf("failed to acquire lock: %w", err)
	}

	l.file = file
	
	// Write lock information
	lockInfo := fmt.Sprintf("PID: %d\nTime: %s\n", os.Getpid(), time.Now().Format(time.RFC3339))
	if _, err := file.WriteString(lockInfo); err != nil {
		// Best effort cleanup
		_ = file.Close()
		_ = os.Remove(l.path)
		return fmt.Errorf("failed to write lock info: %w", err)
	}

	return nil
}

// Release releases the lock
func (l *FileLock) Release() error {
	if l.file == nil {
		return errors.New("lock not held")
	}

	if err := l.file.Close(); err != nil {
		return fmt.Errorf("failed to close lock file: %w", err)
	}

	if err := os.Remove(l.path); err != nil {
		return fmt.Errorf("failed to remove lock file: %w", err)
	}

	l.file = nil
	return nil
}

// TryAcquire attempts to acquire the lock with a timeout
func (l *FileLock) TryAcquire(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	
	for time.Now().Before(deadline) {
		err := l.Acquire()
		if err == nil {
			return nil
		}
		
		if !errors.Is(err, os.ErrExist) {
			return err
		}
		
		time.Sleep(100 * time.Millisecond)
	}
	
	return errors.New("failed to acquire lock within timeout")
}

// IsLocked checks if the lock is currently held
func (l *FileLock) IsLocked() bool {
	_, err := os.Stat(l.path)
	return err == nil
}