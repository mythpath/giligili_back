package lock

import "selfText/giligili_back/libcommon/orm"

// LockService providers distruted lock on db
type LockService struct {
	DB            orm.DBService     `inject:"DB"`
	ModelRegistry orm.ModelRegistry `inject:"DB"`
}

func (p *LockService) AfterNew() {
	p.ModelRegistry.Put("Lock", LockDesc())
}

// GetLock returns a lock
func (p *LockService) GetLock(objType string, objID uint) *Lock {
	return &Lock{Type: objType, ObjID: objID, LockID: 1, db: p.DB}
}
