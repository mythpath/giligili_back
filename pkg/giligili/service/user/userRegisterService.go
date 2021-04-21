package user

import (
	"context"
	"selfText/giligili_back/libcommon/brick"
	"selfText/giligili_back/libcommon/logging"
	"selfText/giligili_back/libcommon/orm"
	"selfText/giligili_back/pkg/giligili/model"
	"selfText/giligili_back/pkg/giligili/protocal"
	"selfText/giligili_back/service/giligili/alias"
	"selfText/giligili_back/service/giligili/password"
	"selfText/giligili_back/service/giligili/serializer"
)

// UserRegisterService 管理用户注册服务
type UserRegisterService struct {
	Config brick.Config           `inject:"config"`
	Orm    *orm.OrmService        `inject:"OrmService"`
	Logger *logging.LoggerService `inject:"LoggerService"`
}

// Valid 验证表单
func (s *UserRegisterService) Valid(ctx context.Context, input protocal.RegisterInput) *serializer.Response {
	if input.PasswordConfirm != input.Password {
		return &serializer.Response{
			Status: 40001,
			Msg:    "两次输入的密码不相同",
		}
	}

	var count int64
	if s.Orm.GetDB().Model(&model.User{}).Where("nickname = ?", input.Nickname).Count(&count); count > 0 {
		return &serializer.Response{
			Status: 40001,
			Msg:    "昵称被占用",
		}
	}

	count = 0
	if s.Orm.GetDB().Model(&model.User{}).Where("user_name = ?", input.UserName).Count(&count); count > 0 {
		return &serializer.Response{
			Status: 40001,
			Msg:    "用户名已经注册",
		}
	}

	return nil
}

// Register 用户注册
func (s *UserRegisterService) Register(ctx context.Context, input protocal.RegisterInput) (model.User, *serializer.Response) {
	// 表单验证
	if err := s.Valid(ctx, input); err != nil {
		return model.User{}, err
	}

	user := model.User{
		Nickname: input.Nickname,
		UserName: input.UserName,
		Status:   alias.Active,
	}

	// 加密密码
	if passwordEncrypted, err := password.SetPassword(input.Password, 0); err != nil {
		return user, &serializer.Response{
			Status: 40002,
			Msg:    "密码加密失败",
		}
	} else {
		user.PasswordDigest = passwordEncrypted
	}

	// 创建用户
	if err := s.Orm.CreateCtx(ctx, "User", &user); err != nil {
		return user, &serializer.Response{
			Status: 40002,
			Msg:    "注册失败",
			Error:  err.Error(),
		}
	}

	return user, nil
}
