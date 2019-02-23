package vm

import (
	"unsafe"
)

// Lexical block data in LexicalFunction for finding
type lexicalBlock struct {
	Parent *lexicalBlock
	// Local names
	// Same names are the same instance String, so using String
	// pointer as key is fine
	Names map[*String]bool // as set[*String]
}

func newLexicalBlock() *lexicalBlock {
	return &lexicalBlock{}
}

// Lexical function data for name finding
type lexicalFunction struct {
	Parent       *lexicalFunction
	CurrentBlock *lexicalBlock
	CurrentLoop  SyntaxTree
	HasVararg    bool
}

func newLexicalFunction() *lexicalFunction {
	return &lexicalFunction{}
}

type semanticAnalysisVisitor struct {
	state           *State
	currentFunction *lexicalFunction // Current lexical function for all names finding
}

func (sav *semanticAnalysisVisitor) VisitChunk(chunk *Chunk, data unsafe.Pointer) {
	NewGuard(func() { sav.EnterFunction() }, func() { sav.LeaveFunction() })

	{
		NewGuard(func() { sav.EnterBlock() }, func() { sav.LeaveBlock() })
		chunk.Block.Accept(sav, nil)
	}
}

func (sav *semanticAnalysisVisitor) VisitBlock(block *Block, data unsafe.Pointer) {
	for i := range block.Statements {
		block.Statements[i].Accept(sav, nil)
	}
	if block.ReturnStmt != nil {
		block.ReturnStmt.Accept(sav, nil)
	}
}

func (sav *semanticAnalysisVisitor) VisitReturnStatement(retStmt *ReturnStatement, data unsafe.Pointer) {
	if retStmt.ExpList != nil {
		var expListData expListData
		retStmt.ExpList.Accept(sav, unsafe.Pointer(&expListData))
		retStmt.ExpValueCount = expListData.ExpValueCount
	}
}

func (sav *semanticAnalysisVisitor) VisitBreakStatement(breakStmt *BreakStatement, data unsafe.Pointer) {
	breakStmt.Loop = sav.GetLoopAST()
	if breakStmt.Loop == nil {
		panic(NewSemanticError("not in any loop", breakStmt.Break))
	}
}

func (sav *semanticAnalysisVisitor) VisitDoStatement(doStmt *DoStatement, data unsafe.Pointer) {
	NewGuard(func() { sav.EnterBlock() }, func() { sav.LeaveBlock() })
	doStmt.Block.Accept(sav, nil)
}

func (sav *semanticAnalysisVisitor) VisitWhileStatement(whileStmt *WhileStatement, data unsafe.Pointer) {
	oldLoop := sav.GetLoopAST()
	NewGuard(func() { sav.SetLoopAST(whileStmt) }, func() { sav.SetLoopAST(oldLoop) })
	eVarData := newExpVarData(SemanticOpRead)
	whileStmt.Exp.Accept(sav, unsafe.Pointer(eVarData))

	NewGuard(func() { sav.EnterBlock() }, func() { sav.LeaveBlock() })
	whileStmt.Block.Accept(sav, nil)
}

func (sav *semanticAnalysisVisitor) VisitRepeatStatement(repeatStmt *RepeatStatement, data unsafe.Pointer) {
	oldLoop := sav.GetLoopAST()
	NewGuard(func() { sav.SetLoopAST(repeatStmt) }, func() { sav.SetLoopAST(oldLoop) })
	NewGuard(func() { sav.EnterBlock() }, func() { sav.LeaveBlock() })

	eVarData := newExpVarData(SemanticOpRead)
	repeatStmt.Block.Accept(sav, nil)
	repeatStmt.Exp.Accept(sav, unsafe.Pointer(eVarData))
}

func (sav *semanticAnalysisVisitor) VisitIfStatement(ifStmt *IfStatement, data unsafe.Pointer) {
	eVarData := newExpVarData(SemanticOpRead)
	ifStmt.Exp.Accept(sav, unsafe.Pointer(eVarData))

	{
		NewGuard(func() { sav.EnterBlock() }, func() { sav.LeaveBlock() })
		ifStmt.TrueBranch.Accept(sav, nil)
	}

	if ifStmt.FalseBranch != nil {
		ifStmt.FalseBranch.Accept(sav, nil)
	}
}

