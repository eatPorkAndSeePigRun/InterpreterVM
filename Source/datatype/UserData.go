package datatype

type Destroyer func(*int32)

type UserData struct {
	gcObjectField
	userData  *int32    // Point to user data
	metaTable *Table    // MetaTable of user data
	destroyer Destroyer // User data destroyer, call it when user data destroy
	destroyed bool      // Whether user data destroyed
}

func NewUserData() *UserData {
	return &UserData{}
}

func (u *UserData) Accept(visitor GCObjectVisitor) {
	if visitor.VisitUserData(u) {
		u.metaTable.Accept(visitor)
	}
}

func (u *UserData) Set(userData *int32, metaTable *Table) {
	u.userData = userData
	u.metaTable = metaTable
}

func (u *UserData) SetDestroyer(destroyer Destroyer) {
	u.destroyer = destroyer
}

func (u *UserData) MarkDestroyed() {
	u.destroyed = true
}

func (u *UserData) GetData() *int32 {
	return u.userData
}

func (u *UserData) GetMetaTable() *Table {
	return u.metaTable
}
