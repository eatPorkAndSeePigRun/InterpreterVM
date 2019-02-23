package vm

const (
	prefixExpTypeNormal = iota
	prefixExpTypeVar
	prefixExpTypeFunctionCall
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

func (p *parserImpl) parse() SyntaxTree {
	res, err := p.parseChunk()
	if err != nil {
		panic(err)
	}
	return res
}

func (p *parserImpl) parseChunk() (SyntaxTree, error) {
	block, err := p.parseBlock()
	if err != nil {
		panic(err)
	}
	if p.nextToken().Token != TokenEOF {
		return nil, NewParseError("expect <eof>", p.current)
	}
	return NewChunk(block, p.lexer.GetLexModule()), nil
}

func (p *parserImpl) parseExp(left SyntaxTree, op TokenDetail, leftPriority int) (SyntaxTree, error) {
	var exp SyntaxTree
	var err error
	p.lookAhead()

	if p.lookAhead_.Token == '-' || p.lookAhead_.Token == '#' || p.lookAhead_.Token == TokenNot {
		p.nextToken()
		unexp := &UnaryExpression{}
		unexp.OpToken = p.current
		unexp.Exp, err = p.parseExp(nil, *NewTokenDetail(), 90)
		if err != nil {
			panic(err)
		}
		exp = unexp
	} else if p.isMainExp(p.lookAhead_) {
		exp, err = p.parseMainExp()
		if err != nil {
			panic(err)
		}
	} else {
		return nil, NewParseError("unexpect token for exp.", p.lookAhead_)
	}

	for true {
		rightPriority := p.getOpPriority(*p.lookAhead())
		if (leftPriority < rightPriority) ||
			(leftPriority == rightPriority && p.isRightAssociation(*p.lookAhead())) {
			exp, err = p.parseExp(exp, *p.nextToken(), rightPriority)
			if err != nil {
				panic(err)
			}
		} else if leftPriority == rightPriority {
			if leftPriority == 0 {
				return exp, nil
			}
			if left == nil {
				panic("assert")
			}
			exp := NewBinaryExpression(left, exp, op)
			return p.parseExp(exp, *p.nextToken(), rightPriority)
		} else {
			if left != nil {
				exp = NewBinaryExpression(left, exp, op)
			}
			return exp, nil
		}
	}

	return nil, nil
}

func (p *parserImpl) parseMainExp() (SyntaxTree, error) {
	var exp SyntaxTree
	var err error

	switch p.lookAhead().Token {
	case TokenNil, TokenFalse, TokenTrue, TokenNumber, TokenString, TokenVarArg:
		exp = NewTerminator(*p.nextToken())
	case TokenFunction:
		exp, err = p.parseFunctionDef()
		if err != nil {
			panic(err)
		}
	case TokenId, '(':
		exp, err = p.parsePrefixExp(nil)
		if err != nil {
			panic(err)
		}
	case '{':
		exp, err = p.parseTableConstructor()
		if err != nil {
			panic(err)
		}
	default:
		return nil, NewParseError("unexpect token for exp.", p.lookAhead_)
	}

	return exp, nil
}

func (p *parserImpl) parseFunctionDef() (SyntaxTree, error) {
	p.nextToken()
	if p.current.Token != TokenFunction {
		panic("assert")
	}
	return p.parseFunctionBody()
}

func (p *parserImpl) parseFunctionBody() (SyntaxTree, error) {
	line := p.lookAhead().Line
	if p.nextToken().Token != '(' {
		return nil, NewParseError("unexpect token after 'function', expect '('", p.current)
	}

	var paramList SyntaxTree

	if p.lookAhead().Token != ')' {
		var err error
		paramList, err = p.parseParamList()
		if err != nil {
			panic(err)
		}
	}

	if p.nextToken().Token != ')' {
		return nil, NewParseError("unexpect token after param list, expect ')'", p.current)
	}

	block, err := p.parseBlock()
	if err != nil {
		panic(err)
	}

	if p.nextToken().Token != TokenEnd {
		return nil, NewParseError("unexpect token after function body, expect 'end'", p.current)
	}

	return NewFunctionBody(paramList, block, line), nil
}

func (p *parserImpl) parseParamList() (SyntaxTree, error) {
	vararg := false
	var nameList SyntaxTree

	if p.lookAhead().Token == TokenId {
		names := NewNameList()
		names.Names = append(names.Names, *p.nextToken())

		for p.lookAhead().Token == ',' {
			p.nextToken() // skip ','
			if p.lookAhead().Token == TokenId {
				names.Names = append(names.Names, *p.nextToken())
			} else if p.lookAhead().Token == TokenVarArg {
				p.nextToken() // skip Token_VarArg
				vararg = true
			} else {
				return nil, NewParseError("unexpect token in param list", p.lookAhead_)
			}

			nameList = names
		}
	} else if p.lookAhead().Token == TokenVarArg {
		p.nextToken() // skip Token_VarArg
		vararg = true
	} else {
		return nil, NewParseError("unexpect token in param list", p.lookAhead_)
	}

	return NewParamList(nameList, vararg), nil
}

func (p *parserImpl) parseBlock() (SyntaxTree, error) {
	block := NewBlock()

	hasReturn := false
	for p.lookAhead().Token != TokenEOF &&
		p.lookAhead().Token != TokenEnd &&
		p.lookAhead().Token != TokenUntil &&
		p.lookAhead().Token != TokenElseif &&
		p.lookAhead().Token != TokenElse {
		if hasReturn {
			return nil, NewParseError("unexpect statement after return statement", p.lookAhead_)
		}

		if p.lookAhead().Token == TokenReturn {
			block.ReturnStmt = p.parseReturnStatement()
			hasReturn = true
		} else {
			statement, err := p.parseStatement()
			if err != nil {
				panic(err)
			}
			if statement != nil {
				block.Statements = append(block.Statements, statement)
			}
		}
	}

	return block, nil
}

func (p *parserImpl) parseReturnStatement() SyntaxTree {
	p.nextToken() // skip 'return'
	if p.current.Token != TokenReturn {
		panic("assert")
	}

	returnStmt := NewReturnStatement(p.current.Line)

	if p.lookAhead().Token == TokenEOF ||
		p.lookAhead().Token == TokenEnd ||
		p.lookAhead().Token == TokenUntil ||
		p.lookAhead().Token == TokenElseif ||
		p.lookAhead().Token == TokenElse {
		return returnStmt
	}

	if p.lookAhead().Token != ';' {
		returnStmt.ExpList = p.parseExpList()
	} else {
		p.nextToken()
	}

	return returnStmt
}

func (p *parserImpl) parseStatement() (SyntaxTree, error) {
	switch p.lookAhead().Token {
	case ';':
		p.nextToken()
	case TokenBreak:
		return p.parseBreakStatement(), nil
	case TokenDo:
		return p.parseDoStatement()
	case TokenWhile:
		return p.parseWhileStatement()
	case TokenRepeat:
		return p.parseRepeatStatement()
	case TokenIf:
		return p.parseIfStatement()
	case TokenFunction:
		return p.parseFunctionStatement(), nil
	case TokenFor:
		return p.parseForStatement()
	case TokenLocal:
		return p.parseLocalStatement()
	default:
		return p.parseOtherStatement()
	}

	return nil, nil
}

func (p *parserImpl) parseBreakStatement() SyntaxTree {
	return NewBreakStatement(*p.nextToken())
}

func (p *parserImpl) parseDoStatement() (SyntaxTree, error) {
	p.nextToken() // skip 'while'
	if p.current.Token != TokenDo {
		panic("assert")
	}

	block, err := p.parseBlock()
	if err != nil {
		panic(err)
	}
	if p.nextToken().Token != TokenEnd {
		return nil, NewParseError("expect 'end' for do-statement", p.current)
	}

	return NewDoStatement(block), nil
}

func (p *parserImpl) parseWhileStatement() (SyntaxTree, error) {
	p.nextToken() // skip 'while'
	if p.current.Token != TokenWhile {
		panic("assert")
	}

	firstLine := p.current.Line

	exp, err := p.parseExp(nil, *NewTokenDetail(), 0)
	if err != nil {
		panic(err)
	}

	if p.nextToken().Token != TokenDo {
		return nil, NewParseError("expect 'do' for while-statement", p.current)
	}

	block, err := p.parseBlock()
	if err != nil {
		panic(err)
	}

	if p.nextToken().Token != TokenEnd {
		return nil, NewParseError("expect 'end' for while-statement", p.current)
	}

	lastLine := p.current.Line

	return NewWhileStatement(exp, block, firstLine, lastLine), nil
}

func (p *parserImpl) parseRepeatStatement() (SyntaxTree, error) {
	p.nextToken() // skip 'repeat'
	if p.current.Token != TokenRepeat {
		panic("assert")
	}

	block, err := p.parseBlock()
	if err != nil {
		panic(err)
	}

	if p.nextToken().Token != TokenUntil {
		return nil, NewParseError("expect 'until' for repeat-statement", p.current)
	}

	line := p.current.Line
	exp, err := p.parseExp(nil, *NewTokenDetail(), 0)
	if err != nil {
		panic(err)
	}

	return NewRepeatStatement(block, exp, line), nil
}

func (p *parserImpl) parseIfStatement() (SyntaxTree, error) {
	p.nextToken() // skip 'if'
	if p.current.Token != TokenIf {
		panic("assert")
	}
	line := p.current.Line

	exp, err := p.parseExp(nil, *NewTokenDetail(), 0)
	if err != nil {
		panic(err)
	}

	if p.nextToken().Token != TokenThen {
		return nil, NewParseError("expect 'then' for if", p.current)
	}

	trueBranch, err := p.parseBlock()
	if err != nil {
		panic(err)
	}
	blockEndLine := p.lookAhead().Line
	falseBranch, err := p.parseFalseBranchStatement()
	if err != nil {
		panic(err)
	}

	return NewIfStatement(exp, trueBranch, falseBranch, line, blockEndLine), nil
}

func (p *parserImpl) parseElseIfStatement() (SyntaxTree, error) {
	p.nextToken() // skip 'elseif'
	if p.current.Token == TokenElseif {
		panic("assert")
	}
	line := p.current.Line

	exp, err := p.parseExp(nil, *NewTokenDetail(), 0)
	if err != nil {
		panic(err)
	}

	if p.nextToken().Token != TokenThen {
		return nil, NewParseError("expect 'then' for elseif", p.current)
	}

	trueBranch, err := p.parseBlock()
	if err != nil {
		panic(err)
	}
	blockEndLine := p.lookAhead().Line
	falseBranch, err := p.parseFalseBranchStatement()
	if err != nil {
		panic(err)
	}

	return NewElseIfStatement(exp, trueBranch, falseBranch, line, blockEndLine), nil
}

func (p *parserImpl) parseFalseBranchStatement() (SyntaxTree, error) {
	if p.lookAhead().Token == TokenElseif {
		return p.parseElseIfStatement()
	} else if p.lookAhead().Token == TokenElse {
		return p.parseElseStatement()
	} else if p.lookAhead().Token == TokenEnd {
		p.nextToken() // skip 'end'
	} else {
		return nil, NewParseError("expect 'end' for if", p.lookAhead_)
	}

	return nil, nil
}

func (p *parserImpl) parseElseStatement() (SyntaxTree, error) {
	p.nextToken() // skip 'else'
	if p.current.Token != TokenElse {
		panic("assert")
	}

	block, err := p.parseBlock()
	if err != nil {
		panic(err)
	}
	if p.nextToken().Token != TokenEnd {
		return nil, NewParseError("expect 'end' for else", p.current)
	}

	return NewElseStatement(block), nil
}

func (p *parserImpl) parseFunctionStatement() SyntaxTree {
	p.nextToken() // skip 'function'
	if p.current.Token != TokenFunction {
		panic("assert")
	}

	funcName, err := p.parseFunctionName()
	if err != nil {
		panic(err)
	}
	funcBody, err := p.parseFunctionBody()
	if err != nil {
		panic(err)
	}
	return NewFunctionStatement(funcName, funcBody)
}

func (p *parserImpl) parseFunctionName() (SyntaxTree, error) {
	if p.nextToken().Token != TokenId {
		return nil, NewParseError("unexpect token after 'function'", p.current)
	}

	funcName := NewFunctionName()
	funcName.Names = append(funcName.Names, p.current)

	for p.lookAhead().Token == '.' {
		p.nextToken() // skip '.'
		if p.nextToken().Token != TokenId {
			return nil, NewParseError("unexpect token in function name after '.'", p.current)
		}
		funcName.Names = append(funcName.Names, p.current)
	}

	if p.lookAhead().Token == ':' {
		p.nextToken() // skip ':'
		if p.nextToken().Token != TokenId {
			return nil, NewParseError("unexpect token in function name after ':'", p.current)
		}
		funcName.MemberName = p.current
	}

	return funcName, nil
}

func (p *parserImpl) parseForStatement() (SyntaxTree, error) {
	p.nextToken() // skip 'for'
	if p.current.Token != TokenFor {
		panic("assert")
	}

	if p.lookAhead().Token != TokenId {
		return nil, NewParseError("expect 'id' after 'for'", p.lookAhead_)
	}

	if p.lookAhead2().Token == '=' {
		return p.parseNumericForStatement()
	} else {
		return p.parseGenericForStatement()
	}
}

func (p *parserImpl) parseNumericForStatement() (SyntaxTree, error) {
	name := p.nextToken()
	if p.current.Token != TokenId {
		panic("assert")
	}

	p.nextToken() // skip '='
	if p.current.Token != '=' {
		panic("assert")
	}

	exp1, err := p.parseExp(nil, *NewTokenDetail(), 0)
	if err != nil {
		panic(err)
	}
	if p.nextToken().Token != ',' {
		return nil, NewParseError("expect ',' in numeric-for", p.current)
	}

	exp2, err := p.parseExp(nil, *NewTokenDetail(), 0)
	if err != nil {
		panic(err)
	}
	var exp3 SyntaxTree

	if p.lookAhead().Token == ',' {
		p.nextToken() // skip ','
		exp3, err = p.parseExp(nil, *NewTokenDetail(), 0)
		if err != nil {
			panic(err)
		}
	}

	if p.nextToken().Token != TokenDo {
		return nil, NewParseError("expect 'do' to start numeric-for body", p.current)
	}
	block, err := p.parseBlock()
	if err != nil {
		panic(err)
	}

	if p.nextToken().Token != TokenEnd {
		return nil, NewParseError("expect 'end' to complete numeric-for", p.current)
	}
	return NewNumericForStatement(*name, exp1, exp2, exp3, block), nil
}

func (p *parserImpl) parseGenericForStatement() (SyntaxTree, error) {
	line := p.lookAhead().Line
	nameList, err := p.parseNameList()
	if err != nil {
		panic(err)
	}

	if p.nextToken().Token != TokenIn {
		return nil, NewParseError("expect 'in' in generic-for", p.current)
	}

	expList := p.parseExpList()

	if p.nextToken().Token != TokenDo {
		return nil, NewParseError("expect 'do' to start generic-for body", p.current)
	}

	block, err := p.parseBlock()
	if err != nil {
		panic(err)
	}

	if p.nextToken().Token != TokenEnd {
		return nil, NewParseError("expect 'end' to complete generic-for", p.current)
	}

	return NewGenericForStatement(nameList, expList, block, line), nil
}

func (p *parserImpl) parseLocalStatement() (SyntaxTree, error) {
	p.nextToken() // skip 'local'
	if p.current.Token != TokenLocal {
		panic("assert")
	}

	if p.lookAhead().Token == TokenFunction {
		return p.parseLocalFunction()
	} else if p.lookAhead().Token == TokenId {
		return p.parseNameList()
	} else {
		return nil, NewParseError("unexpect token after 'local'", p.lookAhead_)
	}
}

func (p *parserImpl) parseLocalFunction() (SyntaxTree, error) {
	p.nextToken() // skip 'function'
	if p.current.Token != TokenFunction {
		panic("assert")
	}

	if p.nextToken().Token != TokenId {
		return nil, NewParseError("expect 'id' after 'local function'", p.current)
	}

	name := p.current
	body, err := p.parseFunctionBody()
	if err != nil {
		panic(err)
	}

	return NewLocalFunctionStatement(name, body), nil
}

func (p *parserImpl) parseLocalNameList() SyntaxTree {
	startLine := p.lookAhead().Line
	nameList, err := p.parseNameList()
	if err != nil {
		panic(err)
	}
	var expList SyntaxTree

	if p.lookAhead().Token == '=' {
		p.nextToken() // skip '='
		expList = p.parseExpList()
	}

	return NewLocalNameListStatement(nameList, expList, startLine)
}

func (p *parserImpl) parseNameList() (SyntaxTree, error) {
	if p.nextToken().Token != TokenId {
		return nil, NewParseError("expect 'id'", p.current)
	}

	nameList := NewNameList()

	nameList.Names = append(nameList.Names, p.current)
	for p.lookAhead().Token == ',' {
		p.nextToken() // skip ','
		if p.nextToken().Token != TokenId {
			return nil, NewParseError("expect 'id' after ','", p.current)
		}
		nameList.Names = append(nameList.Names, p.current)
	}

	return nameList, nil
}

func (p *parserImpl) parseOtherStatement() (SyntaxTree, error) {
	var prefixExpType int
	startLine := p.lookAhead().Line
	exp, err := p.parsePrefixExp(&prefixExpType)
	if err != nil {
		panic(err)
	}

	if prefixExpType == prefixExpTypeVar {
		varList := NewVarList()
		varList.VarList = append(varList.VarList, exp)

		for p.lookAhead().Token != '=' {
			if p.lookAhead().Token != ',' {
				return nil, NewParseError("expect ',' to split var", p.lookAhead_)
			}
			p.nextToken() // skip ','
			exp, err := p.parsePrefixExp(&prefixExpType)
			if err != nil {
				panic(err)
			}
			if prefixExpType != prefixExpTypeVar {
				return nil, NewParseError("expect var here", p.current)
			}
			varList.VarList = append(varList.VarList, exp)
		}

		p.nextToken() // skip '='
		expList := p.parseExpList()
		return NewAssignmentStatement(varList, expList, startLine), nil
	} else if prefixExpType == prefixExpTypeFunctionCall {
		return exp, nil
	} else {
		return nil, NewParseError("incomplete statement", p.current)
	}
}

func (p *parserImpl) parsePrefixExp(prefixExpType *int) (SyntaxTree, error) {
	p.nextToken()
	if p.current.Token != TokenId && p.current.Token != '(' {
		return nil, NewParseError("unexpect token here", p.current)
	}

	var exp SyntaxTree
	var err error

	if p.current.Token == '(' {
		exp, err = p.parseExp(nil, *NewTokenDetail(), 0)
		if err != nil {
			panic(err)
		}
		if p.nextToken().Token != ')' {
			return nil, NewParseError("expect ')'", p.current)
		}
		if prefixExpType != nil {
			*prefixExpType = prefixExpTypeNormal
		}
	} else {
		exp = NewTerminator(p.current)
		if prefixExpType != nil {
			*prefixExpType = prefixExpTypeVar
		}
	}

	return p.parsePrefixExpTail(exp, prefixExpType), nil
}

func (p *parserImpl) parsePrefixExpTail(exp SyntaxTree, prefixExpType *int) SyntaxTree {
	if p.lookAhead().Token == '[' || p.lookAhead().Token == '.' {
		if prefixExpType != nil {
			*prefixExpType = prefixExpTypeVar
		}
		exp, err := p.parseVar(exp)
		if err != nil {
			panic(err)
		}
		return p.parsePrefixExpTail(exp, prefixExpType)
	} else if p.lookAhead().Token == ':' || p.lookAhead().Token == '(' ||
		p.lookAhead().Token == '{' || p.lookAhead().Token == TokenString {
		if prefixExpType != nil {
			*prefixExpType = prefixExpTypeFunctionCall
		}
		exp, err := p.parseFunctionCall(exp)
		if err != nil {
			panic(err)
		}
		return p.parsePrefixExpTail(exp, prefixExpType)
	} else {
		return exp
	}
}

func (p *parserImpl) parseVar(table SyntaxTree) (SyntaxTree, error) {
	p.nextToken()
	if p.current.Token != '[' && p.current.Token != '.' {
		panic("assert")
	}

	if p.current.Token == '[' {
		line := p.lookAhead().Line
		exp, err := p.parseExp(nil, *NewTokenDetail(), 0)
		if err != nil {
			panic(err)
		}
		if p.nextToken().Token != ']' {
			return nil, NewParseError("expect ']'", p.current)
		}
		return NewIndexAccessor(table, exp, line), nil
	} else {
		if p.nextToken().Token != TokenId {
			return nil, NewParseError("expect 'id' after '.'", p.current)
		}
		return NewMemberAccessor(table, p.current), nil
	}
}

func (p *parserImpl) parseFunctionCall(caller SyntaxTree) (SyntaxTree, error) {
	if p.lookAhead().Token == ':' {
		p.nextToken()
		if p.nextToken().Token != TokenId {
			return nil, NewParseError("expect 'id' after ':'", p.current)
		}

		member := p.current
		line := p.lookAhead().Line
		args, err := p.parseArgs()
		if err != nil {
			panic(err)
		}
		return NewMemberFuncCall(caller, member, args, line), nil
	} else {
		line := p.lookAhead().Line
		args, err := p.parseArgs()
		if err != nil {
			panic(err)
		}
		return NewNormalFuncCall(caller, args, line), nil
	}
}

func (p *parserImpl) parseArgs() (SyntaxTree, error) {
	if p.lookAhead().Token != TokenString &&
		p.lookAhead().Token != '{' &&
		p.lookAhead().Token != '(' {
		panic("assert")
	}

	var arg SyntaxTree
	var argType int
	var err error

	if p.lookAhead().Token == TokenString {
		argType = ArgTypeString
		arg = NewTerminator(*p.nextToken())
	} else if p.lookAhead().Token == '{' {
		argType = ArgTypeTable
		arg, err = p.parseTableConstructor()
		if err != nil {
			panic(err)
		}
	} else {
		argType = ArgTypeExpList
		p.nextToken() // skip '('
		if p.lookAhead().Token != ')' {
			arg = p.parseExpList()
		}

		if p.nextToken().Token != ')' {
			return nil, NewParseError("expect ')' to end function call args", p.current)
		}
	}
	return NewFuncCallArgs(arg, argType), nil
}

func (p *parserImpl) parseExpList() SyntaxTree {
	expList := NewExpressionList(p.lookAhead().Line)

	anymore := true
	for anymore {
		st, err := p.parseExp(nil, *NewTokenDetail(), 0)
		if err != nil {
			panic(err)
		}
		expList.ExpList = append(expList.ExpList, st)

		if p.lookAhead().Token == ',' {
			p.nextToken()
		} else {
			anymore = false
		}
	}

	return expList
}

func (p *parserImpl) parseTableConstructor() (SyntaxTree, error) {
	p.nextToken()
	if p.current.Token != '{' {
		panic("assert")
	}

	table := NewTableDefine(p.current.Line)

	for p.lookAhead().Token != '}' {
		if p.lookAhead().Token == '[' {
			st, err := p.parseTableIndexField()
			if err != nil {
				panic(err)
			}
			table.Fields = append(table.Fields, st)
		} else if p.lookAhead().Token == TokenId && p.lookAhead2().Token == '=' {
			table.Fields = append(table.Fields, p.parseTableNameField())
		} else {
			table.Fields = append(table.Fields, p.parseTableArrayField())
		}

		if p.lookAhead().Token != '}' {
			p.nextToken()
			if p.current.Token != ',' && p.current.Token != ';' {
				return nil, NewParseError("expect ',' or ';' to split table fields", p.current)
			}
		}
	}

	if p.nextToken().Token != '}' {
		return nil, NewParseError("expect '}' for table", p.current)
	}
	return table, nil
}

func (p *parserImpl) parseTableIndexField() (SyntaxTree, error) {
	p.nextToken()
	if p.current.Token != '[' {
		panic("assert")
	}

	line := p.lookAhead_.Line
	index, err := p.parseExp(nil, *NewTokenDetail(), 0)
	if err != nil {
		panic(err)
	}

	if p.nextToken().Token != ']' {
		return nil, NewParseError("expect ']'", p.current)
	}

	if p.nextToken().Token != '=' {
		return nil, NewParseError("expect '='", p.current)
	}

	value, err := p.parseExp(nil, *NewTokenDetail(), 0)
	if err != nil {
		panic(err)
	}

	return NewTableIndexField(index, value, line), nil
}

func (p *parserImpl) parseTableNameField() SyntaxTree {
	name := p.nextToken()

	p.nextToken()
	if p.current.Token != '=' {
		panic("assert")
	}

	value, err := p.parseExp(nil, *NewTokenDetail(), 0)
	if err != nil {
		panic(err)
	}

	return NewTableNameField(*name, value)
}

func (p *parserImpl) parseTableArrayField() SyntaxTree {
	line := p.lookAhead().Line
	value, err := p.parseExp(nil, *NewTokenDetail(), 0)
	if err != nil {
		panic(err)
	}
	return NewTableArrayField(value, line)
}

func (p *parserImpl) nextToken() *TokenDetail {
	if p.lookAhead_.Token != TokenEOF {
		p.current = p.lookAhead_
		p.lookAhead_ = p.lookAhead2_
		if p.lookAhead2_.Token != TokenEOF {
			p.lookAhead2_.Token = TokenEOF
		}
	} else {
		if _, err := p.lexer.GetToken(&p.current); err != nil {
			panic(err)
		}
	}

	return &p.current
}

func (p *parserImpl) lookAhead() *TokenDetail {
	if p.lookAhead_.Token == TokenEOF {
		if _, err := p.lexer.GetToken(&p.lookAhead_); err != nil {
			panic(err)
		}
	}
	return &p.lookAhead_
}

func (p *parserImpl) lookAhead2() *TokenDetail {
	p.lookAhead()
	if p.lookAhead2_.Token == TokenEOF {
		if _, err := p.lexer.GetToken(&p.lookAhead2_); err != nil {
			panic(err)
		}
	}
	return &p.lookAhead_
}

func (p *parserImpl) isMainExp(t TokenDetail) bool {
	token := t.Token
	return token == TokenNil ||
		token == TokenFalse ||
		token == TokenTrue ||
		token == TokenNumber ||
		token == TokenString ||
		token == TokenVarArg ||
		token == TokenFunction ||
		token == TokenId ||
		token == '(' ||
		token == '{'
}

func (p *parserImpl) isRightAssociation(t TokenDetail) bool {
	return t.Token == '^'
}

func (p *parserImpl) getOpPriority(t TokenDetail) int {
	switch t.Token {
	case '^':
		return 100
	case '*', '/', '%':
		return 80
	case '+', '-':
		return 70
	case TokenConcat:
		return 60
	case '>', '<', TokenGreaterEqual, TokenLessEqual, TokenNotEqual, TokenEqual:
		return 50
	case TokenAnd:
		return 40
	case TokenOr:
		return 30
	default:
		return 0
	}
}

func Parse(lexer *Lexer) SyntaxTree {
	impl := newParserImpl(lexer)
	return impl.parse()
}
