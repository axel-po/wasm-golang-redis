//go:build js && wasm

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"syscall/js"
	"time"

	"github.com/axel-po/project-go-clone-redis/internal/command"
	"github.com/axel-po/project-go-clone-redis/internal/engine"
)

var db *engine.Engine

var errUninit = errors.New("moteur non initialisé : appeler init() d'abord")

func main() {
	js.Global().Set("__wasmredis", js.ValueOf(map[string]any{
		"init":      guard(initEngine),
		"execBatch": guard(execBatch),
		"getWhere":  guard(getWhere),
		"getMany":   guard(getMany),
		"entries":   guard(entriesHandler),
		"flush":     guard(flushHandler),
		"snapshot":  guard(snapshotHandler),
		"close":     guard(closeHandler),
	}))
	select {}
}

func guard(fn func(args []js.Value) (any, error)) js.Func {
	return js.FuncOf(func(_ js.Value, args []js.Value) (out any) {
		defer func() {
			if r := recover(); r != nil {
				out = map[string]any{"error": fmt.Sprint(r)}
			}
		}()
		result, err := fn(args)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return result
	})
}

func initEngine(args []js.Value) (any, error) {
	bridge := js.Global().Get("__wasmredisStorage")
	if bridge.IsUndefined() {
		return nil, errors.New("__wasmredisStorage absent : le worker doit préparer OPFS avant init")
	}

	db = engine.NewWithStorage(jsStorage{bridge: bridge})
	if err := db.Restore(); err != nil {
		return nil, fmt.Errorf("restore au démarrage: %w", err)
	}

	cfg := args[0]
	db.StartBackground(
		msInterval(cfg, "flushIntervalMs", time.Second),
		msInterval(cfg, "snapshotIntervalMs", 2*time.Minute),
		msInterval(cfg, "sweepIntervalMs", time.Second),
	)
	return map[string]any{"restored": len(db.All())}, nil
}

func msInterval(cfg js.Value, field string, fallback time.Duration) time.Duration {
	v := cfg.Get(field)
	if v.IsUndefined() || v.IsNull() || v.Int() <= 0 {
		return fallback
	}
	return time.Duration(v.Int()) * time.Millisecond
}

func execBatch(args []js.Value) (any, error) {
	if db == nil {
		return nil, errUninit
	}
	n := args[0].Length()
	out := make([]any, n)
	for i := 0; i < n; i++ {
		res, err := db.Execute(args[0].Index(i).String())
		if err != nil {
			out[i] = map[string]any{"error": err.Error()}
		} else {
			out[i] = map[string]any{"value": res.Value}
		}
	}
	return out, nil
}

var opMap = map[string]command.Operator{
	"=":        command.OpEquals,
	"==":       command.OpEquals,
	">":        command.OpGT,
	">=":       command.OpGTE,
	"<":        command.OpLT,
	"<=":       command.OpLTE,
	"contains": command.OpContains,
}

func getWhere(args []js.Value) (any, error) {
	if db == nil {
		return nil, errUninit
	}
	op, ok := opMap[args[0].String()]
	if !ok {
		return nil, fmt.Errorf("opérateur de filtre inconnu %q", args[0].String())
	}
	entries := db.Filter(command.GetWhere{Op: op, Value: args[1].String()})
	out := make([]any, len(entries))
	for i, entry := range entries {
		out[i] = entry.Key
	}
	return out, nil
}

func getMany(args []js.Value) (any, error) {
	if db == nil {
		return nil, errUninit
	}
	n := args[0].Length()
	found := make([]engine.Entry, 0, n)
	for i := 0; i < n; i++ {
		if entry, ok := db.Lookup(args[0].Index(i).String()); ok {
			found = append(found, entry)
		}
	}
	return marshal(found)
}

func entriesHandler([]js.Value) (any, error) {
	if db == nil {
		return nil, errUninit
	}
	all := db.All()
	sort.Slice(all, func(i, j int) bool { return all[i].Key < all[j].Key })
	return marshal(all)
}

func flushHandler([]js.Value) (any, error) {
	if db == nil {
		return nil, errUninit
	}
	return nil, db.Flush()
}

func snapshotHandler([]js.Value) (any, error) {
	if db == nil {
		return nil, errUninit
	}
	return nil, db.Snapshot()
}

func closeHandler([]js.Value) (any, error) {
	if db == nil {
		return nil, errUninit
	}
	return nil, db.Close()
}

func marshal(v any) (any, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("sérialisation JSON: %w", err)
	}
	return string(data), nil
}
