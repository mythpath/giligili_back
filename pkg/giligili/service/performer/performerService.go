package performer

import (
	"context"
	"fmt"
	"selfText/giligili_back/libcommon/brick"
	"selfText/giligili_back/libcommon/logging"
	"selfText/giligili_back/libcommon/orm"
	"selfText/giligili_back/pkg/giligili/model"
	"selfText/giligili_back/pkg/giligili/protocal"
	"selfText/giligili_back/service/common/util"
	"selfText/giligili_back/service/giligili/serializer"
	"strings"
)

// PerformerService 演员相关服务
type PerformerService struct {
	Config brick.Config           `inject:"config"`
	Orm    *orm.OrmService        `inject:"OrmService"`
	Logger *logging.LoggerService `inject:"LoggerService"`

	SerializerService *serializer.SerializerService `inject:"SerializerService"`

	selectFields []string
}

func (p *PerformerService) Init() {
	p.selectFields = []string{"id", "name", "created_by", "created_at", "updated_by", "updated_at",
		"chinese", "gender", "locate", "avatar", "rank", "comment"}
}

// Create 新增演员信息
func (p *PerformerService) Create(ctx context.Context, input protocal.PerformerCreateInput) (serializer.Response, error) {
	performer := model.Performer{}
	if err := util.DeepCopy(input, &performer); err != nil {
		return serializer.Response{
			Status: 50001,
			Msg:    "演员信息保存失败",
			Error:  err.Error(),
		}, err
	}

	if err := p.Orm.CreateCtx(ctx, model.PerformerM, &performer); err != nil {
		return serializer.Response{
			Status: 50001,
			Msg:    "演员信息保存失败",
			Error:  err.Error(),
		}, err
	}

	return serializer.Response{
		Data: p.SerializerService.BuildPerformer(performer),
	}, nil
}

// Delete 删除演员信息
func (p *PerformerService) Delete(ctx context.Context, id string) (serializer.Response, error) {
	var performer model.Performer
	if err := p.Orm.GetDB().First(&performer, id).Error; err != nil {
		return serializer.Response{
			Status: 404,
			Msg:    "演员不存在",
			Error:  err.Error(),
		}, err
	}

	if _, err := p.Orm.RemoveCtx(ctx, model.PerformerM, performer.ID, false); err != nil {
		return serializer.Response{
			Status: 50000,
			Msg:    "演员删除失败",
			Error:  err.Error(),
		}, err
	}

	return serializer.Response{
		Status: 200,
		Data:   performer,
		Msg:    "删除成功",
	}, nil
}

// Update 修改演员信息
func (p *PerformerService) Update(ctx context.Context, input protocal.PerformerUpdateInput) (serializer.Response, error) {
	var performer model.Performer
	if err := p.Orm.GetDB().First(&performer, input.ID).Error; err != nil {
		return serializer.Response{
			Status: 404,
			Msg:    "演员不存在",
			Error:  err.Error(),
		}, err
	}

	if err := util.DeepCopy(input, &performer); err != nil {
		return serializer.Response{
			Status: 50002,
			Msg:    "演员更新失败",
			Error:  err.Error(),
		}, err
	}

	if err := p.Orm.UpdateCtx(ctx, model.PerformerM, &performer); err != nil {
		return serializer.Response{
			Status: 50002,
			Msg:    "演员更新失败",
			Error:  err.Error(),
		}, err
	}

	return serializer.Response{
		Data: p.SerializerService.BuildPerformer(performer),
	}, nil
}

// Get 获取单个演员信息
func (p *PerformerService) Get(ctx context.Context, input protocal.PerformerGetInput) (serializer.Response, error) {
	var performer model.Performer
	if err := p.Orm.GetDB().First(&performer, input.ID).Error; err != nil {
		return serializer.Response{
			Status: 404,
			Msg:    "演员不存在",
			Error:  err.Error(),
		}, err
	}

	return serializer.Response{
		Data: p.SerializerService.BuildPerformer(performer),
	}, nil
}

// List 获取演员列表
func (p *PerformerService) List(ctx context.Context, input protocal.ListInput) (serializer.Response, error) {
	var total int64
	if err := p.Orm.GetDB().Model(model.Film{}).Count(&total).Error; err != nil {
		return serializer.Response{
			Status: 50000,
			Msg:    "数据库连接错误",
			Error:  err.Error(),
		}, err
	}

	argv := make([]interface{}, 0, len(input.Values))
	for _, v := range input.Values {
		argv = append(argv, v)
	}
	where := strings.TrimSpace(input.Where)

	if input.Search != "" {
		whereF := func() string {
			fields := p.selectFields[1:]
			slen := len(fields)
			where := make([]string, slen)
			l := fmt.Sprintf("%%%s%%", input.Search)

			for i := 0; i < slen; i++ {

				where[i] = fmt.Sprintf("(%s LIKE BINARY '?')", fields[i])
				argv = append(argv, l)
			}

			return strings.Join(where, " OR ")
		}

		if where != "" {
			where = fmt.Sprintf("%s AND (%s)", where, whereF())
		} else {
			where = whereF()
		}
	} else if where == "" {
		where = "1 = 1"
	}

	listdata, err := p.Orm.List(model.PerformerM, p.selectFields, where, argv, input.Order, input.Page, input.PageSize)
	if err != nil {
		return serializer.Response{
			Status: 404,
			Msg:    "获取演员列表失败",
			Error:  fmt.Errorf("failed to list topic").Error(),
		}, err
	}

	return p.SerializerService.BuildListResponse(listdata, uint(total)), nil
}
