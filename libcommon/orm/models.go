package orm

import "time"

type SelfGormModel struct {
	ID        uint   `gorm:"primary_key"`
	CreatedBy string `gorm:"size:50"`
	CreatedAt time.Time
	UpdatedBy string `gorm:"size:50"`
	UpdatedAt time.Time
	DeletedBy string     `gorm:"size:50"`
	DeletedAt *time.Time `sql:"index"`
}

type ModelDescriptor struct {
	Type     interface{}
	New      func() interface{}
	NewSlice func() interface{}
}

// ModelRegistry Register all orm models
type ModelRegistry interface {
	Put(name string, model *ModelDescriptor)
	Get(name string) *ModelDescriptor
	Models() <-chan *ModelDescriptor
}

type ModelRegistryImpl struct {
	models map[string]*ModelDescriptor
}

func (p *ModelRegistryImpl) Put(name string, model *ModelDescriptor) {
	if p.models == nil {
		p.models = map[string]*ModelDescriptor{}
	}
	p.models[name] = model
}

func (p *ModelRegistryImpl) Get(name string) *ModelDescriptor {
	if p.models == nil {
		return nil
	} else {
		return p.models[name]
	}
}

func (p *ModelRegistryImpl) Models() <-chan *ModelDescriptor {
	ch := make(chan *ModelDescriptor)
	go func() {
		for _, m := range p.models {
			ch <- m
		}
		close(ch)
	}()
	return ch
}
