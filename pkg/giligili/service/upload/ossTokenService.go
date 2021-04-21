package upload

import (
	"context"
	"selfText/giligili_back/libcommon/brick"
	"selfText/giligili_back/libcommon/logging"
	"selfText/giligili_back/libcommon/orm"
	"selfText/giligili_back/pkg/giligili/protocal"
	"selfText/giligili_back/service/giligili/serializer"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/google/uuid"
)

// OSSTokenService 获得上传oss token的服务
type OSSTokenService struct {
	Config brick.Config           `inject:"config"`
	Orm    *orm.OrmService        `inject:"OrmService"`
	Logger *logging.LoggerService `inject:"LoggerService"`
}

// Post 创建token
func (o *OSSTokenService) Post(ctx context.Context, input protocal.OSSTokenInput) (serializer.Response, error) {
	ossEndpoint := o.Config.GetMapString("OSS", "endpoint", "")
	ossAccessKeyID := o.Config.GetMapString("OSS", "accessKeyID", "")
	ossAccessKeySecret := o.Config.GetMapString("OSS", "accessKeySecret", "")
	ossBucket := o.Config.GetMapString("OSS", "bucket", "")

	client, err := oss.New(ossEndpoint, ossAccessKeyID, ossAccessKeySecret)
	if err != nil {
		return serializer.Response{
			Status: 50002,
			Data:   nil,
			Msg:    "OSS配置错误",
			Error:  err.Error(),
		}, err
	}

	// 获取存储空间
	bucket, err := client.Bucket(ossBucket)
	if err != nil {
		return serializer.Response{
			Status: 50002,
			Data:   nil,
			Msg:    "OSS配置错误",
			Error:  err.Error(),
		}, err
	}

	// 带可选参数的签名直传
	options := []oss.Option{
		oss.ContentType("image/png"),
	}

	key := "upload/" + input.Filename + "_" + uuid.Must(uuid.NewRandom()).String() + ".png"

	// 签名直传
	signedPutURL, err := bucket.SignURL(key, oss.HTTPPut, 600, options...)
	if err != nil {
		return serializer.Response{
			Status: 50002,
			Data:   nil,
			Msg:    "OSS配置错误",
			Error:  err.Error(),
		}, err
	}

	// 查看图片
	signedGetURL, err := bucket.SignURL(key, oss.HTTPGet, 600)
	if err != nil {
		return serializer.Response{
			Status: 50002,
			Data:   nil,
			Msg:    "OSS配置错误",
			Error:  err.Error(),
		}, err
	}

	return serializer.Response{
		Data: map[string]string{
			"key": key,
			"put": signedPutURL,
			"get": signedGetURL,
		},
	}, nil
}
