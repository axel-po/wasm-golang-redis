package command

import (
	"fmt"
	"strconv"
	"strings"
	"time"
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
		return parseSet(args)

	case "GET":
		if len(args) >= 1 && strings.ToUpper(args[0]) == "WHERE" {
			return parseGetWhere(args[1:])
		}
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

func parseSet(args []string) (Command, error) {
	switch {
	case len(args) == 2:
		return Set{Key: args[0], Value: args[1]}, nil

	case len(args) == 4 && strings.ToUpper(args[2]) == "EX":
		secs, err := strconv.Atoi(args[3])
		if err != nil || secs <= 0 {
			return nil, fmt.Errorf("EX attend un entier positif, reçu %q: %w", args[3], ErrInvalidTTL)
		}
		return Set{Key: args[0], Value: args[1], TTL: time.Duration(secs) * time.Second}, nil

	default:
		return nil, fmt.Errorf(`SET attend <clé> <valeur> [EX <secondes>]: %w`, ErrWrongArgCount)
	}
}

func parseGetWhere(args []string) (Command, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("GET WHERE attend <op> <valeur>: %w", ErrWrongArgCount)
	}
	op, err := parseOperator(args[0])
	if err != nil {
		return nil, err
	}
	return GetWhere{Op: op, Value: args[1]}, nil
}

func parseOperator(s string) (Operator, error) {
	switch strings.ToLower(s) {
	case "equals":
		return OpEquals, nil
	case "contains":
		return OpContains, nil
	case ">":
		return OpGT, nil
	case ">=":
		return OpGTE, nil
	case "<":
		return OpLT, nil
	case "<=":
		return OpLTE, nil
	default:
		return "", fmt.Errorf("opérateur %q: %w", s, ErrUnknownOperator)
	}
}

func expectArgs(verb string, args []string, want int) error {
	if len(args) == want {
		return nil
	}
	return fmt.Errorf("%s attend %d argument(s), reçu %d: %w",
		verb, want, len(args), ErrWrongArgCount)
}
