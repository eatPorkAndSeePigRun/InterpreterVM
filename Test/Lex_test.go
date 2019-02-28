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
	lw.state = *NewState()
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

func TestLex2(t *testing.T) {
	lexer := NewLexerWrapper("-- this is comment\n" +
		"--[[this is long\n comment]]" +
		"--[[this is long\n comment too--]]" +
		"--[[incomplete comment]")
	if _, err := lexer.GetToken(); err != nil {
		t.Error("lex2 should be a error")
	}
}

func TestLex3(t *testing.T) {
	lexer := NewLexerWrapper("[==[long\nlong\nstring]==]'string'\"string\"" +
		"[=[incomplete string]=")
	for i := 0; i < 3; i++ {
		if token, _ := lexer.GetToken(); token != TokenString {
			t.Error("lex3 error")
		}
	}

	if _, err := lexer.GetToken(); err != nil {
		t.Error("lex3 should be a error")
	}
}

func TestLex4(t *testing.T) {
	lexer := NewLexerWrapper("3 3.0 3.1416 314.16e-2 0.31416E1 0xff " +
		"0x0.1E 0xA23p-4 0X1.921FB54442D18P+1 0x")
	for i := 0; i < 9; i++ {
		if token, _ := lexer.GetToken(); token != TokenNumber {
			t.Error("lex4 error")
		}
	}

	if _, err := lexer.GetToken(); err == nil {
		t.Error("lex4 should be a error")
	}
}
