package datatype

import (
	"container/list"
	"log"
	"os"
	"time"
)

// Generational of GC object
type GCGeneration uint64

const (
	GCGen0 = iota // Youngest generation
	GCGen1        // Mesozoic generation
	GCGen2        // Oldest generation
)

// GC flag for mark GC object
type GCFlag int64

const (
	GCFlagWhite = iota
	GCFlagBlack
)

// GC object type allocated by GC
type GCObjectType int64

const (
	GCObjectTypeTable = iota
	GCObjectTypeFunction
	GCObjectTypeClosure
	GCObjectTypeUpValue
	GCObjectTypeString
	GCObjectTypeUserData
)

// Visit for visit all GC objects
type GCObjectVisitor interface {
	// Need visit all GC object members when return true
	VisitTable(table *Table) bool
	VisitFunction(function *Function) bool
	VisitClosure(closure *Closure) bool
	VisitUpValue(value *UpValue) bool
	VisitString(str *String) bool
	VisitUserData(userData *UserData) bool
}

func (obj GCObject) Accept(visitor GCObjectVisitor) {

}

// Base class of GC object, GC use this class to manipulate all GC objects
type GCObject struct {
	next       *GCObject    // Pointing next GCObject in current generation
	generation GCGeneration // Generation flag
	gc         uint64       // GCFlag
	gcObjType  uint64       // GCObjectType
}

// GC object barrier checker
func CheckBarrier(obj *GCObject) bool {
	return obj.generation != GCGen0
}

func CheckBarrier_(gc *GC, obj *GCObject) {
	if CheckBarrier(obj) {
		gc.SetBarrier(obj)
	}
}

type GC struct {
	gen0 genInfo // Youngest generation
	gen1 genInfo // Mesozoic generation
	gen2 genInfo // Oldest generation

	minorTraveller RootTravelType // Minor root traveller
	majorTraveller RootTravelType // Major root traveller

	barriered list.List // Barriered GC objects

	objDeleter GCObjectDeleter // GC object Deleter
	logStream  *os.File        // Log file
}

func NewGC(deleter GCObjectDeleter, log bool) *GC {
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
	return &gc
}

const (
	kGen0InitThresholdCount = 512
	kGen1InitThresholdCount = 512
	kGen0MaxThresholdCount  = 2048
	kGen1MaxThresholdCount  = 102400
)

type genInfo struct {
	gen            *GCObject // Pointing to GC object list
	count          uint64    // Count of GC objects
	thresholdCount uint64    // Current threshold count of GC objects
}

func gcLog(msg string) {
	log.Println(msg)
}

func (gc GC) setObjectGen(obj *GCObject, gen GCGeneration) {
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

	obj.generation = gen
	obj.next = genInfo.gen
	genInfo.gen = obj
	genInfo.count++
}

// Run minor GC
func (gc GC) minorGC() {
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
func (gc GC) majorGC() {
	gc.majorGCMark()
	gc.majorGCSweep()
	clearList(&gc.barriered)
}

func (gc GC) minorGCMark() {
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
		if obj.generation != GCGen0 {
			panic("assert")
		}

		// Mark barriered objects, and visitor can visit member GC o
		obj.gc = GCFlagBlack
		obj.Accept(&barrieredMaker)
	}
}

func (gc GC) minorGCSweep() {
	// Sweep GCGen0
	for gc.gen0.gen != nil {
		obj := gc.gen0.gen
		gc.gen0.gen = gc.gen0.gen.next

		// Move object to GCGen1 generation when object is black
		if obj.gc == GCFlagBlack {
			obj.gc = GCFlagWhite
			obj.generation = GCGen1
			obj.next = gc.gen1.gen
			gc.gen1.count++
		} else {
			gc.objDeleter(obj, obj.gcObjType)
		}
	}

	gc.gen0.count = 0
}

