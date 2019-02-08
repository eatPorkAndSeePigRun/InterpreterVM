package main

import (
	"InterpreterVM/Source/lib/base"
	"InterpreterVM/Source/lib/io"
	"InterpreterVM/Source/lib/math"
	string2 "InterpreterVM/Source/lib/string"
	"InterpreterVM/Source/lib/table"
	"InterpreterVM/Source/luna"
	"fmt"
	"os"
)

func repl(state luna.State) {
	fmt.Println("Luna 2.0 Copyright (C) 2014")

	for {
		fmt.Println("> ")

		var buffer [1024]byte
		n, err := os.Stdin.Read(buffer[:])
		if err != nil {
			fmt.Println(err)	// TODO
		}else if n == 0 {
			break
		}

		state.DoString(string(buffer[:]), "stdin")
	}
}

func executeFile(args []string, state luna.State) {
	state.DoModule(args[1])
}

func main() {
	var state luna.State

	base.RegisterLibBase(&state)
	io.RegisterLibIO(&state)
	math.RegisterLibMath(&state)
	string2.RegisterLibString(&state)
	table.RegisterLibTable(&state)

	if len(os.Args) < 2 {
		repl(state)
	} else {
		executeFile(os.Args, state)
	}


}
