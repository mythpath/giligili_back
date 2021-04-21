package user

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

type UserService struct {
	Config brick.Config           `inject:"config"`
	Orm    *orm.OrmService        `inject:"OrmService"`
	Logger *logging.LoggerService `inject:"LoggerService"`

	SerializerService *serializer.SerializerService `inject:"SerializerService"`

	selectFields []string
}

func (u *UserService) Init() {
	u.selectFields = []string{"username", "created_by", "created_at", "updated_by", "updated_at",
		"password_digest", "nickname", "alias", "avatar"}
}

func (u *UserService) Create(ctx context.Context, input protocal.UserCreateInput) (serializer.Response, error) {
	var user model.User
	if err := util.DeepCopy(input, &user); err != nil {
		return serializer.Response{
			Status: 50001,
			Msg:    "用户信息保存失败",
			Error:  err.Error(),
		}, err
	}

	if err := u.Orm.CreateCtx(ctx, model.UserM, &user); err != nil {
		return serializer.Response{
			Status: 50001,
			Msg:    "用户信息保存失败",
			Error:  err.Error(),
		}, err
	}

	return serializer.Response{
		Data: u.SerializerService.BuildUser(user),
	}, nil
}

func (u *UserService) Delete(ctx context.Context, input protocal.UserDeleteInput) (serializer.Response, error) {
	var user model.User
	if err := u.Orm.GetDB().First(&user, input.ID).Error; err != nil {
		return serializer.Response{
			Status: 404,
			Msg:    "用户不存在",
			Error:  err.Error(),
		}, err
	}

	if _, err := u.Orm.RemoveCtx(ctx, model.UserM, user.ID, false); err != nil {
		return serializer.Response{
			Status: 50000,
			Msg:    "用户删除失败",
			Error:  err.Error(),
		}, err
	}

	return serializer.Response{
		Status: 200,
		Msg:    "删除成功",
	}, nil
}

func (u *UserService) Update(ctx context.Context, input protocal.UserUpdateInput) (serializer.Response, error) {
	var user model.User
	if err := u.Orm.GetDB().First(&user, input.ID).Error; err != nil {
		return serializer.Response{
			Status: 404,
			Msg:    "影片不存在",
			Error:  err.Error(),
		}, err
	}

	if err := util.DeepCopy(input, &user); err != nil {
		return serializer.Response{
			Status: 50002,
			Msg:    "影片保存失败",
			Error:  err.Error(),
		}, err
	}

	if err := u.Orm.UpdateCtx(ctx, model.UserM, &user); err != nil {
		return serializer.Response{
			Status: 50002,
			Msg:    "影片保存失败",
			Error:  err.Error(),
		}, err
	}

	return serializer.Response{
		Data: u.SerializerService.BuildUser(user),
	}, nil
}

func (u *UserService) Get(ctx context.Context, input protocal.UserGetInput) (serializer.Response, error) {
	var user model.User
	if err := u.Orm.GetDB().First(&user, input.ID).Error; err != nil {
		return serializer.Response{
			Status: 404,
			Msg:    "影片不存在",
			Error:  err.Error(),
		}, err
	}

	return serializer.Response{
		Data: u.SerializerService.BuildUser(user),
	}, nil
}

func (u *UserService) List(ctx context.Context, input protocal.ListInput) (serializer.Response, error) {
	var total int64

	if err := u.Orm.GetDB().Model(model.User{}).Count(&total).Error; err != nil {
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
			fields := u.selectFields[1:]
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

	listdata, err := u.Orm.List(model.UserM, u.selectFields, where, argv, input.Order, input.Page, input.PageSize)
	if err != nil {
		return serializer.Response{
			Status: 404,
			Msg:    "获取影片列表失败",
			Error:  fmt.Errorf("failed to list topic").Error(),
		}, err
	}

	return u.SerializerService.BuildListResponse(listdata, uint(total)), nil
}

// FuzzyQueryExist 用户模糊查询
func (u *UserService) FuzzyQueryExist(item interface{}) []uint {
	var users []model.User
	var userIDs []uint
	selectFields := []string{"username"}
	whereList := make([]string, len(selectFields))
	argv := make([]interface{}, len(selectFields))
	for i := 0; i < len(selectFields); i++ {
		whereList[i] = fmt.Sprintf("%s like binary '?'", selectFields[i])
		argv = append(argv, fmt.Sprintf("%%%s%%", item))
	}
	where := strings.Join(whereList, " or ")
	if err := u.Orm.GetDB().Model(model.User{}).Where(where, argv...).Find(&users).Error; err != nil {
		for _, user := range users {
			userIDs = append(userIDs, user.ID)
		}
	}
	return userIDs
}

// ExactlyQueryExistByID 根据ID精确查找用户
func (u *UserService) ExactlyQueryExistByID(id int) model.User {
	var user model.User
	if err := u.Orm.GetDB().Where("id = ?", id).First(&user).Error; err != nil {
		return user
	}
	return model.User{}
}
