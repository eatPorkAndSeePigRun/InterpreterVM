package luna

// Base exception for luna, all exception throw by luna
// are derived from this class
type Exception struct {
	what string
}

func (exception Exception) What() string {
	return exception.what
}

func (exception Exception) setWhat() {
	// TODO
}

// Module file open failed, this exception will be throw
type OpenFileFail struct {
}

// For lexer report error of token
type LexException struct {
	exception Exception
}

func (lexException LexException) NewLexException() LexException {
	// TODO
	return LexException{}
}

// For parser report grammar error
type ParseException struct {
}

// For semantic analyser report semantic error
type SemanticException struct {
}

// For code generator report error
type CodeGenerateException struct {
}

// Report error of call c function
type CallCFuncException struct {
}

// For VM report runtime error
type RuntimeException struct {
}
