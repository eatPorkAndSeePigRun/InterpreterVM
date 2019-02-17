package compiler

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

func newLocalNameInfo(registerId, beginPc int) localNameInfo {
	return localNameInfo{registerId, beginPc}
}

// Loop AST info data in GenerateBlock
type loopInfo struct {
	LoopAst    *SyntaxTree // Loop AST
	StartIndex int         // Start instruction index
}

func newLoopInfo() loopInfo {
	return loopInfo{}
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

func newGenerateBlock() generateBlock {
	return generateBlock{}
}

// Jump info for loop AST
type loopJumpInfo struct {
	LoopAst          *SyntaxTree // Owner loop AST
	JumpType         int         // Jump to AST head or tail
	InstructionIndex int         // Instruction need to be filled
}

const (
	jumpHead = iota
	jumpTail
)

func newLoopJumpInfo(loopAst *SyntaxTree, jumpType int, instructionIndex int) loopJumpInfo {
	return loopJumpInfo{loopAst, jumpType, instructionIndex}
}

// Lexical function struct for code generator
type generateFunction struct {
	Parent       *generateFunction
	CurrentBlock *generateBlock // Current block
	Function_    *Function      // Current function for code generate
	FuncIndex    int            // Index of current function in parent
	RegisterId   int            // Register id generator
	RegisterMax  int            // Max register count used in current function
	LoopJumps    list.List      // To be filled loop jump info
}

func newGenerateFunction() generateFunction {
	return generateFunction{}
}

type codeGenerateVisitor struct {
	state           *State
	currentFunction *generateFunction // Current code generating function
}

func newCodeGenerateVisitor(state *State) codeGenerateVisitor {
	return codeGenerateVisitor{state: state}
}

// TODO

func (cgv codeGenerateVisitor) VisitChunk(*Chunk, unsafe.Pointer) {

}
func (cgv codeGenerateVisitor) VisitBlock(*Block, unsafe.Pointer) {

}
func (cgv codeGenerateVisitor) VisitReturnStatement(*ReturnStatement, unsafe.Pointer) {

}
func (cgv codeGenerateVisitor) VisitBreakStatement(*BreakStatement, unsafe.Pointer) {

}
func (cgv codeGenerateVisitor) VisitDoStatement(*DoStatement, unsafe.Pointer) {

}
func (cgv codeGenerateVisitor) VisitWhileStatement(*WhileStatement, unsafe.Pointer) {

}
func (cgv codeGenerateVisitor) VisitRepeatStatement(*RepeatStatement, unsafe.Pointer) {

}
func (cgv codeGenerateVisitor) VisitIfStatement(*IfStatement, unsafe.Pointer) {

}
func (cgv codeGenerateVisitor) VisitElseIfStatement(*ElseStatement, unsafe.Pointer) {

}
func (cgv codeGenerateVisitor) VisitElseStatement(*ElseStatement, unsafe.Pointer) {

}
func (cgv codeGenerateVisitor) VisitNumericForStatement(*NumericForStatement, unsafe.Pointer) {

}
func (cgv codeGenerateVisitor) VisitGenericForStatement(*GenericForStatement, unsafe.Pointer) {

}
func (cgv codeGenerateVisitor) VisitFunctionStatement(*FunctionStatement, unsafe.Pointer) {

}
func (cgv codeGenerateVisitor) VisitFunctionName(*FunctionName, unsafe.Pointer) {

}
func (cgv codeGenerateVisitor) VisitLocalFunctionStatement(*LocalFunctionStatement, unsafe.Pointer) {

}
func (cgv codeGenerateVisitor) VisitLocalNameListStatement(*LocalNameListStatement, unsafe.Pointer) {

}
func (cgv codeGenerateVisitor) VisitAssignmentStatement(*AssignmentStatement, unsafe.Pointer) {

}
func (cgv codeGenerateVisitor) VisitVarList(*VarList, unsafe.Pointer) {

}
func (cgv codeGenerateVisitor) VisitTerminator(*Terminator, unsafe.Pointer) {

}
func (cgv codeGenerateVisitor) VisitBinaryExpression(*BinaryExpression, unsafe.Pointer) {

}
func (cgv codeGenerateVisitor) VisitUnaryExpression(*UnaryExpression, unsafe.Pointer) {

}
func (cgv codeGenerateVisitor) VisitFunctionBody(*FunctionBody, unsafe.Pointer) {

}
func (cgv codeGenerateVisitor) VisitParamList(*ParamList, unsafe.Pointer) {

}
func (cgv codeGenerateVisitor) VisitNameList(*NameList, unsafe.Pointer) {

}
func (cgv codeGenerateVisitor) VisitTableDefine(*TableDefine, unsafe.Pointer) {

}
func (cgv codeGenerateVisitor) VisitTableIndexField(*TableIndexField, unsafe.Pointer) {

}
func (cgv codeGenerateVisitor) VisitTableNameField(*TableNameField, unsafe.Pointer) {

}
func (cgv codeGenerateVisitor) VisitTableArrayField(*TableArrayField, unsafe.Pointer) {

}
func (cgv codeGenerateVisitor) VisitIndexAccessor(*IndexAccessor, unsafe.Pointer) {

}
func (cgv codeGenerateVisitor) VisitMemberAccessor(*MemberAccessor, unsafe.Pointer) {

}
func (cgv codeGenerateVisitor) VisitNormalFuncCall(*NormalFuncCall, unsafe.Pointer) {

}
func (cgv codeGenerateVisitor) VisitMemberFuncCall(*MemberFuncCall, unsafe.Pointer) {

}
func (cgv codeGenerateVisitor) VisitFuncCallArgs(*FuncCallArgs, unsafe.Pointer) {

}
func (cgv codeGenerateVisitor) VisitExpressionList(*ExpressionList, unsafe.Pointer) {

}

// For NameList AST
type cgNameListData struct {
	NeedInit bool // NameList need init itself or not
}

func newCgNameListData(needInit bool) cgNameListData {
	return cgNameListData{needInit}
}

// For ExpList AST
type cgExpListData struct {
	// ExpList need fill into range [StartRegister, EndRegister)
	// when EndRegister != ExpValueCountAny, otherwise fill any
	// count registers begin with StartRegister
	StartRegister int
	EndRegister   int
}

func newCgExpListData(startRegister, endRegister int) cgExpListData {
	return cgExpListData{startRegister, endRegister}
}

// For expression and variable
type cgExpVarData struct {
	// Need fill into range [StartRegister, EndRegister)
	// when EndRegister != ExpValueCountAny, otherwise fill any
	// count registers begin with StartRegister
	StartRegister int
	EndRegister   int
}

func newcgExpVarData(startRegister, endRegister int) cgExpVarData {
	return cgExpVarData{startRegister, endRegister}
}

// For VarList AST
type cgVarListData struct {
	// VarList get results from range [StartRegister, EndRegister)
	StartRegister int
	EndRegister   int
}

func newCgVarListData(startRegister, endRegister int) cgVarListData {
	return cgVarListData{startRegister, endRegister}
}

// For table field
type cgTableFieldData struct {
	TableRegister int  // Table register
	ArrayIndex    uint // Array part index, start from 1
}

func newCgTableFieldData(tableRegister int) cgTableFieldData {
	return cgTableFieldData{tableRegister, 1}
}

// For FuncCallArgs AST
type cgFuncCallArgsData struct {
	ArgValueCount int
}

func newCgFuncCallArgsData() cgFuncCallArgsData {
	return cgFuncCallArgsData{0}
}

// For FunctionName AST
type cgFunctionNameData struct {
	FuncRegister int
}

func newCgFunctionNameData(funcRegister int) cgFunctionNameData {
	return cgFunctionNameData{funcRegister}
}

func CodeGenerate(root SyntaxTree, state *State) {
	if root == nil && state == nil {
		panic("assert")
	}

	codeGenerator := newCodeGenerateVisitor(state)
	root.Accept(&codeGenerator, nil)
}
