package Test

import (
	"InterpreterVM/Source/io/text"
	"InterpreterVM/Source/luna"
	"testing"
)

type lexerWrapper struct {
	iss   text.InStringStream
	state luna.State
	name  luna.String
	lexer luna.Lexer
}

func NewLexerWrapper(str string) lexerWrapper {
	var lw lexerWrapper
	lw.iss = text.NewInStringStream(str)
	lw.name = luna.NewString("lex")
	lw.lexer = luna.NewLexer(&lw.state, &lw.name, lw.iss.GetChar)
	return lw
}

func (lw lexerWrapper) GetToken() int {
	var token luna.TokenDetail
	t, err := lw.lexer.GetToken(&token)
	if err != nil {
		panic(err)
	}
	return t
}

func TestLex1(t *testing.T) {
	lexer := NewLexerWrapper("\r\n\t\v\f")
	if lexer.GetToken() == luna.TokenEOF {
		t.Error("lex1 error")
	}
}
