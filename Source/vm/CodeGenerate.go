package vm

import (
	"container/list"
	"unsafe"
)

const maxFunctionRegisterCount = 250
const maxClosureUpvalueCount = 250

type localNameInfo struct {
	RegisterId int // Name register id
	BeginPc    int // Name begin instruction
}

func newLocalNameInfo(registerId, beginPc int) *localNameInfo {
	return &localNameInfo{registerId, beginPc}
}

// Loop AST info data in GenerateBlock
type loopInfo struct {
	LoopAst    SyntaxTree // Loop AST
	StartIndex int        // Start instruction index
}

func newLoopInfo() *loopInfo {
	return &loopInfo{}
}

// Lexical block struct for code generator
type generateBlock struct {
	Parent *generateBlock
	// Current block register start id
	RegisterStartId int
	// Local names
	// Same names are the same instance String, so using String
	// pointer as key is fine
	Names map[*String]localNameInfo
	// Current loop ast info
	CurrentLoop loopInfo
}

func newGenerateBlock() *generateBlock {
	return &generateBlock{}
}

// Jump info for loop AST
type loopJumpInfo struct {
	LoopAst          SyntaxTree // Owner loop AST
	JumpType         int        // Jump to AST head or tail
	InstructionIndex int        // Instruction need to be filled
}

const (
	jumpHead = iota
	jumpTail
)

func newLoopJumpInfo(loopAst SyntaxTree, jumpType int, instructionIndex int) *loopJumpInfo {
	return &loopJumpInfo{loopAst, jumpType, instructionIndex}
}

// Lexical function struct for code generator
type generateFunction struct {
	Parent       *generateFunction
	CurrentBlock *generateBlock // Current block
	Function_    *Function      // Current function for code generate
	FuncIndex    int            // Index of current function in parent
	RegisterId   int            // Register id generator
	RegisterMax  int            // Max register count used in current function
	LoopJumps    list.List      // To be filled loop jump info, and its element.value is *loopJumpInfo
}

func newGenerateFunction() *generateFunction {
	return &generateFunction{}
}

type codeGenerateVisitor struct {
	state           *State
	currentFunction *generateFunction // Current code generating function
}

func newCodeGenerateVisitor(state *State) *codeGenerateVisitor {
	return &codeGenerateVisitor{state: state}
}

func (cgv *codeGenerateVisitor) VisitChunk(chunk *Chunk, data unsafe.Pointer) {
	NewGuard(func() { cgv.EnterFunction() }, func() { cgv.LeaveFunction() })
	{
		// Generate function code
		function := cgv.GetCurrentFunction()
		function.SetModuleName(chunk.Module)
		function.SetLine(1)

		NewGuard(func() { cgv.EnterBlock() }, func() { cgv.LeaveBlock() })
		chunk.Block.Accept(cgv, nil)

		// New one closure
		closure := cgv.state.NewClosure()
		closure.SetPrototype(function)

		// Put closure on stack
		top := cgv.state.stack.Top
		cgv.state.stack.Top = vPointerAdd(cgv.state.stack.Top, 1)
		top.Closure = closure
		top.Type = ValueTClosure
	}
}

func (cgv *codeGenerateVisitor) VisitBlock(block *Block, data unsafe.Pointer) {
	for i := range block.Statements {
		block.Statements[i].Accept(cgv, nil)
	}
	if block.ReturnStmt != nil {
		block.ReturnStmt.Accept(cgv, nil)
	}
}

func (cgv *codeGenerateVisitor) VisitReturnStatement(retStmt *ReturnStatement, data unsafe.Pointer) {
	registerId := cgv.GetNextRegisterId()
	if retStmt.ExpList != nil {
		registerId, err := cgv.GenerateRegisterId()
		if err != nil {
			panic(err)
		}
		eListData := newCgExpListData(registerId, ExpValueCountAny)
		retStmt.ExpList.Accept(cgv, unsafe.Pointer(eListData))
	}

	function := cgv.GetCurrentFunction()
	instruction := AsBxCode(OpTypeRet, registerId, retStmt.ExpValueCount)
	function.AddInstruction(instruction, retStmt.Line)
}

func (cgv *codeGenerateVisitor) VisitBreakStatement(breakStmt *BreakStatement, data unsafe.Pointer) {
	if breakStmt.Loop != nil {
		panic("assert")
	}
	function := cgv.GetCurrentFunction()
	instruction := AsBxCode(OpTypeJmp, 0, 0)
	index := function.AddInstruction(instruction, breakStmt.Break.Line)
	cgv.AddLoopJumpInfo(breakStmt.Loop, index, jumpTail)
}

func (cgv *codeGenerateVisitor) VisitDoStatement(doStmt *DoStatement, data unsafe.Pointer) {
	NewGuard(func() { cgv.EnterBlock() }, func() { cgv.LeaveBlock() })
	doStmt.Block.Accept(cgv, nil)
}

func (cgv *codeGenerateVisitor) VisitWhileStatement(whileStmt *WhileStatement, data unsafe.Pointer) {
	NewGuard(func() { cgv.EnterBlock() }, func() { cgv.LeaveBlock() })
	NewGuard(func() { cgv.EnterLoop(whileStmt) }, func() { cgv.LeaveLoop() })

	registerId, err := cgv.GenerateRegisterId()
	if err != nil {
		panic(err)
	}
	eVarData := newCgExpVarData(registerId, registerId+1)
	whileStmt.Exp.Accept(cgv, unsafe.Pointer(eVarData))

	// Jump to loop tail when expression is false
	function := cgv.GetCurrentFunction()
	instruction := AsBxCode(OpTypeJmpFalse, registerId, 0)
	index := function.AddInstruction(instruction, whileStmt.FirstLine)
	cgv.AddLoopJumpInfo(whileStmt, index, jumpTail)

	whileStmt.Block.Accept(cgv, nil)

	// Jump to loop head
	instruction = AsBxCode(OpTypeJmp, 0, 0)
	index = function.AddInstruction(instruction, whileStmt.LastLine)
	cgv.AddLoopJumpInfo(whileStmt, index, jumpHead)
}

func (cgv *codeGenerateVisitor) VisitRepeatStatement(repeatStmt *RepeatStatement, data unsafe.Pointer) {
	NewGuard(func() { cgv.EnterBlock() }, func() { cgv.LeaveBlock() })
	NewGuard(func() { cgv.EnterLoop(repeatStmt) }, func() { cgv.LeaveLoop() })
	{
		r := cgv.GetNextRegisterId()
		NewGuard(func() {}, func() { cgv.ResetRegisterIdGenerator(r) })
		repeatStmt.Block.Accept(cgv, nil)
	}

	// Get exp value
	registerId, err := cgv.GenerateRegisterId()
	if err != nil {
		panic(err)
	}
	eVarData := newCgExpVarData(registerId, registerId+1)
	repeatStmt.Exp.Accept(cgv, unsafe.Pointer(eVarData))

	// Jump to head when exp value is true
	function := cgv.GetCurrentFunction()
	instruction := AsBxCode(OpTypeJmpFalse, registerId, 0)
	index := function.AddInstruction(instruction, repeatStmt.Line)
	cgv.AddLoopJumpInfo(repeatStmt, index, jumpHead)
}

func (cgv *codeGenerateVisitor) VisitIfStatement(ifStmt *IfStatement, data unsafe.Pointer) {
	cgv.ifStatementGenerateCode(ifStmt)
}

func (cgv *codeGenerateVisitor) VisitElseIfStatement(elseifStmt *ElseIfStatement, data unsafe.Pointer) {
	cgv.ifStatementGenerateCode(elseifStmt)
}

func (cgv *codeGenerateVisitor) VisitElseStatement(elseStmt *ElseStatement, data unsafe.Pointer) {
	elseStmt.Block.Accept(cgv, nil)
}

