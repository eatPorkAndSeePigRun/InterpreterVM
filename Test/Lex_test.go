package Test

import (
	"InterpreterVM/Source/compiler"
	"InterpreterVM/Source/datatype"
	"InterpreterVM/Source/io/text"
	"InterpreterVM/Source/vm"
	"testing"
)

type lexerWrapper struct {
	iss   text.InStringStream
	state vm.State
	name  datatype.String
	lexer compiler.Lexer
}

func NewLexerWrapper(str string) lexerWrapper {
	var lw lexerWrapper
	lw.iss = text.NewInStringStream(str)
	lw.name = *datatype.NewString("lex")
	lw.lexer = compiler.NewLexer(&lw.state, &lw.name, lw.iss.GetChar)
	return lw
}

func (lw lexerWrapper) GetToken() int {
	var token compiler.TokenDetail
	t, err := lw.lexer.GetToken(&token)
	if err != nil {
		panic(err)
	}
	return t
}

func TestLex1(t *testing.T) {
	lexer := NewLexerWrapper("\r\n\t\v\f")
	if lexer.GetToken() == compiler.TokenEOF {
		t.Error("lex1 error")
	}
}
