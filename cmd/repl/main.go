package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"github.com/axel-po/project-go-clone-redis/internal/engine"
)

func main() {
	db := engine.New()
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("clone mini redis — SET k \"v\" / GET k / DELETE k / Ctrl+D pour quitter")

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				fmt.Fprintln(os.Stderr, "erreur de lecture:", err)
				os.Exit(1)
			}
			return
		}

		result, err := db.Execute(scanner.Text())
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
