package compiler

import (
	"strconv"
	"unicode"
)

const EOF = 0

var keyword = []string{
	"and", "break", "do", "else", "elseif", "end",
	"false", "for", "function", "if", "in",
	"local", "nil", "not", "or", "repeat",
	"return", "then", "true", "until", "while",
}

func isKeyWord(name string, token *int) bool {
	if token == nil {
		panic("assert")
	}

	for i, v := range keyword {
		if name == v {
			*token = TokenAnd + i
			return true
		}
	}
	return false
}

func isHexChar(c uint8) bool {
	return (c >= '0' && c <= '9') ||
		(c >= 'a' && c <= 'f') ||
		(c >= 'A' && c <= 'F')
}

func (l Lexer) normalTokenDetail(detail *TokenDetail, token int) int {
	detail.token = token
	detail.line = l.line
	detail.column = l.column
	detail.module = l.module
	return token
}

func (l Lexer) numberTokenDetail(detail *TokenDetail, number float64) int {
	detail.number = number
	return l.normalTokenDetail(detail, TokenNumber)
}

func (l Lexer) tokenDetail(detail *TokenDetail, str string, token int) int {
	detail.str = l.state.GetString(str)
	return l.normalTokenDetail(detail, token)
}

func (l Lexer) setEofTokenDetail(detail *TokenDetail) {
	detail.str = nil
	detail.token = TokenEOF
	detail.line = l.line
	detail.column = l.column
	detail.module = l.module
}

type CharInStream func() uint8

type Lexer struct {
	state    *State
	module   *String
	inStream CharInStream

	current uint8
	line    int
	column  int

	tokenBuffer string
}

func NewLexer(state *State, module *String, in CharInStream) Lexer {
	var l Lexer
	l.state = state
	l.module = module
	l.inStream = in
	l.current = EOF
	l.line = 1
	l.column = 0
	return l
}

// Get next token, 'detail' store next token detail information,
// return value is next token type.
func (l *Lexer) GetToken(detail *TokenDetail) (int, error) {
	if detail == nil {
		panic("assert")
	}

	l.setEofTokenDetail(detail)
	if l.current == EOF {
		l.current = l.next()
	}

	for l.current != EOF {
		switch l.current {
		case ' ', '\t', '\v', '\f':
			l.current = l.next()
		case '\r', '\n':
			l.lexNewLine()
		case '-':
			next := l.next()
			if next == '-' {
				return -1, l.lexComment()
			} else {
				l.current = next
				return l.normalTokenDetail(detail, '-'), nil
			}
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			return l.lexNumber(detail)
		case '+', '*', '/', '%', '^', '#', '(', ')', '{', '}',
			']', ';', ':', ',':
			token := int(l.current)
			l.current = l.next()
			return l.normalTokenDetail(detail, token), nil
		case '.':
			next := l.next()
			if next == '.' {
				preNext := l.next()
				if preNext == '.' {
					l.current = l.next()
					return l.normalTokenDetail(detail, TokenVarArg), nil
				} else {
					l.current = preNext
					return l.normalTokenDetail(detail, TokenConcat), nil
				}
			} else if unicode.IsDigit(rune(next)) {
				l.tokenBuffer = string(l.current)
				l.current = next
				return l.lexNUmberXFractional(detail, false, true,
					func(c uint8) bool { return unicode.IsDigit(rune(c)) },
					func(c uint8) bool { return c == 'e' || c == 'E' })
			} else {
				l.current = l.next()
				return l.normalTokenDetail(detail, '.'), nil
			}
		case '~':
			next := l.next()
			if next != '=' {
				return -1, NewLexError(l.module.GetCStr(), l.line, l.column, "expect '=' after '~'")
			}
			l.current = l.next()
			return l.normalTokenDetail(detail, TokenNotEqual), nil
		case '=':
			return l.lexXEqual(detail, TokenEqual), nil
		case '>':
			return l.lexXEqual(detail, TokenGreaterEqual), nil
		case '<':
			return l.lexXEqual(detail, TokenLessEqual), nil
		case '[':
			l.current = l.next()
			if l.current == '[' || l.current == '=' {
				return l.lexMultiLineString(detail)
			} else {
				return l.normalTokenDetail(detail, '['), nil
			}
		case '\'', '"':
			return l.lexSingleLineString(detail)
		default:
			return l.lexId(detail)
		}
	}

	return TokenEOF, nil
}

// Get current lex module name.
func (l Lexer) GetLexModule() *String {
	return l.module
}

func (l *Lexer) next() uint8 {
	c := l.inStream()
	if c != EOF {
		l.column++
	}
	return c
}

func (l *Lexer) lexNewLine() {
	next := l.next()
	if (next == '\r' || next == '\n') && (next != l.current) {
		l.current = l.next()
	} else {
		l.current = next
	}
	l.line++
	l.column = 0
}

