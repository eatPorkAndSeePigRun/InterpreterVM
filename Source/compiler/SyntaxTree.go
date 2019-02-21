package compiler

import (
	"InterpreterVM/Source/datatype"
	"unsafe"
)

// Expression or variable operation semantic
const (
	SemanticOpNone = iota
	SemanticOpRead
	SemanticOpWrite
)

// Expression or variable lexical scoping
const (
	LexicalScopingUnknown = iota
	LexicalScopingGlobal  // Expression or variable in global table
	LexicalScopingUpvalue // Expression or variable in upvalue
	LexicalScopingLocal   // Expression or variable in current function
)

// AST base class, all AST node derived from this class and
// provide Visitor to Accept itself.
type SyntaxTree interface {
	Accept(v Visitor, data unsafe.Pointer)
}

type Chunk struct {
	Block  SyntaxTree
	Module *datatype.String
}

func NewChunk(block SyntaxTree, module *datatype.String) *Chunk {
	return &Chunk{block, module}
}

func (c *Chunk) Accept(v Visitor, data unsafe.Pointer) {
	v.VisitChunk(c, data)
}

type Block struct {
	Statements []SyntaxTree
	ReturnStmt SyntaxTree
}

func NewBlock() *Block {
	return &Block{}
}

func (b *Block) Accept(v Visitor, data unsafe.Pointer) {
	v.VisitBlock(b, data)
}

type ReturnStatement struct {
	ExpList       SyntaxTree
	Line          int
	ExpValueCount int
}

func NewReturnStatement(line int) *ReturnStatement {
	return &ReturnStatement{Line: line}
}

func (r *ReturnStatement) Accept(v Visitor, data unsafe.Pointer) {
	v.VisitReturnStatement(r, data)
}

type BreakStatement struct {
	Break TokenDetail
	Loop  SyntaxTree // For semantic
}

func NewBreakStatement(b TokenDetail) *BreakStatement {
	return &BreakStatement{Break: b}
}

func (b *BreakStatement) Accept(v Visitor, data unsafe.Pointer) {
	v.VisitBreakStatement(b, data)
}

type DoStatement struct {
	Block SyntaxTree
}

func NewDoStatement(block SyntaxTree) *DoStatement {
	return &DoStatement{Block: block}
}

func (d *DoStatement) Accept(v Visitor, data unsafe.Pointer) {
	v.VisitDoStatement(d, data)
}

type WhileStatement struct {
	Exp   SyntaxTree
	Block SyntaxTree

	FirstLine int
	LastLine  int
}

func NewWhileStatement(exp, block SyntaxTree, firstLine, lastLine int) *WhileStatement {
	return &WhileStatement{exp, block, firstLine, lastLine}
}

func (w *WhileStatement) Accept(v Visitor, data unsafe.Pointer) {
	v.VisitWhileStatement(w, data)
}

type RepeatStatement struct {
	Block SyntaxTree
	Exp   SyntaxTree

	Line int // Line of until
}

func NewRepeatStatement(block, exp SyntaxTree, line int) *RepeatStatement {
	return &RepeatStatement{block, exp, line}
}

func (r *RepeatStatement) Accept(v Visitor, data unsafe.Pointer) {
	v.VisitRepeatStatement(r, data)
}

type IfStatement struct {
	Exp         SyntaxTree
	TrueBranch  SyntaxTree
	FalseBranch SyntaxTree

	Line         int // Line of if
	BlockEndLine int // End line of block
}

func NewIfStatement(exp, trueBranch, falseBranch SyntaxTree, line, blockEndline int) *IfStatement {
	return &IfStatement{exp, trueBranch, falseBranch, line, blockEndline}
}

func (i *IfStatement) Accept(v Visitor, data unsafe.Pointer) {
	v.VisitIfStatement(i, data)
}

type ElseIfStatement struct {
	Exp         SyntaxTree
	TrueBranch  SyntaxTree
	FalseBranch SyntaxTree

	Line         int // Line of elseif
	BlockEndLine int // End line of block
}

func NewElseIfStatement(exp, trueBranch, falseBranch SyntaxTree, line, blockEndLind int) *ElseIfStatement {
	return &ElseIfStatement{exp, trueBranch, falseBranch, line, blockEndLind}
}

func (e *ElseIfStatement) Accept(v Visitor, data unsafe.Pointer) {
	v.VisitElseIfStatement(e, data)
}

type ElseStatement struct {
	Block SyntaxTree
}

func NewElseStatement(block SyntaxTree) *ElseStatement {
	return &ElseStatement{block}
}

func (e *ElseStatement) Accept(v Visitor, data unsafe.Pointer) {
	v.VisitElseStatement(e, data)
}

