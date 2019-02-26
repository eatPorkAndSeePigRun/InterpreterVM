package Test

import (
	"InterpreterVM/Source/io/text"
	. "InterpreterVM/Source/vm"
	"testing"
)

type lexerWrapper struct {
	iss   text.InStringStream
	state State
	name  String
	lexer Lexer
}

func NewLexerWrapper(str string) *lexerWrapper {
	var lw lexerWrapper
	lw.iss = text.NewInStringStream(str)
	lw.name = *NewString("lex")
	lw.lexer = NewLexer(&lw.state, &lw.name, lw.iss.GetChar)
	return &lw
}

func (lw *lexerWrapper) GetToken() (int, error) {
	var token TokenDetail
	return lw.lexer.GetToken(&token)
}

func TestLex1(t *testing.T) {
	lexer := NewLexerWrapper("\r\n\t\v\f")
	if token, _ := lexer.GetToken(); token != TokenEOF {
		t.Error("lex1 error")
	}
}
