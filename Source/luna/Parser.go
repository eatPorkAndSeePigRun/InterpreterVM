package luna

const (
	prefixExpTypeNormal = iota
	prefixExpTypeVar
	prefixExpTypeFunctioncall
)

type parserImpl struct {
	lexer       *Lexer
	current     TokenDetail
	lookAhead_  TokenDetail
	lookAhead2_ TokenDetail
}

func newParserImpl(lexer *Lexer) parserImpl {
	return parserImpl{lexer: lexer}
}

func (p parserImpl) parse() (SyntaxTree, error) {
	return p.parseChunk()
}

func (p parserImpl) parseChunk() (SyntaxTree, error) {
	block := p.parseBlock()
	if p.nextToken().token != TokenEOF {
		return nil, NewParseError("expect <eof>", p.current)
	}
	return NewChunk(block, p.lexer.GetLexModule()), nil
}

func (p parserImpl) parseExp(left SyntaxTree, op TokenDetail, leftPriority int64) SyntaxTree {

}

func (p parserImpl) parseMainExp() SyntaxTree {

}

func (p parserImpl) parseFunctionDef() SyntaxTree {

}

func (p parserImpl) parseFunctionBody() SyntaxTree {

}

func (p parserImpl) parseParamList() SyntaxTree {

}

func (p parserImpl) parseBlock() SyntaxTree {
	block :=
}

func (p parserImpl) parseReturnStatement() SyntaxTree {

}

func (p parserImpl) parseStatement() SyntaxTree {

}

func (p parserImpl) parseBreakStatement() SyntaxTree {

}

func (p parserImpl) parseDoStatement() SyntaxTree {

}

func (p parserImpl) parseWhileStatement() SyntaxTree {

}

func (p parserImpl) parseRepeatStatement() SyntaxTree {

}

func (p parserImpl) parseIfStatement() SyntaxTree {

}

func (p parserImpl) parseElseIfStatement() SyntaxTree {

}

func (p parserImpl) parseFalseBranchStatement() SyntaxTree {

}

func (p parserImpl) parseElseStatement() SyntaxTree {

}

func (p parserImpl) parseFunctionStatement() SyntaxTree {

}

func (p parserImpl) parseFunctionName() SyntaxTree {

}

func (p parserImpl) parseForStatement() SyntaxTree {

}

func (p parserImpl) parseNumericForStatement() SyntaxTree {

}

func (p parserImpl) parseGenericForStatement() SyntaxTree {

}

func (p parserImpl) parseLocalStatement() SyntaxTree {

}

func (p parserImpl) parseLocalFunction() SyntaxTree {

}

func (p parserImpl) parseLocalNameList() SyntaxTree {

}

func (p parserImpl) parseNameList() SyntaxTree {

}

func (p parserImpl) parseOtherStatement() SyntaxTree {

}

func (p parserImpl) parsePrefixExp(prefixExpType *int) SyntaxTree {

}

func (p parserImpl) parsePrefixExpTail(exp SyntaxTree, prefixExpType *int) SyntaxTree {

}

func (p parserImpl) parseVar(table SyntaxTree) SyntaxTree {

}

func (p parserImpl) parseFunctionCall(caller SyntaxTree) SyntaxTree {

}

func (p parserImpl) parseArgs() SyntaxTree {

}

func (p parserImpl) parseExpList() SyntaxTree {

}

func (p parserImpl) parseTableConstructor() SyntaxTree {

}

func (p parserImpl) parseTableIndexField() SyntaxTree {

}

func (p parserImpl) parseTableNameField() SyntaxTree {

}

func (p parserImpl) parseTableArrayField() SyntaxTree {

}

func (p parserImpl) nextToken() *TokenDetail {

}

func (p parserImpl) lookAhead() *TokenDetail {

}

func (p parserImpl) lookAhead2() *TokenDetail {

}

func (p parserImpl) isMainExp(t TokenDetail) bool {

}

func (p parserImpl) isRightAssociation(t TokenDetail) bool {

}

func (p parserImpl) getOpPriority(t TokenDetail) int {

}

func Parse(lexer *Lexer) SyntaxTree {
	impl := newParserImpl(lexer)
	return impl.parse()
}