func (cgv *codeGenerateVisitor) VisitNumericForStatement(numFor *NumericForStatement, data unsafe.Pointer) {
	NewGuard(func() { cgv.EnterBlock() }, func() { cgv.LeaveBlock() })

	varRegister, err := cgv.GenerateRegisterId()
	if err != nil {
		panic(err)
	}
	limitRegister, err := cgv.GenerateRegisterId()
	if err != nil {
		panic(err)
	}
	stepRegister, err := cgv.GenerateRegisterId()
	if err != nil {
		panic(err)
	}
	function := cgv.GetCurrentFunction()
	line := numFor.Name.Line

	// Init name, limit, step
	{
		r := cgv.GetNextRegisterId()
		NewGuard(func() {}, func() { cgv.ResetRegisterIdGenerator(r) })
		nameExpData := newCgExpVarData(varRegister, varRegister+1)
		numFor.Exp1.Accept(cgv, unsafe.Pointer(nameExpData))
	}
	{
		r := cgv.GetNextRegisterId()
		NewGuard(func() {}, func() { cgv.ResetRegisterIdGenerator(r) })
		limitExpData := newCgExpVarData(limitRegister, limitRegister+1)
		numFor.Exp2.Accept(cgv, unsafe.Pointer(limitExpData))
	}
	{
		r := cgv.GetNextRegisterId()
		NewGuard(func() {}, func() { cgv.ResetRegisterIdGenerator(r) })
		if numFor.Exp3 != nil {
			stepExpData := newCgExpVarData(stepRegister, stepRegister+1)
			numFor.Exp3.Accept(cgv, unsafe.Pointer(stepExpData))
		} else {
			// Default step is 1
			instruction := ACode(OpTypeLoadInt, stepRegister)
			function.AddInstruction(instruction, line)
			// Int value 1
			instruction.OpCode = 1
			function.AddInstruction(instruction, line)
		}
	}

	// Init 'for' var, limit, step value
	instruction := ABCCode(OpTypeFillNil, varRegister, limitRegister, stepRegister)
	function.AddInstruction(instruction, line)

	NewGuard(func() { cgv.EnterLoop(numFor) }, func() { cgv.LeaveLoop() })
	{
		NewGuard(func() { cgv.EnterBlock() }, func() { cgv.LeaveBlock() })

		// Check 'for', continu loop or not
		instruction := ABCCode(OpTypeForStep, varRegister, limitRegister, stepRegister)
		function.AddInstruction(instruction, line)

		// Break loop, prepare to jump to the end of the loop
		instruction.OpCode = 0
		index := function.AddInstruction(instruction, line)
		cgv.AddLoopJumpInfo(numFor, index, jumpTail)

		nameRegister, err := cgv.GenerateRegisterId()
		if err != nil {
			panic(err)
		}
		cgv.InsertName(numFor.Name.Str, nameRegister)

		// Prepare name value
		instruction = ABCode(OpTypeMove, nameRegister, varRegister)
		function.AddInstruction(instruction, line)

		numFor.Block.Accept(cgv, nil)

		// var = var + step
		instruction = ABCCode(OpTypeAdd, varRegister, varRegister, stepRegister)
		function.AddInstruction(instruction, line)
	}
	// Jump to the begin of the loop
	instruction = AsBxCode(OpTypeJmpFalse, 0, 0)
	index := function.AddInstruction(instruction, line)
	cgv.AddLoopJumpInfo(numFor, index, jumpHead)
}

func (cgv *codeGenerateVisitor) VisitGenericForStatement(genFor *GenericForStatement, data unsafe.Pointer) {
	NewGuard(func() { cgv.EnterBlock() }, func() { cgv.LeaveBlock() })

	// Init generic for statement data
	funcRegister, err := cgv.GenerateRegisterId()
	if err != nil {
		panic(err)
	}
	stateRegister, err := cgv.GenerateRegisterId()
	if err != nil {
		panic(err)
	}
	varRegister, err := cgv.GenerateRegisterId()
	eListData := newCgExpListData(funcRegister, varRegister+1)
	genFor.ExpList.Accept(cgv, unsafe.Pointer(eListData))

	function := cgv.GetCurrentFunction()
	line := genFor.Line
	NewGuard(func() { cgv.EnterLoop(genFor) }, func() { cgv.LeaveLoop() })
	{
		NewGuard(func() { cgv.EnterBlock() }, func() { cgv.LeaveBlock() })

		// Alloc registers for names
		nameStart := cgv.GetNextRegisterId()
		nameListData := newCgNameListData(false)
		genFor.NameList.Accept(cgv, unsafe.Pointer(nameListData))
		nameEnd := cgv.GetNextRegisterId()
		if nameStart >= nameEnd {
			panic("assert")
		}

		// Alloc temp registers for call iterate function
		tempFunc, err := cgv.GenerateRegisterId()
		if err != nil {
			panic(err)
		}
		tempState, err := cgv.GenerateRegisterId()
		if err != nil {
			panic(err)
		}
		tempVar, err := cgv.GenerateRegisterId()
		if err != nil {
			panic(err)
		}

		// Call iterate function
		move := func(dst, src int) {
			instruction := ABCode(OpTypeMove, dst, src)
			function.AddInstruction(instruction, line)
		}
		move(tempFunc, funcRegister)
		move(tempState, stateRegister)
		move(tempVar, varRegister)

		instruction := ABCCode(OpTypeCall, tempFunc, 2+1, nameEnd-nameStart+1)
		function.AddInstruction(instruction, line)

		// Copy results to registers of names
		for name := nameStart; name < nameEnd; name++ {
			move(name, tempFunc)
			tempFunc++
		}

		// Break the loop when the first name value is nil
		instruction = AsBxCode(OpTypeJmpNil, nameStart, 0)
		index := function.AddInstruction(instruction, line)
		cgv.AddLoopJumpInfo(genFor, index, jumpTail)

		// Copy first name value to varRegister
		move(varRegister, nameStart)

		genFor.Block.Accept(cgv, nil)
	}

	// Jump to loop start
	instruction := AsBxCode(OpTypeJmp, 0, 0)
	index := function.AddInstruction(instruction, line)
	cgv.AddLoopJumpInfo(genFor, index, jumpHead)
}

func (cgv *codeGenerateVisitor) VisitFunctionStatement(funcStmt *FunctionStatement, data unsafe.Pointer) {
	r := cgv.GetNextRegisterId()
	NewGuard(func() {}, func() { cgv.ResetRegisterIdGenerator(r) })
	funcRegister, err := cgv.GenerateRegisterId()
	if err != nil {
		panic(err)
	}
	eVarData := newCgExpVarData(funcRegister, funcRegister+1)
	funcStmt.FuncBody.Accept(cgv, unsafe.Pointer(eVarData))

	nameData := newCgFunctionNameData(funcRegister)
	funcStmt.FuncName.Accept(cgv, unsafe.Pointer(nameData))
}

func (cgv *codeGenerateVisitor) VisitFunctionName(funcName *FunctionName, data unsafe.Pointer) {
	if len(funcName.Names) == 0 {
		panic("assert")
	}
	funcRegister := (*cgFunctionNameData)(data).FuncRegister
	function := cgv.GetCurrentFunction()

	hasMember := len(funcName.Names) > 1 || funcName.MemberName.Token == TokenId

	firstName := funcName.Names[0].Str
	firstLine := funcName.Names[0].Line

	if !hasMember {
		if len(funcName.Names) != 1 {
			panic("assert")
		}
		var instruction Instruction
		switch funcName.Scoping {
		case LexicalScopingGlobal:
			// Define a global function
			index := function.AddConstString(firstName)
			instruction = ABxCode(OpTypeSetGlobal, funcRegister, index)
		case LexicalScopingUpvalue:
			// Change a upvalue to a function
			index, err := cgv.PrepareUpvalue(firstName)
			if err != nil {
				panic(err)
			}
			instruction = ABCode(OpTypeSetUpvalue, funcRegister, index)
		case LexicalScopingLocal:
			// Change a local variable to a function
			localName := cgv.SearchLocalName(firstName)
			instruction = ABCode(OpTypeMove, localName.RegisterId, funcRegister)
		}
		function.AddInstruction(instruction, firstLine)
	} else {
		var instruction Instruction
		tableRegister, err := cgv.GenerateRegisterId()
		if err != nil {
			panic(err)
		}
		switch funcName.Scoping {
		case LexicalScopingGlobal:
			// Load global variable to table register
			index := function.AddConstString(firstName)
			instruction = ABxCode(OpTypeGetGlobal, tableRegister, index)
		case LexicalScopingUpvalue:
			// Load upvalue to table register
			index, err := cgv.PrepareUpvalue(firstName)
			if err != nil {
				panic(err)
			}
			instruction = ABCode(OpTypeMove, tableRegister, index)
		case LexicalScopingLocal:
			// Load local variable to table register
			localName := cgv.SearchLocalName(firstName)
			instruction = ABCode(OpTypeMove, tableRegister, localName.RegisterId)
		}
		function.AddInstruction(instruction, firstLine)

		member := funcName.MemberName.Token == TokenId
		size := len(funcName.Names)
		var count int
		if member {
			count = size
		} else {
			count = size - 1
		}
		keyRegister, err := cgv.GenerateRegisterId()
		if err != nil {
			panic(err)
		}

		loadKey := func(name *String, line int) {
			index := function.AddConstString(name)
			instruction = ABxCode(OpTypeLoadConst, keyRegister, index)
			function.AddInstruction(instruction, line)
		}

		for i := 0; i < count; i++ {
			// Get value from table by key
			name := funcName.Names[i].Str
			line := funcName.Names[i].Line
			loadKey(name, line)
			instruction = ABCCode(OpTypeGetTable, tableRegister, keyRegister, tableRegister)
			function.AddInstruction(instruction, line)
		}

		// Set function as value of table by key 'token'
		var token *TokenDetail
		if member {
			token = &funcName.MemberName
		} else {
			token = &funcName.Names[len(funcName.Names)-1]
		}
		loadKey(token.Str, token.Line)
		instruction = ABCCode(OpTypeSetTable, tableRegister, keyRegister, funcRegister)

		function.AddInstruction(instruction, token.Line)
	}
}

