package lock

import (
	"errors"
	"gorm.io/gorm"
	"selfText/giligili_back/libcommon/orm"
)

func LockDesc() *orm.ModelDescriptor {
	return &orm.ModelDescriptor{
		Type: &Lock{},
		New: func() interface{} {
			return &Lock{}
		},
		NewSlice: func() interface{} {
			return &[]Lock{}
		},
	}
}

//Lock for mutex
type Lock struct {
	Type   string        `gorm:"unique_index:idx_lock_tol"` //object type
	ObjID  uint          `gorm:"unique_index:idx_lock_tol"` //object id
	LockID uint          `gorm:"unique_index:idx_lock_tol"` //lock id
	db     orm.DBService `gorm:"-"`
}

func GetLock(objType string, objID uint, db orm.DBService) *Lock {
	return &Lock{Type: objType, ObjID: objID, LockID: 1, db: db}
}

//TryLock return true if the lock has been acquired
func (p *Lock) TryLock() bool {

	if err := p.db.GetDB().Create(p).Error; err != nil {
		return false
	}
	return true
}

//Unlock release the lock.Do nothing if no lock
func (p *Lock) Unlock() {
	p.db.GetDB().Where("type=? and obj_id=? and lock_id=?", p.Type, p.ObjID, p.LockID).Delete(p)
}

// Locked return true if the lock has been locked
func (p *Lock) Locked() (bool, error) {
	db := p.db.GetDB().Where("type=? and obj_id=? and lock_id=?", p.Type, p.ObjID, p.LockID).First(p)
	if errors.Is(db.Error, gorm.ErrRecordNotFound) {
		return false, nil
	}

	if db.Error != nil {
		return false, db.Error
	}

	return true, nil
}
