package luna

import (
	"strconv"
	"unicode"
)

var keyword = []string{
	"and", "break", "do", "else", "elseif", "end",
	"false", "for", "function", "if", "in",
	"local", "nil", "not", "or", "repeat",
	"return", "then", "true", "until", "while",
}

func isKeyWord(name string, token *int32) bool {
	i := 0
	for ; i < len(keyword); i++ {
		if name == keyword[i] {
			*token = int32(TokenAnd + i)
			return true
		}
	}
	return false
}

func isHexChar(c int32) bool {
	return (c >= '0' && c <= '9') ||
		(c >= 'a' && c <= 'f') ||
		(c >= 'A' && c <= 'F')
}

func (lexer Lexer) normalTokenDetail(detail *TokenDetail, token int32) int32 {
	detail.token = token
	detail.line = lexer.line
	detail.column = lexer.column
	detail.module = lexer.module
	return token
}

func (lexer Lexer) numberTokenDetail(detail *TokenDetail, number float64) int32 {
	detail.number = number
	return lexer.normalTokenDetail(detail, TokenNumber)
}

func (lexer Lexer) tokenDetail(detail *TokenDetail, str string, token int32) int32 {
	detail.str = lexer.state.GetString(str)
	return lexer.normalTokenDetail(detail, token)
}

func (lexer Lexer) setEofTokenDetail(detail *TokenDetail) {
	detail.str = nil
	detail.token = TokenEOF
	detail.line = lexer.line
	detail.column = lexer.column
	detail.module = lexer.module
}

type CharInStream func() int32

type Lexer struct {
	state    *State
	module   *String
	inStream CharInStream

	current int32
	line    int64
	column  int64

	tokenBuffer string
}

// Get next token, 'detail' store next token detail information,
// return value is next token type.
func (lexer Lexer) GetToken(detail *TokenDetail) int32 {
	lexer.setEofTokenDetail(detail)
	if lexer.current == -1 {
		lexer.current = lexer.next()
	}

	for ;lexer.current != -1; {
		switch lexer.current {
		case ' ', '\t', '\v', '\f':
			lexer.current = lexer.next()
		case '\r', '\n':
			lexer.lexNewLine()
		case '-':
			next := lexer.next()
			if next == '-' {
				lexer.lexComment()
			} else {
				lexer.current = next
				return lexer.normalTokenDetail(detail, '-')
			}
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			return lexer.lexNumber(detail)
		case '+', '*', '/', '%', '^', '#', '(', ')', '{', '}',
			']', ';', ':', ',':
			token := lexer.current
			lexer.current = lexer.next()
			return lexer.normalTokenDetail(detail, token)
		case '.':
			next := lexer.next()
			if next == '.' {
				preNext := lexer.next()
				if preNext == '.' {
					lexer.current = lexer.next()
					return lexer.normalTokenDetail(detail, TokenVarArg)
				} else {
					lexer.current = preNext
					return lexer.normalTokenDetail(detail, TokenConcat)
				}
			} else if unicode.IsDigit(next) {
				lexer.tokenBuffer = string(lexer.current)
				lexer.current = next
				return lexer.lexNUmberXFractional(detail, false, true,
					func(c int32) bool { return unicode.IsDigit(c) },
					func(c int32) bool { return c == 'e' || c == 'E' })
			} else {
				lexer.current = lexer.next()
				return lexer.normalTokenDetail(detail, '.')
			}
		case '~':
			next := lexer.next()
			if next != '=' {
				// TODO
			}
		case '=':
			return lexer.lexXEqual(detail, TokenEqual)
		case '>':
			return lexer.lexXEqual(detail, TokenGreaterEqual)
		case '<':
			return lexer.lexXEqual(detail, TokenLessEqual)
		case '[':
			lexer.current = lexer.next()
			if lexer.current == '[' || lexer.current == '=' {
				return lexer.lexMultiLineString(detail)
			} else {
				return lexer.normalTokenDetail(detail, '[')
			}
		case '\'', '"':
			return lexer.lexSingleLineString(detail)
		default:
			return lexer.lexId(detail)
		}
	return TokenEOF
	}
}

