package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// TokenType ...
// Type for types of token
type TokenType uint8

// The types of token in the program
const (
	TokenLet TokenType = iota
	TokenPrint
	TokenExit
	TokenGoto
	TokenIf
	TokenThen
	TokenElse
	TokenEq
	TokenGt
	TokenLt
	TokenGtEq
	TokenLtEq
	TokenIdentStr
	TokenIdentInt
	TokenConstStr
	TokenConstInt
)

// Token ...
type Token struct {
	Type       TokenType
	IntData    int
	StringData string
}

func validIdentifierStrP(word string) (bool, bool) {
	if word == "" {
		return false, false
	}
	lw := len(word)
	for i, ru := range word {
		if ru > 0xEF || (!unicode.IsLetter(ru) && !(i == lw-1 && ru == '$')) {
			return false, false
		}
	}
	return true, word[lw-1] == '$'
}

var errInvalidIf = fmt.Errorf("IF statements must be in the form IF...THEN or IF...THEN...ELSE")

func lexExpr(words []string) ([]Token, error) {
	ret := make([]Token, 0, len(words))
	for _, word := range words {
		switch word {
		case "<":
			ret = append(ret, Token{Type: TokenLt})
		case "<=":
			ret = append(ret, Token{Type: TokenLtEq})
		case ">":
			ret = append(ret, Token{Type: TokenGt})
		case ">=":
			ret = append(ret, Token{Type: TokenGtEq})
		case "=", "==":
			ret = append(ret, Token{Type: TokenEq})
		default:
			valid, stringp := validIdentifierStrP(word)
			if valid {
				if stringp {
					return nil, fmt.Errorf("Strings are not allowed in comparisons")
				}
				ret = append(ret, Token{Type: TokenIdentInt, StringData: word})
			}
			num, err := strconv.Atoi(word)
			if err != nil {
				return nil, fmt.Errorf("Bad number \"%s\": %s", word, err.Error())
			}
			ret = append(ret, Token{Type: TokenConstInt, IntData: num})
		}
	}
	return ret, nil
}