func (sav *semanticAnalysisVisitor) VisitElseIfStatement(elseifStmt *ElseIfStatement, data unsafe.Pointer) {
	eVarData := newExpVarData(SemanticOpRead)
	elseifStmt.Exp.Accept(sav, unsafe.Pointer(eVarData))

	{
		NewGuard(func() { sav.EnterBlock() }, func() { sav.LeaveBlock() })
		elseifStmt.TrueBranch.Accept(sav, nil)
	}

	if elseifStmt.FalseBranch != nil {
		elseifStmt.FalseBranch.Accept(sav, nil)
	}
}

func (sav *semanticAnalysisVisitor) VisitElseStatement(elseStmt *ElseStatement, data unsafe.Pointer) {
	NewGuard(func() { sav.EnterBlock() }, func() { sav.LeaveBlock() })
	elseStmt.Block.Accept(sav, nil)
}

func (sav *semanticAnalysisVisitor) VisitNumericForStatement(numFor *NumericForStatement, data unsafe.Pointer) {
	oldLoop := sav.GetLoopAST()
	NewGuard(func() { sav.SetLoopAST(numFor) }, func() { sav.SetLoopAST(oldLoop) })
	eVarData := newExpVarData(SemanticOpRead)
	numFor.Exp1.Accept(sav, unsafe.Pointer(eVarData))
	numFor.Exp2.Accept(sav, unsafe.Pointer(eVarData))
	if numFor.Exp3 != nil {
		numFor.Exp3.Accept(sav, unsafe.Pointer(eVarData))
	}

	NewGuard(func() { sav.EnterBlock() }, func() { sav.LeaveBlock() })
	sav.InsertName(numFor.Name.Str)
	numFor.Block.Accept(sav, nil)
}

func (sav *semanticAnalysisVisitor) VisitGenericForStatement(genFor *GenericForStatement, data unsafe.Pointer) {
	oldLoop := sav.GetLoopAST()
	NewGuard(func() { sav.SetLoopAST(genFor) }, func() { sav.SetLoopAST(oldLoop) })
	var eListData expListData
	genFor.ExpList.Accept(sav, unsafe.Pointer(&eListData))

	NewGuard(func() { sav.EnterBlock() }, func() { sav.LeaveBlock() })
	var nameListData nameListData
	genFor.NameList.Accept(sav, unsafe.Pointer(&nameListData))
	genFor.Block.Accept(sav, nil)
}

func (sav *semanticAnalysisVisitor) VisitFunctionStatement(funcStmt *FunctionStatement, data unsafe.Pointer) {
	var nameData functionNameData
	funcStmt.FuncName.Accept(sav, unsafe.Pointer(&nameData))

	// Set FunctionBody has 'self' param when FunctionName has member token
	if nameData.HasMemberToken {
		// TODO
	}

	funcStmt.FuncBody.Accept(sav, nil)
}

func (sav *semanticAnalysisVisitor) VisitFunctionName(funcName *FunctionName, data unsafe.Pointer) {
	if len(funcName.Names) == 0 {
		panic("assert")
	}
	// Get the scoping of first token of FunctionName
	funcName.Scoping = sav.SearchName(funcName.Names[0].Str)

	// Set FunctionNameData
	(*functionNameData)(data).HasMemberToken = funcName.MemberName.Token == TokenId
}

func (sav *semanticAnalysisVisitor) VisitLocalFunctionStatement(lFuncStmt *LocalFunctionStatement, data unsafe.Pointer) {
	sav.InsertName(lFuncStmt.Name.Str)
	lFuncStmt.FuncBody.Accept(sav, nil)
}

func (sav *semanticAnalysisVisitor) VisitLocalNameListStatement(lNameListStmt *LocalNameListStatement, data unsafe.Pointer) {
	if lNameListStmt.ExpList != nil {
		var eListData expListData
		lNameListStmt.ExpList.Accept(sav, unsafe.Pointer(&eListData))
	}

	var nameListData nameListData
	lNameListStmt.NameList.Accept(sav, unsafe.Pointer(&nameListData))
	lNameListStmt.NameCount = nameListData.NameCount
}