type NumericForStatement struct {
	Name  TokenDetail
	Exp1  SyntaxTree
	Exp2  SyntaxTree
	Exp3  SyntaxTree
	Block SyntaxTree
}

func NewNumericForStatement(name TokenDetail, exp1, exp2, exp3, block SyntaxTree) *NumericForStatement {
	return &NumericForStatement{name, exp1, exp2, exp3, block}
}

func (n *NumericForStatement) Accept(v Visitor, data unsafe.Pointer) {
	v.VisitNumericForStatement(n, data)
}

type GenericForStatement struct {
	NameList SyntaxTree
	ExpList  SyntaxTree
	Block    SyntaxTree

	Line int
}

func NewGenericForStatement(nameList, expList, block SyntaxTree, line int) *GenericForStatement {
	return &GenericForStatement{nameList, expList, block, line}
}

func (g *GenericForStatement) Accept(v Visitor, data unsafe.Pointer) {
	v.VisitGenericForStatement(g, data)
}

type FunctionStatement struct {
	FuncName SyntaxTree
	FuncBody SyntaxTree
}

func NewFunctionStatement(funcName, funcBody SyntaxTree) *FunctionStatement {
	return &FunctionStatement{funcName, funcBody}
}

func (f *FunctionStatement) Accept(v Visitor, data unsafe.Pointer) {
	v.VisitFunctionStatement(f, data)
}

type FunctionName struct {
	Names      []TokenDetail
	MemberName TokenDetail
	Scoping    int // First token scoping
}

func NewFunctionName() *FunctionName {
	return &FunctionName{Scoping: LexicalScopingUnknown}
}

func (f *FunctionName) Accept(v Visitor, data unsafe.Pointer) {
	v.VisitFunctionName(f, data)
}

type LocalFunctionStatement struct {
	Name     TokenDetail
	FuncBody SyntaxTree
}

func NewLocalFunctionStatement(name TokenDetail, funcBody SyntaxTree) *LocalFunctionStatement {
	return &LocalFunctionStatement{name, funcBody}
}

func (l *LocalFunctionStatement) Accept(v Visitor, data unsafe.Pointer) {
	v.VisitLocalFunctionStatement(l, data)
}

type LocalNameListStatement struct {
	NameList  SyntaxTree
	ExpList   SyntaxTree
	Line      int // Start Line
	NameCount int // For semantic and code generate
}

func NewLocalNameListStatement(nameList, ExpList SyntaxTree, line int) *LocalNameListStatement {
	return &LocalNameListStatement{nameList, ExpList, line, 0}
}

func (l *LocalNameListStatement) Accept(v Visitor, data unsafe.Pointer) {
	v.VisitLocalNameListStatement(l, data)
}

type AssignmentStatement struct {
	VarList  SyntaxTree
	ExpList  SyntaxTree
	Line     int // Start line
	VarCount int // For semantic
}

func NewAssignmentStatement(varList, expList SyntaxTree, line int) *AssignmentStatement {
	return &AssignmentStatement{varList, expList, line, 0}
}

func (a *AssignmentStatement) Accept(v Visitor, data unsafe.Pointer) {
	v.VisitAssignmentStatement(a, data)
}

type VarList struct {
	VarList []SyntaxTree
}

func NewVarList() *VarList {
	return &VarList{}
}

func (vl *VarList) Accept(v Visitor, data unsafe.Pointer) {
	v.VisitVarList(vl, data)
}

type Terminator struct {
	Token    TokenDetail
	Semantic int
	Scoping  int
}

func NewTerminator(token TokenDetail) *Terminator {
	return &Terminator{token, SemanticOpNone, LexicalScopingUnknown}
}

func (t *Terminator) Accept(v Visitor, data unsafe.Pointer) {
	v.VisitTerminator(t, data)
}

type BinaryExpression struct {
	Left    SyntaxTree
	Right   SyntaxTree
	OpToken TokenDetail
}

func NewBinaryExpression(left, right SyntaxTree, op TokenDetail) *BinaryExpression {
	return &BinaryExpression{left, right, op}
}

func (b *BinaryExpression) Accept(v Visitor, data unsafe.Pointer) {
	v.VisitBinaryExpression(b, data)
}

type UnaryExpression struct {
	Exp     SyntaxTree
	OpToken TokenDetail
}

func NewUnaryExpression(exp SyntaxTree, op TokenDetail) *UnaryExpression {
	return &UnaryExpression{exp, op}
}

func (u *UnaryExpression) Accept(v Visitor, data unsafe.Pointer) {
	v.VisitUnaryExpression(u, data)
}

type FunctionBody struct {
	ParamList SyntaxTree
	BLock     SyntaxTree
	HasSelf   bool // For code generate, has 'self' param or not
	Line      int
}

func NewFunctionBody(paramList, block SyntaxTree, line int) *FunctionBody {
	return &FunctionBody{paramList, block, false, line}
}

