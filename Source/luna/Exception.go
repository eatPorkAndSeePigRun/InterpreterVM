package luna

import "fmt"

// Module file open failed, this Error will be throw
type OpenFileFail struct {
	what string
}

func NewOpenFileFail(file string) error {
	return &OpenFileFail{file}
}

func (o OpenFileFail) Error() string {
	return o.what
}

// For lexer report error of token
type LexError struct {
	what string
}

func NewLexError(module string, line, column int, args ...interface{}) error {
	what := module + ":" + string(line) + ":" + string(column) + " " + fmt.Sprint(args)
	return &LexError{what}
}

func (l LexError) Error() string {
	return l.what
}

// For parser report grammar error
type ParseError struct {
	what string
}

func NewParseError(str string, t TokenDetail) ParseError {
	what := t.module.GetCStr() + ":" + string(t.line) + ":" + string(t.column) +
		" '" + GetTokenStr(t) + "' " + str
	return ParseError{what}
}

func (p ParseError) Error() string {
	return p.what
}

// For semantic analyser report semantic error
type SemanticError struct {
}

// For code generator report error
type CodeGenerateError struct {
}

// Report error of call c function
type CallCFuncError struct {
	what string
}

func NewCallCFuncError(args ...interface{}) error {
	return &CallCFuncError{fmt.Sprint(args)}
}

func (c CallCFuncError) Error() string {
	return c.what
}

// For VM report runtime error
type RuntimeError struct {
}