func (cgv *codeGenerateVisitor) VisitLocalFunctionStatement(lFuncStmt *LocalFunctionStatement, data unsafe.Pointer) {
	registerId, err := cgv.GenerateRegisterId()
	if err != nil {
		panic(err)
	}
	cgv.InsertName(lFuncStmt.Name.Str, registerId)
	eVarData := newCgExpVarData(registerId, registerId+1)
	lFuncStmt.FuncBody.Accept(cgv, unsafe.Pointer(eVarData))
}

func (cgv *codeGenerateVisitor) VisitLocalNameListStatement(lNameListStmt *LocalNameListStatement, data unsafe.Pointer) {
	// Generate code for expression list first, then expression list can get
	// variables which has the same name with variables defined in NameList
	// e.g.
	//     local i = 1
	//     local i = i -- i value is 1
	if lNameListStmt.ExpList != nil {
		// Reserve registers for NameList
		startRegister := cgv.GetNextRegisterId()
		endRegister := startRegister + lNameListStmt.NameCount
		NewGuard(func() { cgv.ResetRegisterIdGenerator(endRegister) },
			func() { cgv.ResetRegisterIdGenerator(startRegister) })

		eListData := newCgExpListData(startRegister, endRegister)
		lNameListStmt.ExpList.Accept(cgv, unsafe.Pointer(eListData))
	}

	// NameList need init itself when ExpList is not existed
	nameListData := newCgNameListData(lNameListStmt.ExpList == nil)
	lNameListStmt.NameList.Accept(cgv, unsafe.Pointer(nameListData))
}

func (cgv *codeGenerateVisitor) VisitAssignmentStatement(assignStmt *AssignmentStatement, data unsafe.Pointer) {
	r := cgv.GetNextRegisterId()
	NewGuard(func() {}, func() { cgv.ResetRegisterIdGenerator(r) })

	// Reserve rigisters for var list
	registerId := cgv.GetNextRegisterId()
	endRegister := registerId + assignStmt.VarCount
	cgv.ResetRegisterIdGenerator(endRegister)
	if cgv.IsRegisterCountOverflow() {
		panic(NewCodeGenerateError(cgv.GetCurrentFunction().GetModule().GetCStr(),
			assignStmt.Line, "assignment statement is too complex"))
	}

	// Get exp list results placed into [registerId, endRegister)
	eListData := newCgExpListData(registerId, endRegister)
	assignStmt.ExpList.Accept(cgv, unsafe.Pointer(eListData))

	// Assign results to var list
	varListData := newCgVarListData(registerId, endRegister)
	assignStmt.VarList.Accept(cgv, unsafe.Pointer(varListData))
}

func (cgv *codeGenerateVisitor) VisitVarList(varList *VarList, data unsafe.Pointer) {
	vListData := (*cgVarListData)(data)
	registerId := vListData.StartRegister
	endRegister := vListData.EndRegister
	varCount := len(varList.VarList)
	if endRegister-registerId != varCount {
		panic("assert")
	}

	// Assign results to each variable
	for i := 0; i < varCount; i++ {
		eVarData := newCgExpVarData(registerId, registerId+1)
		varList.VarList[i].Accept(cgv, unsafe.Pointer(eVarData))
		registerId++
	}
}

func (cgv *codeGenerateVisitor) VisitTerminator(term *Terminator, data unsafe.Pointer) {
	eVarData := (*cgExpVarData)(data)
	registerId := eVarData.StartRegister
	endRegister := eVarData.EndRegister
	function := cgv.GetCurrentFunction()

	// Generate code for SemanticOp Write
	if term.Semantic == SemanticOpWrite {
		if term.Token.Token != TokenId {
			panic("assert")
		}
		if registerId+1 != endRegister {
			panic("assert")
		}
		switch term.Scoping {
		case LexicalScopingGlobal:
			index := function.AddConstString(term.Token.Str)
			instruction := ABxCode(OpTypeSetGlobal, registerId, index)
			function.AddInstruction(instruction, term.Token.Line)
		case LexicalScopingLocal:
			local := cgv.SearchLocalName(term.Token.Str)
			if local == nil {
				panic("assert")
			}
			instruction := ABCode(OpTypeMove, local.RegisterId, registerId)
			function.AddInstruction(instruction, term.Token.Line)
		case LexicalScopingUpvalue:
			index, err := cgv.PrepareUpvalue(term.Token.Str)
			if err != nil {
				panic("assert")
			}
			instruction := ABCode(OpTypeSetUpvalue, registerId, index)
			function.AddInstruction(instruction, term.Token.Line)
		}
		return
	}

	// Generate code for SemanticOpRead
	// Just return when term is SemanticOpRead and no registers to fill
	if term.Semantic == SemanticOpRead &&
		endRegister != ExpValueCountAny &&
		registerId >= endRegister {
		return
	}

	switch term.Token.Token {
	case TokenNumber, TokenString:
		// Load const to register
		index := 0
		if term.Token.Token == TokenNumber {
			index = function.AddConstNumber(term.Token.Number)
		} else {
			index = function.AddConstString(term.Token.Str)
		}
		instruction := ABxCode(OpTypeLoadConst, registerId, index)
		registerId++
		function.AddInstruction(instruction, term.Token.Line)
	case TokenId:
		switch term.Scoping {
		case LexicalScopingGlobal:
			// Get value from global table by key index
			index := function.AddConstString(term.Token.Str)
			instruction := ABxCode(OpTypeGetGlobal, registerId, index)
			registerId++
			function.AddInstruction(instruction, term.Token.Line)
		case LexicalScopingLocal:
			// Load local variable value to dst register
			local := cgv.SearchLocalName(term.Token.Str)
			if local == nil {
				panic("assert")
			}
			instruction := ABCode(OpTypeMove, registerId, local.RegisterId)
			registerId++
			function.AddInstruction(instruction, term.Token.Line)
		case LexicalScopingUpvalue:
			// Get upvalue index
			index, err := cgv.PrepareUpvalue(term.Token.Str)
			if err != nil {
				panic("assert")
			}
			instruction := ABCode(OpTypeGetUpvalue, registerId, index)
			registerId++
			function.AddInstruction(instruction, term.Token.Line)
		}
	case TokenTrue, TokenFalse:
		var bvalue int
		if term.Token.Token == TokenTrue {
			bvalue = 1
		} else {
			bvalue = 0
		}
		instruction := ABCode(OpTypeLoadBool, registerId, bvalue)
		registerId++
		function.AddInstruction(instruction, term.Token.Line)
	case TokenNil:
		instruction := ACode(OpTypeLoadNil, registerId)
		registerId++
		function.AddInstruction(instruction, term.Token.Line)
	case TokenVarArg:
		// Copy vararg to registers which start from registerId
		var expectResults int
		if endRegister == ExpValueCountAny {
			expectResults = ExpValueCountAny
		} else {
			expectResults = endRegister - registerId
		}
		instruction := AsBxCode(OpTypeVarArg, registerId, expectResults)
		function.AddInstruction(instruction, term.Token.Line)

		// All registers will be filled when executing, so do not fill nil to remain registers
		registerId = endRegister
	}

	cgv.fillRemainRegisterNil(registerId, endRegister, term.Token.Line)
}

