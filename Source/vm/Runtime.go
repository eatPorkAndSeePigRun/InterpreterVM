package vm

import (
	"unsafe"
)

// Used for Value pointer addition operations
func vPointerAdd(v *Value, i int) *Value {
	return (*Value)(unsafe.Pointer(uintptr(unsafe.Pointer(v)) + uintptr(i)*unsafe.Sizeof(Value{})))
}

// Used for Instruction pointer addition operations
func iPointerAdd(v *Instruction, i int) *Instruction {
	return (*Instruction)(unsafe.Pointer(uintptr(unsafe.Pointer(v)) + uintptr(i)*unsafe.Sizeof(Instruction{})))
}

const KBaseStackSize int = 10000

// Runtime stack, registers of each function is one part of stack.
type Stack struct {
	ValueStack []Value
	Top        *Value
}

func NewStack() *Stack {
	s := new(Stack)
	s.ValueStack = make([]Value, 1, KBaseStackSize)
	s.Top = &s.ValueStack[0]
	return s
}

// Set new top pointer, and [new top, old top) will be set nil
func (s *Stack) SetNewTop(top *Value) {
	old := s.Top
	s.Top = top

	// Clear values between new top to old
	for uintptr(unsafe.Pointer(top)) <= uintptr(unsafe.Pointer(old)) {
		top.SetNil()
		top = vPointerAdd(top, 1)
	}
}

// Function call stack info
type CallInfo struct {
	Register     *Value       // register base pointer which points to Stack
	Func         *Value       // current closure, pointer to stack Value
	Instruction  *Instruction // current Instruction
	End          *Instruction // Instruction end
	ExpectResult int          // expect result of this function call
}

func NewCallInfo() *CallInfo {
	return &CallInfo{}
}