func (sav *semanticAnalysisVisitor) VisitAssignmentStatement(assignStmt *AssignmentStatement, data unsafe.Pointer) {
	var vListData varListData
	var eListData expListData
	assignStmt.VarList.Accept(sav, unsafe.Pointer(&vListData))
	assignStmt.ExpList.Accept(sav, unsafe.Pointer(&eListData))
	assignStmt.VarCount = vListData.VarCount
}

func (sav *semanticAnalysisVisitor) VisitVarList(varList *VarList, data unsafe.Pointer) {
	eVarData := newExpVarData(SemanticOpWrite)
	for i := range varList.VarList {
		varList.VarList[i].Accept(sav, unsafe.Pointer(eVarData))
	}
	(*varListData)(data).VarCount = len(varList.VarList)
}

func (sav *semanticAnalysisVisitor) VisitTerminator(term *Terminator, data unsafe.Pointer) {
	// Current term read and write semantic
	eVarData := (*expVarData)(data)
	term.Semantic = eVarData.SemanticOp
	if term.Token.Token != TokenId {
		if eVarData.SemanticOp != SemanticOpRead {
			panic("assert")
		}
	}

	// Set Expression type
	switch term.Token.Token {
	case TokenNil:
		eVarData.ExpType = ExpTypeNil
	case TokenId:
		eVarData.ExpType = ExpTypeUnknown
	case TokenNumber:
		eVarData.ExpType = ExpTypeNumber
	case TokenString:
		eVarData.ExpType = ExpTypeString
	case TokenVarArg:
		eVarData.ExpType = ExpTypeVarArg
	case TokenTrue, TokenFalse:
		eVarData.ExpType = ExpTypeBool
	}

	// Search lexical scoping of name
	if term.Token.Token == TokenId {
		term.Scoping = sav.SearchName(term.Token.Str)
	}

	// Check function has vararg
	if term.Token.Token == TokenVarArg && !sav.HasVararg() {
		panic(NewSemanticError("function has no '...' param", term.Token))
	}

	// Set results any count
	if term.Token.Token == TokenVarArg {
		eVarData.ResultsAnyCount = true
	}
}

func (sav *semanticAnalysisVisitor) VisitBinaryExpression(binaryExp *BinaryExpression, data unsafe.Pointer) {
	// Binary expression is read semantic
	lExpVarData := newExpVarData(SemanticOpRead)
	rExpVarData := newExpVarData(SemanticOpRead)
	binaryExp.Left.Accept(sav, unsafe.Pointer(lExpVarData))
	binaryExp.Right.Accept(sav, unsafe.Pointer(rExpVarData))

	parentExpVarData := (*expVarData)(data)
	switch binaryExp.OpToken.Token {
	case '+', '-', '*', '/', '^', '%':
		if lExpVarData.ExpType != ExpTypeUnknown && lExpVarData.ExpType != ExpTypeNumber {
			panic(NewSemanticError("left expression of binary operator is not number",
				binaryExp.OpToken))
		}
		if rExpVarData.ExpType != ExpTypeUnknown && rExpVarData.ExpType != ExpTypeNumber {
			panic(NewSemanticError("right expression of binary operator is not number",
				binaryExp.OpToken))
		}
		parentExpVarData.ExpType = ExpTypeNumber
	case '<', '>', TokenLessEqual, TokenGreaterEqual:
		if lExpVarData.ExpType != ExpTypeUnknown && rExpVarData.ExpType != ExpTypeUnknown {
			if lExpVarData.ExpType != ExpTypeUnknown && rExpVarData.ExpType != ExpTypeUnknown {
				if lExpVarData.ExpType != rExpVarData.ExpType {
					panic(NewSemanticError("compare different expression type", binaryExp.OpToken))
				} else if lExpVarData.ExpType != ExpTypeNumber && lExpVarData.ExpType != ExpTypeString {
					panic(NewSemanticError("can not compare operands", binaryExp.OpToken))
				}
			}
		}
		parentExpVarData.ExpType = ExpTypeBool
	case TokenConcat:
		if lExpVarData.ExpType != ExpTypeUnknown && rExpVarData.ExpType != ExpTypeUnknown {
			if !((lExpVarData.ExpType == ExpTypeString && rExpVarData.ExpType == ExpTypeString) ||
				(lExpVarData.ExpType == ExpTypeString && rExpVarData.ExpType == ExpTypeNumber) ||
				(lExpVarData.ExpType == ExpTypeNumber && rExpVarData.ExpType == ExpTypeString)) {
				panic(NewSemanticError("can not concat operands", binaryExp.OpToken))
			}
		}
		parentExpVarData.ExpType = ExpTypeString
	case TokenNotEqual, TokenEqual:
		parentExpVarData.ExpType = ExpTypeBool
	case TokenAnd, TokenOr:
		// Do nothing
	default:
		// Do nothing
	}
}

