package vm

const (
	OpTypeLoadNil      = iota + 1 // A    A: register
	OpTypeFillNil                 // AB   A: start reg B: end reg [A,B)
	OpTypeLoadBool                // AB   A: register B: 1 true 0 false
	OpTypeLoadInt                 // A    A: register Next instruction opcode is const unsigned int
	OpTypeLoadConst               // ABx  A: register Bx: const index
	OpTypeMove                    // AB   A: dst register B: src register
	OpTypeGetUpvalue              // AB   A: register B: upvalue index
	OpTypeSetUpvalue              // AB   A: register B: upvalue index
	OpTypeGetGlobal               // ABx  A: value register Bx: const index
	OpTypeSetGlobal               // ABx  A: value register Bx: const index
	OpTypeClosure                 // ABx  A: register Bx: proto index
	OpTypeCall                    // ABC  A: register B: arg value count + 1 C: expected result count + 1
	OpTypeVarArg                  // AsBx A: register sBx: expected result count
	OpTypeRet                     // AsBx A: return value start register sBx: return value count
	OpTypeJmpFalse                // AsBx A: register sBx: diff of instruction index
	OpTypeJmpTrue                 // AsBx A: register sBx: diff of instruction index
	OpTypeJmpNil                  // AsBx A: register sBx: diff of instruction index
	OpTypeJmp                     // sBx  sBx: diff of instruction index
	OpTypeNeg                     // A    A: operand register and dst register
	OpTypeNot                     // A    A: operand register and dst register
	OpTypeLen                     // A    A: operand register and dst register
	OpTypeAdd                     // ABC  A: dst register B: operand1 register C: operand2 register
	OpTypeSub                     // ABC  A: dst register B: operand1 register C: operand2 register
	OpTypeMul                     // ABC  A: dst register B: operand1 register C: operand2 register
	OpTypeDiv                     // ABC  A: dst register B: operand1 register C: operand2 register
	OpTypePow                     // ABC  A: dst register B: operand1 register C: operand2 register
	OpTypeMod                     // ABC  A: dst register B: operand1 register C: operand2 register
	OpTypeConcat                  // ABC  A: dst register B: operand1 register C: operand2 register
	OpTypeLess                    // ABC  A: dst register B: operand1 register C: operand2 register
	OpTypeGreater                 // ABC  A: dst register B: operand1 register C: operand2 register
	OpTypeEqual                   // ABC  A: dst register B: operand1 register C: operand2 register
	OpTypeUnEqual                 // ABC  A: dst register B: operand1 register C: operand2 register
	OpTypeLessEqual               // ABC  A: dst register B: operand1 register C: operand2 register
	OpTypeGreaterEqual            // ABC  A: dst register B: operand1 register C: operand2 register
	OpTypeNewTable                // A    A: register of table
	OpTypeSetTable                // ABC  A: register of table B: key register C: value register
	OpTypeGetTable                // ABC  A: register of table B: key register C: value register
	OpTypeForInit                 // ABC  A: var register B: limit register    C: step register
	OpTypeForStep                 // ABC  ABC same with OpType_ForInit, next instruction sBx: diff of instruction index
)

type Instruction struct {
	OpCode int
}

func newInstruction1() Instruction {
	return Instruction{}
}

func newInstruction2(opType, a, b, c int) Instruction {
	opCode := (opType << 24) | ((a & 0xFF) << 16) | ((b & 0xFF) << 8) | (c & 0xFF)
	return Instruction{opCode}
}

func newInstruction3(opType, a int, b int16) Instruction {
	opCode := (opType << 24) | ((a & 0xFF) << 16) | (int(b) & 0xFFFF)
	return Instruction{opCode}
}

func newInstruction4(opType, a int, b uint16) Instruction {
	opCode := (opType << 24) | ((a & 0xFF) << 16) | (int(b) & 0xFFFF)
	return Instruction{opCode}
}

func (i Instruction) RefillsBx(b int) {
	i.OpCode = (i.OpCode & 0xFFFF0000) | (b & 0xFFFF)
}

func GetOpCode(i Instruction) int {
	return (i.OpCode >> 24) & 0xFF
}

func GetParamA(i Instruction) int {
	return (i.OpCode >> 16) & 0xFF
}

func GetParamB(i Instruction) int {
	return (i.OpCode >> 8) & 0xFF
}

func GetParamC(instruction Instruction) int {
	return instruction.OpCode & 0xFF
}

func GetParamsBx(i Instruction) int16 {
	return int16(i.OpCode & 0xFFFF)
}

func GetParamBx(i Instruction) uint16 {
	return uint16(i.OpCode & 0xFFFF)
}

func ABCCode(opType, a, b, c int) Instruction {
	return newInstruction2(opType, a, b, c)
}

func ABCode(opType, a, b int) Instruction {
	return newInstruction2(opType, a, b, 0)
}

func ACode(opType, a int) Instruction {
	return newInstruction2(opType, a, 0, 0)
}

func AsBxCode(opType, a, b int) Instruction {
	return newInstruction3(opType, a, int16(b))
}

func ABxCode(opType, a, b int) Instruction {
	return newInstruction4(opType, a, uint16(b))
}
