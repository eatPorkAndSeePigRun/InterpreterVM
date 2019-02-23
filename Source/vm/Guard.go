package vm

import "runtime"

// Guard struct, using for RAII operations
type Guard struct {
	leave func()
}

func NewGuard(enter, leave func()) *Guard {
	guard := &Guard{leave}
	enter()
	runtime.SetFinalizer(guard, guard.leave)
	return guard
}