func (cgv *codeGenerateVisitor) VisitBinaryExpression(binaryExp *BinaryExpression, data unsafe.Pointer) {
	eVarData := (*cgExpVarData)(data)
	registerId := eVarData.StartRegister
	endRegister := eVarData.EndRegister

	if endRegister != ExpValueCountAny && registerId >= endRegister {
		return
	}

	function := cgv.GetCurrentFunction()
	line := binaryExp.OpToken.Line
	token := binaryExp.OpToken.Token
	if token == TokenAnd || token == TokenOr {
		// Calculate left expression
		leftData := newCgExpVarData(registerId, registerId+1)
		binaryExp.Left.Accept(cgv, unsafe.Pointer(leftData))

		// Do not calculate right expression when the result of left expression
		// satisfy semantic of operator
		var opType int
		if token == TokenAnd {
			opType = OpTypeJmpFalse
		} else {
			opType = OpTypeJmpTrue
		}
		instruction := AsBxCode(opType, registerId, 0)
		index := function.AddInstruction(instruction, line)

		// Calculate right expression
		rightData := newCgExpVarData(registerId, registerId+1)
		binaryExp.Right.Accept(cgv, unsafe.Pointer(rightData))

		dstIndex := function.OpCodeSize()
		function.GetMutableInstruction(index).RefillsBx(dstIndex - index)

		cgv.fillRemainRegisterNil(registerId+1, endRegister, line)
		return
	}

	leftRegister := 0
	// Generate code to calculate left expression
	{
		eVarData := newCgExpVarData(registerId, registerId+1)
		binaryExp.Left.Accept(cgv, unsafe.Pointer(eVarData))
		leftRegister = registerId
	}

	rightRegister := 0
	// Generate code to calculate right expression
	{
		if endRegister != ExpValueCountAny && registerId+1 < endRegister {
			// If parent AST provide more than one register, then use the second
			// register as temp register of right expression
			eVarData := newCgExpVarData(registerId+1, registerId+2)
			binaryExp.Right.Accept(cgv, unsafe.Pointer(eVarData))
			rightRegister = registerId + 1
		} else {
			// No more register, then generate a new register as temp register of
			// right expression
			r := cgv.GetNextRegisterId()
			NewGuard(func() {}, func() { cgv.ResetRegisterIdGenerator(r) })
			rightRegister = cgv.GetNextRegisterId()
			eVarData := newCgExpVarData(rightRegister, rightRegister+1)
			binaryExp.Right.Accept(cgv, unsafe.Pointer(eVarData))
		}
	}

	// Choose OpType by operator
	var opType int
	switch token {
	case '+':
		opType = OpTypeAdd
	case '-':
		opType = OpTypeSub
	case '*':
		opType = OpTypeMul
	case '/':
		opType = OpTypeDiv
	case '^':
		opType = OpTypePow
	case '%':
		opType = OpTypeMod
	case '<':
		opType = OpTypeLess
	case '>':
		opType = OpTypeGreater
	case TokenConcat:
		opType = OpTypeConcat
	case TokenEqual:
		opType = OpTypeEqual
	case TokenNotEqual:
		opType = OpTypeUnEqual
	case TokenLessEqual:
		opType = OpTypeLessEqual
	case TokenGreaterEqual:
		opType = OpTypeGreaterEqual
	default:
		panic("assert")
	}

	// Generate instruction to calculate
	instruction := ABCCode(opType, registerId, leftRegister, rightRegister)
	leftRegister++
	function.AddInstruction(instruction, line)
	cgv.fillRemainRegisterNil(registerId, endRegister, line)
}

func (cgv *codeGenerateVisitor) VisitUnaryExpression(unaryExp *UnaryExpression, data unsafe.Pointer) {
	eVarData := (*cgExpVarData)(data)
	registerId := eVarData.StartRegister
	endRegister := eVarData.EndRegister

	if endRegister != ExpValueCountAny && registerId >= endRegister {
		return
	}

	unaryExp.Exp.Accept(cgv, unsafe.Pointer(eVarData))

	// Choose OpType by operator
	var opType int
	switch unaryExp.OpToken.Token {
	case '-':
		opType = OpTypeNeg
	case '#':
		opType = OpTypeLen
	case TokenNot:
		opType = OpTypeNot
	default:
		panic("assert")
	}

	// Generate instruction
	function := cgv.GetCurrentFunction()
	instruction := ACode(opType, registerId)
	registerId++
	function.AddInstruction(instruction, unaryExp.OpToken.Line)

	cgv.fillRemainRegisterNil(registerId, endRegister, unaryExp.OpToken.Line)
}

func (cgv *codeGenerateVisitor) VisitFunctionBody(funcBody *FunctionBody, data unsafe.Pointer) {
	childIndex := 0
	{
		NewGuard(func() { cgv.EnterFunction() }, func() { cgv.LeaveFunction() })
		function := cgv.GetCurrentFunction()
		function.SetLine(funcBody.Line)
		childIndex = cgv.currentFunction.FuncIndex

		{
			NewGuard(func() { cgv.EnterBlock() }, func() { cgv.LeaveBlock() })
			// Child function generate code
			if funcBody.HasSelf {
				registerId, err := cgv.GenerateRegisterId()
				if err != nil {
					panic(err)
				}
				self := cgv.state.GetString("self")
				cgv.InsertName(self, registerId)

				function := cgv.GetCurrentFunction()
				function.AddFixedArgCount(1)
			}

			if funcBody.ParamList != nil {
				funcBody.ParamList.Accept(cgv, nil)
			}
			funcBody.BLock.Accept(cgv, nil)
		}
	}

	// Generate closure
	eVarData := (*cgExpVarData)(data)
	registerId := eVarData.StartRegister
	endRegister := eVarData.EndRegister
	if endRegister == ExpValueCountAny || registerId < endRegister {
		function := cgv.GetCurrentFunction()
		i := ABxCode(OpTypeClosure, registerId, childIndex)
		registerId++
		function.AddInstruction(i, funcBody.Line)
	}

	cgv.fillRemainRegisterNil(registerId, endRegister, funcBody.Line)
}

func (cgv *codeGenerateVisitor) VisitParamList(paramList *ParamList, data unsafe.Pointer) {
	function := cgv.GetCurrentFunction()
	function.AddFixedArgCount(paramList.FixArgCount)
	if paramList.Vararg {
		function.SetHasVararg()
	}

	if paramList.NameList != nil {
		nameListData := newCgNameListData(false)
		paramList.NameList.Accept(cgv, unsafe.Pointer(nameListData))
	}
}

func (cgv *codeGenerateVisitor) VisitNameList(nameList *NameList, data unsafe.Pointer) {
	needInit := (*cgNameListData)(data).NeedInit

	size := len(nameList.Names)
	for i := 0; i < size; i++ {
		registerId, err := cgv.GenerateRegisterId()
		if err != nil {
			panic(err)
		}
		cgv.InsertName(nameList.Names[i].Str, registerId)

		// Add init instructions when need
		if needInit {
			function := cgv.GetCurrentFunction()
			instruction := ACode(OpTypeLoadNil, registerId)
			function.AddInstruction(instruction, nameList.Names[i].Line)
		}
	}
}

func (cgv *codeGenerateVisitor) VisitTableDefine(tableDef *TableDefine, data unsafe.Pointer) {
	eVarData := (*cgExpVarData)(data)
	registerId := eVarData.StartRegister
	endRegister := eVarData.EndRegister

	// No register, then do not generate code
	if endRegister != ExpValueCountAny && registerId >= endRegister {
		return
	}

	// New table
	function := cgv.GetCurrentFunction()
	instruction := ACode(OpTypeNewTable, registerId)
	function.AddInstruction(instruction, tableDef.Line)

	if len(tableDef.Fields) != 0 {
		// Init table value
		fieldData := newCgTableFieldData(registerId)
		for i := range tableDef.Fields {
			r := cgv.GetNextRegisterId()
			NewGuard(func() {}, func() { cgv.ResetRegisterIdGenerator(r) })
			tableDef.Fields[i].Accept(cgv, unsafe.Pointer(fieldData))
		}
	}

	cgv.fillRemainRegisterNil(registerId+1, endRegister, tableDef.Line)
}