func (sav *semanticAnalysisVisitor) VisitUnaryExpression(unaryExp *UnaryExpression, data unsafe.Pointer) {
	// Unary expression is read semantic
	eVarData := newExpVarData(SemanticOpRead)
	unaryExp.Exp.Accept(sav, unsafe.Pointer(eVarData))

	// Expression type
	if eVarData.ExpType != ExpTypeUnknown {
		switch unaryExp.OpToken.Token {
		case '-':
			if eVarData.ExpType != ExpTypeNumber {
				panic(NewSemanticError("operand is not number", unaryExp.OpToken))
			}
		case '#':
			if eVarData.ExpType != ExpTypeTable && eVarData.ExpType != ExpTypeString {
				panic(NewSemanticError("operand is not table or string", unaryExp.OpToken))
			}
		default:
		}
	}

	parentExpVarData := (*expVarData)(data)
	if unaryExp.OpToken.Token == '-' || unaryExp.OpToken.Token == '#' {
		parentExpVarData.ExpType = ExpTypeNumber
	} else if unaryExp.OpToken.Token == TokenNot {
		parentExpVarData.ExpType = ExpTypeBool
	} else {
		panic("unexpect unary operator")
	}
}

func (sav *semanticAnalysisVisitor) VisitFunctionBody(funcBody *FunctionBody, data unsafe.Pointer) {
	NewGuard(func() { sav.EnterFunction() }, func() { sav.LeaveFunction() })

	{
		NewGuard(func() { sav.EnterBlock() }, func() { sav.LeaveBlock() })
		if funcBody.HasSelf {
			self := sav.state.GetString("self")
			sav.InsertName(self)
		}

		if funcBody.ParamList != nil {
			funcBody.ParamList.Accept(sav, nil)
		}

		funcBody.BLock.Accept(sav, nil)
	}
}

func (sav *semanticAnalysisVisitor) VisitParamList(parList *ParamList, data unsafe.Pointer) {
	if parList.NameList != nil {
		var nameListData nameListData
		parList.NameList.Accept(sav, unsafe.Pointer(&nameListData))
		parList.FixArgCount = nameListData.NameCount
	}

	if parList.Vararg {
		sav.SetFunctionVararg()
	}
}

func (sav *semanticAnalysisVisitor) VisitNameList(nameList *NameList, data unsafe.Pointer) {
	size := len(nameList.Names)
	(*nameListData)(data).NameCount = size

	for i := 0; i < size; i++ {
		sav.InsertName(nameList.Names[i].Str)
	}
}

func (sav *semanticAnalysisVisitor) VisitTableDefine(tableDef *TableDefine, data unsafe.Pointer) {
	for i := range tableDef.Fields {
		tableDef.Fields[i].Accept(sav, nil)
	}

	// Set Expression type
	(*expVarData)(data).ExpType = ExpTypeTable
}

func (sav *semanticAnalysisVisitor) VisitTableIndexField(tableIField *TableIndexField, data unsafe.Pointer) {
	// Table Index and Value expressions are read semantic
	eVarData := newExpVarData(SemanticOpRead)
	tableIField.Index.Accept(sav, unsafe.Pointer(eVarData))
	tableIField.Value.Accept(sav, unsafe.Pointer(eVarData))
}

func (sav *semanticAnalysisVisitor) VisitTableNameField(tableNField *TableNameField, data unsafe.Pointer) {
	// Table Value expression is read semantic
	eVarData := newExpVarData(SemanticOpRead)
	tableNField.Value.Accept(sav, unsafe.Pointer(eVarData))
}

