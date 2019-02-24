package math

import (
	. "InterpreterVM/Source/vm"
	"math"
)

func abs(state *State) int {
	api := NewStackAPI(state)
	if !api.CheckArgs(1, ValueTNumber) {
		return 0
	}
	api.PushNumber(math.Abs(api.GetNumber(0)))
	return 1
}

func acos(state *State) int {
	api := NewStackAPI(state)
	if !api.CheckArgs(1, ValueTNumber) {
		return 0
	}
	api.PushNumber(math.Acos(api.GetNumber(0)))
	return 1
}

func asin(state *State) int {
	api := NewStackAPI(state)
	if !api.CheckArgs(1, ValueTNumber) {
		return 0
	}
	api.PushNumber(math.Asin(api.GetNumber(0)))
	return 1
}

func atan(state *State) int {
	api := NewStackAPI(state)
	if !api.CheckArgs(1, ValueTNumber) {
		return 0
	}
	api.PushNumber(math.Atan(api.GetNumber(0)))
	return 1
}

func ceil(state *State) int {
	api := NewStackAPI(state)
	if !api.CheckArgs(1, ValueTNumber) {
		return 0
	}
	api.PushNumber(math.Ceil(api.GetNumber(0)))
	return 1
}

func cos(state *State) int {
	api := NewStackAPI(state)
	if !api.CheckArgs(1, ValueTNumber) {
		return 0
	}
	api.PushNumber(math.Cos(api.GetNumber(0)))
	return 1
}

func cosh(state *State) int {
	api := NewStackAPI(state)
	if !api.CheckArgs(1, ValueTNumber) {
		return 0
	}
	api.PushNumber(math.Cosh(api.GetNumber(0)))
	return 1
}

func exp(state *State) int {
	api := NewStackAPI(state)
	if !api.CheckArgs(1, ValueTNumber) {
		return 0
	}
	api.PushNumber(math.Exp(api.GetNumber(0)))
	return 1
}

func floor(state *State) int {
	api := NewStackAPI(state)
	if !api.CheckArgs(1, ValueTNumber) {
		return 0
	}
	api.PushNumber(math.Floor(api.GetNumber(0)))
	return 1
}

func sin(state *State) int {
	api := NewStackAPI(state)
	if !api.CheckArgs(1, ValueTNumber) {
		return 0
	}
	api.PushNumber(math.Sin(api.GetNumber(0)))
	return 1
}

func sinh(state *State) int {
	api := NewStackAPI(state)
	if !api.CheckArgs(1, ValueTNumber) {
		return 0
	}
	api.PushNumber(math.Sinh(api.GetNumber(0)))
	return 1
}

func sqrt(state *State) int {
	api := NewStackAPI(state)
	if !api.CheckArgs(1, ValueTNumber) {
		return 0
	}
	api.PushNumber(math.Sqrt(api.GetNumber(0)))
	return 1
}

func tan(state *State) int {
	api := NewStackAPI(state)
	if !api.CheckArgs(1, ValueTNumber) {
		return 0
	}
	api.PushNumber(math.Tan(api.GetNumber(0)))
	return 1
}

func tanh(state *State) int {
	api := NewStackAPI(state)
	if !api.CheckArgs(1, ValueTNumber) {
		return 0
	}
	api.PushNumber(math.Tanh(api.GetNumber(0)))
	return 1
}

func atan2(state *State) int {
	api := NewStackAPI(state)
	if !api.CheckArgs3(2, ValueTNumber, ValueTNumber) {
		return 0
	}
	api.PushNumber(math.Atan2(api.GetNumber(0), api.GetNumber(1)))
	return 1
}

func fmod(state *State) int {
	api := NewStackAPI(state)
	if !api.CheckArgs3(2, ValueTNumber, ValueTNumber) {
		return 0
	}
	api.PushNumber(math.Mod(api.GetNumber(0), api.GetNumber(1)))
	return 1
}

func ldexp(state *State) int {
	api := NewStackAPI(state)
	if !api.CheckArgs3(2, ValueTNumber, ValueTNumber) {
		return 0
	}
	api.PushNumber(math.Ldexp(api.GetNumber(0), int(api.GetNumber(1))))
	return 1
}

func pow(state *State) int {
	api := NewStackAPI(state)
	if !api.CheckArgs3(2, ValueTNumber, ValueTNumber) {
		return 0
	}
	api.PushNumber(math.Pow(api.GetNumber(0), api.GetNumber(1)))
	return 1
}

func deg(state *State) int {
	api := NewStackAPI(state)
	if !api.CheckArgs(1, ValueTNumber) {
		return 0
	}

	api.PushNumber(api.GetNumber(0) / math.Pi * 180)
	return 1
}

