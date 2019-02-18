package vm

import (
	"InterpreterVM/Source/datatype"
	"unsafe"
)

// Used for Value pointer addition operations
func vPointerAdd(v *datatype.Value, i int) *datatype.Value {
	return (*datatype.Value)(unsafe.Pointer(uintptr(unsafe.Pointer(v)) + uintptr(i)*unsafe.Sizeof(datatype.Value{})))
}

// Used for Instruction pointer addition operations
func iPointerAdd(v *Instruction, i int) *Instruction {
	return (*Instruction)(unsafe.Pointer(uintptr(unsafe.Pointer(v)) + uintptr(i)*unsafe.Sizeof(Instruction{})))
}

const KBaseStackSize int = 10000

// Runtime stack, registers of each function is one part of stack.
type Stack struct {
	ValueStack []datatype.Value
	Top        *datatype.Value
}

func NewStack() *Stack {
	return &Stack{make([]datatype.Value, 0, KBaseStackSize), nil}
}

// Set new top pointer, and [new top, old top) will be set nil
func (s *Stack) SetNewTop(top *datatype.Value) {
	old := s.Top
	s.Top = top

	// Clear values between new top to old
	for uintptr(unsafe.Pointer(top)) <= uintptr(unsafe.Pointer(old)) {
		top.SetNil()
		//top = (*datatype.Value)(unsafe.Pointer(uintptr(unsafe.Pointer(top)) + unsafe.Sizeof(datatype.Value{})))
		top = vPointerAdd(top, 1)
	}
}

// Function call stack info
type CallInfo struct {
	Register     *datatype.Value // register base pointer which points to Stack
	Func         *datatype.Value // current closure, pointer to stack Value
	Instruction  *Instruction    // current Instruction
	End          *Instruction    // Instruction end
	ExpectResult int             // expect result of this function call
}

func NewCallInfo() *CallInfo {
	return &CallInfo{}
}
