package main

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

type FileLock struct {
	path string
	file *os.File
}

func NewFileLock(projectRoot string) *FileLock {
	return &FileLock{
		path: filepath.Join(projectRoot, ".devslot.lock"),
	}
}

func (l *FileLock) Lock() error {
	file, err := os.OpenFile(l.path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open lock file: %w", err)
	}

	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		file.Close()
		if err == syscall.EWOULDBLOCK {
			return fmt.Errorf("another devslot process is already running")
		}
		return fmt.Errorf("failed to acquire lock: %w", err)
	}

	l.file = file

	// Write PID and timestamp to lock file
	content := fmt.Sprintf("PID: %d\nTime: %s\n", os.Getpid(), time.Now().Format(time.RFC3339))
	if err := file.Truncate(0); err != nil {
		l.Unlock()
		return fmt.Errorf("failed to truncate lock file: %w", err)
	}
	if _, err := file.WriteAt([]byte(content), 0); err != nil {
		l.Unlock()
		return fmt.Errorf("failed to write to lock file: %w", err)
	}

	return nil
}

func (l *FileLock) Unlock() error {
	if l.file == nil {
		return nil
	}

	err := syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN)
	if err != nil {
		return fmt.Errorf("failed to unlock: %w", err)
	}

	if err := l.file.Close(); err != nil {
		return fmt.Errorf("failed to close lock file: %w", err)
	}

	l.file = nil
	return nil
}

