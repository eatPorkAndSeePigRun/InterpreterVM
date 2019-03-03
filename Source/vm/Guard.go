package vm

import "runtime"

// Guard struct, using for RAII operations
type Guard struct {
	leave func()
}

func NewGuard(enter, leave func()) *Guard {
	guard := &Guard{leave}
	enter()
	l := func(obj *Guard) { obj.leave() }
	runtime.SetFinalizer(guard, l)
	return guard
}