func (l *Lexer) lexComment() error {
	l.current = l.next()
	if l.current == '[' {
		l.current = l.next()
		if l.current == '[' {
			return l.lexMultiLineComment()
		} else {
			l.lexSingleLineComment()
		}
	} else {
		l.lexSingleLineComment()
	}
	return nil
}

func (l *Lexer) lexMultiLineComment() error {
	var isCommentEnd bool
	for !isCommentEnd {
		if l.current == ']' {
			l.current = l.next()
			if l.current == ']' {
				isCommentEnd = true
				l.current = l.next()
			}
		} else if l.current == EOF {
			// uncompleted multi-line comment
			return NewLexError(l.module.GetCStr(), l.line, l.column,
				"expect complete multi-line comment before <eof>")
		} else if l.current == '\r' || l.current == '\n' {
			l.lexNewLine()
		} else {
			l.current = l.next()
		}
	}

	return nil
}

func (l *Lexer) lexSingleLineComment() {
	for l.current != '\r' && l.current != '\n' && l.current != EOF {
		l.current = l.next()
	}
}

func (l *Lexer) lexNumber(detail *TokenDetail) (int, error) {
	var integerPart bool
	l.tokenBuffer = ""
	if l.current == '0' {
		next := l.next()
		if next == 'x' || next == 'X' {
			l.tokenBuffer = l.tokenBuffer + string(l.current)
			l.tokenBuffer = l.tokenBuffer + string(next)
			l.current = l.next()

			return l.lexNumberX(detail, false, isHexChar,
				func(c uint8) bool { return c == 'p' || c == 'P' })
		} else {
			l.tokenBuffer = l.tokenBuffer + string(l.current)
			l.current = next
			integerPart = true
		}
	}

	return l.lexNumberX(detail, integerPart,
		func(c uint8) bool { return !unicode.IsDigit(rune(c)) },
		func(c uint8) bool { return c == 'e' || c == 'E' })
}

func (l *Lexer) lexNumberX(detail *TokenDetail, integerPart bool,
	isNumberChar func(uint8) bool, isExponent func(uint8) bool) (int, error) {
	for isNumberChar(l.current) {
		l.tokenBuffer = l.tokenBuffer + string(l.current)
		l.current = l.next()
		integerPart = true
	}

	var point = false
	if l.current == '.' {
		l.tokenBuffer = l.tokenBuffer + string(l.current)
		l.current = l.next()
		point = true
	}

	return l.lexNUmberXFractional(detail, integerPart, point, isNumberChar, isExponent)
}

func (l *Lexer) lexNUmberXFractional(detail *TokenDetail, integerPart bool, point bool,
	isNumberChar func(uint8) bool, isExponent func(uint8) bool) (int, error) {
	fractionalPart := false
	for isNumberChar(l.current) {
		l.tokenBuffer = l.tokenBuffer + string(l.current)
		l.current = l.next()
		fractionalPart = true
	}

	if point && !integerPart && !fractionalPart {
		return -1, NewLexError(l.module.GetCStr(), l.line, l.column, "unexpect '.'")
	} else if !point && !integerPart && !fractionalPart {
		return -1, NewLexError(l.module.GetCStr(), l.line, l.column,
			"unexpect incomplete number'", l.tokenBuffer, "'")
	}

	if isExponent(l.current) {
		l.tokenBuffer = l.tokenBuffer + string(l.current)
		l.current = l.next()
		if l.current == '-' || l.current == '+' {
			l.tokenBuffer = l.tokenBuffer + string(l.current)
			l.current = l.next()
		}

		if !unicode.IsDigit(rune(l.current)) {
			return -1, NewLexError(l.module.GetCStr(), l.line, l.column,
				"expect exponent after '", l.tokenBuffer, "'")
		}

		for unicode.IsDigit(rune(l.current)) {
			l.tokenBuffer = l.tokenBuffer + string(l.current)
			l.current = l.next()
		}
	}

	number, err := strconv.ParseFloat(l.tokenBuffer, 64)
	panic(err)
	return l.numberTokenDetail(detail, number), nil
}

func (l *Lexer) lexXEqual(detail *TokenDetail, equalToken int) int {
	token := int(l.current)

	next := l.next()
	if next == '=' {
		l.current = l.next()
		return l.normalTokenDetail(detail, equalToken)
	} else {
		l.current = next
		return l.normalTokenDetail(detail, token)
	}
}

