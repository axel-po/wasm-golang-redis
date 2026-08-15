package command

import "errors"

var (
	ErrEmptyInput      = errors.New("commande vide")
	ErrUnknownCommand  = errors.New("commande inconnue")
	ErrWrongArgCount   = errors.New("nombre d'arguments invalide")
	ErrUnclosedQuote   = errors.New("guillemet non fermé")
	ErrUnknownOperator = errors.New("opérateur inconnu")
)