func (f *FunctionBody) Accept(v Visitor, data unsafe.Pointer) {
	v.VisitFunctionBody(f, data)
}

type ParamList struct {
	NameList    SyntaxTree
	Vararg      bool
	FixArgCount int // For semantic and code generate
}

func NewParamList(nameList SyntaxTree, vararg bool) *ParamList {
	return &ParamList{nameList, vararg, 0}
}

func (p *ParamList) Accept(v Visitor, data unsafe.Pointer) {
	v.VisitParamList(p, data)
}

type NameList struct {
	Names []TokenDetail
}

func NewNameList() *NameList {
	return &NameList{}
}

func (n *NameList) Accept(v Visitor, data unsafe.Pointer) {
	v.VisitNameList(n, data)
}

type TableDefine struct {
	Fields []SyntaxTree
	Line   int
}

func NewTableDefine(line int) *TableDefine {
	return &TableDefine{Line: line}
}

func (t *TableDefine) Accept(v Visitor, data unsafe.Pointer) {
	v.VisitTableDefine(t, data)
}

type TableIndexField struct {
	Index SyntaxTree
	Value SyntaxTree
	Line  int
}

func NewTableIndexField(index, value SyntaxTree, line int) *TableIndexField {
	return &TableIndexField{index, value, line}
}

func (t *TableIndexField) Accept(v Visitor, data unsafe.Pointer) {
	v.VisitTableIndexField(t, data)
}

type TableNameField struct {
	Name  TokenDetail
	Value SyntaxTree
}

func NewTableNameField(name TokenDetail, value SyntaxTree) *TableNameField {
	return &TableNameField{name, value}
}

func (t *TableNameField) Accept(v Visitor, data unsafe.Pointer) {
	v.VisitTableNameField(t, data)
}

type TableArrayField struct {
	Value SyntaxTree
	Line  int
}

func NewTableArrayField(value SyntaxTree, line int) *TableArrayField {
	return &TableArrayField{value, line}
}

func (t *TableArrayField) Accept(v Visitor, data unsafe.Pointer) {
	v.VisitTableArrayField(t, data)
}

type IndexAccessor struct {
	Table    SyntaxTree
	Index    SyntaxTree
	Line     int
	Semantic int // For semantic
}

func NewIndexAccessor(table, index SyntaxTree, line int) *IndexAccessor {
	return &IndexAccessor{table, index, line, SemanticOpNone}
}

func (i *IndexAccessor) Accept(v Visitor, data unsafe.Pointer) {
	v.VisitIndexAccessor(i, data)
}

type MemberAccessor struct {
	Table  SyntaxTree
	Member TokenDetail
}

func NewMemberAccessor(table SyntaxTree, member TokenDetail) *MemberAccessor {
	return &MemberAccessor{table, member}
}

func (m *MemberAccessor) Accept(v Visitor, data unsafe.Pointer) {
	v.VisitMemberAccessor(m, data)
}

type NormalFuncCall struct {
	Caller SyntaxTree
	Args   SyntaxTree
	Line   int // Function call line in source
}

func NewNormalFuncCall(caller, args SyntaxTree, line int) *NormalFuncCall {
	return &NormalFuncCall{caller, args, line}
}

func (n *NormalFuncCall) Accept(v Visitor, data unsafe.Pointer) {
	v.VisitNormalFuncCall(n, data)
}

type MemberFuncCall struct {
	Caller SyntaxTree
	Member TokenDetail
	Args   SyntaxTree
	Line   int // Function call line in source
}

func NewMemberFuncCall(caller SyntaxTree, member TokenDetail, args SyntaxTree, line int) *MemberFuncCall {
	return &MemberFuncCall{caller, member, args, line}
}

func (m *MemberFuncCall) Accept(v Visitor, data unsafe.Pointer) {
	v.VisitMemberFuncCall(m, data)
}

type FuncCallArgs struct {
	Arg           SyntaxTree
	Type          int
	ArgValueCount int // For code generate
}

const (
	ArgTypeExpList = iota
	ArgTypeTable
	ArgTypeString
)

func NewFuncCallArgs(arg SyntaxTree, argType int) *FuncCallArgs {
	return &FuncCallArgs{arg, argType, 0}
}

func (f *FuncCallArgs) Accept(v Visitor, data unsafe.Pointer) {
	v.VisitFuncCallArgs(f, data)
}

type ExpressionList struct {
	ExpList []SyntaxTree
	Line    int // Start line
}

func NewExpressionList(startLine int) *ExpressionList {
	return &ExpressionList{Line: startLine}
}

func (e *ExpressionList) Accept(v Visitor, data unsafe.Pointer) {
	v.VisitExpressionList(e, data)
}
