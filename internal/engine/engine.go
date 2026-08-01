package engine

type Engine struct {
	state map[string]string
}

func New() *Engine {
	return &Engine{state: make(map[string]string)}
}




