package command

import "strings"

func tokenize(input string) ([]string, error) {
	var (
		tokens  []string
		current strings.Builder
		inQuote bool
		started bool
	)

	flush := func() {
		if started {
			tokens = append(tokens, current.String())
			current.Reset()
			started = false
		}
	}

	for _, r := range input {
		switch {
		case r == '"':
			inQuote = !inQuote
			started = true
		case (r == ' ' || r == '\t') && !inQuote:
			flush()
		default:
			current.WriteRune(r)
			started = true
		}
	}

	if inQuote {
		return nil, ErrUnclosedQuote
	}
	flush()

	return tokens, nil
}