// Get current lex module name.
func (lexer Lexer) GetLexModule() *String {
	return lexer.module
}

func (lexer Lexer) next() int32 {
	c := lexer.inStream()
	if c != -1 {
		lexer.column++
	}
	return c
}

func (lexer Lexer) lexNewLine() {
	next := lexer.next()
	if (next == '\r' || next == '\n') && (next != lexer.current) {
		lexer.current = lexer.next()
	} else {
		lexer.current = next
	}
	lexer.line++
	lexer.column = 0
}

func (lexer Lexer) lexComment() {
	lexer.current = lexer.next()
	if lexer.current == '[' {
		lexer.current = lexer.next()
		if lexer.current == '[' {
			lexer.lexMultiLineComment()
		} else {
			lexer.lexSingleLineComment()
		}
	} else {
		lexer.lexSingleLineComment()
	}
}

func (lexer Lexer) lexMultiLineComment() LexException {
	isCommentEnd := false
	for ; isCommentEnd; {
		if lexer.current == ']' {
			lexer.current = lexer.next()
			if lexer.current == ']' {
				isCommentEnd = true
				lexer.current = lexer.next()
			}
		} else if lexer.current == -1 {
			// uncompleted multi-line comment
			return LexException{lexer.module.GetCStr(), lexer.line, lexer.column,
				"expect complete multi-line comment before <eof>"}
		} else if lexer.current == '\r' || lexer.current == '\n' {
			lexer.lexNewLine()
		} else {
			lexer.current = lexer.next()
		}
	}
}

func (lexer Lexer) lexSingleLineComment() {
	for ; lexer.current != '\r' &&
		lexer.current != '\n' &&
		lexer.current != -1; {
		lexer.current = lexer.next()
	}
}

func (lexer Lexer) lexNumber(detail *TokenDetail) int32 {
	integerPart := false
	lexer.tokenBuffer = ""
	if lexer.current == '0' {
		next := lexer.next()
		if next == 'x' || next == 'X' {
			lexer.tokenBuffer = lexer.tokenBuffer + string(lexer.current)
			lexer.tokenBuffer = lexer.tokenBuffer + string(next)
			lexer.current = lexer.next()

			return lexer.lexNumberX(detail, false, isHexChar,
				)
		}
	}
}

func (lexer Lexer) lexNumberX(detail *TokenDetail, integerPart bool,
	isNumberChar func(int32) bool,
	isExponent func(int64) bool) int64 {
	for ;isNumberChar(lexer.current);{
		lexer.tokenBuffer = lexer.tokenBuffer + string(lexer.current)
		lexer.current = lexer.next()
		integerPart = true
	}

	point := false
	if lexer.current == '.' {
		lexer.tokenBuffer = lexer.tokenBuffer + string(lexer.current)
		lexer.current = lexer.next()
		point  = true
	}

	return lexer.lexNUmberXFractional(detail, integerPart, point, &isNumberChar, &isExponent)
}

func (lexer Lexer) lexNUmberXFractional(detail *TokenDetail,
	integerPart bool, point bool,
	isNumberChar *func(int64) bool,
	isExponent *func(int64) bool) int64 {
		fractionalPart := false
		for ;isNumberChar(lexer.current); {
			lexer.tokenBuffer = lexer.tokenBuffer + string(lexer.current)
			lexer.current = lexer.next()
			fractionalPart = true
		}

		if point && !integerPart && !fractionalPart {
			// TODO
		} else if !point && !integerPart && !fractionalPart {
			// TODO
		}

		if isExponent(lexer.current) {
			lexer.tokenBuffer = lexer.tokenBuffer + string(lexer.current)
			lexer.current = lexer.next()
			if lexer.current == '-' || lexer.current == '+' {
				lexer.tokenBuffer = lexer.tokenBuffer + string(lexer.current)
				lexer.current = lexer.next()
			}

			if !unicode.IsDigit(lexer.current) {
				// TODO
			}

			for ;unicode.IsDigit(lexer.current); {
				lexer.tokenBuffer = lexer.tokenBuffer + string(lexer.current)
				lexer.current = lexer.next()
			}
		}

		number := strconv.ParseFloat(lexer.tokenBuffer, 64)
		return lexer.numberTokenDetail(detail, number)
}

