package datatype

import (
	"container/list"
	"fmt"
	"os"
	"time"
)

// Generational of GC object
const (
	GCGen0 = iota + 1 // Youngest generation
	GCGen1            // Mesozoic generation
	GCGen2            // Oldest generation
)

// GC flag for mark GC object
const (
	GCFlagWhite = iota
	GCFlagBlack
)

// GC object type allocated by GC
const (
	GCObjectTypeTable = iota
	GCObjectTypeFunction
	GCObjectTypeClosure
	GCObjectTypeUpvalue
	GCObjectTypeString
	GCObjectTypeUserData
)

// Visit for visit all GC objects
type GCObjectVisitor interface {
	// Need visit all GC object members when return true
	VisitTable(table *Table) bool
	VisitFunction(function *Function) bool
	VisitClosure(closure *Closure) bool
	VisitUpvalue(value *Upvalue) bool
	VisitString(str *String) bool
	VisitUserData(userData *UserData) bool
}

// Base class of GC object, GC use this class to manipulate all GC objects
type GCObject interface {
	Accept(visitor GCObjectVisitor)
}

type gcObjectField struct {
	next       GCObject // Pointing next GCObject in current generation
	generation int      // Generation flag
	gc         int      // GCFlag
	gcObjType  int      // GCObjectType
}

func newGCObjectField() *gcObjectField {
	return &gcObjectField{generation: GCGen0}
}

// GC object barrier checker
func CheckBarrier(obj GCObject) bool {
	switch object := obj.(type) {
	case *Table:
		return object.generation != GCGen0
	case *Function:
		return object.generation != GCGen0
	case *Closure:
		return object.generation != GCGen0
	case *Upvalue:
		return object.generation != GCGen0
	case *String:
		return object.generation != GCGen0
	case *UserData:
		return object.generation != GCGen0
	default:
		panic("Unrecognizable data type")
	}
}

type GC struct {
	gen0 genInfo // Youngest generation
	gen1 genInfo // Mesozoic generation
	gen2 genInfo // Oldest generation

	minorTraveller RootTravelType // Minor root traveller
	majorTraveller RootTravelType // Major root traveller

	barriered list.List // Barriered GC objects, and its element.value is GCObject

	objDeleter GCObjectDeleter // GC object Deleter
	logStream  *os.File        // Log file
}

type RootTravelType func(GCObjectVisitor)
type GCObjectDeleter func(GCObject, int)

func NewGC(deleter GCObjectDeleter, log bool) GC {
	gc := GC{objDeleter: deleter}
	gc.gen0.thresholdCount = kGen0InitThresholdCount
	gc.gen1.thresholdCount = kGen1InitThresholdCount

	if log {
		f, err := os.OpenFile("gc.log", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			panic(err)
		}
		gc.logStream = f
	}
	return gc
}

func (gc *GC) ResetDeleter(objDeleter GCObjectDeleter) {
	gc.objDeleter = objDeleter
}

// Set minor and major root travel functions
func (gc *GC) SetRootTraveller(minor, major RootTravelType) {
	gc.minorTraveller = minor
	gc.majorTraveller = major
}

// Alloc GC objects
func (gc *GC) NewTAble(gen int) *Table {
	t := &Table{}
	t.gcObjType = GCObjectTypeTable
	gc.setObjectGen(t, gen)
	return t
}

// Alloc GC objects
func (gc *GC) NewFunction(gen int) *Function {
	f := &Function{}
	f.gcObjType = GCObjectTypeFunction
	gc.setObjectGen(f, gen)
	return f
}

// Alloc GC objects
func (gc *GC) NewClosure(gen int) *Closure {
	c := &Closure{}
	c.gcObjType = GCObjectTypeClosure
	gc.setObjectGen(c, gen)
	return c
}

// Alloc GC objects
func (gc *GC) NewUpvalue(gen int) *Upvalue {
	u := &Upvalue{}
	u.gcObjType = GCObjectTypeUpvalue
	gc.setObjectGen(u, gen)
	return u
}

