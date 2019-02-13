package luna

type OpType int64

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
	OpCode uint64
}

func (i Instruction) GetOpCode(instruction Instruction) int {
	return int((instruction.OpCode >> 24) & 0xff)
}
