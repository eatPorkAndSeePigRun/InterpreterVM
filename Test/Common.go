package Test

import (
	"InterpreterVM/Source/io/text"
	. "InterpreterVM/Source/vm"
	"unsafe"
)

type ParserWrapper struct {
	iss   text.InStringStream
	state State
	name  String
	lexer Lexer
}

func newParserWrapper(str string) *ParserWrapper {
	var pw ParserWrapper
	pw.iss = text.NewInStringStream(str)
	pw.state = *NewState()
	pw.name = *NewString("parser")
	pw.lexer = NewLexer(&pw.state, &pw.name, pw.iss.GetChar)
	return &pw
}

func (pw *ParserWrapper) SetInput(input string) {
	pw.iss.SetInputString(input)
}

func (pw *ParserWrapper) IsEOF() bool {
	var detail TokenDetail
	token, err := pw.lexer.GetToken(&detail)
	if err != nil {
		panic(err)
	}
	return token == TokenEOF
}

func (pw *ParserWrapper) Parse() SyntaxTree {
	return Parse(&pw.lexer)
}

func (pw *ParserWrapper) GetState() *State {
	return &pw.state
}

type ASTFinder struct {
	theASTNode SyntaxTree
	finder     interface{}
}

func NewASTFinder(finder interface{}) *ASTFinder {
	return &ASTFinder{theASTNode: nil, finder: finder}
}

func (af *ASTFinder) GetResult() SyntaxTree {
	return af.theASTNode
}

func (af *ASTFinder) VisitChunk(ast *Chunk, data unsafe.Pointer) {
	if af.theASTNode == nil {
		f := func() { ast.Block.Accept(af, nil) }
		af.setResult(false, ast, f)
	}
}

func (af *ASTFinder) VisitBlock(ast *Block, data unsafe.Pointer) {
	if af.theASTNode == nil {
		f := func() {
			for i := range ast.Statements {
				ast.Statements[i].Accept(af, nil)
			}
			if ast.ReturnStmt != nil {
				ast.ReturnStmt.Accept(af, nil)
			}
		}
		af.setResult(false, ast, f)
	}
}

func (af *ASTFinder) VisitReturnStatement(ast *ReturnStatement, data unsafe.Pointer) {
	if af.theASTNode == nil {
		f := func() {
			if ast.ExpList != nil {
				ast.ExpList.Accept(af, nil)
			}
		}
		af.setResult(false, ast, f)
	}
}

func (af *ASTFinder) VisitBreakStatement(ast *BreakStatement, data unsafe.Pointer) {
	// TODO
	panic("pass")
}

func (af *ASTFinder) VisitDoStatement(ast *DoStatement, data unsafe.Pointer) {
	// TODO
	panic("pass")
}

func (af *ASTFinder) VisitWhileStatement(ast *WhileStatement, data unsafe.Pointer) {
	// TODO
	panic("pass")
}

func (af *ASTFinder) VisitRepeatStatement(ast *RepeatStatement, data unsafe.Pointer) {
	// TODO
	panic("pass")
}

func (af *ASTFinder) VisitIfStatement(ast *IfStatement, data unsafe.Pointer) {
	// TODO
	panic("pass")
}

func (af *ASTFinder) VisitElseIfStatement(ast *ElseIfStatement, data unsafe.Pointer) {
	// TODO
	panic("pass")
}

func (af *ASTFinder) VisitElseStatement(ast *ElseStatement, data unsafe.Pointer) {
	// TODO
	panic("pass")
}

func (af *ASTFinder) VisitNumericForStatement(ast *NumericForStatement, data unsafe.Pointer) {
	// TODO
	panic("pass")
}

func (af *ASTFinder) VisitGenericForStatement(ast *GenericForStatement, data unsafe.Pointer) {
	// TODO
	panic("pass")
}

func (af *ASTFinder) VisitFunctionStatement(ast *FunctionStatement, data unsafe.Pointer) {
	// TODO
	panic("pass")
}

func (af *ASTFinder) VisitFunctionName(ast *FunctionName, data unsafe.Pointer) {
	// TODO
	panic("pass")
}

func (af *ASTFinder) VisitLocalFunctionStatement(ast *LocalFunctionStatement, data unsafe.Pointer) {
	// TODO
	panic("pass")
}

