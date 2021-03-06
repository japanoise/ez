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
	TokenFieldSep
	TokenInput
	TokenPrint
	TokenExit
	TokenGoto
	TokenIf
	TokenThen
	TokenElse
	TokenEq
	TokenNe
	TokenGt
	TokenLt
	TokenGtEq
	TokenLtEq
	TokenIdentStr
	TokenIdentInt
	TokenConstStr
	TokenConstInt
	TokenAdd
	TokenSub
	TokenMul
	TokenDiv
	TokenAnd
	TokenOr
	TokenXor
)

// Token ...
type Token struct {
	Type       TokenType
	IntData    int
	StringData string
}

func isOperatorType(t TokenType) bool {
	return t >= TokenAdd
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

var errInvalidInput = fmt.Errorf("INPUT statements must be in the form INPUT PROMPT VAR")

func lexOp(word string) *Token {
	switch word {
	case "+":
		return &Token{Type: TokenAdd}
	case "-":
		return &Token{Type: TokenSub}
	case "*":
		return &Token{Type: TokenMul}
	case "/":
		return &Token{Type: TokenMul}
	case "&":
		return &Token{Type: TokenAnd}
	case "|":
		return &Token{Type: TokenOr}
	case "^":
		return &Token{Type: TokenXor}
	case "<":
		return &Token{Type: TokenLt}
	case "<=":
		return &Token{Type: TokenLtEq}
	case ">":
		return &Token{Type: TokenGt}
	case ">=":
		return &Token{Type: TokenGtEq}
	case "=", "==":
		return &Token{Type: TokenEq}
	case "!=":
		return &Token{Type: TokenNe}
	default:
		return nil
	}
}

func lexExpr(words []string) ([]Token, error) {
	ret := make([]Token, 0, len(words))
	snarf := -1
	for i, word := range words {
		if snarf != -1 {
			ret[snarf].StringData += " " + word
			wl := len(ret[snarf].StringData)
			if ret[snarf].StringData[wl-1] == '"' {
				ret[snarf].StringData = ret[snarf].StringData[:wl-1]
				snarf = -1
			}
		}
		op := lexOp(word)
		if op != nil {
			ret = append(ret, *op)
			continue
		}

		valid, stringp := validIdentifierStrP(word)
		if valid {
			if stringp {
				ret = append(ret, Token{Type: TokenIdentStr, StringData: word})
			} else {
				ret = append(ret, Token{Type: TokenIdentInt, StringData: word})
			}
			continue
		}

		if len(word) > 0 && word[0] == '"' {
			lw := len(word)
			if lw == 1 {
				ret = append(ret, Token{Type: TokenConstStr, StringData: ""})
			} else if word[lw-1] == '"' {
				ret = append(ret, Token{Type: TokenConstStr, StringData: word[1 : len(word)-1]})
			} else {
				ret = append(ret, Token{Type: TokenConstStr, StringData: word[1:]})
				snarf = i
			}
			continue
		}

		num, err := strconv.Atoi(word)
		if err != nil {
			return nil, fmt.Errorf("Bad number \"%s\": %s", word, err.Error())
		}
		ret = append(ret, Token{Type: TokenConstInt, IntData: num})
	}
	return ret, nil
}

// Lex ...
// Lexes the list of words. Returns a list of tokens, or non-nil error if it can't lex.
func Lex(words []string) ([]Token, error) {
	ret := make([]Token, 0, len(words))
	switch strings.ToUpper(words[0]) {
	case "IF":
		lw := len(words)
		if lw < 4 {
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
			if thenPos == lw-1 {
				return nil, errInvalidIf
			}
			thenexpr, err := Lex(words[thenPos+1:])
			if err != nil {
				return nil, err
			}
			ret = append(ret, Token{Type: TokenThen})
			ret = append(ret, thenexpr...)
		} else {
			if elsePos == lw-1 || thenPos == lw-1 {
				return nil, errInvalidIf
			}
			thenexpr, err := Lex(words[thenPos+1 : elsePos])
			if err != nil {
				return nil, err
			}
			ret = append(ret, Token{Type: TokenThen})
			ret = append(ret, thenexpr...)
			elseexpr, err := Lex(words[elsePos+1:])
			if err != nil {
				return nil, err
			}
			ret = append(ret, Token{Type: TokenElse})
			ret = append(ret, elseexpr...)
		}
		return ret, nil
	case "GOTO":
		if len(words) == 1 {
			return nil, fmt.Errorf("GOTO statement requires a line number")
		}

		valid, string := validIdentifierStrP(words[1])
		if valid {
			if string {
				return nil, fmt.Errorf("GOTO statement cannot use string variables")
			}
			ret = append(ret, Token{Type: TokenGoto, StringData: words[1]})
			break
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
				op := lexOp(word)
				if op != nil {
					ret = append(ret, *op)
					snarf++
					state = 'c'
				} else if word == ";" {
					ret = append(ret, Token{Type: TokenFieldSep})
					snarf++
					state = 'i'
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
						state = 'e'
					} else {
						ret = append(ret, Token{Type: TokenConstStr, StringData: word[1:]})
						state = 's'
					}
				} else {
					valid, string := validIdentifierStrP(word)
					if valid {
						if string {
							ret = append(ret, Token{Type: TokenIdentStr, StringData: word})
						} else {
							ret = append(ret, Token{Type: TokenIdentInt, StringData: word})
						}
						snarf++
						state = 'e'
						continue
					}
					num, err := strconv.Atoi(word)
					if err != nil {
						return nil, fmt.Errorf("Integer constant %s not parsed: %s", word, err.Error())
					}
					ret = append(ret, Token{Type: TokenConstInt, IntData: num})
					snarf++
					state = 'e'
				}
			case 's':
				ret[snarf].StringData += " " + word
				wl := len(ret[snarf].StringData)
				if ret[snarf].StringData[wl-1] == '"' {
					ret[snarf].StringData = ret[snarf].StringData[:wl-1]
					state = 'e'
					snarf++
				}
			}
		}
		if !(state == 'e') {
			switch state {
			case 'i':
				return nil, fmt.Errorf("Incomplete LET expression; expected identifier")
			case 'c':
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
	case "INPUT":
		if len(words) < 3 {
			return nil, errInvalidInput
		}
		ret = append(ret, Token{Type: TokenInput})
		i := 2

		valid, string := validIdentifierStrP(words[1])

		if valid {
			if string {
				ret = append(ret, Token{Type: TokenIdentStr, StringData: words[1]})
			} else {
				return nil, fmt.Errorf("Cannot use integer variable as a prompt in INPUT")
			}
		} else if len(words[1]) > 0 && words[1][0] == '"' {
			str := words[1][1:]
			done := false
			for i = 2; i < len(words); i++ {
				if words[i] == "\"" {
					str += " "
					done = true
					i++
					break
				}
				wl := len(words[i])
				if words[i][wl-1] == '"' {
					str += " " + words[i][:wl-1]
					done = true
					i++
					break
				}
				str += " " + words[i]
			}
			if !done {
				return nil, fmt.Errorf("Unterminated string")
			}
			if i >= len(words) {
				return nil, errInvalidInput
			}
			ret = append(ret, Token{Type: TokenConstStr, StringData: str})
		} else {
			return nil, errInvalidInput
		}

		valid, string = validIdentifierStrP(words[i])
		if valid {
			if string {
				ret = append(ret, Token{Type: TokenIdentStr, StringData: words[i]})
			} else {
				ret = append(ret, Token{Type: TokenIdentInt, StringData: words[i]})
			}
		} else {
			return nil, fmt.Errorf("Invalid identifier %s", words[i])
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