// Alloc GC objects
func (gc *GC) NewString(gen int) *String {
	s := &String{}
	s.gcObjType = GCObjectTypeString
	gc.setObjectGen(s, gen)
	return s
}

// Alloc GC objects
func (gc *GC) NewUserData(gen int) *UserData {
	u := &UserData{}
	u.gcObjType = GCObjectTypeUserData
	gc.setObjectGen(u, gen)
	return u
}

// Set Gc object barrier
func (gc *GC) SetBarrier(obj GCObject) {
	gc.barriered.PushBack(obj)
}

// Check run GC
func (gc *GC) CheckGC() {
	if gc.gen0.count >= gc.gen0.thresholdCount {
		gen0Count := gc.gen0.count
		gen0Threshold := gc.gen0.thresholdCount
		gen1Count := gc.gen1.count
		gen1Threshold := gc.gen1.thresholdCount
		gen2Count := gc.gen2.count
		gen2Threshold := gc.gen2.thresholdCount

		var gcName string
		start := time.Now()
		if gc.gen1.count >= gc.gen1.thresholdCount {
			gcName = "major"
			gc.majorGC()
		} else {
			gcName = "minor"
			gc.minorGC()
		}

		duration := time.Since(start)
		_, err := fmt.Fprintf(gc.logStream, "%s[%v]: %d %d | %d %d | %d %d"+
			" - %d %d | %d %d | %d %d", gcName, duration, gen0Count, gen0Threshold,
			gen1Count, gen1Threshold, gen2Count, gen2Threshold, gc.gen0.count,
			gc.gen0.thresholdCount, gc.gen1.count, gc.gen1.thresholdCount,
			gc.gen2.count, gc.gen2.thresholdCount)
		if err != nil {
			panic(err)
		}
	}
}

const (
	kGen0InitThresholdCount = 512
	kGen1InitThresholdCount = 512
	kGen0MaxThresholdCount  = 2048
	kGen1MaxThresholdCount  = 102400
)

type genInfo struct {
	gen            GCObject // Pointing to GC object list
	count          uint     // Count of GC objects
	thresholdCount uint     // Current threshold count of GC objects
}

func newGenInfo() *genInfo {
	return &genInfo{}
}

func (gc *GC) setObjectGen(obj GCObject, gen int) {
	var genInfo *genInfo
	switch gen {
	case GCGen0:
		genInfo = &gc.gen0
	case GCGen1:
		genInfo = &gc.gen1
	case GCGen2:
		genInfo = &gc.gen2
	}

	if genInfo == nil {
		panic("assert")
	}

	switch object := obj.(type) {
	case *Table:
		object.generation = gen
		object.next = genInfo.gen
	case *Function:
		object.generation = gen
		object.next = genInfo.gen
	case *Closure:
		object.generation = gen
		object.next = genInfo.gen
	case *Upvalue:
		object.generation = gen
		object.next = genInfo.gen
	case *String:
		object.generation = gen
		object.next = genInfo.gen
	case *UserData:
		object.generation = gen
		object.next = genInfo.gen
	}
	genInfo.gen = obj
	genInfo.count++
}

// Run minor GC
func (gc *GC) minorGC() {
	oldGen1Count := gc.gen1.count

	gc.minorGCMark()
	gc.minorGCSweep()

	clearList(&gc.barriered)

	// Calculate objects count from gen0_ to gen1_, which is how many alived
	// objects in gen0_ after mark-sweep, and adjust gen0_'s threshold count
	// by the alivedGen0Count
	alivedGen0Count := gc.gen1.count - oldGen1Count
	gc.adjustThreshold(alivedGen0Count, &gc.gen0, kGen0InitThresholdCount,
		kGen0MaxThresholdCount)
}

// Run major GC
func (gc *GC) majorGC() {
	gc.majorGCMark()
	gc.majorGCSweep()
	clearList(&gc.barriered)
}