func (gc GC) majorGCMark() {
	if gc.majorTraveller != nil {
		panic("assert")
	}

	// Visit all major GC root objects
	var marker majorMarkVisitor
	gc.majorTraveller(&marker)
}

func (gc GC) majorGCSweep() {
	// Sweep all generations
	gc.sweepGeneration(&gc.gen2)
	gc.sweepGeneration(&gc.gen1)
	gc.sweepGeneration(&gc.gen0)

	// Move all GCGen0 objects to GCGen1
	for gc.gen0.gen != nil {
		obj := gc.gen0.gen
		gc.gen0.gen = gc.gen0.gen.next

		obj.generation = GCGen1
		obj.next = gc.gen1.gen
		gc.gen1.gen = obj
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

func (gc GC) sweepGeneration(gen *genInfo) {
	var alived *GCObject
	for gen.gen != nil {
		obj := gen.gen
		gen.gen = obj.next

		if obj.gc == GCFlagBlack {
			obj.gc = GCFlagWhite
			obj.next = alived
			alived = obj
		} else {
			gc.objDeleter(obj, obj.gcObjType)
			gen.count--
		}
	}

	gen.gen = alived
}

// Adjust GenInfo's thresholdCount by alivedCount
func (gc GC) adjustThreshold(alivedCount uint64, gen *genInfo, minThreshold, maxThreshold uint64) {
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
func (gc GC) destroyGeneration(gen *genInfo) {
	for gen.gen != nil {
		obj := gen.gen
		gen.gen = gen.gen.next
		gc.objDeleter(obj, obj.gcObjType)
	}
	gen.count = 0
}

type RootTravelType func(GCObjectVisitor)
type GCObjectDeleter func(*GCObject, int)

type DefaultDeleter struct {
}

func (gc GC) ResetDeleter(objDeleter GCObjectDeleter) {
	gc.objDeleter = objDeleter
}

// Set minor and major root travel functions
func (gc GC) SetRootTraveller(minor, major RootTravelType) {
	gc.minorTraveller = minor
	gc.majorTraveller = major
}

// Alloc GC objects
func (gc GC) NewTAble(gen GCGeneration) *Table {
	t := Table{}
	t.gcObjType = GCObjectTypeTable
	gc.setObjectGen(&t.GCObject, gen)
	return &t
}

// Alloc GC objects
func (gc GC) NewFunction(gen GCGeneration) *Function {
	f := Function{}
	f.gcObjType = GCObjectTypeFunction
	gc.setObjectGen(&f.GCObject, gen)
	return &f
}

// Alloc GC objects
func (gc GC) NewClosure(gen GCGeneration) *Closure {
	c := Closure{}
	c.gcObjType = GCObjectTypeClosure
	gc.setObjectGen(&c.GCObject, gen)
	return &c
}

// Alloc GC objects
func (gc GC) NewUpValue(gen GCGeneration) *UpValue {
	u := UpValue{}
	u.gcObjType = GCObjectTypeUpValue
	gc.setObjectGen(&u.GCObject, gen)
	return &u
}

// Alloc GC objects
func (gc GC) NewString(gen GCGeneration) *String {
	s := String{}
	s.gcObjType = GCObjectTypeString
	gc.setObjectGen(&s.GCObject, gen)
	return &s
}

// Alloc GC objects
func (gc GC) NewUserData(gen GCGeneration) *UserData {
	u := UserData{}
	u.gcObjType = GCObjectTypeUserData
	gc.setObjectGen(&u.GCObject, gen)
	return &u
}

// Set Gc object barrier
func (gc GC) SetBarrier(obj *GCObject) {
	if obj.generation != GCGen0 {
		panic("assert")
	}
	gc.barriered.PushBack(obj)
}

// Check run GC
func (gc GC) CheckGC() {
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
		log.Print(gcName, "[", duration, "]")
		log.Print(gen0Count, " ", gen0Threshold, " | ")
		log.Print(gen1Count, " ", gen1Threshold, " | ")
		log.Print(gen2Count, " ", gen2Threshold, " | ")
		log.Print(gc.gen0.count, " ", gc.gen0.thresholdCount, " | ")
		log.Print(gc.gen1.count, " ", gc.gen1.thresholdCount, " | ")
		log.Print(gc.gen2.count, " ", gc.gen2.thresholdCount)
	}
}

type minorMarkVisitor struct {
}

func (minor minorMarkVisitor) visitObj(obj *GCObject) bool {
	if obj.generation == GCGen0 && obj.gc == GCFlagWhite {
		obj.gc = GCFlagBlack
		return true
	}
	return false
}

func (minor minorMarkVisitor) VisitTable(table *Table) bool {
	return minor.visitObj(&table.GCObject)
}

func (minor minorMarkVisitor) VisitFunction(function *Function) bool {
	return minor.visitObj(&function.GCObject)
}

func (minor minorMarkVisitor) VisitClosure(closure *Closure) bool {
	return minor.visitObj(&closure.GCObject)
}

func (minor minorMarkVisitor) VisitUpValue(value *UpValue) bool {
	return minor.visitObj(&value.GCObject)
}

func (minor minorMarkVisitor) VisitString(str *String) bool {
	return minor.visitObj(&str.GCObject)
}

func (minor minorMarkVisitor) VisitUserData(userData *UserData) bool {
	return minor.visitObj(&userData.GCObject)
}

type barrieredMarkVisitor struct {
}

func (bmv barrieredMarkVisitor) visitObj(obj *GCObject) bool {
	// Visit member GC objects of obj when it is barriered object
	if obj.generation == GCGen0 && obj.gc == GCFlagBlack {
		obj.gc = GCFlagWhite
		return true
	}
	// Visit GCGen0 generation object
	if obj.generation == GCGen0 && obj.gc == GCFlagWhite {
		obj.gc = GCFlagBlack
		return true
	}
	return false
}

func (bmv barrieredMarkVisitor) VisitTable(table *Table) bool {
	return bmv.visitObj(&table.GCObject)
}

func (bmv barrieredMarkVisitor) VisitFunction(function *Function) bool {
	return bmv.visitObj(&function.GCObject)
}

func (bmv barrieredMarkVisitor) VisitClosure(closure *Closure) bool {
	return bmv.visitObj(&closure.GCObject)
}

func (bmv barrieredMarkVisitor) VisitUpValue(value *UpValue) bool {
	return bmv.visitObj(&value.GCObject)
}

func (bmv barrieredMarkVisitor) VisitString(str *String) bool {
	return bmv.visitObj(&str.GCObject)
}

func (bmv barrieredMarkVisitor) VisitUserData(userData *UserData) bool {
	return bmv.visitObj(&userData.GCObject)
}

type majorMarkVisitor struct {
}

func (major majorMarkVisitor) visitObj(obj *GCObject) bool {
	if obj.gc == GCFlagWhite {
		obj.gc = GCFlagBlack
		return true
	}
	return false
}

func (major majorMarkVisitor) VisitTable(table *Table) bool {
	return major.visitObj(&table.GCObject)
}

func (major majorMarkVisitor) VisitFunction(function *Function) bool {
	return major.visitObj(&function.GCObject)
}

func (major majorMarkVisitor) VisitClosure(closure *Closure) bool {
	return major.visitObj(&closure.GCObject)
}

func (major majorMarkVisitor) VisitUpValue(value *UpValue) bool {
	return major.visitObj(&value.GCObject)
}

func (major majorMarkVisitor) VisitString(str *String) bool {
	return major.visitObj(&str.GCObject)
}

func (major majorMarkVisitor) VisitUserData(userData *UserData) bool {
	return major.visitObj(&userData.GCObject)
}

func clearList(l *list.List) {
	var next *list.Element
	for e := l.Front(); e != nil; e = next {
		next = e.Next()
		l.Remove(e)
	}
}
