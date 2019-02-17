package vm

import (
	"InterpreterVM/Source/datatype"
)

const KBaseStackSize int = 10000

// Runtime stack, registers of each function is one part of stack.
type Stack struct {
	ValueStack []datatype.Value
	Top        int
}

func NewStack() *Stack {
	return &Stack{make([]datatype.Value, 0, KBaseStackSize), 0}
}

// Set new top pointer, and [new top, old top) will be set nil
func (s *Stack) SetNewTop(top int) {
	old := s.Top
	s.Top = top

	// Clear values between new top to old
	for ; top <= old; top++ {
		s.ValueStack[top].SetNil()
	}
}

// Function call stack info
type CallInfo struct {
	Register     *datatype.Value // register base pointer which points to Stack
	Func         *datatype.Value // current closure, pointer to stack Value
	Instruction  *Instruction    // current Instruction
	End          *Instruction    // Instruction end
	ExpectResult int64           // expect result of this function call
}

func NewCallInfo() *CallInfo {
	return &CallInfo{}
}
