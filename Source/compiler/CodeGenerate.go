package compiler

import (
	"InterpreterVM/Source/datatype"
	"InterpreterVM/Source/vm"
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
	Names map[*datatype.String]localNameInfo
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
	CurrentBlock *generateBlock     // Current block
	Function_    *datatype.Function // Current function for code generate
	FuncIndex    int                // Index of current function in parent
	RegisterId   int                // Register id generator
	RegisterMax  int                // Max register count used in current function
	LoopJumps    list.List          // To be filled loop jump info, and its element.value is *loopJumpInfo
}

func newGenerateFunction() *generateFunction {
	return &generateFunction{}
}

type codeGenerateVisitor struct {
	state           *vm.State
	currentFunction *generateFunction // Current code generating function
}

func newCodeGenerateVisitor(state *vm.State) *codeGenerateVisitor {
	return &codeGenerateVisitor{state: state}
}

func (cgv *codeGenerateVisitor) VisitChunk(chunk *Chunk, data unsafe.Pointer) {

}
func (cgv *codeGenerateVisitor) VisitBlock(block *Block, data unsafe.Pointer) {

}
func (cgv *codeGenerateVisitor) VisitReturnStatement(retStmt *ReturnStatement, data unsafe.Pointer) {

}
func (cgv *codeGenerateVisitor) VisitBreakStatement(breakStmt *BreakStatement, data unsafe.Pointer) {

}
func (cgv *codeGenerateVisitor) VisitDoStatement(doStmt *DoStatement, data unsafe.Pointer) {

}
func (cgv *codeGenerateVisitor) VisitWhileStatement(whileStmt *WhileStatement, data unsafe.Pointer) {

}
func (cgv *codeGenerateVisitor) VisitRepeatStatement(repeatStmt *RepeatStatement, data unsafe.Pointer) {

}
func (cgv *codeGenerateVisitor) VisitIfStatement(ifStmt *IfStatement, data unsafe.Pointer) {

}
func (cgv *codeGenerateVisitor) VisitElseIfStatement(elseifStmt *ElseIfStatement, data unsafe.Pointer) {

}
func (cgv *codeGenerateVisitor) VisitElseStatement(elseStmt *ElseStatement, data unsafe.Pointer) {

}
func (cgv *codeGenerateVisitor) VisitNumericForStatement(numFor *NumericForStatement, data unsafe.Pointer) {

}
func (cgv *codeGenerateVisitor) VisitGenericForStatement(genFor *GenericForStatement, data unsafe.Pointer) {

}
func (cgv *codeGenerateVisitor) VisitFunctionStatement(funcStmt *FunctionStatement, data unsafe.Pointer) {

}
func (cgv *codeGenerateVisitor) VisitFunctionName(funcName *FunctionName, data unsafe.Pointer) {

}
func (cgv *codeGenerateVisitor) VisitLocalFunctionStatement(lFuncStmt *LocalFunctionStatement, data unsafe.Pointer) {

}
func (cgv *codeGenerateVisitor) VisitLocalNameListStatement(lNameListStmt *LocalNameListStatement, data unsafe.Pointer) {

}
func (cgv *codeGenerateVisitor) VisitAssignmentStatement(assignStmt *AssignmentStatement, data unsafe.Pointer) {

}
func (cgv *codeGenerateVisitor) VisitVarList(varList *VarList, data unsafe.Pointer) {

}
func (cgv *codeGenerateVisitor) VisitTerminator(term *Terminator, data unsafe.Pointer) {

}
func (cgv *codeGenerateVisitor) VisitBinaryExpression(binaryExp *BinaryExpression, data unsafe.Pointer) {

}
func (cgv *codeGenerateVisitor) VisitUnaryExpression(unaryExp *UnaryExpression, data unsafe.Pointer) {

}
func (cgv *codeGenerateVisitor) VisitFunctionBody(funcBody *FunctionBody, data unsafe.Pointer) {

}
func (cgv *codeGenerateVisitor) VisitParamList(parList *ParamList, data unsafe.Pointer) {

}
func (cgv *codeGenerateVisitor) VisitNameList(nameList *NameList, data unsafe.Pointer) {

}
func (cgv *codeGenerateVisitor) VisitTableDefine(tableDef *TableDefine, data unsafe.Pointer) {

}
func (cgv *codeGenerateVisitor) VisitTableIndexField(tableIField *TableIndexField, data unsafe.Pointer) {

}
func (cgv *codeGenerateVisitor) VisitTableNameField(tableNField *TableNameField, data unsafe.Pointer) {

}
func (cgv *codeGenerateVisitor) VisitTableArrayField(tableAField *TableArrayField, data unsafe.Pointer) {

}
func (cgv *codeGenerateVisitor) VisitIndexAccessor(iAccessor *IndexAccessor, data unsafe.Pointer) {

}
func (cgv *codeGenerateVisitor) VisitMemberAccessor(mAccessor *MemberAccessor, data unsafe.Pointer) {

}
func (cgv *codeGenerateVisitor) VisitNormalFuncCall(nFuncCall *NormalFuncCall, data unsafe.Pointer) {

}
func (cgv *codeGenerateVisitor) VisitMemberFuncCall(mFuncCall *MemberFuncCall, data unsafe.Pointer) {

}
func (cgv *codeGenerateVisitor) VisitFuncCallArgs(callArgs *FuncCallArgs, data unsafe.Pointer) {

}
func (cgv *codeGenerateVisitor) VisitExpressionList(expList *ExpressionList, data unsafe.Pointer) {

}

