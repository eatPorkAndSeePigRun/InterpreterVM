package luna

const ExpValueCountAny = -1

const (
	ValueTNil = iota;
	ValueTBool;
	ValueTNumber;
	ValueTObj;
	ValueTString;
	ValueTClosure;
	ValueTUpValue;
	ValueTTable;
	ValueTUserDate;
	ValueTCFunction;
)

type Value struct {
	Obj      *GCObject
	Str      *String
	Closure  *Closure
	UpValue  *UpValue
	Table    *Table
	UserDate *UserData
	//CFunc    *CFunctionType
	Num      float64
	BValue   bool

	Type	int64
}

func (value Value) SetNil()  {
	
}

func (value Value) SetBool(bvalue bool)  {
	
}

func (value Value) IsNil()  {
	
}

func (value Value) IsFalse() {
	
}

func (value Value) Accept()  {
	
}