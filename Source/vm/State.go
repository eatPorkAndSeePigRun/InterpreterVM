package vm

import (
	"container/list"
	"math"
	"unsafe"
)

// Error type reported by called c function
const (
	CFunctionErrorTypeNoError = iota + 1
	CFunctionErrorTypeArgCount
	CFunctionErrorTypeArgType
)

const (
	metaTables   = "__metaTables"
	modulesTable = "__modules"
)

// Error reported by called c function
type CFunctionError struct {
	eType          int
	ExpectArgCount int
	ArgIndex       int
	ExpectType     int
}

func NewCFunctionError() CFunctionError {
	return CFunctionError{eType: CFunctionErrorTypeNoError}
}

type State struct {
	moduleManager *ModuleManager // Manage all modules
	stringPool    *StringPool    // All strings in the pool
	gc            *GC            // The GC

	cFuncError CFunctionError // Error of call c function

	stack  Stack     // Stack data
	calls  list.List // Stack frames, and its element.value is CallInfo
	global Value     // Global table
}

func NewState() *State {
	var s State

	s.stringPool = NewStringPool()

	// Init GC
	deleter := func(obj GCObject, objType int) {
		if objType == GCObjectTypeString {
			s.stringPool.DeleteString(obj.(*String))
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
	k := NewValueString(s.GetString(metaTables))
	v := NewValueTable(s.NewTable())
	s.global.Table.SetValue(k, v)

	// New table for store modules
	k = NewValueString(s.GetString(modulesTable))
	v = NewValueTable(s.NewTable())
	s.global.Table.SetValue(k, v)

	// Init module manager
	s.moduleManager = NewModuleManager(&s, v.Table)

	return &s
}

// Full GC root
func (s *State) fullGCRoot(v GCObjectVisitor) {
	// Visit global table
	s.global.Accept(v)

	// Visit stack values
	for index := range s.stack.ValueStack {
		s.stack.ValueStack[index].Accept(v)
	}

	// Visit call info
	for e := s.calls.Front(); e != nil; e = e.Next() {
		call := e.Value.(*CallInfo)
		call.Register.Accept(v)
		if call.Func != nil {
			call.Func.Accept(v)
		}
	}
}

// For CallFunction
func (s *State) callClosure(f *Value, expectResult int) {
	var callee CallInfo
	calleeProto := f.Closure.GetPrototype()
	callee.Func = f
	callee.Instruction = calleeProto.GetOpCodes()
	callee.End = iPointerAdd(callee.Instruction, calleeProto.OpCodeSize())
	callee.ExpectResult = expectResult

	arg := vPointerAdd(f, 1)
	fixedArgs := calleeProto.FixedArgCount()

	// Fixed arg start from base register
	if calleeProto.HasVararg() {
		top := s.stack.Top
		callee.Register = top
		count := int((uintptr(unsafe.Pointer(top)) - uintptr(unsafe.Pointer(arg))) /
			unsafe.Sizeof(Value{}))
		for i := 0; i < count && i < int(fixedArgs); i++ {
			*top = *arg
			top = vPointerAdd(top, 1)
			arg = vPointerAdd(arg, 1)
		}
	} else {
		callee.Register = arg
		// fill nil for overflow args
		newTop := vPointerAdd(callee.Register, fixedArgs)
		arg := s.stack.Top
		for uintptr(unsafe.Pointer(arg)) < uintptr(unsafe.Pointer(newTop)) {
			arg.SetNil()
			arg = vPointerAdd(arg, 1)
		}
	}

	s.stack.SetNewTop(vPointerAdd(callee.Register, fixedArgs))
	s.calls.PushBack(&callee)
}

func (s *State) callCFunction(f *Value, expectResult int) {
	// Push the c function CallInfo
	callee := CallInfo{Register: vPointerAdd(f, 1), Func: f, ExpectResult: expectResult}
	s.calls.PushBack(&callee)

	// Call c function
	cfunc := f.CFunc
	if err := s.checkCFunctionError(); err != nil {
		panic(err)
	}
	resCount := cfunc(s)
	if err := s.checkCFunctionError(); err != nil {
		panic(err)
	}

	var src *Value
	if resCount > 0 {
		src = vPointerAdd(s.stack.Top, resCount)
	}

	// Copy c function result to caller stack
	dst := f
	if expectResult == ExpValueCountAny {
		for i := 0; i < int(resCount); i++ {
			*dst = *src
			dst = vPointerAdd(dst, 1)
			src = vPointerAdd(src, 1)
		}
	} else {
		count := int(math.Min(float64(expectResult), float64(resCount)))
		for i := 0; i < count; i++ {
			dst.SetNil()
			dst = vPointerAdd(dst, 1)
		}
	}

	// Set registers which after dst to nil
	// and set new stack top pointer
	s.stack.SetNewTop(dst)

	// Pop the c function CallInfo
	s.calls.Remove(s.calls.Back())
}

func (s *State) checkCFunctionError() error {
	e := s.GetCFunctionErrorData()
	if e.eType == CFunctionErrorTypeNoError {
		return nil
	}

	var exp error
	if e.eType == CFunctionErrorTypeArgCount {
		exp = NewCallCFuncError("expect ", e.ExpectArgCount, " arguments")
	} else if e.eType == CFunctionErrorTypeArgType {
		call := s.calls.Back().Value.(*CallInfo)
		arg := vPointerAdd(call.Register, e.ArgIndex)
		exp = NewCallCFuncError("argument #", e.ArgIndex+1,
			" is a ", arg.TypeName(), " value, expect a ",
			(e.ExpectType), " value")
		// TODO
	}

	// Pop the c function CallInfo
	s.calls.Remove(s.calls.Back())
	panic(exp)
}

// Get the table which stores all metaTables
func (s *State) getMetaTables() *Table {
	k := NewValueString(s.GetString(metaTables))
	v := s.global.Table.GetValue(k)
	if v.Type != ValueTTable {
		panic("assert")
	}
	return v.Table
}

// Check module loaded or not
func (s *State) IsModuleLoaded(moduleName string) bool {
	return s.moduleManager.IsLoaded(moduleName)
}

// Load module, if load success, then push a module closure on stack,
// otherwise throw Exception
func (s *State) LoadModule(moduleName string) {
	value := s.moduleManager.GetModuleClosure(moduleName)
	if value.IsNil() {
		panic(s.moduleManager.LoadModule(moduleName))
	} else {
		*s.stack.Top = value
		s.stack.Top = vPointerAdd(s.stack.Top, 1)
	}
}

// Load module and call the module function when the module loaded success.
func (s *State) DoModule(moduleName string) {
	s.LoadModule(moduleName)
	isTrue, err := s.CallFunction(vPointerAdd(s.stack.Top, -1), 0, 0)
	if err != nil {
		panic(err)
	}
	if isTrue {
		vm := NewVM(s)
		vm.Execute()
	}
}

// Load string and call the string function when the string loaded success.
func (s *State) DoString(str, name string) {
	s.moduleManager.LoadString(str, name)
	isTrue, err := s.CallFunction(vPointerAdd(s.stack.Top, -1), 0, 0)
	if err != nil {
		panic(err)
	}
	if isTrue {
		vm := NewVM(s)
		vm.Execute()
	}
}

// Call an in stack function
// If f is a closure, then create a stack frame and return true,
// call VM::Execute() to execute the closure instructions.
// Return false when f is a c function.
func (s *State) CallFunction(f *Value, argCount int, expectResult int) (bool, error) {
	if f.Type != ValueTClosure && f.Type != ValueTCFunction {
		panic("assert")
	}

	// Set stack top when argCount is fixed
	if argCount != ExpValueCountAny {
		s.stack.Top = vPointerAdd(f, 1+argCount)
	}

	if f.Type == ValueTClosure {
		// We need enter next ExecuteFrame
		s.callClosure(f, expectResult)
		return true, nil
	} else {
		s.callCFunction(f, expectResult)
		return false, nil
	}
}

// New GCObjects
func (s *State) GetString(str string) *String {
	str2 := s.stringPool.GetString(str)
	if str2 == nil {
		str2 = s.gc.NewString(GCGen0)
		str2.SetValue(str)
		s.stringPool.AddString(str2)
	}
	return str2
}

// New GCObjects
func (s *State) NewTable() *Table {
	return s.gc.NewTAble(GCGen0)
}

// New GCObjects
func (s *State) NewFunction() *Function {
	return s.gc.NewFunction(GCGen2)
}

// New GCObjects
func (s *State) NewClosure() *Closure {
	return s.gc.NewClosure(GCGen0)
}

// New GCObjects
func (s *State) NewUpvalue() *Upvalue {
	return s.gc.NewUpvalue(GCGen0)
}

// New GCObjects
func (s *State) NewString() *String {
	return s.gc.NewString(GCGen0)
}

// New GCObjects
func (s *State) NewUserData() *UserData {
	return s.gc.NewUserData(GCGen0)
}

// Get current CallInfo
func (s *State) GetCurrentCall() *CallInfo {
	if s.calls.Len() == 0 {
		return nil
	}
	return s.calls.Back().Value.(*CallInfo)
}

// Get the global table value
func (s *State) GetGlobal() *Value {
	return &s.global
}

// Return metaTable, create when metaTable not existed
func (s *State) GetMetaTable(metaTableName string) *Table {
	k := NewValueString(s.GetString(metaTableName))
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
func (s *State) EraseMetaTable(metaTableName string) {
	k := NewValueString(s.GetString(metaTableName))
	var null Value
	metaTables := s.getMetaTables()
	metaTables.SetValue(k, null)
}

// For call c function
func (s *State) ClearCFunctionError() {
	s.cFuncError.eType = CFunctionErrorTypeNoError
}

// Error data for call c function
func (s *State) GetCFunctionErrorData() *CFunctionError {
	return &s.cFuncError
}

// Get the GC
func (s *State) GetGC() *GC {
	return s.gc
}

// Check and run GC
func (s *State) CheckRunGC() {
	s.gc.CheckGC()
}
