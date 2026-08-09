package storage

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
)

type FileStorage struct {
	logPath      string
	snapshotPath string
}

func NewFileStorage(dir string) (*FileStorage, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &FileStorage{
		logPath:      filepath.Join(dir, "aof.log"),
		snapshotPath: filepath.Join(dir, "snapshot.json"),
	}, nil
}

func (s *FileStorage) AppendLog(lines [][]byte) error {
	f, err := os.OpenFile(s.logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, line := range lines {
		if _, err := f.Write(line); err != nil {
			return err
		}
		if _, err := f.WriteString("\n"); err != nil {
			return err
		}
	}
	return nil
}

func (s *FileStorage) ReadLog() ([][]byte, error) {
	data, err := os.ReadFile(s.logPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var lines [][]byte
	for _, line := range bytes.Split(data, []byte("\n")) {
		if len(line) > 0 {
			lines = append(lines, line)
		}
	}
	return lines, nil
}

func (s *FileStorage) ClearLog() error {
	return os.WriteFile(s.logPath, nil, 0o644)
}

func (s *FileStorage) WriteSnapshot(data []byte) error {
	tmp := s.snapshotPath + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.snapshotPath)
}

func (s *FileStorage) ReadSnapshot() ([]byte, error) {
	data, err := os.ReadFile(s.snapshotPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	return data, err
}