func (gc *GC) minorGCMark() {
	if gc.majorTraveller == nil {
		panic("assert")
	}

	// Visit all minor GC root objects
	var marker minorMarkVisitor
	gc.minorTraveller(&marker)

	// Visit all barriered GC objects
	var barrieredMaker barrieredMarkVisitor
	for e := gc.barriered.Front(); e != nil; e = e.Next() {
		// All barriered objects must be GCGen1 or GCGen2
		obj := e.Value.(GCObject)
		switch object := obj.(type) {
		case *Table:
			if object.generation == GCGen0 {
				panic("assert")
			}
			// Mark barriered objects, and visitor can visit member GC o
			object.gc = GCFlagBlack
			object.Accept(&barrieredMaker)
		case *Function:
			if object.generation == GCGen0 {
				panic("assert")
			}
			// Mark barriered objects, and visitor can visit member GC o
			object.gc = GCFlagBlack
			object.Accept(&barrieredMaker)
		case *Closure:
			if object.generation == GCGen0 {
				panic("assert")
			}
			// Mark barriered objects, and visitor can visit member GC o
			object.gc = GCFlagBlack
			object.Accept(&barrieredMaker)
		case *Upvalue:
			if object.generation == GCGen0 {
				panic("assert")
			}
			// Mark barriered objects, and visitor can visit member GC o
			object.gc = GCFlagBlack
			object.Accept(&barrieredMaker)
		case *String:
			if object.generation == GCGen0 {
				panic("assert")
			}
			// Mark barriered objects, and visitor can visit member GC o
			object.gc = GCFlagBlack
			object.Accept(&barrieredMaker)
		case *UserData:
			if object.generation == GCGen0 {
				panic("assert")
			}
			// Mark barriered objects, and visitor can visit member GC o
			object.gc = GCFlagBlack
			object.Accept(&barrieredMaker)
		default:
			panic("Unrecognizable data type")
		}

	}
}

func (gc *GC) minorGCSweep() {

	// Sweep GCGen0
	for gc.gen0.gen != nil {
		obj := gc.gen0.gen
		switch object := obj.(type) {
		case *Table:
			gc.gen0.gen = object.next

			// Move object to GCGen1 generation when object is black
			if object.gc == GCFlagBlack {
				object.gc = GCFlagWhite
				object.generation = GCGen1
				object.next = gc.gen1.gen
				gc.gen1.count++
			} else {
				gc.objDeleter(object, object.gcObjType)
			}
		case *Function:
			gc.gen0.gen = object.next

			// Move object to GCGen1 generation when object is black
			if object.gc == GCFlagBlack {
				object.gc = GCFlagWhite
				object.generation = GCGen1
				object.next = gc.gen1.gen
				gc.gen1.count++
			} else {
				gc.objDeleter(object, object.gcObjType)
			}
		case *Closure:
			gc.gen0.gen = object.next

			// Move object to GCGen1 generation when object is black
			if object.gc == GCFlagBlack {
				object.gc = GCFlagWhite
				object.generation = GCGen1
				object.next = gc.gen1.gen
				gc.gen1.count++
			} else {
				gc.objDeleter(object, object.gcObjType)
			}
		case *Upvalue:
			gc.gen0.gen = object.next

			// Move object to GCGen1 generation when object is black
			if object.gc == GCFlagBlack {
				object.gc = GCFlagWhite
				object.generation = GCGen1
				object.next = gc.gen1.gen
				gc.gen1.count++
			} else {
				gc.objDeleter(object, object.gcObjType)
			}
		case *String:
			gc.gen0.gen = object.next

			// Move object to GCGen1 generation when object is black
			if object.gc == GCFlagBlack {
				object.gc = GCFlagWhite
				object.generation = GCGen1
				object.next = gc.gen1.gen
				gc.gen1.count++
			} else {
				gc.objDeleter(object, object.gcObjType)
			}
		case *UserData:
			gc.gen0.gen = object.next

			// Move object to GCGen1 generation when object is black
			if object.gc == GCFlagBlack {
				object.gc = GCFlagWhite
				object.generation = GCGen1
				object.next = gc.gen1.gen
				gc.gen1.count++
			} else {
				gc.objDeleter(object, object.gcObjType)
			}
		default:
			panic("Unrecognizable data type")
		}
	}

	gc.gen0.count = 0
}

