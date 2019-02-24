package vm

import "unsafe"

// This class is API for library to manipulate stack,
// stack index value is:
// -1 ~ -n is top to bottom,
// 0 ~ n is bottom to top.
type StackAPI struct {
	state *State
	stack *Stack
}

func NewStackAPI(state *State) *StackAPI {
	return &StackAPI{state: state, stack: &state.stack}
}

func (s *StackAPI) CheckArgs(index, params int) bool {
	// No more expect argument to check, success
	return true
}

func (s *StackAPI) CheckArgs3(index, params, vType int, valueTypes ...interface{}) bool {
	// All arguments check success
	if index == params {
		return true
	}

	// Check type of the index + 1 argument
	if s.GetValueType(index) != vType {
		s.ArgTypeError(index, vType)
		return false
	}

	// Check remain arguments
	index++
	return s.CheckArgs1(index, params, valueTypes)
}

func (s *StackAPI) CheckArgs1(minCount int, valueTypes ...interface{}) bool {
	// Check count of arguments
	params := s.GetStackSize()
	if params < minCount {
		s.ArgCountError(minCount)
		return false
	}

	if len(valueTypes) == 2 {
		return s.CheckArgs(0, params)
	} else {
		return s.CheckArgs3(0, params, valueTypes[0].(int), valueTypes[1:])
	}
}

// Get count of value in this function stack
func (s *StackAPI) GetStackSize() int {
	begin := s.stack.Top
	end := s.state.calls.Back().Value.(*CallInfo).Register
	return int((uintptr(unsafe.Pointer(begin)) - uintptr(unsafe.Pointer(end))) /
		unsafe.Sizeof(Value{}))
}

// Get value type by index of stack
func (s *StackAPI) GetValueType(index int) int {
	v := s.GetValue(index)
	if v != nil {
		return v.Type
	} else {
		return ValueTNil
	}
}

// Check value type by index of stack
func (s *StackAPI) IsNumber(index int) bool {
	return s.GetValueType(index) == ValueTNumber
}

// Check value type by index of stack
func (s *StackAPI) IsString(index int) bool {
	return s.GetValueType(index) == ValueTString
}

// Check value type by index of stack
func (s *StackAPI) IsBool(index int) bool {
	return s.GetValueType(index) == ValueTBool
}

// Check value type by index of stack
func (s *StackAPI) IsClosure(index int) bool {
	return s.GetValueType(index) == ValueTCFunction
}

// Check value type by index of stack
func (s *StackAPI) IsTable(index int) bool {
	return s.GetValueType(index) == ValueTTable
}

// Check value type by index of stack
func (s *StackAPI) IsUserData(index int) bool {
	return s.GetValueType(index) == ValueTUserData
}

// Check value type by index of stack
func (s *StackAPI) IsCFunction(index int) bool {
	return s.GetValueType(index) == ValueTCFunction
}

// Get value from stack by index
func (s *StackAPI) GetNumber(index int) float64 {
	v := s.GetValue(index)
	if v != nil {
		return v.Num
	} else {
		return 0.0
	}
}

// Get value from stack by index
func (s *StackAPI) GetCString(index int) string {
	v := s.GetValue(index)
	if v != nil {
		return v.Str.GetCStr()
	} else {
		return ""
	}
}

// Get value from stack by index
func (s *StackAPI) GetString(index int) *String {
	v := s.GetValue(index)
	if v != nil {
		return v.Str
	} else {
		return nil
	}
}

// Get value from stack by index
func (s *StackAPI) GetBool(index int) bool {
	v := s.GetValue(index)
	if v != nil {
		return v.BValue
	} else {
		return false
	}
}

// Get value from stack by index
func (s *StackAPI) GetClosure(index int) *Closure {
	v := s.GetValue(index)
	if v != nil {
		return v.Closure
	} else {
		return nil
	}
}

// Get value from stack by index
func (s *StackAPI) GetTable(index int) *Table {
	v := s.GetValue(index)
	if v != nil {
		return v.Table
	} else {
		return nil
	}
}

// Get value from stack by index
func (s *StackAPI) GetUserData(index int) *UserData {
	v := s.GetValue(index)
	if v != nil {
		return v.UserDate
	} else {
		return nil
	}
}

// Get value from stack by index
func (s *StackAPI) GetCFunction(index int) CFunctionType {
	v := s.GetValue(index)
	if v != nil {
		return v.CFunc
	} else {
		return nil
	}
}

// Get value from stack by index
func (s *StackAPI) GetValue(index int) *Value {
	if s.state.calls.Len() == 0 {
		panic("assert")
	}
	var v *Value
	if index < 0 {
		v = vPointerAdd(s.stack.Top, index)
	} else {
		v = vPointerAdd(s.state.calls.Back().Value.(*CallInfo).Register, index)
	}

	if (uintptr(unsafe.Pointer(v)) >= uintptr(unsafe.Pointer(s.stack.Top))) ||
		(uintptr(unsafe.Pointer(v)) < uintptr(unsafe.Pointer(s.state.calls.Back().Value.(*CallInfo).Register))) {
		return nil
	} else {
		return v
	}
}

