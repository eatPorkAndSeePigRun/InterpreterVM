package luna

type String struct {
	GCObject
	inHeap    int32   // String in heap or not
	strBuffer string  // Buffer for short string
	str       *string // Pointer to heap which stored long string
	length    uint64  // Length of string
	hash_     int64   // Hash value of string
}

// Calculate hash of string
func (s String) hash(str string) {
	s.hash_ = 5381
	// TODO
}

func (s String) Accept(v GCObjectVisitor) {
	v.VisitString(&s)
}

func (s String) GetHash() int64 {
	return s.hash_
}

func (s String) GetLength() uint64 {
	return s.length
}

func (s String) GetCStr() string {
	if s.inHeap != 0 {
		return *s.str
	} else {
		return s.strBuffer
	}
}

// Convert to string
func (s String) GetStdString() string {
	// TODO
	//if s.inHeap != 0 {
	//	return (*s.str) * s.length
	//} else {
	//	return s.strBuffer * s.length
	//}
	return ""
}

// Change context of string
func (s String) SetValue(str string) {
	if s.inHeap != 0 {
		s.str = nil
	}
	s.length = uint64(len(str))
	// TODO
}
