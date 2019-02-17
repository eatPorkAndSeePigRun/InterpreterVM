package vm

import (
	"InterpreterVM/Source/datatype"
	"container/list"
	"math"
	"unsafe"
)

// Error type reported by called c function
const (
	CFunctionErrorTypeNoError = iota
	CFunctionErrorTypeArgCount
	CFunctionErrorTypeArgType
)

const (
	metaTables   = "__metaTables"
	modulesTable = "__modules"
)

// Error reported by called c function
type CFunctionError struct {
	type_ int
	//ExpectArgCount int64
	//ArgIndex       int64
	//ExpectType     ValueT
}

type State struct {
	moduleManager *ModuleManager       // Manage all modules
	stringPool    *datatype.StringPool // All strings in the pool
	gc            *datatype.GC         // The GC

	cFuncError CFunctionError // Error of call c function

	stack  Stack          // Stack data
	calls  list.List      // Stack frames, and its element.value is CallInfo
	global datatype.Value // Global table
}

func NewState() *State {
	var s State

	s.stringPool = new(StringPool)

	// Init GC
	deleter := func(obj *GCObject, type_ int) {
		if type_ == GCObjectTypeString {
			s.stringPool.DeleteString((*String)(unsafe.Pointer(obj)))
		}
		obj = nil
	}
	s.gc = NewGC(deleter, false)
	root := s.fullGCRoot
	s.gc.SetRootTraveller(root, root)

	// New global table
	s.global.Table = s.NewTable()
	s.global.Type = ValueTTable

	// New table for store metaTables
	k := Value{Type: ValueTString, Str: s.GetString(metaTables)}
	v := Value{Type: ValueTTable, Table: s.NewTable()}
	s.global.Table.SetValue(k, v)

	// New table for store modules
	k = Value{Type: ValueTString, Str: s.GetString(modulesTable)}
	v = Value{Type: ValueTTable, Table: s.NewTable()}
	s.global.Table.SetValue(k, v)

	// Init module manager
	s.moduleManager = NewModuleManager(&s, v.Table)

	return s
}

// Full GC root
func (s State) fullGCRoot(v GCObjectVisitor) {
	// Visit global table
	s.global.Accept(v)

	// Visit stack values
	for _, value := range s.stack.Stack_ {
		value.Accept(v)
	}

	// Visit call info
	for e := s.calls.Front(); e != nil; e = e.Next() {
		call := e.Value.(CallInfo)
		call.Register.Accept(v)
		if call.Func_ != nil {
			call.Func_.Accept(v)
		}
	}
}

// For CallFunction
func (s State) callClosure(f *Value, expectResult int64) {
	var callee CallInfo
	calleeProto := f.Closure.GetPrototype()
	callee.Func_ = f
	callee.Instruction = calleeProto.GetOpCodes()
	callee.End = callee.Instruction + calleeProto.OpCodeSize()
	callee.ExpectResult = expectResult

	arg := f + 1
	fixedArgs := calleeProto.FixedArgCount()

	// Fixed arg start from base register
	if calleeProto.HasVararg() {
		top := s.stack.Top
		callee.Register = top
		count := top - arg
		for i := 0; i < count && i < int(fixedArgs); i++ {
			*top = *arg
			top++
			arg++
		}
	} else {
		callee.Register = arg
		// fill nil for overflow args
		newTop := callee.Register + fixedArgs
		for arg := s.stack.Top; arg < newTop; arg++ {
			arg.SetNil()
		}
	}

	s.stack.SetNewTop(callee.Register + fixedArgs)
	s.calls.PushBack(callee)
}

func (s State) callCFunction(f *Value, expectResult int64) {
	// Push the c function CallInfo
	callee := CallInfo{Register: f + 1, Func_: f, ExpectResult: expectResult}
	s.calls.PushBack(callee)

	// Call c function
	cfunc := f.CFunc
	s.checkCFunctionError()
	resCount := cfunc(&s)
	s.checkCFunctionError()

	var src *Value
	if resCount > 0 {
		src = s.stack.Top - resCount
	}

	// Copy c function result to caller stack
	dst := f
	if expectResult == ExpValueCountAny {
		for i := 0; i < int(resCount); i++ {
			*dst = *src
			dst++
			src++
		}
	} else {
		count := int(math.Min(float64(expectResult), float64(resCount)))
		for i := 0; i < count; i++ {
			dst.SetNil()
			dst++
		}
	}

	// Set registers which after dst to nil
	// and set new stack top pointer
	s.stack.SetNewTop(dst)

	// Pop the c function CallInfo
	s.calls.Back() // TODO
}

