package pluralforms

//go:generate goyacc -o parser.go -v "" parser.y

import (
	"fmt"
	"strconv"
)

type lexer struct {
	data string
	pos  int

	exp Expression
	err string
}

func (l *lexer) Lex(lval *yySymType) int {
	for {
		if l.pos >= len(l.data) {
			return 0
		}
		if l.data[l.pos] != ' ' && l.data[l.pos] != '\t' {
			break
		}
		l.pos += 1
	}

	pos := l.pos
	result := int(l.data[pos])
	l.pos += 1
	switch result {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		for l.pos < len(l.data) && l.data[l.pos] >= '0' && l.data[l.pos] <= '9' {
			l.pos += 1
		}
		if num, err := strconv.ParseInt(l.data[pos:l.pos], 10, 32); err == nil {
			lval.num = int(num)
			return numTok
		}
		return invalidTok
	case '=':
		if l.pos < len(l.data) && l.data[l.pos] == '=' {
			l.pos += 1
			return result
		}
		return invalidTok
	case '!':
		if l.pos < len(l.data) && l.data[l.pos] == '=' {
			l.pos += 1
			return neTok
		}
		return result
	case '&', '|':
		if l.pos < len(l.data) && l.data[l.pos] == l.data[pos] {
			l.pos += 1
			return result
		}
		return invalidTok
	case '<':
		if l.pos < len(l.data) && l.data[l.pos] == '=' {
			l.pos += 1
			return lteTok
		} else {
			return ltTok
		}
	case '>':
		if l.pos < len(l.data) && l.data[l.pos] == '=' {
			l.pos += 1
			return gteTok
		} else {
			return gtTok
		}
	case 'n', '?', ':', '(', ')', '*', '/', '%', '+', '-':
		// Return as is
		return result
	case ';', '\n':
		return 0
	default:
		return invalidTok
	}
}

func (l *lexer) Error(err string) {
	l.err = err
}

// Compile a string containing a plural form expression to a Expression object.
func Compile(expr string) (Expression, error) {
	l := lexer{data: expr}
	if yyParse(&l) != 0 {
		return nil, fmt.Errorf("cannot parse expression: %s", l.err)
	}
	return l.exp, nil
}
