package gormHook

import "selfText/giligili_back/libcommon/logging"

type AfterHook struct {
	Logger *logging.LoggerService `inject:"LoggerService"`

	defaultFields []string
}