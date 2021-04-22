package time

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"selfText/giligili_back/libcommon/brick"
	"selfText/giligili_back/libcommon/logging"
	"selfText/giligili_back/libcommon/orm"
	"selfText/giligili_back/pkg/ntpServer/model"
	"selfText/giligili_back/pkg/ntpServer/protocal"
	"selfText/giligili_back/service/common/util"
	"selfText/giligili_back/service/ntp/serializer"
	"strings"
	"time"
)

type TimeService struct {
	Config brick.Config           `inject:"config"`
	Orm    *orm.OrmService        `inject:"OrmService"`
	Logger *logging.LoggerService `inject:"LoggerService"`

	SerializerService *serializer.SerializerService `inject:"SerializerService"`

	selectFields []string
}

func (t *TimeService) Init() {
	t.selectFields = []string{"id", "name", "created_by", "created_at", "updated_by", "updated_at",
		"url", "comment"}
}

// Create 创建时间服务器记录
func (t *TimeService) Create(ctx context.Context, input protocal.TimeCreateInput) (serializer.Response, error) {
	var timeServer model.TimeServer
	if err := util.DeepCopy(input, &timeServer); err != nil {
		return serializer.Response{
			Status: 50001,
			Msg:    "时间服务器记录保存失败",
			Error:  err.Error(),
		}, err
	}

	if err := t.Orm.CreateCtx(ctx, model.TimeServerM, &timeServer); err != nil {
		return serializer.Response{
			Status: 50001,
			Msg:    "时间服务器记录保存失败",
			Error:  err.Error(),
		}, err
	}

	return serializer.Response{
		Data: t.SerializerService.BuildTimeServer(timeServer),
	}, nil
}

// Delete 删除时间服务器记录
func (t *TimeService) Delete(ctx context.Context, input protocal.TimeDeleteInput) (serializer.Response, error) {
	var timeServer model.TimeServer
	if err := t.Orm.GetDB().First(&timeServer, input.ID).Error; err != nil {
		return serializer.Response{
			Status: 404,
			Msg:    "时间服务器记录不存在",
			Error:  err.Error(),
		}, err
	}

	if _, err := t.Orm.RemoveCtx(ctx, model.TimeServerM, timeServer.ID, false); err != nil {
		return serializer.Response{
			Status: 50000,
			Msg:    "时间服务器记录删除失败",
			Error:  err.Error(),
		}, err
	}

	return serializer.Response{}, nil
}

// Get 获取指定时间服务器记录
func (t *TimeService) Get(ctx context.Context, input protocal.TimeGetInput) (serializer.Response, error) {
	var timeServer model.TimeServer
	if err := t.Orm.GetDB().First(&timeServer, input.ID).Error; err != nil {
		return serializer.Response{
			Status: 404,
			Msg:    "时间服务器记录不存在",
			Error:  err.Error(),
		}, err
	}

	return serializer.Response{
		Data: t.SerializerService.BuildTimeServer(timeServer),
	}, nil
}

// List 列举所有时间服务器
func (t *TimeService) List(ctx context.Context, input protocal.ListInput) (serializer.Response, error) {
	var total int64

	if err := t.Orm.GetDB().Model(model.TimeServer{}).Count(&total).Error; err != nil {
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
			fields := t.selectFields[1:]
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

	listdata, err := t.Orm.List(model.TimeServerM, t.selectFields, where, argv, input.Order, input.Page, input.PageSize)
	if err != nil {
		return serializer.Response{
			Status: 404,
			Msg:    "获取时间服务器记录列表失败",
			Error:  fmt.Errorf("failed to list topic").Error(),
		}, err
	}

	return t.SerializerService.BuildListResponse(listdata, uint(total)), err
}

func (t *TimeService) Now(ctx context.Context, input protocal.TimeNowInput) (serializer.Response, error) {
	var timeServer model.TimeServer
	if err := t.Orm.GetDB().Where("id = ?", input.ID).First(&timeServer).Error; err != nil {
		return serializer.Response{
			Status: 50001,
			Msg:    "获取时间服务器记录失败",
			Error:  err.Error(),
		}, err
	}

	conn, err := net.Dial("udp", timeServer.URL+":123")
	if err != nil {
		return serializer.Response{
			Status: 50001,
			Msg:    "查询指定时间服务记录失败",
			Error:  err.Error(),
		}, err
	}
	t.Logger.Infoln("after dial udp server.")

	defer func() {
		if err := recover(); err != nil {
			t.Logger.Errorln(err)
		}
		conn.Close()
	}()

	if err := conn.SetDeadline(time.Now().Add(15 * time.Second)); err != nil {
		return serializer.Response{
			Status: 50001,
			Msg:    "查询指定时间服务记录失败",
			Error:  err.Error(),
		}, err
	}
	// configure request settings by specifying the first byte as
	// 00 011 011 (or 0x1B)
	// |  |   +-- client mode (3)
	// |  + ----- version (3)
	// + -------- leap year indicator, 0 no warning
	req := &model.NtpPacket{Settings: 0x1B}

	// send time request
	if err := binary.Write(conn, binary.BigEndian, req); err != nil {
		return serializer.Response{
			Status: 50001,
			Data:   "查询指定时间服务记录失败",
			Msg:    fmt.Sprintf("failed to send request: %v", err),
		}, err
	}

	// block to receive server response
	rsp := &model.NtpPacket{}
	if err := binary.Read(conn, binary.BigEndian, rsp); err != nil {
		return serializer.Response{
			Status: 50001,
			Data:   "查询指定时间服务记录失败",
			Msg:    fmt.Sprintf("failed to read server response: %v", err),
		}, err
	}

	// On POSIX-compliant OS, time is expressed
	// using the Unix time epoch (or secs since year 1970).
	// NTP seconds are counted since 1900 and therefore must
	// be corrected with an epoch offset to convert NTP seconds
	// to Unix time by removing 70 yrs of seconds (1970-1900)
	// or 2208988800 seconds.
	secs := float64(rsp.TxTimeSec) - model.UNIX_STA_TIMESTAMP
	nanos := (int64(rsp.TxTimeFrac) * 1e9) >> 32 // convert fractional to nanos
	//fmt.Printf("%v\n", time.Unix(int64(secs), nanos))

	return serializer.Response{
		Status: 200,
		Data:   fmt.Sprintf("%v\n", time.Unix(int64(secs), nanos)),
		Msg:    "target time server: " + timeServer.URL,
	}, nil
}
