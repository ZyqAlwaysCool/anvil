package logging

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type DailyWriter struct {
	dir           string
	filePrefix    string
	fileSuffix    string
	retentionDays int

	mu          sync.Mutex
	currentDate string
	file        *os.File
}

func NewDailyWriter(dir, filePrefix, fileSuffix string, retentionDays int) (*DailyWriter, error) {
	if dir == "" {
		return nil, fmt.Errorf("log dir is required")
	}
	if filePrefix == "" {
		return nil, fmt.Errorf("log file prefix is required")
	}

	writer := &DailyWriter{
		dir:           dir,
		filePrefix:    filePrefix,
		fileSuffix:    fileSuffix,
		retentionDays: retentionDays,
	}
	if err := writer.rotate(time.Now()); err != nil {
		return nil, err
	}
	return writer, nil
}

func (w *DailyWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if err := w.rotate(time.Now()); err != nil {
		return 0, err
	}
	return w.file.Write(p)
}

func (w *DailyWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file == nil {
		return nil
	}
	err := w.file.Close()
	w.file = nil
	return err
}

func (w *DailyWriter) rotate(now time.Time) error {
	date := now.Format("2006-01-02")
	if w.file != nil && w.currentDate == date {
		return nil
	}

	if err := os.MkdirAll(w.dir, 0o755); err != nil {
		return fmt.Errorf("create log dir: %w", err)
	}
	if w.file != nil {
		if err := w.file.Close(); err != nil {
			return fmt.Errorf("close previous log file: %w", err)
		}
	}

	filePath := filepath.Join(w.dir, w.buildFileName(date))
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}

	w.file = file
	w.currentDate = date
	return w.cleanupOldFiles(now)
}

func (w *DailyWriter) cleanupOldFiles(now time.Time) error {
	if w.retentionDays <= 0 {
		return nil
	}
	entries, err := os.ReadDir(w.dir)
	if err != nil {
		return fmt.Errorf("read log dir: %w", err)
	}

	expireBefore := now.AddDate(0, 0, -w.retentionDays)
	prefix := w.filePrefix + "-"
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, ".log") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return fmt.Errorf("read log file info: %w", err)
		}
		if info.ModTime().Before(expireBefore) {
			if err := os.Remove(filepath.Join(w.dir, name)); err != nil {
				return fmt.Errorf("remove expired log file: %w", err)
			}
		}
	}
	return nil
}

func (w *DailyWriter) buildFileName(date string) string {
	if w.fileSuffix == "" {
		return fmt.Sprintf("%s-%s.log", w.filePrefix, date)
	}
	return fmt.Sprintf("%s-%s-%s.log", w.filePrefix, date, w.fileSuffix)
}

var _ io.WriteCloser = (*DailyWriter)(nil)