// Push value to stack
func (s *StackAPI) PushNil() {
	s.pushValue().Type = ValueTNil
}

// Push value to stack
func (s *StackAPI) PushNumber(num float64) {
	v := s.pushValue()
	v.Type = ValueTNumber
	v.Num = num
}

// Push value to stack
func (s *StackAPI) PushString(str string) {
	v := s.pushValue()
	v.Type = ValueTString
	v.Str = s.state.GetString(str)
}

// Push value to stack
func (s *StackAPI) PushBool(value bool) {
	v := s.pushValue()
	v.Type = ValueTBool
	v.BValue = value
}

// Push value to stack
func (s *StackAPI) PushTable(table *Table) {
	v := s.pushValue()
	v.Type = ValueTTable
	v.Table = table
}

// Push value to stack
func (s *StackAPI) PushUserData(userData *UserData) {
	v := s.pushValue()
	v.Type = ValueTUserData
	v.UserDate = userData
}

// Push value to stack
func (s *StackAPI) PushCFunction(function CFunctionType) {
	v := s.pushValue()
	v.Type = ValueTCFunction
	v.CFunc = function
}

// Push value to stack
func (s *StackAPI) PushValue(value Value) {
	*s.pushValue() = value
}

// For report argument error
func (s *StackAPI) ArgCountError(expectCount int) {
	cFuncError := s.state.GetCFunctionErrorData()
	cFuncError.eType = CFunctionErrorTypeArgCount
	cFuncError.ExpectArgCount = expectCount
}

// For report argument error
func (s *StackAPI) ArgTypeError(argIndex, expectType int) {
	cFuncError := s.state.GetCFunctionErrorData()
	cFuncError.eType = CFunctionErrorTypeArgType
	cFuncError.ArgIndex = argIndex
	cFuncError.ExpectType = expectType
}

// Push value to stack, and return the value
func (s *StackAPI) pushValue() *Value {
	res := s.stack.Top
	s.stack.Top = vPointerAdd(s.stack.Top, 1)
	return res
}

// For register table member
type TableMemberReg struct {
	Name   string        // Member name
	CFunc  CFunctionType // Member value
	Number float64       // Member value
	Str    string        // Member value
	VType  int           // Member value type
}

func NewTableMemberRegCFunction(name string, cFunc CFunctionType) *TableMemberReg {
	return &TableMemberReg{Name: name, CFunc: cFunc}
}

func NewTableMemberRegNumber(name string, number float64) *TableMemberReg {
	return &TableMemberReg{Name: name, Number: number}
}

func NewTableMemberRegString(name, str string) *TableMemberReg {
	return &TableMemberReg{Name: name, Str: str}
}

// This class provide register C function/data to vm
type Library struct {
	state  *State
	global *Table
}

func NewLibrary(state *State) *Library {
	return &Library{state: state, global: state.global.Table}
}

// Register global function 'func' as 'name'
func (l *Library) RegisterFunc(name string, cFunc CFunctionType) {
	l.registerFunc(l.global, name, cFunc)
}

// Register a table of functions
func (l *Library) RegisterTableFunction(name string, table *TableMemberReg, size int) {
	k := NewValueString(l.state.GetString(name))
	t := l.state.NewTable()
	v := NewValueTable(t)
	l.global.SetValue(k, v)

	l.registerToTable(t, table, size)
}

// Register a metatable
func (l *Library) RegisterMetatable(name string, table *TableMemberReg, size int) {
	t := l.state.GetMetaTable(name)
	l.registerToTable(t, table, size)
}

func (l *Library) registerToTable(table *Table, tableReg *TableMemberReg, size int) {
	for i := 0; i < size; i++ {
		tr := (*TableMemberReg)(unsafe.Pointer(uintptr(unsafe.Pointer(tableReg)) +
			uintptr(i)*unsafe.Sizeof(TableMemberReg{})))
		switch tr.VType {
		case ValueTCFunction:
			l.registerFunc(table, tr.Name, tr.CFunc)
		case ValueTNumber:
			l.registerNumber(table, tr.Name, tr.Number)
		case ValueTString:
			l.registerString(table, tr.Name, tr.Str)
		default:
		}
	}
}

func (l *Library) registerFunc(table *Table, name string, cFunc CFunctionType) {
	k := NewValueString(l.state.GetString(name))
	v := NewValueCFunction(cFunc)
	table.SetValue(k, v)
}

func (l *Library) registerNumber(table *Table, name string, number float64) {
	k := NewValueString(l.state.GetString(name))
	v := NewValueNum(number)
	table.SetValue(k, v)
}

func (l *Library) registerString(table *Table, name, str string) {
	k := NewValueString(l.state.GetString(name))
	v := NewValueString(l.state.GetString(str))
	table.SetValue(k, v)
}