func (sav *semanticAnalysisVisitor) VisitTableArrayField(tableAField *TableArrayField, data unsafe.Pointer) {
	// Table Value expression is read semantic
	eVarData := newExpVarData(SemanticOpRead)
	tableAField.Value.Accept(sav, unsafe.Pointer(eVarData))
}

func (sav *semanticAnalysisVisitor) VisitIndexAccessor(iAccessor *IndexAccessor, data unsafe.Pointer) {
	// Set this IndexAccessor semantic from parent's semantic data
	iAccessor.Semantic = (*expVarData)(data).SemanticOp

	// Table and Index expression are read semantic
	eVarData := newExpVarData(SemanticOpRead)
	iAccessor.Table.Accept(sav, unsafe.Pointer(eVarData))
	iAccessor.Index.Accept(sav, unsafe.Pointer(eVarData))
}

func (sav *semanticAnalysisVisitor) VisitMemberAccessor(mAccessor *MemberAccessor, data unsafe.Pointer) {
	// Set this MemberAccessor semantic from parent's semantic data
	mAccessor.Semantic = (*expVarData)(data).SemanticOp

	// Table expression is read semantic
	eVarData := newExpVarData(SemanticOpRead)
	mAccessor.Table.Accept(sav, unsafe.Pointer(eVarData))
}

func (sav *semanticAnalysisVisitor) VisitNormalFuncCall(nFuncCall *NormalFuncCall, data unsafe.Pointer) {
	// Function call must be read semantic
	eVarData := newExpVarData(SemanticOpRead)
	nFuncCall.Caller.Accept(sav, unsafe.Pointer(eVarData))
	nFuncCall.Args.Accept(sav, unsafe.Pointer(eVarData))

	if data != nil {
		(*expVarData)(data).ResultsAnyCount = true
	}
}

func (sav *semanticAnalysisVisitor) VisitMemberFuncCall(mFuncCall *MemberFuncCall, data unsafe.Pointer) {
	// Function call must be read semantic
	eVarData := newExpVarData(SemanticOpRead)
	mFuncCall.Caller.Accept(sav, unsafe.Pointer(eVarData))
	mFuncCall.Args.Accept(sav, unsafe.Pointer(eVarData))

	if data != nil {
		(*expVarData)(data).ResultsAnyCount = true
	}
}

func (sav *semanticAnalysisVisitor) VisitFuncCallArgs(callArgs *FuncCallArgs, data unsafe.Pointer) {
	if callArgs.Type == ArgTypeExpList {
		if callArgs.Arg != nil {
			var expListData expListData
			callArgs.Arg.Accept(sav, unsafe.Pointer(&expListData))
			callArgs.ArgValueCount = expListData.ExpValueCount
		}
	} else {
		expVarData := newExpVarData(SemanticOpRead)
		callArgs.Arg.Accept(sav, unsafe.Pointer(expVarData))
		callArgs.ArgValueCount = 1
	}
}

func (sav *semanticAnalysisVisitor) VisitExpressionList(expList *ExpressionList, data unsafe.Pointer) {
	if len(expList.ExpList) == 0 {
		panic("assert")
	}

	// Expressions in ExpressionList must be read semantic
	size := len(expList.ExpList) - 1
	for i := 0; i < size; i++ {
		expVarData := newExpVarData(SemanticOpRead)
		expList.ExpList[i].Accept(sav, unsafe.Pointer(expVarData))
	}

	// If the last expression in list which has any count value results,
	// then this expression list has any count value results also
	expVarData := newExpVarData(SemanticOpRead)
	expList.ExpList[size].Accept(sav, unsafe.Pointer(expVarData))
	if expVarData.ResultsAnyCount {
		(*expListData)(data).ExpValueCount = ExpValueCountAny
	} else {
		(*expListData)(data).ExpValueCount = size + 1
	}
}

func newSemanticAnalysisVisitor(state *State) *semanticAnalysisVisitor {
	return &semanticAnalysisVisitor{state, nil}
}

// Enter a new function AST, and add a new LexicalFunction data structure
func (sav *semanticAnalysisVisitor) EnterFunction() {
	function := newLexicalFunction()
	function.Parent = sav.currentFunction
	sav.currentFunction = function
}

// Leave a function AST, and delete current LexicalFunction data structure
func (sav *semanticAnalysisVisitor) LeaveFunction() {
	sav.deleteCurrentFunction()
}

