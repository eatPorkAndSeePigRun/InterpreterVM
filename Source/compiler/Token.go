package compiler

import "strconv"

const (
	TokenAnd = 256 + iota
	TokenBreak
	TokenDo
	TokenElse
	TokenElseif
	TokenEnd
	TokenFalse
	TokenFor
	TokenFunction
	TokenIf
	TokenIn
	TokenLocal
	TokenNil
	TokenNot
	TokenOr
	TokenRepeat
	TokenId
	TokenString
	TokenNumber
	TokenEqual
	TokenNotEqual
	TokenLessEqual
	TokenGreaterEqual
	TokenConcat
	TokenVarArg
	TokenEOF
)

var tokenStr = []string{
	"and", "break", "do", "else", "elseif", "end",
	"false", "for", "function", "if", "in",
	"local", "nil", "not", "or", "repeat",
	"return", "then", "true", "false", "until", "while",
}

type TokenDetail struct {
	number float64 // number for TokenNumber
	str    *String // string for TokenId, TokenKeyWord and TokenString

	module *String // module name of this token belongs to
	line   int     // token line number in module
	column int     // token column number at 'line'
	token  int     // token value
}

func GetTokenStr(t TokenDetail) string {
	var str string

	token := t.token
	if token == TokenNumber {
		str = strconv.FormatFloat(t.number, 'f', 6, 64)
	} else if (token == TokenId) || (token == TokenString) {
		str = t.str.GetStdString()
	} else if (token >= TokenAnd) && (token <= TokenEOF) {
		str = tokenStr[token-TokenAnd]
	} else {
		str = str + string(token)
	}

	return str
}
