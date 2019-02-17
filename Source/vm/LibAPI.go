package vm

// This class is API for library to manipulate stack,
// stack index value is:
// -1 ~ -n is top to bottom,
// 0 ~ n is bottom to top.
type StackAPI struct {
	state *State
	stack *Stack
}

// For register table member
type TableMemberReg struct {
}

// This class provide register C function/data to vm
type Library struct {
}
