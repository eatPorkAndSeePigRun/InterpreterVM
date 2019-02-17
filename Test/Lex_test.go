package Test

import (
	"InterpreterVM/Source/io/text"
	"InterpreterVM/Source/vm"
	"testing"
)

type lexerWrapper struct {
	iss   text.InStringStream
	state vm.State
	name  vm.String
	lexer vm.Lexer
}

func NewLexerWrapper(str string) lexerWrapper {
	var lw lexerWrapper
	lw.iss = text.NewInStringStream(str)
	lw.name = vm.NewString("lex")
	lw.lexer = vm.NewLexer(&lw.state, &lw.name, lw.iss.GetChar)
	return lw
}

func (lw lexerWrapper) GetToken() int {
	var token vm.TokenDetail
	t, err := lw.lexer.GetToken(&token)
	if err != nil {
		panic(err)
	}
	return t
}

func TestLex1(t *testing.T) {
	lexer := NewLexerWrapper("\r\n\t\v\f")
	if lexer.GetToken() == vm.TokenEOF {
		t.Error("lex1 error")
	}
}
