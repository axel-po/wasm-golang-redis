package engine

import (
	"github.com/axel-po/project-go-clone-redis/internal/command"
	"github.com/samber/lo"
)

type BatchResult struct {
	Result Result
	Err    error
}

func (e *Engine) Batch(cmds []command.Command) []BatchResult {
	return lo.Map(cmds, func(cmd command.Command, _ int) BatchResult {
		res, err := e.Apply(cmd)
		return BatchResult{Result: res, Err: err}
	})
}