func (af *ASTFinder) VisitLocalNameListStatement(ast *LocalNameListStatement, data unsafe.Pointer) {
	// TODO
	panic("pass")
}

func (af *ASTFinder) VisitAssignmentStatement(ast *AssignmentStatement, data unsafe.Pointer) {
	if af.theASTNode == nil {
		f := func() {
			ast.VarList.Accept(af, nil)
			ast.ExpList.Accept(af, nil)
		}
		af.setResult(false, ast, f)
	}
}

func (af *ASTFinder) VisitVarList(ast *VarList, data unsafe.Pointer) {
	if af.theASTNode == nil {
		f := func() {
			for i := range ast.VarList {
				ast.VarList[i].Accept(af, nil)
			}
		}
		af.setResult(false, ast, f)
	}
}

func (af *ASTFinder) VisitTerminator(ast *Terminator, data unsafe.Pointer) {
	if af.theASTNode == nil {
		af.setResult(false, ast, func() {})
	}
}

func (af *ASTFinder) VisitBinaryExpression(ast *BinaryExpression, data unsafe.Pointer) {
	// TODO
	panic("pass")
}

func (af *ASTFinder) VisitUnaryExpression(ast *UnaryExpression, data unsafe.Pointer) {
	// TODO
	panic("pass")
}

func (af *ASTFinder) VisitFunctionBody(ast *FunctionBody, data unsafe.Pointer) {
	// TODO
	panic("pass")
}

func (af *ASTFinder) VisitParamList(ast *ParamList, data unsafe.Pointer) {
	// TODO
	panic("pass")
}

func (af *ASTFinder) VisitNameList(ast *NameList, data unsafe.Pointer) {
	// TODO
	panic("pass")
}

func (af *ASTFinder) VisitTableDefine(ast *TableDefine, data unsafe.Pointer) {
	// TODO
	panic("pass")
}

func (af *ASTFinder) VisitTableIndexField(ast *TableIndexField, data unsafe.Pointer) {
	// TODO
	panic("pass")
}

func (af *ASTFinder) VisitTableNameField(ast *TableNameField, data unsafe.Pointer) {
	// TODO
	panic("pass")
}

func (af *ASTFinder) VisitTableArrayField(ast *TableArrayField, data unsafe.Pointer) {
	// TODO
	panic("pass")
}

func (af *ASTFinder) VisitIndexAccessor(ast *IndexAccessor, data unsafe.Pointer) {
	// TODO
	panic("pass")
}

func (af *ASTFinder) VisitMemberAccessor(ast *MemberAccessor, data unsafe.Pointer) {
	// TODO
	panic("pass")
}

func (af *ASTFinder) VisitNormalFuncCall(ast *NormalFuncCall, data unsafe.Pointer) {
	// TODO
	panic("pass")
}

func (af *ASTFinder) VisitMemberFuncCall(ast *MemberFuncCall, data unsafe.Pointer) {
	// TODO
	panic("pass")
}

func (af *ASTFinder) VisitFuncCallArgs(ast *FuncCallArgs, data unsafe.Pointer) {
	// TODO
	panic("pass")
}

func (af *ASTFinder) VisitExpressionList(ast *ExpressionList, data unsafe.Pointer) {
	if af.theASTNode == nil {
		f := func() {
			for i := range ast.ExpList {
				ast.ExpList[i].Accept(af, nil)
			}
		}
		af.setResult(true, ast, f)
	}
}

func (af *ASTFinder) setResult(isSameType bool, theType SyntaxTree, op func()) {
	if isSameType {
		switch finder := af.finder.(type) {
		case func(SyntaxTree) bool:
			if finder(theType) {
				af.theASTNode = theType
			}
		}
	} else {
		op()
	}
}

func ASTFind(root SyntaxTree, finderType interface{}) interface{} {
	switch finder := finderType.(type) {
	case func(SyntaxTree) bool:
		astFinder := NewASTFinder(finder)
		root.Accept(astFinder, nil)
		return astFinder.GetResult()
	default:
		panic("assert")
	}
}

func FindName(name string, term *Terminator) bool {
	if term.Token.Token == TokenId {
		return term.Token.Str.GetStdString() == name
	} else {
		return false
	}
}
