package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"github.com/axel-po/project-go-clone-redis/internal/command"
	"github.com/axel-po/project-go-clone-redis/internal/config"
	"github.com/axel-po/project-go-clone-redis/internal/engine"
	"github.com/axel-po/project-go-clone-redis/internal/storage"
)

func main() {
	cfg := config.Load()

	store, err := storage.NewFileStorage(cfg.DataDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "erreur de stockage:", err)
		os.Exit(1)
	}

	db := engine.NewWithStorage(store)
	if err := db.Restore(); err != nil {
		fmt.Fprintln(os.Stderr, "erreur de restauration:", err)
		os.Exit(1)
	}
	db.StartBackground(cfg.FlushInterval, cfg.SnapshotInterval)
	defer db.Close()

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println(`clone mini redis — SET k "v" / GET k / DELETE k / GET WHERE <op> v / Ctrl+D pour quitter`)

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				fmt.Fprintln(os.Stderr, "erreur de lecture:", err)
			}
			return
		}

		cmd, err := command.Parse(scanner.Text())
		if err != nil {
			fmt.Println("ERR", err)
			continue
		}

		if q, ok := cmd.(command.GetWhere); ok {
			entries := db.Filter(q)
			if len(entries) == 0 {
				fmt.Println("(vide)")
				continue
			}
			for _, entry := range entries {
				fmt.Printf("%s = %q\n", entry.Key, entry.Value)
			}
			continue
		}

		result, err := db.Apply(cmd)
		switch {
		case errors.Is(err, engine.ErrKeyNotFound):
			fmt.Println("(nil)")
		case err != nil:
			fmt.Println("ERR", err)
		case result.Found:
			fmt.Printf("%q\n", result.Value)
		default:
			fmt.Println("OK")
		}
	}
}
