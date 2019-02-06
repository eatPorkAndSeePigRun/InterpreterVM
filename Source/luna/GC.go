package luna

// Generational of GC object
const (
	GCGen0 = iota; // Youngest generation
	GCGen1 ;       // Mesozoic generation
	GCGen2 ;       // Oldest generation
)

// GC flag for mark GC object
const (
	GCFlagWhite = iota;
	GCFlagBlack;
)

// GC object type allocated by GC
const (
	GCObjectTypeTable = iota;
	GCObjectTypeFunction;
	GCObjectTypeClosure;
	GCObjectTypeUpValue;
	GCObjectTypeString;
	GCObjectTypeUserData;
)

type GCObject struct {
	next       *GCObject
	generation uint64
	gc         uint64
	gcObjType  uint64
}

func (gcObject GCObject) Accept() {

}
