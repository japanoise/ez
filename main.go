package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// MaxLines ...
// Maximum number of lines in the program
const MaxLines int = 0xFFFF

var stringVars = make(map[string]string)
var intVars = make(map[string]int)
var lines []*Line = make([]*Line, MaxLines)

func listLines() {
	for i, line := range lines {
		if line != nil && line.Used {
			fmt.Printf("%d: %s\n", i, line.Content)
		}
	}
}

func listLinesDebug() {
	for i, line := range lines {
		if line != nil && line.Used {
			fmt.Printf("%d:", i)
			for _, t := range line.Tokens {
				fmt.Print(t.String())
			}
			fmt.Println()
		}
	}
}

func predicateTrue(l []Token) (bool, error) {
	if len(l) < 3 {
		return false, fmt.Errorf("Predicate too small")
	}

	ints := false
	intl, intr := 0, 0
	strl, strr := "", ""

	switch l[0].Type {
	case TokenConstInt:
		intl = l[0].IntData
		ints = true
	case TokenConstStr:
		strl = l[0].StringData
		ints = false
	case TokenIdentInt:
		intl = intVars[l[0].StringData]
		ints = true
	case TokenIdentStr:
		strl = stringVars[l[0].StringData]
		ints = false
	default:
		return false, fmt.Errorf("Unexpected token on left hand side of IF statement %s", l[0].String())
	}

	switch l[2].Type {
	case TokenConstInt:
		if !ints {
			return false, fmt.Errorf("Cannot compare integer and string")
		}
		intr = l[2].IntData
	case TokenConstStr:
		if ints {
			return false, fmt.Errorf("Cannot compare integer and string")
		}
		strr = l[2].StringData
	case TokenIdentInt:
		if !ints {
			return false, fmt.Errorf("Cannot compare integer and string")
		}
		intr = intVars[l[2].StringData]
	case TokenIdentStr:
		if ints {
			return false, fmt.Errorf("Cannot compare integer and string")
		}
		strr = stringVars[l[2].StringData]
	default:
		return false, fmt.Errorf("Unexpected token on left hand side of IF statement %s", l[2].String())
	}

	switch l[1].Type {
	case TokenEq:
		if ints {
			return intl == intr, nil
		}
		return strl == strr, nil
	case TokenNe:
		if ints {
			return intl != intr, nil
		}
		return strl != strr, nil
	case TokenGt:
		if ints {
			return intl > intr, nil
		}
		return strl > strr, nil
	case TokenLt:
		if ints {
			return intl < intr, nil
		}
		return strl < strr, nil
	case TokenGtEq:
		if ints {
			return intl >= intr, nil
		}
		return strl >= strr, nil
	case TokenLtEq:
		if ints {
			return intl <= intr, nil
		}
		return strl <= strr, nil
	default:
		return false, fmt.Errorf("Not a comparison operator: %s", l[1].String())
	}
}

func execTokenList(l []Token) ([]Token, error) {
	switch l[0].Type {
	case TokenExit, TokenGoto:
		return l, nil
	case TokenIf:
		thenPos := -1
		elsePos := -1
		for i, token := range l[1:] {
			if token.Type == TokenThen {
				thenPos = i + 1
			} else if token.Type == TokenElse {
				elsePos = i + 1
			}
		}

		pred, err := predicateTrue(l[1:thenPos])
		if err != nil {
			return nil, err
		}

		if pred {
			if elsePos == -1 {
				return execTokenList(l[thenPos+1:])
			}
			return execTokenList(l[thenPos+1 : elsePos])
		} else if elsePos != -1 {
			return execTokenList(l[elsePos+1:])
		}
	case TokenLet:
		for i := 1; i < len(l); i += 2 {
			if l[i+1].Type == TokenConstInt {
				intVars[l[i].StringData] = l[i+1].IntData
			} else if l[i+1].Type == TokenConstStr {
				stringVars[l[i].StringData] = l[i+1].StringData
			} else {
				return nil, fmt.Errorf("Bad assignment in LET clause")
			}
		}
	case TokenPrint:
		for _, token := range l[1:] {
			switch token.Type {
			case TokenConstInt:
				fmt.Print(token.IntData)
			case TokenConstStr:
				fmt.Print(token.StringData)
			case TokenIdentInt:
				fmt.Print(intVars[token.StringData])
			case TokenIdentStr:
				fmt.Print(stringVars[token.StringData])
			}
		}
		fmt.Println()
	default:
		return nil, fmt.Errorf("Unexpected token in this context: %s", l[0].String())
	}
	return nil, nil
}

func execute(line *Line) ([]Token, error) {
	return execTokenList(line.Tokens)
}

func execLines(lines []*Line) {
	index := 0
	ll := len(lines)
	for index < ll {
		if lines[index] != nil && lines[index].Used {
			switch lines[index].Tokens[0].Type {
			case TokenExit:
				return
			case TokenGoto:
				index = lines[index].Tokens[0].IntData
				continue
			default:
				extraTokens, err := execute(lines[index])
				if err != nil {
					fmt.Println(err.Error())
					return
				}

				if extraTokens != nil {
					switch extraTokens[0].Type {
					case TokenExit:
						return
					case TokenGoto:
						index = extraTokens[0].IntData
						continue
					}
				}
			}
		}
		index++
	}
}

func main() {
	var reader io.Reader
	if len(os.Args) > 1 {
		file, err := os.Open(os.Args[1])
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		defer file.Close()
		reader = file
	} else {
		reader = os.Stdin
	}

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			continue
		}
		words := strings.Split(text, " ")
		switch strings.ToUpper(words[0]) {
		case "EXIT":
			os.Exit(0)
		case "RUN":
			execLines(lines)
			continue
		case "REM":
			continue
		case "LISTDEBUG":
			listLinesDebug()
			continue
		case "LIST":
			listLines()
			continue
		case "VARS":
			fmt.Println("Strings:", stringVars, "Integers:", intVars)
			continue
		}

		num, err := strconv.Atoi(words[0])
		if err == nil {
			if num < MaxLines && num >= 0 {
				lines[num], err = MakeLine(strings.Join(words[1:], " "))
				if err != nil {
					fmt.Printf("%d: %s\n", num, err.Error())
				}
			} else {
				fmt.Printf("Line number %d isn't in range 0-%d\n", num, MaxLines)
			}
		} else {
			line, err := MakeLine(text)
			if err != nil {
				fmt.Println(err.Error())
			} else {
				_, err = execute(line)
				if err != nil {
					fmt.Println(err.Error())
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
