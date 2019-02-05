package luna

type Exception struct {
	what string
}

func (exception Exception) What() string {
	return exception.what
}

func (exception Exception) setWhat() {
	// TODO
}

type LexException struct {
	exception Exception
}

func (lexException LexException) NewLexException() LexException {
	// TODO
	return LexException{}
}