func rad(state *State) int {
	api := NewStackAPI(state)
	if !api.CheckArgs(1, ValueTNumber) {
		return 0
	}

	api.PushNumber(api.GetNumber(0) / 180 * math.Pi)
	return 1
}

func log(state *State) int {
	api := NewStackAPI(state)
	if !api.CheckArgs3(1, ValueTNumber, ValueTNumber) {
		return 0
	}

	l := math.Log(api.GetNumber(0))
	if api.GetStackSize() > 1 {
		b := math.Log(api.GetNumber(1))
		l /= b
	}

	api.PushNumber(l)
	return 1
}

func min(state *State) int {
	api := NewStackAPI(state)
	if !api.CheckArgs(1, ValueTNumber) {
		return 0
	}

	min := api.GetNumber(0)
	params := api.GetStackSize()
	for i := 1; i < params; i++ {
		if !api.IsNumber(i) {
			api.ArgTypeError(i, ValueTNumber)
			return 0
		}

		n := api.GetNumber(i)
		if n < min {
			min = n
		}
	}

	api.PushNumber(min)
	return 1
}

func max(state *State) int {
	api := NewStackAPI(state)
	if !api.CheckArgs(1, ValueTNumber) {
		return 0
	}

	max := api.GetNumber(0)
	params := api.GetStackSize()
	for i := 1; i < params; i++ {
		if !api.IsNumber(i) {
			api.ArgTypeError(i, ValueTNumber)
			return 0
		}

		n := api.GetNumber(i)
		if n > max {
			max = n
		}
	}

	api.PushNumber(max)
	return 1
}

func frexp(state *State) int {
	api := NewStackAPI(state)
	if !api.CheckArgs(1, ValueTNumber) {
		return 0
	}

	m, exp := math.Frexp(api.GetNumber(0))
	api.PushNumber(m)
	api.PushNumber(float64(exp))
	return 2
}

func modf(state *State) int {
	api := NewStackAPI(state)
	if !api.CheckArgs(1, ValueTNumber) {
		return 0
	}

	fpart, ipart := math.Modf(api.GetNumber(0))
	api.PushNumber(ipart)
	api.PushNumber(fpart)
	return 2
}

// Rand engine for math.random function
type randEngine struct {
}

type resultType uint

func random(state *State) int {
	// TODO
	api := NewStackAPI(state)
	if !api.CheckArgs3(0, ValueTNumber, ValueTNumber) {
		return 0
	}

	params := api.GetStackSize()
	if params == 0 {

	} else if params == 1 {

	} else if params == 2 {

	}

	return 1
}

func randomSeed(state *State) int {
	// TODO
	api := NewStackAPI(state)
	if !api.CheckArgs(1, ValueTNumber) {
		return 0
	}
	return 1
}

func RegisterLibMath(state *State) {
	lib := NewLibrary(state)
	libmath := []TableMemberReg{
		*NewTableMemberRegCFunction("abs", abs),
		*NewTableMemberRegCFunction("acos", acos),
		*NewTableMemberRegCFunction("asin", asin),
		*NewTableMemberRegCFunction("atan", atan),
		*NewTableMemberRegCFunction("atan2", atan2),
		*NewTableMemberRegCFunction("ceil", ceil),
		*NewTableMemberRegCFunction("cos", cos),
		*NewTableMemberRegCFunction("cosh", cosh),
		*NewTableMemberRegCFunction("deg", deg),
		*NewTableMemberRegCFunction("exp", exp),
		*NewTableMemberRegCFunction("floor", floor),
		*NewTableMemberRegCFunction("fmod", fmod),
		*NewTableMemberRegCFunction("frexp", frexp),
		*NewTableMemberRegCFunction("ldexp", ldexp),
		*NewTableMemberRegCFunction("log", log),
		*NewTableMemberRegCFunction("max", max),
		*NewTableMemberRegCFunction("min", min),
		*NewTableMemberRegCFunction("modf", modf),
		*NewTableMemberRegCFunction("pow", pow),
		*NewTableMemberRegCFunction("rad", rad),
		*NewTableMemberRegCFunction("random", random),
		*NewTableMemberRegCFunction("randomseed", randomSeed),
		*NewTableMemberRegCFunction("sin", sin),
		*NewTableMemberRegCFunction("sinh", sinh),
		*NewTableMemberRegCFunction("sqrt", sqrt),
		*NewTableMemberRegCFunction("tan", tan),
		*NewTableMemberRegCFunction("tanh", tanh),
		*NewTableMemberRegNumber("pi", math.Pi),
	}

	lib.RegisterTableFunction("math", &libmath[0], len(libmath))
}
