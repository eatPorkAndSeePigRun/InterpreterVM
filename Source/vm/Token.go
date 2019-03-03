package vm

import (
	"strconv"
)

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
	TokenReturn
	TokenThen
	TokenTrue
	TokenUntil
	TokenWhile
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
	"return", "then", "true", "until", "while",
	"<id>", "<string>", "<number>",
	"==", "~=", "<=", ">=", "..", "...", "<EOF>",
}

type TokenDetail struct {
	Number float64 // number for TokenNumber
	Str    *String // string for TokenId, TokenKeyWord and TokenString

	Module *String // module name of this token belongs to
	Line   int     // token line number in module
	Column int     // token column number at 'line'
	Token  int     // token value
}

func NewTokenDetail() *TokenDetail {
	return &TokenDetail{Token: TokenEOF}
}

func GetTokenStr(t TokenDetail) string {
	var str string

	token := t.Token
	if token == TokenNumber {
		str = strconv.FormatFloat(t.Number, 'f', 6, 64)
	} else if (token == TokenId) || (token == TokenString) {
		str = t.Str.GetStdString()
	} else if (token >= TokenAnd) && (token <= TokenEOF) {
		str = tokenStr[token-TokenAnd]
	} else {
		str = str + string(token)
	}

	return str
}
