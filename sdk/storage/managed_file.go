package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/Aldiwildan77/inspectd/sdk/types"
)

// ManagedFileStorage is a production-ready file storage with automatic cleanup.
// It manages file retention based on age and count limits.
// Suitable for production environments where file-based storage is needed.
type ManagedFileStorage struct {
	*FileStorage
	mu           sync.RWMutex
	maxFiles     int
	maxAge       time.Duration
	cleanupTicker *time.Ticker
	stopCleanup  chan struct{}
	cleanupDone  chan struct{}
}

// ManagedFileStorageConfig configures managed file storage behavior.
type ManagedFileStorageConfig struct {
	// MaxFiles is the maximum number of files to retain (0 = no limit).
	MaxFiles int

	// MaxAge is the maximum age of files to retain (0 = no age limit).
	MaxAge time.Duration

	// CleanupInterval is how often to run cleanup (default: 1 hour).
	CleanupInterval time.Duration
}

// NewManagedFileStorage creates a new managed file storage instance.
// The baseDir will be created if it doesn't exist.
// Cleanup runs automatically in the background.
func NewManagedFileStorage(baseDir string, config ManagedFileStorageConfig) (*ManagedFileStorage, error) {
	fs, err := NewFileStorage(baseDir)
	if err != nil {
		return nil, err
	}

	if config.CleanupInterval == 0 {
		config.CleanupInterval = 1 * time.Hour
	}

	mfs := &ManagedFileStorage{
		FileStorage:    fs,
		maxFiles:       config.MaxFiles,
		maxAge:         config.MaxAge,
		cleanupTicker:   time.NewTicker(config.CleanupInterval),
		stopCleanup:    make(chan struct{}),
		cleanupDone:    make(chan struct{}),
	}

	// Start cleanup goroutine
	go mfs.cleanupLoop()

	return mfs, nil
}

// cleanupLoop runs periodic cleanup in the background.
func (m *ManagedFileStorage) cleanupLoop() {
	defer close(m.cleanupDone)

	for {
		select {
		case <-m.cleanupTicker.C:
			if err := m.cleanup(); err != nil {
				// Log error but continue
				_ = err
			}
		case <-m.stopCleanup:
			return
		}
	}
}

// cleanup removes old files based on retention policies.
func (m *ManagedFileStorage) cleanup() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Read all files
	entries, err := os.ReadDir(m.baseDir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	// Collect file info with timestamps
	type fileInfo struct {
		path      string
		timestamp time.Time
		modTime   time.Time
	}

	files := make([]fileInfo, 0)
	now := time.Now()

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		filePath := filepath.Join(m.baseDir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Try to parse timestamp from filename or use mod time
		timestamp := info.ModTime()
		if snapshot, err := m.parseFile(filePath); err == nil {
			if ts, err := snapshot.ParseTimestamp(); err == nil {
				timestamp = ts
			}
		}

		files = append(files, fileInfo{
			path:      filePath,
			timestamp: timestamp,
			modTime:   info.ModTime(),
		})
	}

	// Sort by timestamp (oldest first)
	sort.Slice(files, func(i, j int) bool {
		return files[i].timestamp.Before(files[j].timestamp)
	})

	// Apply retention policies
	toDelete := make([]string, 0)

	// Remove files older than maxAge
	if m.maxAge > 0 {
		cutoff := now.Add(-m.maxAge)
		for _, file := range files {
			if file.timestamp.Before(cutoff) {
				toDelete = append(toDelete, file.path)
			}
		}
	}

	// Remove oldest files if over maxFiles limit
	if m.maxFiles > 0 {
		remaining := len(files) - len(toDelete)
		if remaining > m.maxFiles {
			// Mark oldest files for deletion
			deleteCount := remaining - m.maxFiles
			for _, file := range files {
				if deleteCount <= 0 {
					break
				}
				// Skip if already marked for deletion
				alreadyMarked := false
				for _, delPath := range toDelete {
					if delPath == file.path {
						alreadyMarked = true
						break
					}
				}
				if !alreadyMarked {
					toDelete = append(toDelete, file.path)
					deleteCount--
				}
			}
		}
	}

	// Delete files
	for _, path := range toDelete {
		os.Remove(path)
	}

	return nil
}

// parseFile reads and parses a snapshot file.
func (m *ManagedFileStorage) parseFile(filePath string) (*types.Snapshot, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return types.FromJSON(data)
}

// Close stops cleanup and releases resources.
func (m *ManagedFileStorage) Close() error {
	// Stop cleanup goroutine
	close(m.stopCleanup)
	m.cleanupTicker.Stop()

	// Wait for cleanup to finish
	select {
	case <-m.cleanupDone:
	case <-time.After(5 * time.Second):
		// Timeout waiting for cleanup
	}

	// Run final cleanup
	m.cleanup()

	return m.FileStorage.Close()
}

// Stats returns storage statistics.
func (m *ManagedFileStorage) Stats() (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entries, err := os.ReadDir(m.baseDir)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			count++
		}
	}

	return count, nil
}

