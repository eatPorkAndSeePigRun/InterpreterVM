package vm

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
	what := fmt.Sprintf("%s:%d:%d ", module, line, column) + fmt.Sprintln(args)
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
	what := fmt.Sprintf("%s:%d:%d '%s' %s",
		t.module.GetCStr(), t.line, t.column, GetTokenStr(t), str)
	return ParseError{what}
}

func (p ParseError) Error() string {
	return p.what
}

// For semantic analyser report semantic error
type SemanticError struct {
	what string
}

func NewSemanticError(str string, t TokenDetail) error {
	what := fmt.Sprintf("%s:%d:%d '%s' %s",
		t.module.GetCStr(), t.line, t.column, GetTokenStr(t), str)
	return &SemanticError{what}
}

func (s SemanticError) Error() string {
	return s.what
}

// For code generator report error
type CodeGenerateError struct {
	what string
}

func NewCodeGenerateError(module string, line int, args ...interface{}) error {
	what := fmt.Sprintf("%s:%d ", module, line) + fmt.Sprintln(args)
	return &CodeGenerateError{what}
}

func (c CodeGenerateError) Error() string {
	return c.what
}

// Report error of call c function
type CallCFuncError struct {
	what string
}

func NewCallCFuncError(args ...interface{}) error {
	return &CallCFuncError{fmt.Sprintln(args)}
}

func (c CallCFuncError) Error() string {
	return c.what
}

// For VM report runtime error
type RuntimeError struct {
	what string
}

func NewRuntimeError1(module string, line int, desc string) error {
	what := fmt.Sprintf("%s:%d %s", module, line, desc)
	return &RuntimeError{what}
}

func NewRuntimeError2(module string, line int, v Value, vName, expectType string) error {
	what := fmt.Sprintf("%s:%d %s is a %s value, expect a %s value",
		module, line, vName, v.TypeName(), expectType)
	return &RuntimeError{what}
}

func NewRuntimeError3(module string, line int, v Value, vName, vScope, op string) error {
	what := fmt.Sprintf("%s:%d attempt to %s %s '%s' (a %s value)",
		module, line, op, vScope, vName, v.TypeName())
	return &RuntimeError{what}
}

func NewRuntimeError4(module string, line int, v1, v2 Value, op string) error {
	what := fmt.Sprintf("%s:%d attempt to %s %s with %s",
		module, line, op, v1.TypeName(), v2.TypeName())
	return &RuntimeError{what}
}

func (r RuntimeError) Error() string {
	return r.what
}