// Add a new LexicalBlock for entering a new block scope
func (sav *semanticAnalysisVisitor) EnterBlock() {
	if sav.currentFunction == nil {
		panic("assert")
	}
	block := newLexicalBlock()
	block.Parent = sav.currentFunction.CurrentBlock
	sav.currentFunction.CurrentBlock = block
}

// Delete current LexicalBlock when leaving a block scope
func (sav *semanticAnalysisVisitor) LeaveBlock() {
	sav.deleteCurrentBlock()
}

// Insert a name into current block, replace its info when existed
func (sav *semanticAnalysisVisitor) InsertName(name *String) {
	if sav.currentFunction == nil || sav.currentFunction.CurrentBlock == nil {
		panic("assert")
	}
	sav.currentFunction.CurrentBlock.Names[name] = true
}

// Search LexicalScoping of a name
func (sav *semanticAnalysisVisitor) SearchName(str *String) int {
	if sav.currentFunction == nil || sav.currentFunction.CurrentBlock == nil {
		panic("assert")
	}

	function := sav.currentFunction
	for function != nil {
		block := function.CurrentBlock
		for block != nil {
			if block.Names[str] {
				if function == sav.currentFunction {
					return LexicalScopingLocal
				} else {
					return LexicalScopingUpvalue
				}
			}

			block = block.Parent
		}

		function = function.Parent
	}

	return LexicalScopingGlobal
}

// Set current loop AST
func (sav *semanticAnalysisVisitor) SetLoopAST(loop SyntaxTree) {
	sav.currentFunction.CurrentLoop = loop
}

// Get current loop AST
func (sav *semanticAnalysisVisitor) GetLoopAST() SyntaxTree {
	return sav.currentFunction.CurrentLoop
}

// Set current function has a vararg param
func (sav *semanticAnalysisVisitor) SetFunctionVararg() {
	sav.currentFunction.HasVararg = true
}

// Current function has a vararg or not
func (sav *semanticAnalysisVisitor) HasVararg() bool {
	return sav.currentFunction.HasVararg
}

func (sav *semanticAnalysisVisitor) deleteCurrentFunction() {
	if sav.currentFunction == nil {
		panic("assert")
	}
	// delete all blocks in current function
	sav.deleteBlocks()

	// delete current function
	function := sav.currentFunction
	sav.currentFunction = function.Parent
	function = nil
}

func (sav *semanticAnalysisVisitor) deleteCurrentBlock() {
	if sav.currentFunction == nil || sav.currentFunction.CurrentBlock == nil {
		panic("assert")
	}
	block := sav.currentFunction.CurrentBlock
	sav.currentFunction.CurrentBlock = block.Parent
	block = nil
}

func (sav *semanticAnalysisVisitor) deleteBlocks() {
	for sav.currentFunction.CurrentBlock != nil {
		sav.deleteCurrentBlock()
	}
}

// For NameList AST
type nameListData struct {
	NameCount int
}

func newNameListData() *nameListData {
	return &nameListData{}
}

// For VarList AST
type varListData struct {
	VarCount int
}

func newVarListData() *varListData {
	return &varListData{}
}

// For ExpList AST
type expListData struct {
	ExpValueCount int
}

func newExpListData() *expListData {
	return &expListData{}
}

// Expression type for semantic
const (
	ExpTypeUnknown = iota // unknown type at semantic analysis phase
	ExpTypeNil
	ExpTypeBool
	ExpTypeNumber
	ExpTypeString
	ExpTypeVarArg
	ExpTypeTable
)

// Expression or variable data for semantic analysis
type expVarData struct {
	SemanticOp      int
	ExpType         int
	ResultsAnyCount bool
}

func newExpVarData(semanticOp int) *expVarData {
	return &expVarData{semanticOp, ExpTypeUnknown, false}
}

// For FunctionName
type functionNameData struct {
	HasMemberToken bool
}

func newFunctionNameData() *functionNameData {
	return &functionNameData{}
}

func SemanticAnalysis(root SyntaxTree, state *State) {
	if root == nil || state == nil {
		panic("assert")
	}
	semanticAnalysis := *newSemanticAnalysisVisitor(state)
	root.Accept(&semanticAnalysis, nil)
}
