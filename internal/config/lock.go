package config

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// StaleAfter is the age from which a lock file is taken as left behind by a dead
// process. The critical sections of mu make at most a couple of short requests.
const StaleAfter = 30 * time.Second

const lockRetry = 25 * time.Millisecond

// Lock takes the exclusive lock next to the configuration file, waiting for another
// process to release it until ctx ends. A lock older than StaleAfter is removed.
// The returned function releases the lock.
func (s *Store) Lock(ctx context.Context) (func() error, error) {
	if s.pathErr != nil {
		return nil, s.pathErr
	}
	if err := os.MkdirAll(filepath.Dir(s.path), dirMode); err != nil {
		return nil, fmt.Errorf("creating config directory %s: %w", filepath.Dir(s.path), err)
	}
	lockPath := s.path + ".lock"
	if err := s.acquire(ctx, lockPath); err != nil {
		return nil, err
	}
	return func() error { return releaseLock(lockPath) }, nil
}

func (s *Store) acquire(ctx context.Context, lockPath string) error {
	for {
		taken, err := tryLock(lockPath)
		if taken || err != nil {
			return err
		}
		s.removeIfStale(lockPath)
		if err := waitRetry(ctx); err != nil {
			return fmt.Errorf("waiting for config lock %s: %w", lockPath, err)
		}
	}
}

func tryLock(lockPath string) (bool, error) {
	file, err := os.OpenFile(filepath.Clean(lockPath), os.O_CREATE|os.O_EXCL|os.O_WRONLY, fileMode)
	if errors.Is(err, fs.ErrExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("creating config lock %s: %w", lockPath, err)
	}
	_, writeErr := file.WriteString(strconv.Itoa(os.Getpid()) + "\n")
	if err := errors.Join(writeErr, file.Close()); err != nil {
		_ = os.Remove(lockPath)
		return false, fmt.Errorf("writing config lock %s: %w", lockPath, err)
	}
	return true, nil
}

// removeIfStale removes a lock its holder abandoned. Two processes finding the same
// stale lock may both remove it; that only happens after a process died holding it.
func (s *Store) removeIfStale(lockPath string) {
	info, err := os.Stat(lockPath)
	if err != nil || s.now().Sub(info.ModTime()) <= StaleAfter {
		return
	}
	_ = os.Remove(lockPath)
}

func waitRetry(ctx context.Context) error {
	timer := time.NewTimer(lockRetry)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func releaseLock(lockPath string) error {
	if err := os.Remove(lockPath); err != nil {
		return fmt.Errorf("releasing config lock %s: %w", lockPath, err)
	}
	return nil
}