func (cgv *codeGenerateVisitor) VisitTableIndexField(tableIField *TableIndexField, data unsafe.Pointer) {
	fieldData := (*cgTableFieldData)(data)
	tableRegister := fieldData.TableRegister

	// Load key
	keyRegister, err := cgv.GenerateRegisterId()
	if err != nil {
		panic(err)
	}
	expVarData := newCgExpVarData(keyRegister, keyRegister+1)
	tableIField.Index.Accept(cgv, unsafe.Pointer(expVarData))

	cgv.setTableFieldValue(tableIField, tableRegister, keyRegister, tableIField.Line)
}

func (cgv *codeGenerateVisitor) VisitTableNameField(tableNField *TableNameField, data unsafe.Pointer) {
	fieldData := (*cgTableFieldData)(data)
	tableRegister := fieldData.TableRegister

	// Load key
	function := cgv.GetCurrentFunction()
	keyIndex := function.AddConstString(tableNField.Name.Str)
	keyRegister, err := cgv.GenerateRegisterId()
	if err != nil {
		panic(err)
	}
	instruction := AsBxCode(OpTypeLoadConst, keyRegister, keyIndex)
	function.AddInstruction(instruction, tableNField.Name.Line)

	cgv.setTableFieldValue(tableNField, tableRegister, keyRegister, tableNField.Name.Line)
}

func (cgv *codeGenerateVisitor) VisitTableArrayField(tableAField *TableArrayField, data unsafe.Pointer) {
	fieldData := (*cgTableFieldData)(data)
	tableRegister := fieldData.TableRegister

	// Load key
	function := cgv.GetCurrentFunction()
	keyRegister, err := cgv.GenerateRegisterId()
	if err != nil {
		panic(err)
	}
	instruction := ACode(OpTypeLoadInt, keyRegister)
	function.AddInstruction(instruction, tableAField.Line)
	instruction.OpCode = int(fieldData.ArrayIndex)
	fieldData.ArrayIndex++
	function.AddInstruction(instruction, tableAField.Line)

	cgv.setTableFieldValue(tableAField, tableRegister, keyRegister, tableAField.Line)
}

func (cgv *codeGenerateVisitor) VisitIndexAccessor(iAccessor *IndexAccessor, data unsafe.Pointer) {
	cgv.accessTableField(iAccessor, data, iAccessor.Line,
		func(keyRegister int) {
			data := newCgExpVarData(keyRegister, keyRegister+1)
			iAccessor.Index.Accept(cgv, unsafe.Pointer(data))
		})
}

func (cgv *codeGenerateVisitor) VisitMemberAccessor(mAccessor *MemberAccessor, data unsafe.Pointer) {
	cgv.accessTableField(mAccessor, data, mAccessor.Member.Line,
		func(keyRegister int) {
			function := cgv.GetCurrentFunction()
			keyIndex := function.AddConstString(mAccessor.Member.Str)
			instruction := ABxCode(OpTypeLoadConst, keyRegister, keyIndex)
			function.AddInstruction(instruction, mAccessor.Member.Line)
		})
}

func (cgv *codeGenerateVisitor) VisitNormalFuncCall(nFuncCall *NormalFuncCall, data unsafe.Pointer) {
	cgv.functionCall(nFuncCall, data, func(int) int { return 0 })
}

func (cgv *codeGenerateVisitor) VisitMemberFuncCall(mFuncCall *MemberFuncCall, data unsafe.Pointer) {
	cgv.functionCall(mFuncCall, data, func(callerRegister int) int {
		function := cgv.GetCurrentFunction()
		// Copy table to argRegister as first argument
		argRegister, err := cgv.GenerateRegisterId()
		if err != nil {
			panic(err)
		}
		instruction := ABCode(OpTypeMove, argRegister, callerRegister)
		function.AddInstruction(instruction, mFuncCall.Member.Line)

		{
			r := cgv.GetNextRegisterId()
			NewGuard(func() {}, func() { cgv.ResetRegisterIdGenerator(r) })
			// Get key
			index := function.AddConstString(mFuncCall.Member.Str)
			keyRegister, err := cgv.GenerateRegisterId()
			if err != nil {
				panic(err)
			}
			instruction := ABxCode(OpTypeLoadConst, keyRegister, index)
			function.AddInstruction(instruction, mFuncCall.Member.Line)
		}

		return 1
	})
}

func (cgv *codeGenerateVisitor) VisitFuncCallArgs(callArgs *FuncCallArgs, data unsafe.Pointer) {
	(*cgFuncCallArgsData)(data).ArgValueCount = callArgs.ArgValueCount

	if callArgs.Type == ArgTypeExpList {
		if callArgs.Arg != nil {
			startRegister, err := cgv.GenerateRegisterId()
			if err != nil {
				panic(err)
			}
			eListData := newCgExpListData(startRegister, ExpValueCountAny)
			callArgs.Arg.Accept(cgv, unsafe.Pointer(eListData))
		}
	} else {
		//
		startRegister, err := cgv.GenerateRegisterId()
		if err != nil {
			panic(err)
		}
		eVarData := newCgExpVarData(startRegister, startRegister+1)
		callArgs.Arg.Accept(cgv, unsafe.Pointer(eVarData))
	}
}

func (cgv *codeGenerateVisitor) VisitExpressionList(expList *ExpressionList, data unsafe.Pointer) {
	eListData := (*cgExpListData)(data)
	registerId := eListData.StartRegister
	endRegister := eListData.EndRegister

	// When parent do not limit register count, reset register id generator as consume some registers,
	// and check register count overflow or not
	registerConsumer := func(id int) error {
		if endRegister == ExpValueCountAny {
			cgv.ResetRegisterIdGenerator(id)
		}
		if cgv.IsRegisterCountOverflow() {
			return NewCodeGenerateError(cgv.GetCurrentFunction().GetModule().GetCStr(),
				expList.Line, "too many local variables or too complex expression")
		}
		return nil
	}

	if len(expList.ExpList) == 0 {
		panic("assert")
	}
	count := len(expList.ExpList) - 1

	// Each expression consume one register
	i := 0
	var maxRegister int
	if endRegister == ExpValueCountAny {
		maxRegister = int(^uint(0) >> 1) // Maximum Constant of Type int
	} else {
		maxRegister = endRegister
	}
	for ; i < count && registerId < maxRegister; i++ {
		err := registerConsumer(registerId + 1)
		if err != nil {
			panic(err)
		}

		r := cgv.GetNextRegisterId()
		NewGuard(func() {}, func() { cgv.ResetRegisterIdGenerator(r) })
		eVarData := newCgExpVarData(registerId, registerId+1)
		expList.ExpList[i].Accept(cgv, unsafe.Pointer(eVarData))
		registerId++
	}

	// No more register
	for ; i < count; i++ {
		r := cgv.GetNextRegisterId()
		NewGuard(func() {}, func() { cgv.ResetRegisterIdGenerator(r) })
		eVarData := newCgExpVarData(0, 0)
		expList.ExpList[i].Accept(cgv, unsafe.Pointer(eVarData))
	}

	// Last expression consume all remain registers
	err := registerConsumer(registerId + 1)
	if err != nil {
		panic(err)
	}
	r := cgv.GetNextRegisterId()
	NewGuard(func() {}, func() { cgv.ResetRegisterIdGenerator(r) })
	eVarData := newCgExpVarData(registerId, endRegister)
	expList.ExpList[len(expList.ExpList)-1].Accept(cgv, unsafe.Pointer(eVarData))
}

