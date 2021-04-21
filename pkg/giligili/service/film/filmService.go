package film

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

type FilmService struct {
	Config brick.Config           `inject:"config"`
	Orm    *orm.OrmService        `inject:"OrmService"`
	Logger *logging.LoggerService `inject:"LoggerService"`

	SerializerService *serializer.SerializerService `inject:"SerializerService"`

	selectFields []string
}

func (f *FilmService) Init() {
	f.selectFields = []string{"id", "name", "created_by", "created_at", "updated_by", "updated_at",
		"code", "performer_id", "post", "rank", "comment"}
}

// Create 新增演员信息
func (f *FilmService) Create(ctx context.Context, input protocal.FilmCreateInput) (serializer.Response, error) {
	film := model.Film{}
	if err := util.DeepCopy(input, &film); err != nil {
		return serializer.Response{
			Status: 50001,
			Msg:    "影片信息保存失败",
			Error:  err.Error(),
		}, err
	}

	if err := f.Orm.CreateCtx(ctx, model.FilmM, &film); err != nil {
		return serializer.Response{
			Status: 50001,
			Msg:    "影片信息保存失败",
			Error:  err.Error(),
		}, err
	}

	return serializer.Response{
		Data: f.SerializerService.BuildFilm(film),
	}, nil
}

// Delete 删除演员信息
func (f *FilmService) Delete(ctx context.Context, input protocal.FilmDeleteInput) (serializer.Response, error) {
	var film model.Film
	if err := f.Orm.GetDB().First(&film, input.ID).Error; err != nil {
		return serializer.Response{
			Status: 404,
			Msg:    "影片不存在",
			Error:  err.Error(),
		}, err
	}

	if _, err := f.Orm.RemoveCtx(ctx, model.FilmM, film.ID, false); err != nil {
		return serializer.Response{
			Status: 50000,
			Msg:    "影片删除失败",
			Error:  err.Error(),
		}, err
	}

	return serializer.Response{}, nil
}

// Update 修改演员信息
func (f *FilmService) Update(ctx context.Context, input protocal.FilmUpdateInput) (serializer.Response, error) {
	var film model.Film
	if err := f.Orm.GetDB().First(&film, input.ID).Error; err != nil {
		return serializer.Response{
			Status: 404,
			Msg:    "影片不存在",
			Error:  err.Error(),
		}, err
	}

	if err := util.DeepCopy(input, &film); err != nil {
		return serializer.Response{
			Status: 50002,
			Msg:    "影片保存失败",
			Error:  err.Error(),
		}, err
	}

	if err := f.Orm.UpdateCtx(ctx, model.FilmM, &film); err != nil {
		return serializer.Response{
			Status: 50002,
			Msg:    "影片保存失败",
			Error:  err.Error(),
		}, err
	}

	return serializer.Response{
		Data: f.SerializerService.BuildFilm(film),
	}, nil
}

// Get 获取单个演员信息
func (f *FilmService) Get(ctx context.Context, input protocal.FilmGetInput) (serializer.Response, error) {
	var film model.Film
	if err := f.Orm.GetDB().First(&film, input.ID).Error; err != nil {
		return serializer.Response{
			Status: 404,
			Msg:    "影片不存在",
			Error:  err.Error(),
		}, err
	}

	return serializer.Response{
		Data: f.SerializerService.BuildFilm(film),
	}, nil
}

// List 获取演员列表
func (f *FilmService) List(ctx context.Context, input protocal.ListInput) (serializer.Response, error) {
	var total int64

	if err := f.Orm.GetDB().Model(model.Film{}).Count(&total).Error; err != nil {
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
			fields := f.selectFields[1:]
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

	listdata, err := f.Orm.List(model.FilmM, f.selectFields, where, argv, input.Order, input.Page, input.PageSize)
	if err != nil {
		return serializer.Response{
			Status: 404,
			Msg:    "获取影片列表失败",
			Error:  fmt.Errorf("failed to list topic").Error(),
		}, err
	}

	return f.SerializerService.BuildListResponse(listdata, uint(total)), nil
}
