package video

import (
	"context"
	"fmt"
	"selfText/giligili_back/libcommon/brick"
	"selfText/giligili_back/libcommon/logging"
	"selfText/giligili_back/libcommon/orm"
	"selfText/giligili_back/pkg/giligili/model"
	"selfText/giligili_back/pkg/giligili/protocal"
	"selfText/giligili_back/pkg/giligili/service/user"
	"selfText/giligili_back/service/common/util"
	"selfText/giligili_back/service/giligili/cache"
	"selfText/giligili_back/service/giligili/ossService"
	"selfText/giligili_back/service/giligili/serializer"
	"strconv"
	"strings"
)

type VideoService struct {
	Config brick.Config           `inject:"config"`
	Orm    *orm.OrmService        `inject:"OrmService"`
	Logger *logging.LoggerService `inject:"LoggerService"`

	UserService       *user.UserService             `inject:"UserService"`
	RedisService      *cache.RedisService           `inject:"RedisService"`
	OSSService        *ossService.OSSServcie        `inject:"OSSService"`
	SerializerService *serializer.SerializerService `inject:"SerializerService"`

	selectFields []string
}

func (v *VideoService) Init() {
	v.selectFields = []string{"id", "title", "created_by", "created_at", "updated_by", "updated_at",
		"info", "url", "avatar"}
}

// Create 创建视频
func (v *VideoService) Create(ctx context.Context, input protocal.VideoCreateInput) (serializer.Response, error) {
	var video model.Video
	if err := util.DeepCopy(input, &video); err != nil {
		return serializer.Response{
			Status: 50001,
			Msg:    "视频保存失败",
			Error:  err.Error(),
		}, err
	}

	if video.Avatar == "" {
		video.Avatar = v.OSSService.DefaultAvatarUrl()
	}

	if err := v.Orm.CreateCtx(ctx, model.VideoM, &video); err != nil {
		return serializer.Response{
			Status: 50001,
			Msg:    "视频保存失败",
			Error:  err.Error(),
		}, err
	}

	var relatedUser model.User
	if err := v.Orm.GetDB().Where("id = ?", video.UserID).First(&relatedUser).Error; err != nil {
		return serializer.Response{
			Status: 50001,
			Msg:    "用户不存在",
			Error:  err.Error(),
		}, err
	}

	return serializer.Response{
		Data: v.SerializerService.BuildVideo(video),
	}, nil
}

// Delete 删除视频
func (v *VideoService) Delete(ctx context.Context, input protocal.VideoDeleteInput) (serializer.Response, error) {
	var video model.Video
	err := v.Orm.GetDB().First(&video, input.ID).Error
	if err != nil {
		return serializer.Response{
			Status: 404,
			Msg:    "视频不存在",
			Error:  err.Error(),
		}, err
	}

	if _, err = v.Orm.RemoveCtx(ctx, model.VideoM, &video, false); err != nil {
		return serializer.Response{
			Status: 50000,
			Msg:    "视频删除失败",
			Error:  err.Error(),
		}, err
	}

	return serializer.Response{}, nil
}

// Update 更新视频
func (v *VideoService) Update(ctx context.Context, input protocal.VideoUpdateInput) (serializer.Response, error) {
	var video model.Video
	if err := v.Orm.GetDB().First(&video, input.ID).Error; err != nil {
		return serializer.Response{
			Status: 404,
			Msg:    "视频不存在",
			Error:  err.Error(),
		}, err
	}

	if err := util.DeepCopy(input, &video); err != nil {
		return serializer.Response{
			Status: 50002,
			Msg:    "视频保存失败",
			Error:  err.Error(),
		}, err
	}
	if err := v.Orm.UpdateCtx(ctx, model.VideoM, &video); err != nil {
		return serializer.Response{
			Status: 50002,
			Msg:    "视频保存失败",
			Error:  err.Error(),
		}, err
	}

	var relatedUser model.User
	if err := v.Orm.GetDB().Where("id = ?", video.UserID).First(&relatedUser).Error; err != nil {
		return serializer.Response{
			Status: 50001,
			Msg:    "用户不存在",
			Error:  err.Error(),
		}, err
	}

	return serializer.Response{
		Data: v.SerializerService.BuildVideo(video),
	}, nil
}

// Show 视频
func (v *VideoService) Show(ctx context.Context, input protocal.VideoGetInput) (serializer.Response, error) {
	var video model.Video
	err := v.Orm.GetDB().First(&video, input.ID).Error
	if err != nil {
		return serializer.Response{
			Status: 404,
			Msg:    "视频不存在",
			Error:  err.Error(),
		}, err
	}

	// 处理视频被观看的一系列问题
	v.AddView(input.ID)

	var relatedUser model.User
	if err := v.Orm.GetDB().Where("id = ?", video.UserID).First(&relatedUser).Error; err != nil {
		return serializer.Response{
			Status: 50001,
			Msg:    "用户不存在",
			Error:  err.Error(),
		}, err
	}

	return serializer.Response{
		Data: v.SerializerService.BuildVideo(video),
	}, nil
}

// List 视频列表
func (v *VideoService) List(ctx context.Context, input protocal.ListInput) (serializer.Response, error) {
	var total int64
	search := strings.TrimSpace(input.Search)
	selectFields := []string{"user_id", "title", "info"}
	searchValue := fmt.Sprintf("%%%s%%", search)
	where := ""
	whereList := make([]string, len(selectFields))
	argv := make([]interface{}, len(selectFields))

	if err := v.Orm.GetDB().Model(model.Video{}).Count(&total).Error; err != nil {
		return serializer.Response{
			Status: 50000,
			Msg:    "数据库连接错误",
			Error:  err.Error(),
		}, err
	}

	if search != "" {
		relatedUserIDs := v.UserService.FuzzyQueryExist(search)
		if len(relatedUserIDs) > 0 {
			relatedUserID := strings.Join(util.ConvertUint2String(relatedUserIDs), ",")
			whereList[0] = fmt.Sprintf("%s in (?)", selectFields[0])
			argv[0] = relatedUserID
		} else {
			whereList[0] = ""
			argv[0] = ""
		}
		for index := 1; index < len(selectFields); index++ {
			whereList[index] = fmt.Sprintf("%s like binary ?", selectFields[index])
			argv[index] = searchValue
		}

		if whereList[0] == "" {
			whereList = whereList[1:]
			argv = argv[1:]
		}
		where = strings.Join(whereList, " or ")
	} else if where == "" {
		where = "1 = 1"
	}

	listdata, err := v.Orm.List(model.VideoM, v.selectFields, where, argv, input.Order, input.Page, input.PageSize)
	if err != nil {
		return serializer.Response{
			Status: 500,
			Msg:    "获取影片列表失败",
			Error:  fmt.Errorf("failed to list topic").Error(),
		}, err
	}

	return v.SerializerService.BuildListResponse(listdata, uint(total)), nil
}

// AddView 视频浏览
func (v *VideoService) AddView(ID uint) {
	// 增加视频点击数
	v.RedisService.RedisClient.Incr(cache.VideoViewKey(ID))
	// 增加排行榜点击数
	v.RedisService.RedisClient.ZIncrBy(cache.DailyRankKey, 1, strconv.Itoa(int(ID)))
}