func (l *Lexer) lexMultiLineString(detail *TokenDetail) (int, error) {
	equals := 0
	for l.current == '=' {
		equals++
		l.current = l.next()
	}

	if l.current != '[' {
		return -1, NewLexError(l.module.GetCStr(), l.line, l.column,
			"incomplete multi-line string at '", l.tokenBuffer, "'")
	}

	l.current = l.next()
	l.tokenBuffer = ""

	if l.current == '\r' || l.current == '\n' {
		l.lexNewLine()
		if equals == 0 { // "[[]]" keeps first '\n'
			l.tokenBuffer = l.tokenBuffer + string('\n')
		}
	}

	for l.current != EOF {
		if l.current == ']' {
			l.current = l.next()
			i := 0
			for ; i < equals; i++ {
				if l.current != '=' {
					break
				}
				l.current = l.next()
			}

			if i == equals && l.current == ']' {
				l.current = l.next()
				l.tokenDetail(detail, l.tokenBuffer, TokenString)
			} else {
				l.tokenBuffer = l.tokenBuffer + "]"
				for j := 0; j < i; j++ {
					l.tokenBuffer = l.tokenBuffer + string('=')
				}
			}
		} else if l.current == '\r' || l.current == '\n' {
			l.lexNewLine()
			l.tokenBuffer = l.tokenBuffer + string('\n')
		} else {
			l.tokenBuffer = l.tokenBuffer + string(l.current)
			l.current = l.next()
		}
	}

	return -1, NewLexError(l.module.GetCStr(), l.line, l.column, "incomplete multi-line string at <eof>")
}

func (l *Lexer) lexSingleLineString(detail *TokenDetail) (int, error) {
	quote := l.current
	l.current = l.next()
	l.tokenBuffer = ""

	for l.current != quote {
		if l.current == EOF {
			return -1, NewLexError(l.module.GetCStr(), l.line, l.column,
				"incomplete string at <eof>")
		}
		if l.current == '\r' || l.current == '\n' {
			return -1, NewLexError(l.module.GetCStr(), l.line, l.column,
				"incomplete string at this line")
		}
		return -1, l.lexStringChar()
	}

	l.current = l.next()
	return l.tokenDetail(detail, l.tokenBuffer, TokenString), nil
}

func (l *Lexer) lexStringChar() error {
	if l.current == '\\' {
		l.current = l.next()
		if l.current == 'a' {
			l.tokenBuffer = l.tokenBuffer + string('\a')
		} else if l.current == 'b' {
			l.tokenBuffer = l.tokenBuffer + string('\b')
		} else if l.current == 'f' {
			l.tokenBuffer = l.tokenBuffer + string('\f')
		} else if l.current == 'n' {
			l.tokenBuffer = l.tokenBuffer + string('\n')
		} else if l.current == 'r' {
			l.tokenBuffer = l.tokenBuffer + string('\r')
		} else if l.current == 't' {
			l.tokenBuffer = l.tokenBuffer + string('\t')
		} else if l.current == 'v' {
			l.tokenBuffer = l.tokenBuffer + string('\v')
		} else if l.current == '\\' {
			l.tokenBuffer = l.tokenBuffer + string('\\')
		} else if l.current == '"' {
			l.tokenBuffer = l.tokenBuffer + string('"')
		} else if l.current == '\'' {
			l.tokenBuffer = l.tokenBuffer + string('\'')
		} else if l.current == 'x' {
			l.current = l.next()
			var hex string
			var i int
			for ; i < 2 && isHexChar(l.current); i++ {
				hex = hex + string(l.current)
				l.current = l.next()
			}
			if i == 0 {
				return NewLexError(l.module.GetCStr(), l.line, l.column,
					"unexpect character after '\\x'")
			}
			num, err := strconv.ParseInt(hex, 16, 8)
			if err != nil {
				panic(err)
			}
			l.tokenBuffer = string(num)
			return nil
		} else if unicode.IsDigit(rune(l.current)) {
			var oct string
			for i := 0; i < 3 && unicode.IsDigit(rune(l.current)); i++ {
				oct = oct + string(l.current)
				l.current = l.next()
			}
			num, err := strconv.ParseInt(oct, 8, 8)
			if err != nil {
				panic(err)
			}
			l.tokenBuffer = l.tokenBuffer + string(num)
			return nil
		} else {
			return NewLexError(l.module.GetCStr(), l.line, l.column,
				"unexpect character after '\\'")
		}
	} else {
		l.tokenBuffer = l.tokenBuffer + string(l.current)
	}

	l.current = l.next()
	return nil
}

func (l *Lexer) lexId(detail *TokenDetail) (int, error) {
	if !unicode.IsLetter(rune(l.current)) && l.current != '_' {
		return -1, NewLexError(l.module.GetCStr(), l.line, l.column, "unexpect character")
	}

	l.tokenBuffer = string(l.current)
	l.current = l.next()

	for unicode.IsLetter(rune(l.current)) || l.current == '_' {
		l.tokenBuffer = l.tokenBuffer + string(l.current)
		l.current = l.next()
	}

	var token int
	if !isKeyWord(l.tokenBuffer, &token) {
		token = TokenId
	}

	return l.tokenDetail(detail, l.tokenBuffer, token), nil
}
