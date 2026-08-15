package command

import "fmt"

type Command interface {
	isCommand()
	String() string
}

type Set struct {
	Key   string
	Value string
}

type Get struct {
	Key string
}

type Delete struct {
	Key string
}

type Operator string

const (
	OpEquals   Operator = "equals"
	OpContains Operator = "contains"
	OpGT       Operator = ">"
	OpGTE      Operator = ">="
	OpLT       Operator = "<"
	OpLTE      Operator = "<="
)

type GetWhere struct {
	Op    Operator
	Value string
}

func (Set) isCommand()      {}
func (Get) isCommand()      {}
func (Delete) isCommand()   {}
func (GetWhere) isCommand() {}

func (c Set) String() string      { return fmt.Sprintf("SET %s %q", c.Key, c.Value) }
func (c Get) String() string      { return fmt.Sprintf("GET %s", c.Key) }
func (c Delete) String() string   { return fmt.Sprintf("DELETE %s", c.Key) }
func (c GetWhere) String() string { return fmt.Sprintf("GET WHERE %s %q", c.Op, c.Value) }
