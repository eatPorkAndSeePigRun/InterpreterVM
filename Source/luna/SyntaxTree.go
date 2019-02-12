package luna

// Expression or variable operation semantic
type SemanticOp int64

const (
	SemanticOpNone = iota
	SemanticOpRead
	SemanticOpWrite
)

// Expression or variable lexical scoping
type LexicalScoping int64

const (
	LexicalScopingUnknown = iota
	LexicalScopingGlobal  // Expression or variable in global table
	LexicalScopingUpValue // Expression or variable in upValue
	LexicalScopingLocal   // Expression or variable in current function
)

// AST base class, all AST node derived from this class and
// provide Visitor to Accept itself.
type SyntaxTree struct {
}

type Chunk struct {
}

type Block struct {
}

type ReturnStatement struct {
}

type BreakStatement struct {
}

type DoStatement struct {
}

type WhileStatement struct {
}

type RepeatStatement struct {
}

type IfStatement struct {
}

type ElseIfStatement struct {
}

type ElseStatement struct {
}

type NumericForStatement struct {
}

type GenericForStatement struct {
}

type FunctionStatement struct {
}

type FunctionName struct {
}

type LocalFunctionStatement struct {
}

type LocalNameListStatement struct {
}

type AssignmentStatement struct {
}

type VarList struct {
}

type Terminator struct {
}

type BinaryExpression struct {
}

type UnaryExpression struct {
}

type FunctionBody struct {
}

type ParamList struct {
}

type NameList struct {
}

type TableDefine struct {
}

type TableIndexField struct {
}

type TableNameField struct {
}

type TableArrayField struct {
}

type IndexAccessor struct {
}

type MemberAccessor struct {
}

type NormalFuncCall struct {
}

type MemberFuncCall struct {
}

type FuncCallArgs struct {
}

type ExpressionList struct {
}
