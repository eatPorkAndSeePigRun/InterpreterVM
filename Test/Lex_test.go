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

func TestLex5(t *testing.T) {
	lexer := NewLexerWrapper("+ - * / % ^ # == ~= <= >= < > = " +
		"( ) { } [ ] ; : , . .. ...")
	if token, _ := lexer.GetToken(); token != '+' {
		t.Error("lex5 error")
	}
	if token, _ := lexer.GetToken(); token != '-' {
		t.Error("lex5 error")
	}
	if token, _ := lexer.GetToken(); token != '*' {
		t.Error("lex5 error")
	}
	if token, _ := lexer.GetToken(); token != '/' {
		t.Error("lex5 error")
	}
	if token, _ := lexer.GetToken(); token != '%' {
		t.Error("lex5 error")
	}
	if token, _ := lexer.GetToken(); token != '^' {
		t.Error("lex5 error")
	}
	if token, _ := lexer.GetToken(); token != '#' {
		t.Error("lex5 error")
	}
	if token, _ := lexer.GetToken(); token != TokenEqual {
		t.Error("lex5 error")
	}
	if token, _ := lexer.GetToken(); token != TokenNotEqual {
		t.Error("lex5 error")
	}
	if token, _ := lexer.GetToken(); token != TokenLessEqual {
		t.Error("lex5 error")
	}
	if token, _ := lexer.GetToken(); token != TokenGreaterEqual {
		t.Error("lex5 error")
	}
	if token, _ := lexer.GetToken(); token != '<' {
		t.Error("lex5 error")
	}
	if token, _ := lexer.GetToken(); token != '>' {
		t.Error("lex5 error")
	}
	if token, _ := lexer.GetToken(); token != '=' {
		t.Error("lex5 error")
	}
	if token, _ := lexer.GetToken(); token != '(' {
		t.Error("lex5 error")
	}
	if token, _ := lexer.GetToken(); token != ')' {
		t.Error("lex5 error")
	}
	if token, _ := lexer.GetToken(); token != '{' {
		t.Error("lex5 error")
	}
	if token, _ := lexer.GetToken(); token != '}' {
		t.Error("lex5 error")
	}
	if token, _ := lexer.GetToken(); token != '[' {
		t.Error("lex5 error")
	}
	if token, _ := lexer.GetToken(); token != ']' {
		t.Error("lex5 error")
	}
	if token, _ := lexer.GetToken(); token != ';' {
		t.Error("lex5 error")
	}
	if token, _ := lexer.GetToken(); token != ':' {
		t.Error("lex5 error")
	}
	if token, _ := lexer.GetToken(); token != ',' {
		t.Error("lex5 error")
	}
	if token, _ := lexer.GetToken(); token != '.' {
		t.Error("lex5 error")
	}
	if token, _ := lexer.GetToken(); token != TokenConcat {
		t.Error("lex5 error")
	}
	if token, _ := lexer.GetToken(); token != TokenVarArg {
		t.Error("lex5 error")
	}
	if token, _ := lexer.GetToken(); token != TokenEOF {
		t.Error("lex5 error")
	}
}

func TestLex6(t *testing.T) {
	lexer := NewLexerWrapper("and do else elseif end false for function if " +
		"in local nil not or repeat return then true until while")
	if token, _ := lexer.GetToken(); token != TokenAnd {
		t.Error("lex6 error")
	}
	if token, _ := lexer.GetToken(); token != TokenDo {
		t.Error("lex6 error")
	}
	if token, _ := lexer.GetToken(); token != TokenElse {
		t.Error("lex6 error")
	}
	if token, _ := lexer.GetToken(); token != TokenElseif {
		t.Error("lex6 error")
	}
	if token, _ := lexer.GetToken(); token != TokenEnd {
		t.Error("lex6 error")
	}
	if token, _ := lexer.GetToken(); token != TokenFalse {
		t.Error("lex6 error")
	}
	if token, _ := lexer.GetToken(); token != TokenFor {
		t.Error("lex6 error")
	}
	if token, _ := lexer.GetToken(); token != TokenFunction {
		t.Error("lex6 error")
	}
	if token, _ := lexer.GetToken(); token != TokenIf {
		t.Error("lex6 error")
	}
	if token, _ := lexer.GetToken(); token != TokenIn {
		t.Error("lex6 error")
	}
	if token, _ := lexer.GetToken(); token != TokenLocal {
		t.Error("lex6 error")
	}
	if token, _ := lexer.GetToken(); token != TokenNil {
		t.Error("lex6 error")
	}
	if token, _ := lexer.GetToken(); token != TokenNot {
		t.Error("lex6 error")
	}
	if token, _ := lexer.GetToken(); token != TokenOr {
		t.Error("lex6 error")
	}
	if token, _ := lexer.GetToken(); token != TokenRepeat {
		t.Error("lex6 error")
	}
	if token, _ := lexer.GetToken(); token != TokenReturn {
		t.Error("lex6 error")
	}
	if token, _ := lexer.GetToken(); token != TokenThen {
		t.Error("lex6 error")
	}
	if token, _ := lexer.GetToken(); token != TokenTrue {
		t.Error("lex6 error")
	}
	if token, _ := lexer.GetToken(); token != TokenUntil {
		t.Error("lex6 error")
	}
	if token, _ := lexer.GetToken(); token != TokenWhile {
		t.Error("lex6 error")
	}
	if token, _ := lexer.GetToken(); token != TokenEOF {
		t.Error("lex6 error")
	}
}
