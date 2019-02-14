package luna

import "fmt"

// Module file open failed, this Error will be throw
type OpenFileFail struct {
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
}

// For semantic analyser report semantic error
type SemanticError struct {
}

// For code generator report error
type CodeGenerateError struct {
}

// Report error of call c function
type CallCFuncError struct {
}

// For VM report runtime error
type RuntimeError struct {
}
