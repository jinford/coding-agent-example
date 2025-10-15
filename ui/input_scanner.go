package ui

type InputScanner interface {
	Scan() bool
	Text() string
}
