package engine

import "encoding/json"

type opKind string

const (
	opSet    opKind = "set"
	opDelete opKind = "delete"
)

type Operation struct {
	Kind      opKind `json:"kind"`
	Key       string `json:"key"`
	Value     string `json:"value,omitempty"`
	ExpiresAt int64  `json:"expires_at,omitempty"`
}

func encodeOp(op Operation) ([]byte, error) {
	return json.Marshal(op)
}

func decodeOp(line []byte) (Operation, error) {
	var op Operation
	err := json.Unmarshal(line, &op)
	return op, err
}
