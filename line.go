package main

import (
	"strings"
)

// Line ...
type Line struct {
	Used    bool
	Content string
	Tokens  []Token
}

// MakeLine ...
// Parse line from a string. Returns an error if syntax is bad.
func MakeLine(line string) (*Line, error) {
	ret := &Line{Content: line}
	if len(line) == 0 {
		ret.Used = false
		return ret, nil
	}

	t, err := Lex(strings.Split(line, " "))
	if err != nil {
		return nil, err
	}
	ret.Tokens = t
	ret.Used = true

	return ret, nil
}
