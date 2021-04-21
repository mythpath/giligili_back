package ossService

import (
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"os"
	"selfText/giligili_back/libcommon/brick"
	"selfText/giligili_back/libcommon/logging"
	"selfText/giligili_back/libcommon/orm"
	"selfText/giligili_back/pkg/giligili/model"
)

type OSSServcie struct {
	Config brick.Config           `inject:"config"`
	Orm    *orm.OrmService        `inject:"OrmService"`
	Logger *logging.LoggerService `inject:"LoggerService"`
}

// AvatarURL 封面地址
func (o *OSSServcie) AvatarURL(ID uint) string {
	ossEndPoint := os.Getenv("OSS_END_POINT")
	ossAccessKeyId := os.Getenv("OSS_ACCESS_KEY_ID")
	ossAccessKeySecret := os.Getenv("OSS_ACCESS_KEY_SECRET")
	ossBucket := os.Getenv("OSS_BUCKET")
	defaultAvatar := os.Getenv("DEFAULT_AVATAR")

	if ossEndPoint == "" || ossAccessKeyId == "" || ossAccessKeySecret == "" || ossBucket == "" {
		return defaultAvatar
	}
	var video model.Video
	if err := o.Orm.GetDB().Where("id = ?", ID).First(&video); err != nil {
		return defaultAvatar
	}
	client, _ := oss.New(ossEndPoint, ossAccessKeyId, ossAccessKeySecret)
	bucket, _ := client.Bucket(ossBucket)

	if signedGetURL, err := bucket.SignURL(video.Avatar, oss.HTTPGet, 600); err == nil {
		return signedGetURL
	}

	return defaultAvatar
}

func (o *OSSServcie) DefaultAvatarUrl() string {
	defaultAvatar := o.Config.GetMapString("video", "defaultAvatar", "")

	return defaultAvatar
}
