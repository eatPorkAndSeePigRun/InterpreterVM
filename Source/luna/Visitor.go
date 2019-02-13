package luna

import "unsafe"

type Visitor interface {
	VisitChunk(*Chunk, unsafe.Pointer)
	VisitBlock(*Block, unsafe.Pointer)
	VisitReturnStatement(*ReturnStatement, unsafe.Pointer)
	VisitBreakStatement(*BreakStatement, unsafe.Pointer)
	VisitDoStatement(*DoStatement, unsafe.Pointer)
	VisitWhileStatement(*WhileStatement, unsafe.Pointer)
	VisitRepeatStatement(*RepeatStatement, unsafe.Pointer)
	VisitIfStatement(*IfStatement, unsafe.Pointer)
	VisitElseIfStatement(*ElseStatement, unsafe.Pointer)
	VisitElseStatement(*ElseStatement, unsafe.Pointer)
	VisitNumericForStatement(*NumericForStatement, unsafe.Pointer)
	VisitGenericForStatement(*GenericForStatement, unsafe.Pointer)
	VisitFunctionStatement(*FunctionStatement, unsafe.Pointer)
	VisitFunctionName(*FunctionName, unsafe.Pointer)
	VisitLocalFunctionStatement(*LocalFunctionStatement, unsafe.Pointer)
	VisitLocalNameListStatement(*LocalNameListStatement, unsafe.Pointer)
	VisitAssignmentStatement(*AssignmentStatement, unsafe.Pointer)
	VisitVarList(*VarList, unsafe.Pointer)
	VisitTerminator(*Terminator, unsafe.Pointer)
	VisitBinaryExpression(*BinaryExpression, unsafe.Pointer)
	VisitUnaryExpression(*UnaryExpression, unsafe.Pointer)
	VisitFunctionBody(*FunctionBody, unsafe.Pointer)
	VisitParamList(*ParamList, unsafe.Pointer)
	VisitNameList(*NameList, unsafe.Pointer)
	VisitTableDefine(*TableDefine, unsafe.Pointer)
	VisitTableIndexField(*TableIndexField, unsafe.Pointer)
	VisitTableNameField(*TableNameField, unsafe.Pointer)
	VisitTableArrayField(*TableArrayField, unsafe.Pointer)
	VisitIndexAccessor(*IndexAccessor, unsafe.Pointer)
	VisitMemberAccessor(*MemberAccessor, unsafe.Pointer)
	VisitNormalFuncCall(*NormalFuncCall, unsafe.Pointer)
	VisitMemberFuncCall(*MemberFuncCall, unsafe.Pointer)
	VisitFuncCallArgs(*FuncCallArgs, unsafe.Pointer)
	VisitExpressionList(*ExpressionList, unsafe.Pointer)
}
