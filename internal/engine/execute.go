package engine

import (
	"fmt"

	"github.com/axel-po/project-go-clone-redis/internal/command"
)

type Result struct {
	Value string
	Found bool
}

func (e *Engine) Apply(cmd command.Command) (Result, error) {
	switch c := cmd.(type) {
	case command.Set:
		e.Set(c.Key, c.Value)
		return Result{}, nil

	case command.Get:
		value, err := e.Get(c.Key)
		if err != nil {
			return Result{}, err
		}
		return Result{Value: value, Found: true}, nil

	case command.Delete:
		return Result{Found: e.Delete(c.Key)}, nil

	default:
		return Result{}, fmt.Errorf("apply %v: %w", cmd, ErrUnsupportedCommand)
	}
}

func (e *Engine) Execute(input string) (Result, error) {
	cmd, err := command.Parse(input)
	if err != nil {
		return Result{}, err
	}
	return e.Apply(cmd)
}
