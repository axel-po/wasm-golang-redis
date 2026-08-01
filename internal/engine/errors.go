package engine

import "errors"

var (
	ErrKeyNotFound        = errors.New("clé absente")
	ErrUnsupportedCommand = errors.New("commande non supportée par le moteur")
)