func (gc *GC) majorGCMark() {
	if gc.majorTraveller != nil {
		panic("assert")
	}

	// Visit all major GC root objects
	var marker majorMarkVisitor
	gc.majorTraveller(&marker)
}

func (gc *GC) majorGCSweep() {
	// Sweep all generations
	gc.sweepGeneration(&gc.gen2)
	gc.sweepGeneration(&gc.gen1)
	gc.sweepGeneration(&gc.gen0)

	// Move all GCGen0 objects to GCGen1
	for gc.gen0.gen != nil {
		obj := gc.gen0.gen
		switch object := obj.(type) {
		case *Table:
			gc.gen0.gen = object.next

			object.generation = GCGen1
			object.next = gc.gen1.gen
			gc.gen1.gen = object
		case *Function:
			gc.gen0.gen = object.next

			object.generation = GCGen1
			object.next = gc.gen1.gen
			gc.gen1.gen = object
		case *Closure:
			gc.gen0.gen = object.next

			object.generation = GCGen1
			object.next = gc.gen1.gen
			gc.gen1.gen = object
		case *Upvalue:
			gc.gen0.gen = object.next

			object.generation = GCGen1
			object.next = gc.gen1.gen
			gc.gen1.gen = object
		case *String:
			gc.gen0.gen = object.next

			object.generation = GCGen1
			object.next = gc.gen1.gen
			gc.gen1.gen = object
		case *UserData:
			gc.gen0.gen = object.next

			object.generation = GCGen1
			object.next = gc.gen1.gen
			gc.gen1.gen = object
		default:
			panic("Unrecognizable data type")
		}
	}

	// Adjust GCGen0 threshold count
	gc.adjustThreshold(gc.gen0.count, &gc.gen0, kGen0InitThresholdCount, kGen0MaxThresholdCount)

	gc.gen1.count += gc.gen0.count
	gc.gen0.count = 0

	// Adjust GCGen1 threshold count
	gc.adjustThreshold(gc.gen1.count, &gc.gen1, kGen1InitThresholdCount, kGen1MaxThresholdCount)
	if gc.gen1.count >= kGen1MaxThresholdCount {
		gc.gen1.thresholdCount = gc.gen1.count + kGen1MaxThresholdCount
	}
}

func (gc *GC) sweepGeneration(gen *genInfo) {
	var alived GCObject
	for gen.gen != nil {
		obj := gen.gen
		switch object := obj.(type) {
		case *Table:
			gen.gen = object.next

			if object.gc == GCFlagBlack {
				object.gc = GCFlagWhite
				object.next = alived
				alived = object
			} else {
				gc.objDeleter(object, object.gcObjType)
				gen.count--
			}
		case *Function:
			gen.gen = object.next

			if object.gc == GCFlagBlack {
				object.gc = GCFlagWhite
				object.next = alived
				alived = object
			} else {
				gc.objDeleter(object, object.gcObjType)
				gen.count--
			}
		case *Closure:
			gen.gen = object.next

			if object.gc == GCFlagBlack {
				object.gc = GCFlagWhite
				object.next = alived
				alived = object
			} else {
				gc.objDeleter(object, object.gcObjType)
				gen.count--
			}
		case *Upvalue:
			gen.gen = object.next

			if object.gc == GCFlagBlack {
				object.gc = GCFlagWhite
				object.next = alived
				alived = object
			} else {
				gc.objDeleter(object, object.gcObjType)
				gen.count--
			}
		case *String:
			gen.gen = object.next

			if object.gc == GCFlagBlack {
				object.gc = GCFlagWhite
				object.next = alived
				alived = object
			} else {
				gc.objDeleter(object, object.gcObjType)
				gen.count--
			}
		case *UserData:
			gen.gen = object.next

			if object.gc == GCFlagBlack {
				object.gc = GCFlagWhite
				object.next = alived
				alived = object
			} else {
				gc.objDeleter(object, object.gcObjType)
				gen.count--
			}
		default:
			panic("Unrecognizable data type")
		}
	}

	gen.gen = alived
}