// Prepare function data when enter each lexical function
func (cgv *codeGenerateVisitor) EnterFunction() {
	function := newGenerateFunction()
	parent := cgv.currentFunction
	function.Parent = parent
	cgv.currentFunction = function

	// New function is default on GCGen2, so barrier it
	cgv.currentFunction.Function_ = cgv.state.NewFunction()
	if CheckBarrier(cgv.currentFunction.Function_) {
		cgv.state.GetGC().SetBarrier(cgv.currentFunction.Function_)
	}

	if parent != nil {
		index := parent.Function_.AddChildFunction(function.Function_)
		function.FuncIndex = index
		function.Function_.SetSuperior(parent.Function_)
		function.Function_.SetModuleName(parent.Function_.GetModule())
	}
}

// Clean up when leave lexical function
func (cgv *codeGenerateVisitor) LeaveFunction() {
	cgv.deleteCurrentFunction()
}

// Prepare some data when enter each lexical block
func (cgv *codeGenerateVisitor) EnterBlock() {
	block := newGenerateBlock()
	block.Parent = cgv.currentFunction.CurrentBlock
	block.RegisterStartId = cgv.currentFunction.RegisterId
	cgv.currentFunction.CurrentBlock = block
}

// Clean up when leave lexical block
func (cgv *codeGenerateVisitor) LeaveBlock() {
	block := cgv.currentFunction.CurrentBlock

	// Add all variables in block to the function local variable list
	function := cgv.currentFunction.Function_
	endPc := function.OpCodeSize()
	for k, v := range block.Names {
		function.AddLocalVar(k, v.RegisterId, v.BeginPc, endPc)
	}

	// add one instruction to close block
	instruction := ABCode(OpTypeFillNil, block.RegisterStartId, cgv.currentFunction.RegisterId)
	function.AddInstruction(instruction, 0)

	cgv.currentFunction.CurrentBlock = block.Parent
	cgv.currentFunction.RegisterId = block.RegisterStartId
	block = nil
}

// Prepare data for loop AST
func (cgv *codeGenerateVisitor) EnterLoop(loopAst SyntaxTree) {
	// Start instruction index of loop
	startIndex := cgv.GetCurrentFunction().OpCodeSize()
	block := cgv.currentFunction.CurrentBlock
	block.CurrentLoop.LoopAst = loopAst
	block.CurrentLoop.StartIndex = startIndex
}

// Complete loop AST in current block
func (cgv *codeGenerateVisitor) LeaveLoop() {
	function := cgv.GetCurrentFunction()
	// Instruction index after loop
	endIndex := function.OpCodeSize()

	loop := &cgv.currentFunction.CurrentBlock.CurrentLoop
	if loop.LoopAst != nil {
		loopJumps := &cgv.currentFunction.LoopJumps
		for e := loopJumps.Front(); e != nil; {
			info := e.Value.(*loopJumpInfo)
			if info.LoopAst == loop.LoopAst {
				// Calculate diff between current index with index of destination
				diff := 0
				if info.JumpType == jumpHead {
					diff = loop.StartIndex - info.InstructionIndex
				} else if info.JumpType == jumpTail {
					diff = endIndex - info.InstructionIndex
				}

				// Get instruction and refill its jump diff
				i := function.GetMutableInstruction(info.InstructionIndex)
				i.RefillsBx(diff)

				// Remove it from loopJumps when it refilled
				nextE := e.Next()
				loopJumps.Remove(e)
				e = nextE
			} else {
				e = e.Next()
			}
		}
	}
}

// Add one LoopJumpInfo, the instruction will be refilled
// when the loop AST complete
func (cgv *codeGenerateVisitor) AddLoopJumpInfo(loopAst SyntaxTree, instructionIndex, jumpType int) {
	cgv.currentFunction.LoopJumps.PushBack(newLoopJumpInfo(loopAst, jumpType, instructionIndex))
}

// Insert name into current local scope, replace its info when existed
func (cgv *codeGenerateVisitor) InsertName(name *String, registerId int) {
	if cgv.currentFunction == nil || cgv.currentFunction.CurrentBlock == nil {
		panic("assert")
	}

	function := cgv.currentFunction.Function_
	block := cgv.currentFunction.CurrentBlock
	beginPc := function.OpCodeSize()

	if info, ok := block.Names[name]; ok {
		// Add the same name variable to the function local variable list
		endPc := function.OpCodeSize()
		function.AddLocalVar(name, info.RegisterId, info.BeginPc, endPc)

		// New variable replace the old one
		info = *newLocalNameInfo(registerId, beginPc)
	} else {
		// Variable not existed, then insert into
		local := *newLocalNameInfo(registerId, beginPc)
		block.Names[name] = local
	}
}

// Search name in current lexical function
func (cgv *codeGenerateVisitor) SearchLocalName(name *String) *localNameInfo {
	return cgv.SearchFunctionLocalName(cgv.currentFunction, name)
}

// Search name in lexical function
func (cgv *codeGenerateVisitor) SearchFunctionLocalName(function *generateFunction, name *String) *localNameInfo {
	block := function.CurrentBlock
	for block != nil {
		if info, ok := block.Names[name]; ok {
			return &info
		} else {
			block = block.Parent
		}
	}

	return nil
}

// Prepare upvalue info when the name upvalue info not existed, and
// return upvalue index, otherwise just return upvalue index
// the name must reference a upvalue, otherwise will assert fail
func (cgv *codeGenerateVisitor) PrepareUpvalue(name *String) (int, error) {
	// If the upvalue info existed, then return the index of the upvalue
	function := cgv.GetCurrentFunction()
	index := function.SearchUpvalue(name)
	if index >= 0 {
		return index, nil
	}

	// Search start form parent
	var parents []*generateFunction                       // Used as a stack
	parents = append(parents, cgv.currentFunction.Parent) // push

	registerIndex := -1
	parentLocal := false
	for len(parents) != 0 {
		current := parents[len(parents)-1]
		if current == nil {
			panic("assert")
		}
		if registerIndex >= 0 {
			// Find it, add it as upvalue to function, and continue backtrack
			index := current.Function_.AddUpvalue(name, parentLocal, registerIndex)
			if index >= maxClosureUpvalueCount {
				return -1, NewCodeGenerateError(current.Function_.GetModule().GetCStr(),
					current.Function_.GetLine(), "too many upvalues in function")
			}
			registerIndex = index
			parentLocal = false
			parents = parents[:len(parents)-2] // pop
		} else {
			// Find name from local names
			nameInfo := cgv.SearchFunctionLocalName(current, name)
			if nameInfo != nil {
				// Find it, get its registerId and start backtrack
				registerIndex = nameInfo.RegisterId
				parentLocal = true
				parents = parents[:len(parents)-2] // pop
			} else {
				// Find it from current function upvalue list
				index := current.Function_.SearchUpvalue(name)
				if index >= 0 {
					// Find it, the name upvalue has been inserted,
					// then get the upvalue index, and start backtrack
					registerIndex = index
					parentLocal = false
					parents = parents[:len(parents)-2] // pop
				} else {
					// Not find it, continue to search its parent
					parents = append(parents, current.Parent)
				}
			}
		}
	}

	// Add it as upvalue to current function
	if registerIndex < 0 {
		panic("assert")
	}
	index = function.AddUpvalue(name, parentLocal, registerIndex)
	if index >= maxClosureUpvalueCount {
		return -1, NewCodeGenerateError(function.GetModule().GetCStr(),
			function.GetLine(), "too many upvalues in function")
	}
	return index, nil
}

// Get current function data
func (cgv *codeGenerateVisitor) GetCurrentFunction() *Function {
	return cgv.currentFunction.Function_
}

// Generate one register id from current function
func (cgv *codeGenerateVisitor) GenerateRegisterId() (int, error) {
	id := cgv.currentFunction.RegisterId
	cgv.currentFunction.RegisterId++
	if cgv.IsRegisterCountOverflow() {
		return -1, NewCodeGenerateError(cgv.GetCurrentFunction().GetModule().GetCStr(),
			cgv.GetCurrentFunction().GetLine(), "too many local variables in function")
	}
	return id, nil
}

// Get next register id, do not change register generator
func (cgv *codeGenerateVisitor) GetNextRegisterId() int {
	return cgv.currentFunction.RegisterId
}

// Reset register id generator, then next GenerateRegisterId
// use the new id generator to generate register id
func (cgv *codeGenerateVisitor) ResetRegisterIdGenerator(generator int) {
	cgv.currentFunction.RegisterId = generator
}