func (lexer Lexer) lexXEqual(detail *TokenDetail, equalToken int32) int32 {
	token := lexer.current

	next := lexer.next()
	if next == '=' {
		lexer.current = lexer.next()
		return lexer.normalTokenDetail(detail, equalToken)
	} else {
		lexer.current = next
		return lexer.normalTokenDetail(detail, token)
	}
}

func (lexer Lexer) lexMultiLineString(detail *TokenDetail) int32 {
	var equals int64 = 0
	for ;lexer.current == '='; {
		equals++
		lexer.current = lexer.next()
	}

	if lexer.current != '['
		// TODO

	lexer.current = lexer.next()
	lexer.tokenBuffer = ""

	if lexer.current == '\r' || lexer.current == '\n' {
		lexer.lexNewLine()
		if equals == 0 {
			lexer.tokenBuffer = lexer.tokenBuffer + string('\n')
		}
	}
}

func (lexer Lexer) lexSingleLineString(detail *TokenDetail) int32 {
	quote := lexer.current
	lexer.current = lexer.next()
	lexer.tokenBuffer = ""

	for ;lexer.current != quote; {
		if lexer.current == -1 {
			// TODO
		}
		if lexer.current == '\r' || lexer.current == '\n' {
			// TODO
		}
		lexer.lexStringChar()
	}

	lexer.current = lexer.next()
	return lexer.tokenDetail(detail, lexer.tokenBuffer, TokenString)
}

func (lexer Lexer) lexStringChar() {
	if lexer.current == '\\' {
		lexer.current = lexer.next()
		if lexer.current == 'a' {
			lexer.tokenBuffer = lexer.tokenBuffer + string('\a')
		} else if lexer.current == 'b' {
			lexer.tokenBuffer = lexer.tokenBuffer + string('\b')
		} else if lexer.current == 'f' {
			lexer.tokenBuffer = lexer.tokenBuffer + string('\f')
		} else if lexer.current == 'n' {
			lexer.tokenBuffer = lexer.tokenBuffer + string('\n')
		} else if lexer.current == 'r' {
			lexer.tokenBuffer = lexer.tokenBuffer + string('\r')
		} else if lexer.current == 't' {
			lexer.tokenBuffer = lexer.tokenBuffer + string('\t')
		} else if lexer.current == 'v' {
			lexer.tokenBuffer = lexer.tokenBuffer + string('\v')
		} else if lexer.current == '\\' {
			lexer.tokenBuffer = lexer.tokenBuffer + string('\\')
		} else if lexer.current == '"' {
			lexer.tokenBuffer = lexer.tokenBuffer + string('"')
		} else if lexer.current == '\'' {
			lexer.tokenBuffer = lexer.tokenBuffer + string('\'')
		} else if lexer.current == 'x' {
			lexer.current = lexer.next()

		} else if unicode.IsDigit(lexer.current) {

		} else {

		}
	}
}

func (lexer Lexer) lexId(detail *TokenDetail) int32 {
	if !unicode.IsLetter(lexer.current) && lexer.current != '_' {
		// TODO panic()
	}

	lexer.tokenBuffer = string(lexer.current)
	lexer.current = lexer.next()

	for ;unicode.IsLetter(lexer.current) || lexer.current == '_'; {
		lexer.tokenBuffer = lexer.tokenBuffer + string(lexer.current)
		lexer.current = lexer.next()
	}

	var token int32 = 0
	if !isKeyWord(lexer.tokenBuffer, &token) {
		token = TokenId
	}

	return lexer.tokenDetail(detail, lexer.tokenBuffer, token)
}
