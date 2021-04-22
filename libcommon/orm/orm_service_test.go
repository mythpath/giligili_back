package orm_test

import (
	"context"
	"selfText/giligili_back/libcommon/logging"
	"testing"

	"selfText/giligili_back/libcommon/brick"
	"selfText/giligili_back/libcommon/orm"
	"selfText/giligili_back/libcommon/orm/dialects/mysql"
)

var ormSvc *orm.OrmService

func init() {
	c := brick.NewContainer()
	c.Add(&brick.JSONConfigService{}, "config", brick.FactoryFunc(func() interface{} {
		return brick.NewJSONConfigService("config/config.json")
	}))

	c.Add(&orm.OrmService{}, "OrmService", nil)
	c.Add(&mysql.MySQLService{}, "DB", nil)
	c.Add(&logging.LoggerService{}, "LoggerService", nil)
	c.Add(&TestModel{}, "TestEntity", nil)
	c.Build()
	ormSvc = c.GetByName("OrmService").(*orm.OrmService)
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

func TestCreate(t *testing.T) {
	entity := &TestEntity{Name: "obj1"}

	//Create
	ctx := context.WithValue(context.Background(), "User", "chaos")
	if err := ormSvc.CreateCtx(ctx, "TestEntity", entity); err != nil {
		t.Error(err)
	}
	if entity.CreatedBy != "chaos" {
		t.Error("CreatedBy is error")
	}
	if entity.UpdatedBy != "chaos" {
		t.Error("UpdatedBy is error")
	}
	if entity.DeletedBy != "" {
		t.Error("DeletedBy is error")
	}

	//Update
	ctx = context.WithValue(context.Background(), "User", "nerv")
	if err := ormSvc.UpdateCtx(ctx, "TestEntity", entity); err != nil {
		t.Error(err)
	}
	if entity.CreatedBy != "chaos" {
		t.Error("CreatedBy is error")
	}
	if entity.UpdatedBy != "nerv" {
		t.Error("UpdatedBy is error")
	}
	if entity.DeletedBy != "" {
		t.Error("DeletedBy is error")
	}

	ctx = context.WithValue(context.Background(), "User", "libnerv")
	if r, err := ormSvc.RemoveCtx(ctx, "TestEntity", entity.ID, true); err != nil {
		t.Error(err)
	} else {
		entity = r.(*TestEntity)
		if entity.CreatedBy != "chaos" {
			t.Error("CreatedBy is error")
		}
		if entity.UpdatedBy != "nerv" {
			t.Error("UpdatedBy is error")
		}
		// if entity.DeletedBy != "libner" {
		// 	t.Error("DeletedBy is error")
		// }
	}
}
