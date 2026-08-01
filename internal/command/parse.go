package command

import (
	"fmt"
	"strings"
)

func Parse(input string) (Command, error) {
	tokens, err := tokenize(input)
	if err != nil {
		return nil, fmt.Errorf("parse %q: %w", input, err)
	}
	if len(tokens) == 0 {
		return nil, fmt.Errorf("parse: %w", ErrEmptyInput)
	}

	verb, args := strings.ToUpper(tokens[0]), tokens[1:]

	switch verb {
	case "SET":
		if err := expectArgs(verb, args, 2); err != nil {
			return nil, err
		}
		return Set{Key: args[0], Value: args[1]}, nil

	case "GET":
		if err := expectArgs(verb, args, 1); err != nil {
			return nil, err
		}
		return Get{Key: args[0]}, nil

	case "DELETE":
		if err := expectArgs(verb, args, 1); err != nil {
			return nil, err
		}
		return Delete{Key: args[0]}, nil

	default:
		return nil, fmt.Errorf("parse %q: %w", tokens[0], ErrUnknownCommand)
	}
}

func expectArgs(verb string, args []string, want int) error {
	if len(args) == want {
		return nil
	}
	return fmt.Errorf("%s attend %d argument(s), reçu %d: %w",
		verb, want, len(args), ErrWrongArgCount)
}