// Is register count overflow
func (cgv *codeGenerateVisitor) IsRegisterCountOverflow() bool {
	if cgv.currentFunction.RegisterId > cgv.currentFunction.RegisterMax {
		cgv.currentFunction.RegisterMax = cgv.currentFunction.RegisterId
	}
	return cgv.currentFunction.RegisterMax > maxFunctionRegisterCount
}

func (cgv *codeGenerateVisitor) deleteCurrentFunction() {
	function := cgv.currentFunction

	// Delete all blocks in function
	for function.CurrentBlock != nil {
		block := function.CurrentBlock
		function.CurrentBlock = block.Parent
		block = nil
	}

	cgv.currentFunction = function.Parent
	function = nil
}

func (cgv *codeGenerateVisitor) fillRemainRegisterNil(registerId, endRegister, line int) {
	// Fill nil into all remain registers
	// when end_register != EXP_VALUE_COUNT_ANY
	function := cgv.GetCurrentFunction()
	if endRegister != ExpValueCountAny {
		for registerId < endRegister {
			instruction := ACode(OpTypeLoadNil, registerId)
			function.AddInstruction(instruction, line)
		}
	}
}

func (cgv *codeGenerateVisitor) ifStatementGenerateCode(stmtType interface{}) {
	switch ifStmt := stmtType.(type) {
	case *IfStatement:
		function := cgv.GetCurrentFunction()
		jmpEndIndex := 0
		{
			r := cgv.GetNextRegisterId()
			NewGuard(func() {}, func() { cgv.ResetRegisterIdGenerator(r) })
			registerId, err := cgv.GenerateRegisterId()
			if err != nil {
				panic(err)
			}
			eVarData := newCgExpVarData(registerId, registerId+1)
			ifStmt.Exp.Accept(cgv, unsafe.Pointer(eVarData))

			instruction := AsBxCode(OpTypeJmpFalse, registerId, 0)
			jmpIndex := function.AddInstruction(instruction, ifStmt.Line)

			{
				// True branch block generate code
				NewGuard(func() { cgv.EnterBlock() }, func() { cgv.LeaveBlock() })
				ifStmt.TrueBranch.Accept(cgv, nil)
			}

			// jmp to the of if-elseif-else statement after execute block
			instruction = AsBxCode(OpTypeJmp, 0, 0)
			jmpEndIndex = function.AddInstruction(instruction, ifStmt.BlockEndLine)

			// Refill OpType JmpFalse instruction
			index := function.OpCodeSize()
			function.GetMutableInstruction(jmpIndex).RefillsBx(index - jmpIndex)
		}

		if ifStmt.FalseBranch != nil {
			ifStmt.FalseBranch.Accept(cgv, nil)
		}

		// Refill OpType Jmp instruction
		endIndex := function.OpCodeSize()
		function.GetMutableInstruction(jmpEndIndex).RefillsBx(endIndex - jmpEndIndex)
	case *ElseIfStatement:
		function := cgv.GetCurrentFunction()
		jmpEndIndex := 0
		{
			r := cgv.GetNextRegisterId()
			NewGuard(func() {}, func() { cgv.ResetRegisterIdGenerator(r) })
			registerId, err := cgv.GenerateRegisterId()
			if err != nil {
				panic(err)
			}
			eVarData := newCgExpVarData(registerId, registerId+1)
			ifStmt.Exp.Accept(cgv, unsafe.Pointer(eVarData))

			instruction := AsBxCode(OpTypeJmpFalse, registerId, 0)
			jmpIndex := function.AddInstruction(instruction, ifStmt.Line)

			{
				// True branch block generate code
				NewGuard(func() { cgv.EnterBlock() }, func() { cgv.LeaveBlock() })
				ifStmt.TrueBranch.Accept(cgv, nil)
			}

			// jmp to the of if-elseif-else statement after execute block
			instruction = AsBxCode(OpTypeJmp, 0, 0)
			jmpEndIndex = function.AddInstruction(instruction, ifStmt.BlockEndLine)

			// Refill OpType JmpFalse instruction
			index := function.OpCodeSize()
			function.GetMutableInstruction(jmpIndex).RefillsBx(index - jmpIndex)
		}

		if ifStmt.FalseBranch != nil {
			ifStmt.FalseBranch.Accept(cgv, nil)
		}

		// Refill OpType Jmp instruction
		endIndex := function.OpCodeSize()
		function.GetMutableInstruction(jmpEndIndex).RefillsBx(endIndex - jmpEndIndex)
	}

}

func (cgv *codeGenerateVisitor) setTableFieldValue(tableFieldType interface{}, tableRegister, keyRegister, line int) {
	switch field := tableFieldType.(type) {
	case *TableIndexField:
		// Load value
		valueRegister, err := cgv.GenerateRegisterId()
		if err != nil {
			panic(err)
		}
		eVarData := newCgExpVarData(valueRegister, valueRegister+1)
		field.Value.Accept(cgv, unsafe.Pointer(eVarData))

		// Set table field
		instruction := ABCCode(OpTypeSetTable, tableRegister, keyRegister, valueRegister)
		cgv.GetCurrentFunction().AddInstruction(instruction, line)
	case *TableNameField:
		// Load value
		valueRegister, err := cgv.GenerateRegisterId()
		if err != nil {
			panic(err)
		}
		eVarData := newCgExpVarData(valueRegister, valueRegister+1)
		field.Value.Accept(cgv, unsafe.Pointer(eVarData))

		// Set table field
		instruction := ABCCode(OpTypeSetTable, tableRegister, keyRegister, valueRegister)
		cgv.GetCurrentFunction().AddInstruction(instruction, line)
	case *TableArrayField:
		// Load value
		valueRegister, err := cgv.GenerateRegisterId()
		if err != nil {
			panic(err)
		}
		eVarData := newCgExpVarData(valueRegister, valueRegister+1)
		field.Value.Accept(cgv, unsafe.Pointer(eVarData))

		// Set table field
		instruction := ABCCode(OpTypeSetTable, tableRegister, keyRegister, valueRegister)
		cgv.GetCurrentFunction().AddInstruction(instruction, line)
	}
}

func (cgv *codeGenerateVisitor) accessTableField(tableAccessorType interface{}, data unsafe.Pointer, line int, loadKeyFunc interface{}) {
	loadKey := loadKeyFunc.(func(int))
	switch accessor := tableAccessorType.(type) {
	case *IndexAccessor:
		eVarData := (*cgExpVarData)(data)
		registerId := eVarData.StartRegister
		endRegister := eVarData.EndRegister
		function := cgv.GetCurrentFunction()

		tableRegister := 0
		keyRegister := 0
		valueRegister := 0
		var opType int
		if accessor.Semantic == SemanticOpRead {
			// No more register, do nothing
			if endRegister != ExpValueCountAny && registerId >= endRegister {
				return
			}

			if endRegister != ExpValueCountAny && registerId+1 < endRegister {
				keyRegister = registerId + 1
			} else {
				keyRegister = cgv.GetNextRegisterId()
			}
			tableRegister = registerId
			valueRegister = registerId
			opType = OpTypeGetTable
		} else {
			if accessor.Semantic != SemanticOpWrite {
				panic("assert")
			}
			if registerId+1 == endRegister {
				panic("assert")
			}

			tableRegister = cgv.GetNextRegisterId()
			keyRegister = cgv.GetNextRegisterId()
			valueRegister = registerId
			opType = OpTypeSetTable
		}

		// Load table
		tableExpVarData := newCgExpVarData(tableRegister, tableRegister+1)
		accessor.Table.Accept(cgv, unsafe.Pointer(tableExpVarData))

		// Load key
		loadKey(keyRegister)

		// Set/Get table value by key
		instruction := ABCCode(opType, tableRegister, keyRegister, valueRegister)
		function.AddInstruction(instruction, line)

		if accessor.Semantic == SemanticOpRead {
			cgv.fillRemainRegisterNil(registerId+1, endRegister, line)
		}
	case *MemberAccessor:
		eVarData := (*cgExpVarData)(data)
		registerId := eVarData.StartRegister
		endRegister := eVarData.EndRegister
		function := cgv.GetCurrentFunction()

		tableRegister := 0
		keyRegister := 0
		valueRegister := 0
		var opType int
		if accessor.Semantic == SemanticOpRead {
			// No more register, do nothing
			if endRegister != ExpValueCountAny && registerId >= endRegister {
				return
			}

			if endRegister != ExpValueCountAny && registerId+1 < endRegister {
				keyRegister = registerId + 1
			} else {
				keyRegister = cgv.GetNextRegisterId()
			}
			tableRegister = registerId
			valueRegister = registerId
			opType = OpTypeGetTable
		} else {
			if accessor.Semantic != SemanticOpWrite {
				panic("assert")
			}
			if registerId+1 == endRegister {
				panic("assert")
			}

			tableRegister = cgv.GetNextRegisterId()
			keyRegister = cgv.GetNextRegisterId()
			valueRegister = registerId
			opType = OpTypeSetTable
		}

		// Load table
		tableExpVarData := newCgExpVarData(tableRegister, tableRegister+1)
		accessor.Table.Accept(cgv, unsafe.Pointer(tableExpVarData))

		// Load key
		loadKey(keyRegister)

		// Set/Get table value by key
		instruction := ABCCode(opType, tableRegister, keyRegister, valueRegister)
		function.AddInstruction(instruction, line)

		if accessor.Semantic == SemanticOpRead {
			cgv.fillRemainRegisterNil(registerId+1, endRegister, line)
		}
	}
}