func (s State) checkCFunctionError() {
	err := s.GetCFunctionErrorData()
	if err.type_ == CFunctionErrorTypeNoError {
		return
	}

	var exp CallCFuncException
	if err.type_ == CFunctionErrorTypeArgCount {
		//TODO exp=
	} else if err.type_ == CFunctionErrorTypeArgType {
		call := s.calls.Back().Value.(*CallInfo)
		arg := call.Register + err.ArgIndex
		// TODO exp =
	}

	// Pop the c function CallInfo
	s.calls.Back() // TODO
	panic("exp")
}

// Get the table which stores all metaTables
func (s State) getMetaTables() *Table {
	k := Value{Type: ValueTString, Str: s.GetString(metaTables)}
	v := s.global.Table.GetValue(k)
	if v.Type != ValueTTable {
		panic("assert")
	}
	return v.Table
}

// Check module loaded or not
func (s State) IsModuleLoaded(moduleName string) bool {
	return s.moduleManager.IsLoaded(moduleName)
}

// Load module, if load success, then push a module closure on stack,
// otherwise throw Exception
func (s *State) LoadModule(moduleName string) {
	value := s.moduleManager.GetModuleClosure(moduleName)
	if value.IsNil() {
		s.moduleManager.LoadModule(moduleName)
	} else {
		*s.stack.Top = value
		s.stack.Top++
	}
}

// Load module and call the module function when the module
// loaded success.
func (s State) DoModule(moduleName string) {
	s.LoadModule(moduleName)
	if s.CallFunction(s.stack.top-1, 0, 0) {
		vm := VM{&s}
		vm.Execute()
	}
}

// Load string and call the string function when the string
// loaded success.
func (s State) DoString(str, name string) {
	s.moduleManager.LoadString(str, name)
	if s.CallFunction(s.stack.top-1, 0, 0) {
		vm := VM{&s}
		vm.Execute()
	}
}

// Call an in stack function
// If f is a closure, then create a stack frame and return true,
// call VM::Execute() to execute the closure instructions.
// Return false when f is a c function.
func (s State) CallFunction(f *Value, argCount int, expectResult int64) (bool, error) {
	if !(f.Type == ValueTClosure || f.Type == ValueTCFunction) {
		panic("assert")
	}

	// Set stack top when argCount is fixed
	if argCount != ExpValueCountAny {
		s.stack.top = f + 1 + argCount
	}

	if f.Type == ValueTClosure {
		// We need enter next ExecuteFrame
		s.callClosure(f, expectResult)
		return true
	} else {
		s.callCFunction(f, expectResult)
		return false
	}
}

// New GCObjects
func (s State) GetString(str string) *String {
	str2 := s.stringPool.GetString(str)
	if str2 == nil {
		str2 = s.gc.NewString()
		str2.SetValue(str)
		s.stringPool.AddString(str2)
	}
	return str2
}

// New GCObjects
func (s State) NewFunction() *Function {
	return s.gc.NewFunction()
}

// New GCObjects
func (s State) NewClosure() *Closure {
	return s.gc.NewClosure()
}

// New GCObjects
func (s State) NewUpValue() *UpValue {
	return s.gc.NewUpValue()
}

// New GCObjects
func (s State) NewTable() *Table {
	return s.gc.NewTAble()
}

// New GCObjects
func (s State) NewUserData() *UserData {
	return s.gc.NewUserData()
}

// Get current CallInfo
func (s State) GetCurrentCall() *CallInfo {
	if s.calls.Len() == 0 {
		return nil
	}
	return s.calls.Back().Value.(*CallInfo)
}

// Get the global table value
func (s State) GetGlobal() *Value {
	return &s.global
}

// Return metaTable, create when metaTable not existed
func (s State) GetMetaTable(metaTableName string) *Table {
	k := Value{Type: ValueTString, Str: s.GetString(metaTableName)}
	metaTables := s.getMetaTables()
	metaTable := metaTables.GetValue(k)

	// Create table when metaTable not existed
	if metaTable.Type == ValueTNil {
		metaTable.Type = ValueTTable
		metaTable.Table = s.NewTable()
		metaTables.SetValue(k, metaTable)
	}

	if metaTable.Type != ValueTTable {
		panic("assert")
	}
	return metaTable.Table
}

// Erase metaTable
func (s State) EraseMetaTable(metaTableName string) {
	k := Value{Type: ValueTString, Str: s.GetString(metaTableName)}
	var null Value
	metaTables := s.getMetaTables()
	metaTables.SetValue(k, null)
}

// For call c function
func (s *State) ClearCFunctionError() {
	s.cFuncError.type_ = CFunctionErrorTypeNoError
}

// Error data for call c function
func (s State) GetCFunctionErrorData() *CFunctionError {
	return &s.cFuncError
}

// Get the GC
func (s State) GetGC() *GC {
	return s.gc
}

// Check and run GC
func (s State) CheckRunGC() {
	s.gc.CheckGC()
}
