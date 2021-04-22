package sqlite

import (
	"selfText/giligili_back/libcommon/brick"
	"selfText/giligili_back/libcommon/orm"
	"testing"
	"time"
)

var sqldb *SQLite3Service

func init() {
	c := brick.NewContainer()
	c.Add(&brick.JSONConfigService{}, "config", brick.FactoryFunc(func() interface{} {
		return brick.NewJSONConfigService("../../config/sqlite.json")
	}))

	c.Add(&SQLite3Service{}, "DB", nil)
	c.Add(&TestModel{}, "TestEntity", nil)
	c.Build()
	sqldb = c.GetByName("DB").(*SQLite3Service)
}

type TestEntity struct {
	orm.SelfGormModel
	Name string
}

type TestModel struct {
	ModelRegistry orm.ModelRegistry `inject:"DB"`
}

func (p *TestModel) AfterNew() {
	p.ModelRegistry.Put("TestEntity", p.desc())
}

func (p *TestModel) desc() *orm.ModelDescriptor {
	return &orm.ModelDescriptor{
		Type: &TestEntity{},
		New: func() interface{} {
			return &TestEntity{}
		},
		NewSlice: func() interface{} {
			return &[]TestEntity{}
		},
	}
}

func TestDatabaseIsLocked(t *testing.T) {
	entity := &TestEntity{
		Name: "test-entity",
	}

	if err := sqldb.GetDB().Create(entity).Error; err != nil {
		t.Fatalf("failed to create the entity: %v", err)
	}

	stopC := make(chan struct{}, 1)
	go func() {
		defer close(stopC)
		ticker := time.NewTicker(1 * time.Millisecond)
		for {
			select {
			case <-ticker.C:
				result := &TestEntity{}
				if err := sqldb.GetDB().Where("id = ?", entity.ID).First(result).Error; err != nil {
					t.Fatalf("failed to first the entity: %v", err)
				}
			}
		}
	}()

	go func() {
		defer close(stopC)
		ticker := time.NewTicker(1 * time.Millisecond)
		for {
			select {
			case <-ticker.C:
				result := &TestEntity{
					Name: "hello",
				}
				if err := sqldb.GetDB().Create(result).Error; err != nil {
					t.Fatalf("failed to create the entity: %v", err)
				}
			}
		}
	}()
	<-stopC
}

func TestDatabaseIsUnlocked(t *testing.T) {
	entity := &TestEntity{
		Name: "test-entity",
	}

	//sqldb.GetDB().DB().SetMaxOpenConns(1)
	if db, err := sqldb.GetDB().DB(); err != nil {
		t.Error("failed to set max open connections")
	} else {
		db.SetMaxOpenConns(1)
	}

	if err := sqldb.GetDB().Create(entity).Error; err != nil {
		t.Fatalf("failed to create the entity: %v", err)
	}

	stopC := make(chan struct{}, 1)
	go func() {
		defer close(stopC)
		ticker := time.NewTicker(1 * time.Millisecond)
		for {
			select {
			case <-ticker.C:
				result := &TestEntity{}
				if err := sqldb.GetDB().Where("id = ?", entity.ID).First(result).Error; err != nil {
					t.Fatalf("failed to first the entity: %v", err)
				}
			}
		}
	}()

	go func() {
		defer close(stopC)
		ticker := time.NewTicker(1 * time.Millisecond)
		for {
			select {
			case <-ticker.C:
				result := &TestEntity{
					Name: "hello",
				}
				if err := sqldb.GetDB().Create(result).Error; err != nil {
					t.Fatalf("failed to create the entity: %v", err)
				}
			}
		}
	}()
	<-stopC
}