// Prepare function data when enter each lexical function
func (cgv *codeGenerateVisitor) EnterFunction() {
	function := newGenerateFunction()
	parent := cgv.currentFunction
	function.Parent = parent
	cgv.currentFunction = function

	// New function is default on GCGen2, so barrier it
	cgv.currentFunction.Function_ = cgv.state.NewFunction()
	if datatype.CheckBarrier(cgv.currentFunction.Function_) {
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
	instruction := vm.ABCode(vm.OpTypeFillNil, block.RegisterStartId, cgv.currentFunction.RegisterId)
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
func (cgv *codeGenerateVisitor) InsertName(name *datatype.String, registerId int) {
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
func (cgv *codeGenerateVisitor) SearchLocalName(name *datatype.String) *localNameInfo {
	return cgv.SearchFunctionLocalName(cgv.currentFunction, name)
}

// Search name in lexical function
func (cgv *codeGenerateVisitor) SearchFunctionLocalName(function *generateFunction, name *datatype.String) *localNameInfo {
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
func (cgv *codeGenerateVisitor) PrepareUpvalue(name *datatype.String) (int, error) {
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
				return -1, vm.NewCodeGenerateError(current.Function_.GetModule().GetCStr(),
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
		return -1, vm.NewCodeGenerateError(function.GetModule().GetCStr(),
			function.GetLine(), "too many upvalues in function")
	}
	return index, nil
}

// Get current function data
func (cgv *codeGenerateVisitor) GetCurrentFunction() *datatype.Function {
	return cgv.currentFunction.Function_
}

// Generate one register id from current function
func (cgv *codeGenerateVisitor) GenerateRegisterId() (int, error) {
	id := cgv.currentFunction.RegisterId
	cgv.currentFunction.RegisterId++
	if cgv.IsRegisterCountOverflow() {
		return -1, vm.NewCodeGenerateError(cgv.GetCurrentFunction().GetModule().GetCStr(),
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
	if endRegister != datatype.ExpValueCountAny {
		for registerId < endRegister {
			instruction := vm.ACode(vm.OpTypeLoadNil, registerId)
			function.AddInstruction(instruction, line)
		}
	}
}

func (cgv *codeGenerateVisitor) ifStatementGenerateCode(ifStmt *interface{}) {

}

func (cgv *codeGenerateVisitor) setTableFieldValue(field *interface{}, tableRegister, keyRegister, line int) {

}

func (cgv *codeGenerateVisitor) accessTableField(accessor *interface{}, data unsafe.Pointer, line int, loadKey *interface{}) {

}

func (cgv *codeGenerateVisitor) functionCall(funcCall *interface{}, data unsafe.Pointer, adjustCallerArg *interface{}) {

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

func CodeGenerate(root SyntaxTree, state *vm.State) {
	if root == nil || state == nil {
		panic("assert")
	}

	codeGenerator := newCodeGenerateVisitor(state)
	root.Accept(codeGenerator, nil)
}
