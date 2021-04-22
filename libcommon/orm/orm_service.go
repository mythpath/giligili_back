package orm

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"strings"
)

// InstanceNotFoundError xxx
type InstanceNotFoundError struct {
	msg string
}

func newInstanceNotFoundError(msg string) *InstanceNotFoundError {
	return &InstanceNotFoundError{msg: msg}
}

func (p *InstanceNotFoundError) Error() string {
	return p.msg
}

type OrmService struct {
	DB            DBService     `inject:"DB"`
	ModelRegistry ModelRegistry `inject:"DB"`
}

func (p *OrmService) GetDB() *gorm.DB {
	return p.DB.GetDB()
}
func (p *OrmService) Get(class string, id interface{}, ass string) (interface{}, error) {
	md := p.ModelRegistry.Get(class)
	if md == nil {
		return nil, fmt.Errorf("class %s isn't exists", class)
	}

	data := md.New()
	d := p.DB.GetDB()
	if ass != "" {
		for _, as := range strings.Split(ass, ",") {
			d = d.Preload(as)
		}
	}

	if errors.Is(d.First(data, id).Error, gorm.ErrRecordNotFound) {
		return nil, newInstanceNotFoundError(fmt.Sprintf("could not found instance %v of class %s", id, class))
	}

	return data, nil
}

func (p *OrmService) List(class string, selectFields []string, where string, whereValues []interface{}, order string, page int, pageSize int) (map[string]interface{}, error) {
	md := p.ModelRegistry.Get(class)
	if md == nil {
		return nil, fmt.Errorf("class %s isn't exists", class)

	}

	//count
	d := p.DB.GetDB().Model(md.NewSlice())
	var count int64
	where, err := ParseSQL(where)
	if err != nil {
		return nil, fmt.Errorf("where sql condition is error:%s", err.Error())
	}
	if where != "" {
		d = d.Where(where, whereValues...)
	}
	if len(selectFields) > 0 {
		d = d.Select(selectFields)
	}
	if err := d.Count(&count).Error; err != nil {
		return nil, err
	}
	//order page
	var pageCount, limit int
	if pageSize > 0 {
		limit = pageSize
	} else {
		limit = 10
	}
	pageCount = (int(count) + limit - 1) / limit

	if page < 0 {
		page = 0
	}

	if page >= pageCount {
		page = pageCount - 1
	}

	if where != "" {
		d = p.DB.GetDB().Where(where, whereValues...)
	}
	if len(selectFields) > 0 {
		d = d.Select(selectFields)
	}
	if order != "" {
		if ok, _ := IsRightOrder(order); !ok {
			return nil, fmt.Errorf("order sql condition is wrong:%s", order)
		}
		d = d.Order(order)
	}
	data := md.NewSlice()
	if errors.Is(d.Offset(page*limit).Limit(limit).Find(data).Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return map[string]interface{}{"data": data, "page": page, "pageSize": limit, "pageCount": pageCount}, nil
}

func (p *OrmService) CreateCtx(ctx context.Context, class string, data interface{}) error {
	//if err := p.DB.GetDB().CreateCtx(ctx, data).Error; err != nil {
	//	return err
	//}

	if err := p.DB.GetDB().WithContext(ctx).Create(data).Error; err != nil {
		return err
	}

	return nil
}

func (p *OrmService) Create(class string, data interface{}) error {
	if err := p.DB.GetDB().Create(data).Error; err != nil {
		return err
	}

	//p.watcher.NotifyCreate(class(class), data)
	return nil
}

type classType string

func (p *OrmService) RemoveCtx(ctx context.Context, className string, id interface{}, soft bool) (interface{}, error) {
	md := p.ModelRegistry.Get(className)
	if md == nil {
		return nil, fmt.Errorf("class %s isn't exists", className)
	}
	data := md.New()
	if err := p.DB.GetDB().First(data, id).Error; err != nil {
		return nil, err
	}

	if !soft {
		if err := p.DB.GetDB().Unscoped().Delete(data).Error; err != nil {
			return nil, err
		}
	} else {
		//if err := p.DB.GetDB().DeleteCtx(ctx, data).Error; err != nil {
		//	return nil, err
		//}
		if err := p.DB.GetDB().WithContext(ctx).Delete(data).Error; err != nil {
			return nil, err
		}
	}

	//p.watcher.NotifyDelete(class(className), id)
	return data, nil
}

func (p *OrmService) Remove(className string, id interface{}) (interface{}, error) {
	md := p.ModelRegistry.Get(className)
	if md == nil {
		return nil, fmt.Errorf("class %s isn't exists", className)
	}
	data := md.New()
	if err := p.DB.GetDB().First(data, id).Error; err != nil {
		return nil, err
	}

	if err := p.DB.GetDB().Unscoped().Delete(data).Error; err != nil {
		return nil, err
	}

	//p.watcher.NotifyDelete(class(className), id)
	return data, nil
}

func (p *OrmService) UpdateCtx(ctx context.Context, className string, data interface{}) error {
	//if err := p.DB.GetDB().SaveCtx(ctx, data).Error; err != nil {
	//	return err
	//}
	if err := p.DB.GetDB().WithContext(ctx).Save(data).Error; err != nil {
		return err
	}
	//p.watcher.NotifyUpdate(class(className), data)
	return nil
}

func (p *OrmService) Update(className string, data interface{}) error {
	if err := p.DB.GetDB().Save(data).Error; err != nil {
		return err
	}
	//p.watcher.NotifyUpdate(class(className), data)
	return nil
}

// func (p *OrmService) WatchObject(wk int64, timeout int, className string, id interface{}, associations string) (int64, chan *WatchEvent) {
// 	return p.watcher.Wait(wk, timeout, class(className), id)
// }
