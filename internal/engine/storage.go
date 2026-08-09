package engine

type Storage interface {
	AppendLog(lines [][]byte) error
	ReadLog() ([][]byte, error)
	ClearLog() error
	WriteSnapshot(data []byte) error
	ReadSnapshot() ([]byte, error)
}
