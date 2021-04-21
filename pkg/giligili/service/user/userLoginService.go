package user

import (
	"context"
	"selfText/giligili_back/libcommon/brick"
	"selfText/giligili_back/libcommon/logging"
	"selfText/giligili_back/libcommon/orm"
	"selfText/giligili_back/pkg/giligili/model"
	"selfText/giligili_back/pkg/giligili/protocal"
	"selfText/giligili_back/service/giligili/password"
	"selfText/giligili_back/service/giligili/serializer"
)

// UserLoginService 管理用户登录的服务
type UserLoginService struct {
	Config brick.Config           `inject:"config"`
	Orm    *orm.OrmService        `inject:"OrmService"`
	Logger *logging.LoggerService `inject:"LoggerService"`
}

// Login 用户登录函数
func (s *UserLoginService) Login(ctx context.Context, input protocal.LoginInput) (model.User, *serializer.Response) {
	var user model.User

	if err := s.Orm.GetDB().Where("user_name = ?", input.UserName).First(&user).Error; err != nil {
		return user, &serializer.Response{
			Status: 40001,
			Msg:    "账号或密码错误",
		}
	}

	if password.CheckPassword(input.Password, user.PasswordDigest) == false {
		return user, &serializer.Response{
			Status: 40001,
			Msg:    "账号或密码错误",
		}
	}
	return user, nil
}