// Lex ...
// Lexes the list of words. Returns a list of tokens, or non-nil error if it can't lex.
func Lex(words []string) ([]Token, error) {
	ret := make([]Token, 0, len(words))
	switch strings.ToUpper(words[0]) {
	case "IF":
		if len(words) < 4 {
			return nil, errInvalidIf
		}
		ret = append(ret, Token{Type: TokenIf})
		thenPos := -1
		elsePos := -1
		for i, word := range words[1:] {
			if strings.ToUpper(word) == "THEN" {
				thenPos = i + 1
			} else if strings.ToUpper(word) == "ELSE" {
				elsePos = i + 1
			}
		}
		if thenPos == -1 || (elsePos != -1 && elsePos < thenPos) {
			return nil, errInvalidIf
		}
		ifexpr, err := lexExpr(words[1:thenPos])
		if err != nil {
			return nil, err
		}
		ret = append(ret, ifexpr...)
		if elsePos == -1 {
			thenexpr, err := Lex(words[thenPos+1 : elsePos])
			if err != nil {
				return nil, err
			}
			ret = append(ret, thenexpr...)
		} else {
			thenexpr, err := Lex(words[thenPos+1 : elsePos])
			if err != nil {
				return nil, err
			}
			ret = append(ret, thenexpr...)
			elseexpr, err := Lex(words[elsePos+1:])
			if err != nil {
				return nil, err
			}
			ret = append(ret, elseexpr...)
		}
		return ret, nil
	case "GOTO":
		if len(words) == 1 {
			return nil, fmt.Errorf("GOTO statement requires a line number")
		}
		num, err := strconv.Atoi(words[1])
		if err != nil {
			return nil, fmt.Errorf("Bad line number \"%s\"; %s", words[1], err.Error())
		} else if 0 <= num && num < MaxLines {
			ret = append(ret, Token{Type: TokenGoto, IntData: num})
		} else {
			return nil, fmt.Errorf("Line number must be in the range 0-%d", MaxLines)
		}
	case "EXIT", "QUIT", "BYE", "END":
		ret = append(ret, Token{Type: TokenExit})
	case "LET":
		if len(words) < 4 {
			return nil, fmt.Errorf("Expected at least one identifier in LET clause")
		}
		ret = append(ret, Token{Type: TokenLet})
		state := 'i'
		snarf := 1
		for _, word := range words[1:] {
			switch state {
			case 'i':
				valid, string := validIdentifierStrP(word)
				if valid {
					if string {
						ret = append(ret, Token{Type: TokenIdentStr, StringData: word})
					} else {
						ret = append(ret, Token{Type: TokenIdentInt, StringData: word})
					}
					snarf++
				} else {
					return nil, fmt.Errorf("Bad identifier %s", word)
				}
				state = 'e'
			case 'e':
				if word == "=" {
					state = 'c'
				} else {
					return nil, fmt.Errorf("Unknown token %s in LET clause", word)
				}
			case 'c':
				if len(word) == 0 {
					return nil, fmt.Errorf("Empty constant in LET clause")
				} else if word[0] == '"' {
					lw := len(word)
					if lw == 1 {
						ret = append(ret, Token{Type: TokenConstStr, StringData: ""})
						state = 's'
					} else if word[lw-1] == '"' {
						ret = append(ret, Token{Type: TokenConstStr, StringData: word[1 : len(word)-1]})
						snarf++
						state = 'i'
					} else {
						ret = append(ret, Token{Type: TokenConstStr, StringData: word[1:]})
						state = 's'
					}
				} else {
					num, err := strconv.Atoi(word)
					if err != nil {
						return nil, fmt.Errorf("Integer constant %s not parsed: %s", word, err.Error())
					}
					ret = append(ret, Token{Type: TokenConstInt, IntData: num})
					snarf++
					state = 'i'
				}
			case 's':
				ret[snarf].StringData += " " + word
				wl := len(ret[snarf].StringData)
				if ret[snarf].StringData[wl-1] == '"' {
					ret[snarf].StringData = ret[snarf].StringData[:wl-1]
					state = 'i'
					snarf++
				}
			}
		}
		if state != 'i' {
			switch state {
			case 'e':
				return nil, fmt.Errorf("Expected constant")
			case 's':
				return nil, fmt.Errorf("Unterminated string")
			default:
				return nil, fmt.Errorf("Lexer entered error state %c", state)
			}
		}
	case "PRINT":
		ret = append(ret, Token{Type: TokenPrint})
		state := 'i'
		snarf := 1
		for _, word := range words[1:] {
			switch state {
			case 'i':
				if word == "" {
					continue
				}

				valid, string := validIdentifierStrP(word)
				if valid {
					if string {
						ret = append(ret, Token{Type: TokenIdentStr, StringData: word})
					} else {
						ret = append(ret, Token{Type: TokenIdentInt, StringData: word})
					}
					snarf++
				} else if word[0] == '"' {
					lw := len(word)
					if lw == 1 {
						ret = append(ret, Token{Type: TokenConstStr, StringData: ""})
						state = 's'
					} else if word[lw-1] == '"' {
						ret = append(ret, Token{Type: TokenConstStr, StringData: word[1 : len(word)-1]})
						snarf++
						state = 'i'
					} else {
						ret = append(ret, Token{Type: TokenConstStr, StringData: word[1:]})
						state = 's'
					}
				} else {
					num, err := strconv.Atoi(word)
					if err != nil {
						return nil, fmt.Errorf("Invalid identifier %s", word)
					}
					ret = append(ret, Token{Type: TokenConstInt, IntData: num})
					snarf++
				}
			case 's':
				ret[snarf].StringData += " " + word
				wl := len(ret[snarf].StringData)
				if ret[snarf].StringData[wl-1] == '"' {
					ret[snarf].StringData = ret[snarf].StringData[:wl-1]
					state = 'i'
					snarf++
				}
			}
		}
		if state == 's' {
			return nil, fmt.Errorf("Unterminated string")
		}
	default:
		return nil, fmt.Errorf("Unknown keyword %s", strings.ToUpper(words[0]))
	}
	return ret, nil
}

func (t Token) String() string {
	switch t.Type {
	case TokenIf:
		return "IF"
	case TokenLet:
		return "LET"
	case TokenPrint:
		return "PRINT"
	case TokenExit:
		return "END"
	case TokenGoto:
		return fmt.Sprintf("GOTO %d", t.IntData)
	case TokenIdentStr:
		return t.StringData
	case TokenIdentInt:
		return t.StringData
	case TokenConstStr:
		return fmt.Sprintf("\"%s\"", t.StringData)
	case TokenConstInt:
		return strconv.Itoa(t.IntData)
	default:
		return fmt.Sprintf("{type: %d, str: %s, int: %d}", t.Type, t.StringData, t.IntData)
	}
}