// Adjust GenInfo's thresholdCount by alivedCount
func (gc *GC) adjustThreshold(alivedCount uint, gen *genInfo, minThreshold, maxThreshold uint) {
	if alivedCount != 0 {
		for gen.thresholdCount < 2*alivedCount {
			gen.thresholdCount *= 2
		}
		for gen.thresholdCount >= 4*alivedCount {
			gen.thresholdCount /= 2
		}
	}

	if gen.thresholdCount < minThreshold {
		gen.thresholdCount = minThreshold
	} else if gen.thresholdCount > maxThreshold {
		gen.thresholdCount = maxThreshold
	}
}

// Delete generation all objects
func (gc *GC) destroyGeneration(gen *genInfo) {
	for gen.gen != nil {
		obj := gen.gen
		switch object := obj.(type) {
		case *Table:
			gen.gen = object.next
			gc.objDeleter(obj, object.gcObjType)
		case *Function:
			gen.gen = object.next
			gc.objDeleter(obj, object.gcObjType)
		case *Closure:
			gen.gen = object.next
			gc.objDeleter(obj, object.gcObjType)
		case *Upvalue:
			gen.gen = object.next
			gc.objDeleter(obj, object.gcObjType)
		case *String:
			gen.gen = object.next
			gc.objDeleter(obj, object.gcObjType)
		case *UserData:
			gen.gen = object.next
			gc.objDeleter(obj, object.gcObjType)
		default:
			panic("Unrecognizable data type")
		}
	}
	gen.count = 0
}

type minorMarkVisitor struct {
}

func (minor *minorMarkVisitor) visitObj(obj GCObject) bool {
	switch object := obj.(type) {
	case *Table:
		if object.generation == GCGen0 && object.gc == GCFlagWhite {
			object.gc = GCFlagBlack
			return true
		}
	case *Function:
		if object.generation == GCGen0 && object.gc == GCFlagWhite {
			object.gc = GCFlagBlack
			return true
		}
	case *Closure:
		if object.generation == GCGen0 && object.gc == GCFlagWhite {
			object.gc = GCFlagBlack
			return true
		}
	case *Upvalue:
		if object.generation == GCGen0 && object.gc == GCFlagWhite {
			object.gc = GCFlagBlack
			return true
		}
	case *String:
		if object.generation == GCGen0 && object.gc == GCFlagWhite {
			object.gc = GCFlagBlack
			return true
		}
	case *UserData:
		if object.generation == GCGen0 && object.gc == GCFlagWhite {
			object.gc = GCFlagBlack
			return true
		}
	default:
		panic("Unrecognizable data type")
	}

	return false
}

func (minor *minorMarkVisitor) VisitTable(table *Table) bool {
	return minor.visitObj(table)
}

func (minor *minorMarkVisitor) VisitFunction(function *Function) bool {
	return minor.visitObj(function)
}

func (minor *minorMarkVisitor) VisitClosure(closure *Closure) bool {
	return minor.visitObj(closure)
}

func (minor *minorMarkVisitor) VisitUpvalue(value *Upvalue) bool {
	return minor.visitObj(value)
}

func (minor *minorMarkVisitor) VisitString(str *String) bool {
	return minor.visitObj(str)
}

func (minor *minorMarkVisitor) VisitUserData(userData *UserData) bool {
	return minor.visitObj(userData)
}

type barrieredMarkVisitor struct {
}