func (cgv *codeGenerateVisitor) functionCall(funcCallType interface{}, data unsafe.Pointer, callerArgAdjuster interface{}) {
	adjustCallerArg := callerArgAdjuster.(func(int) int)
	switch funcCall := funcCallType.(type) {
	case *NormalFuncCall:
		r := cgv.GetNextRegisterId()
		NewGuard(func() {}, func() { cgv.ResetRegisterIdGenerator(r) })
		eVarData := (*cgExpVarData)(data)
		var startRegister, endRegister int
		if eVarData != nil {
			startRegister = eVarData.StartRegister
			endRegister = eVarData.EndRegister
		} else {
			startRegister = 0
			endRegister = 0
		}

		var err error
		// Generate code to get caller
		callerRegister := 0
		if endRegister == ExpValueCountAny {
			callerArgAdjuster = startRegister
		} else {
			callerArgAdjuster, err = cgv.GenerateRegisterId()
			if err != nil {
				panic(err)
			}
		}

		{
			r := cgv.GetNextRegisterId()
			NewGuard(func() {}, func() { cgv.ResetRegisterIdGenerator(r) })
			callerData := newCgExpVarData(callerRegister, callerRegister+1)
			funcCall.Caller.Accept(cgv, unsafe.Pointer(callerData))
		}

		// Adjust caller, and also adjust args, return how many args adjusted
		adjustArgs := adjustCallerArg(callerRegister)

		var argData cgFuncCallArgsData
		funcCall.Args.Accept(cgv, unsafe.Pointer(&argData))

		// Calculate total args
		var totalArgs int
		if argData.ArgValueCount == ExpValueCountAny {
			totalArgs = ExpValueCountAny
		} else {
			totalArgs = argData.ArgValueCount + adjustArgs
		}

		// Calculate expect results count of function call
		var results int
		if endRegister == ExpValueCountAny {
			results = ExpValueCountAny
		} else {
			results = endRegister - startRegister
		}

		// Generate call instruction
		function := cgv.GetCurrentFunction()
		instruction := ABCCode(OpTypeCall, callerRegister, totalArgs+1, results+1)
		function.AddInstruction(instruction, funcCall.Line)

		// Copy results of function call to dst registers
		// if end_register == EXP_VALUE_COUNT_ANY, then do not
		// copy results to dst registers, just keep it
		if endRegister != ExpValueCountAny {
			src := callerRegister
			for dst := startRegister; dst < endRegister; dst++ {
				i := ABCode(OpTypeMove, dst, src)
				function.AddInstruction(i, funcCall.Line)
				src++
			}
		}
	case *MemberFuncCall:
		r := cgv.GetNextRegisterId()
		NewGuard(func() {}, func() { cgv.ResetRegisterIdGenerator(r) })
		eVarData := (*cgExpVarData)(data)
		var startRegister, endRegister int
		if eVarData != nil {
			startRegister = eVarData.StartRegister
			endRegister = eVarData.EndRegister
		} else {
			startRegister = 0
			endRegister = 0
		}

		var err error
		// Generate code to get caller
		callerRegister := 0
		if endRegister == ExpValueCountAny {
			callerArgAdjuster = startRegister
		} else {
			callerArgAdjuster, err = cgv.GenerateRegisterId()
			if err != nil {
				panic(err)
			}
		}

		{
			r := cgv.GetNextRegisterId()
			NewGuard(func() {}, func() { cgv.ResetRegisterIdGenerator(r) })
			callerData := newCgExpVarData(callerRegister, callerRegister+1)
			funcCall.Caller.Accept(cgv, unsafe.Pointer(callerData))
		}

		// Adjust caller, and also adjust args, return how many args adjusted
		adjustArgs := adjustCallerArg(callerRegister)

		var argData cgFuncCallArgsData
		funcCall.Args.Accept(cgv, unsafe.Pointer(&argData))

		// Calculate total args
		var totalArgs int
		if argData.ArgValueCount == ExpValueCountAny {
			totalArgs = ExpValueCountAny
		} else {
			totalArgs = argData.ArgValueCount + adjustArgs
		}

		// Calculate expect results count of function call
		var results int
		if endRegister == ExpValueCountAny {
			results = ExpValueCountAny
		} else {
			results = endRegister - startRegister
		}

		// Generate call instruction
		function := cgv.GetCurrentFunction()
		instruction := ABCCode(OpTypeCall, callerRegister, totalArgs+1, results+1)
		function.AddInstruction(instruction, funcCall.Line)

		// Copy results of function call to dst registers
		// if end_register == EXP_VALUE_COUNT_ANY, then do not
		// copy results to dst registers, just keep it
		if endRegister != ExpValueCountAny {
			src := callerRegister
			for dst := startRegister; dst < endRegister; dst++ {
				i := ABCode(OpTypeMove, dst, src)
				function.AddInstruction(i, funcCall.Line)
				src++
			}
		}
	}
}

// For NameList AST
type cgNameListData struct {
	NeedInit bool // NameList need init itself or not
}

func newCgNameListData(needInit bool) *cgNameListData {
	return &cgNameListData{needInit}
}

// For ExpList AST
type cgExpListData struct {
	// ExpList need fill into range [StartRegister, EndRegister)
	// when EndRegister != ExpValueCountAny, otherwise fill any
	// count registers begin with StartRegister
	StartRegister int
	EndRegister   int
}

func newCgExpListData(startRegister, endRegister int) *cgExpListData {
	return &cgExpListData{startRegister, endRegister}
}

// For expression and variable
type cgExpVarData struct {
	// Need fill into range [StartRegister, EndRegister)
	// when EndRegister != ExpValueCountAny, otherwise fill any
	// count registers begin with StartRegister
	StartRegister int
	EndRegister   int
}

func newCgExpVarData(startRegister, endRegister int) *cgExpVarData {
	return &cgExpVarData{startRegister, endRegister}
}

// For VarList AST
type cgVarListData struct {
	// VarList get results from range [StartRegister, EndRegister)
	StartRegister int
	EndRegister   int
}

func newCgVarListData(startRegister, endRegister int) *cgVarListData {
	return &cgVarListData{startRegister, endRegister}
}

// For table field
type cgTableFieldData struct {
	TableRegister int  // Table register
	ArrayIndex    uint // Array part index, start from 1
}

func newCgTableFieldData(tableRegister int) *cgTableFieldData {
	return &cgTableFieldData{tableRegister, 1}
}

// For FuncCallArgs AST
type cgFuncCallArgsData struct {
	ArgValueCount int
}

func newCgFuncCallArgsData() *cgFuncCallArgsData {
	return &cgFuncCallArgsData{0}
}

// For FunctionName AST
type cgFunctionNameData struct {
	FuncRegister int
}

func newCgFunctionNameData(funcRegister int) *cgFunctionNameData {
	return &cgFunctionNameData{funcRegister}
}

func CodeGenerate(root SyntaxTree, state *State) {
	if root == nil || state == nil {
		panic("assert")
	}

	codeGenerator := newCodeGenerateVisitor(state)
	root.Accept(codeGenerator, nil)
}
