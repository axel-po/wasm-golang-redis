package storage

import (
	"testing"
)

func TestJournalAppendReadClear(t *testing.T) {
	s, err := NewFileStorage(t.TempDir())
	if err != nil {
		t.Fatalf("NewFileStorage: %v", err)
	}

	if lines, err := s.ReadLog(); err != nil || lines != nil {
		t.Fatalf("ReadLog initial = %v, %v ; attendu nil, nil", lines, err)
	}

	if err := s.AppendLog([][]byte{[]byte("un"), []byte("deux")}); err != nil {
		t.Fatalf("AppendLog: %v", err)
	}
	if err := s.AppendLog([][]byte{[]byte("trois")}); err != nil {
		t.Fatalf("AppendLog: %v", err)
	}

	lines, err := s.ReadLog()
	if err != nil {
		t.Fatalf("ReadLog: %v", err)
	}
	if len(lines) != 3 || string(lines[0]) != "un" || string(lines[2]) != "trois" {
		t.Errorf("ReadLog = %q ; attendu [un deux trois]", lines)
	}

	if err := s.ClearLog(); err != nil {
		t.Fatalf("ClearLog: %v", err)
	}
	if lines, _ := s.ReadLog(); len(lines) != 0 {
		t.Errorf("ReadLog après ClearLog = %q ; attendu vide", lines)
	}
}

func TestSnapshotWriteRead(t *testing.T) {
	s, err := NewFileStorage(t.TempDir())
	if err != nil {
		t.Fatalf("NewFileStorage: %v", err)
	}

	if data, err := s.ReadSnapshot(); err != nil || data != nil {
		t.Fatalf("ReadSnapshot initial = %v, %v ; attendu nil, nil", data, err)
	}

	if err := s.WriteSnapshot([]byte(`{"name":"matt"}`)); err != nil {
		t.Fatalf("WriteSnapshot: %v", err)
	}
	data, err := s.ReadSnapshot()
	if err != nil || string(data) != `{"name":"matt"}` {
		t.Errorf("ReadSnapshot = %q, %v", data, err)
	}
}
