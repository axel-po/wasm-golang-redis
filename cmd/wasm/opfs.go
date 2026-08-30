//go:build js && wasm

package main

import (
	"bytes"
	"strings"
	"syscall/js"
)

type jsStorage struct {
	bridge js.Value
}

func (s jsStorage) AppendLog(lines [][]byte) error {
	var buf bytes.Buffer
	for _, line := range lines {
		buf.Write(line)
		buf.WriteByte('\n')
	}
	s.bridge.Call("appendAOF", buf.String())
	return nil
}

func (s jsStorage) ReadLog() ([][]byte, error) {
	raw := s.bridge.Call("readAOF").String()
	if raw == "" {
		return nil, nil
	}
	var lines [][]byte
	for _, line := range strings.Split(raw, "\n") {
		if line != "" {
			lines = append(lines, []byte(line))
		}
	}
	return lines, nil
}

func (s jsStorage) ClearLog() error {
	s.bridge.Call("clearAOF")
	return nil
}

func (s jsStorage) WriteSnapshot(data []byte) error {
	s.bridge.Call("writeSnapshot", string(data))
	return nil
}

func (s jsStorage) ReadSnapshot() ([]byte, error) {
	raw := s.bridge.Call("readSnapshot").String()
	if raw == "" {
		return nil, nil
	}
	return []byte(raw), nil
}