func (bmv *barrieredMarkVisitor) visitObj(obj GCObject) bool {
	switch object := obj.(type) {
	case *Table:
		// Visit member GC objects of obj when it is barriered object
		if object.generation == GCGen0 && object.gc == GCFlagBlack {
			object.gc = GCFlagWhite
			return true
		}
		// Visit GCGen0 generation object
		if object.generation == GCGen0 && object.gc == GCFlagWhite {
			object.gc = GCFlagBlack
			return true
		}
	case *Function:
		// Visit member GC objects of obj when it is barriered object
		if object.generation == GCGen0 && object.gc == GCFlagBlack {
			object.gc = GCFlagWhite
			return true
		}
		// Visit GCGen0 generation object
		if object.generation == GCGen0 && object.gc == GCFlagWhite {
			object.gc = GCFlagBlack
			return true
		}
	case *Closure:
		// Visit member GC objects of obj when it is barriered object
		if object.generation == GCGen0 && object.gc == GCFlagBlack {
			object.gc = GCFlagWhite
			return true
		}
		// Visit GCGen0 generation object
		if object.generation == GCGen0 && object.gc == GCFlagWhite {
			object.gc = GCFlagBlack
			return true
		}
	case *Upvalue:
		// Visit member GC objects of obj when it is barriered object
		if object.generation == GCGen0 && object.gc == GCFlagBlack {
			object.gc = GCFlagWhite
			return true
		}
		// Visit GCGen0 generation object
		if object.generation == GCGen0 && object.gc == GCFlagWhite {
			object.gc = GCFlagBlack
			return true
		}
	case *String:
		// Visit member GC objects of obj when it is barriered object
		if object.generation == GCGen0 && object.gc == GCFlagBlack {
			object.gc = GCFlagWhite
			return true
		}
		// Visit GCGen0 generation object
		if object.generation == GCGen0 && object.gc == GCFlagWhite {
			object.gc = GCFlagBlack
			return true
		}
	case *UserData:
		// Visit member GC objects of obj when it is barriered object
		if object.generation == GCGen0 && object.gc == GCFlagBlack {
			object.gc = GCFlagWhite
			return true
		}
		// Visit GCGen0 generation object
		if object.generation == GCGen0 && object.gc == GCFlagWhite {
			object.gc = GCFlagBlack
			return true
		}
	default:
		panic("Unrecognizable data type")
	}
	return false
}

func (bmv *barrieredMarkVisitor) VisitTable(table *Table) bool {
	return bmv.visitObj(table)
}

func (bmv *barrieredMarkVisitor) VisitFunction(function *Function) bool {
	return bmv.visitObj(function)
}

func (bmv *barrieredMarkVisitor) VisitClosure(closure *Closure) bool {
	return bmv.visitObj(closure)
}

func (bmv *barrieredMarkVisitor) VisitUpvalue(value *Upvalue) bool {
	return bmv.visitObj(value)
}

func (bmv *barrieredMarkVisitor) VisitString(str *String) bool {
	return bmv.visitObj(str)
}

func (bmv *barrieredMarkVisitor) VisitUserData(userData *UserData) bool {
	return bmv.visitObj(userData)
}

type majorMarkVisitor struct {
}

func (major *majorMarkVisitor) visitObj(obj GCObject) bool {
	switch object := obj.(type) {
	case *Table:
		if object.gc == GCFlagWhite {
			object.gc = GCFlagBlack
			return true
		}
	case *Function:
		if object.gc == GCFlagWhite {
			object.gc = GCFlagBlack
			return true
		}
	case *Closure:
		if object.gc == GCFlagWhite {
			object.gc = GCFlagBlack
			return true
		}
	case *Upvalue:
		if object.gc == GCFlagWhite {
			object.gc = GCFlagBlack
			return true
		}
	case *String:
		if object.gc == GCFlagWhite {
			object.gc = GCFlagBlack
			return true
		}
	case *UserData:
		if object.gc == GCFlagWhite {
			object.gc = GCFlagBlack
			return true
		}
		panic("Unrecognizable data type")
	}

	return false
}

func (major *majorMarkVisitor) VisitTable(table *Table) bool {
	return major.visitObj(table)
}

func (major *majorMarkVisitor) VisitFunction(function *Function) bool {
	return major.visitObj(function)
}

func (major *majorMarkVisitor) VisitClosure(closure *Closure) bool {
	return major.visitObj(closure)
}

func (major *majorMarkVisitor) VisitUpvalue(value *Upvalue) bool {
	return major.visitObj(value)
}

func (major *majorMarkVisitor) VisitString(str *String) bool {
	return major.visitObj(str)
}

func (major *majorMarkVisitor) VisitUserData(userData *UserData) bool {
	return major.visitObj(userData)
}

func clearList(l *list.List) {
	var next *list.Element
	for e := l.Front(); e != nil; e = next {
		next = e.Next()
		l.Remove(e)
	}
}
