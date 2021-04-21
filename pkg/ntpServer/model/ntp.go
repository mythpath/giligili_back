package model

import (
	"bytes"
	"encoding/binary"
)

const (
	UNIX_STA_TIMESTAMP = 2208988800
)

/**
NTP协议 http://www.ntp.org/documentation.html
*/
type NTP struct {
	// 1:32bits
	Li        uint8 `json:"li"`
	Vn        uint8 `json:"vn"`
	Mode      uint8 `json:"mode"`
	Stratum   uint8 `json:"stratum"`
	Poll      uint8 `json:"poll"`
	Precision uint8 `json:"precision"`
	// 2:
	RootDelay           int32 `json:"root_delay"`
	RootDispersion      int32 `json:"root_dispersion"`
	ReferenceIdentifier int32 `json:"reference_identifier"`
	// 64位时间戳
	ReferenceTimestamp uint64 `json:"reference_timestamp"` //指示系统时钟最后一次校准的时间
	OriginateTimestamp uint64 `json:"originate_timestamp"` //指示客户向服务器发起请求的时间
	ReceiveTimestamp   uint64 `json:"receive_timestamp"` //指服务器收到客户请求的时间
	TransmitTimestamp  uint64 `json:"transmit_timestamp"` //指示服务器向客户发时间戳的时间
}

func (n *NTP) GetBytes() []byte {
	// 网络中使用大端字节排序
	buf := &bytes.Buffer{}
	head := (n.Li << 6) | (n.Vn << 3) | ((n.Mode << 5) >> 5)
	binary.Write(buf, binary.BigEndian, uint8(head))
	binary.Write(buf, binary.BigEndian, n.Stratum)
	binary.Write(buf, binary.BigEndian, n.Poll)
	binary.Write(buf, binary.BigEndian, n.Precision)
	//写入其他字节数据
	binary.Write(buf, binary.BigEndian, n.RootDelay)
	binary.Write(buf, binary.BigEndian, n.RootDispersion)
	binary.Write(buf, binary.BigEndian, n.ReferenceIdentifier)
	binary.Write(buf, binary.BigEndian, n.ReferenceTimestamp)
	binary.Write(buf, binary.BigEndian, n.OriginateTimestamp)
	binary.Write(buf, binary.BigEndian, n.ReceiveTimestamp)
	binary.Write(buf, binary.BigEndian, n.TransmitTimestamp)
	//[27 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
	return buf.Bytes()
}

func (n *NTP) Parse(bf []byte, useUnixSec bool) {
	var (
		bit8  uint8
		bit32 int32
		bit64 uint64
		rb    *bytes.Reader
	)
	//貌似这binary.Read只能顺序读，不能跳着读，想要跳着读只能使用切片bf
	rb = bytes.NewReader(bf)
	binary.Read(rb, binary.BigEndian, &bit8)
	//向右偏移6位得到前两位LI即可
	n.Li = bit8 >> 6
	//向右偏移2位,向右偏移5位,得到前中间3位
	n.Vn = (bit8 << 2) >> 5
	//向左偏移5位，然后右偏移5位得到最后3位
	n.Mode = (bit8 << 5) >> 5
	binary.Read(rb, binary.BigEndian, &bit8)
	n.Stratum = bit8
	binary.Read(rb, binary.BigEndian, &bit8)
	n.Poll = bit8
	binary.Read(rb, binary.BigEndian, &bit8)
	n.Precision = bit8

	//32bits
	binary.Read(rb, binary.BigEndian, &bit32)
	n.RootDelay = bit32
	binary.Read(rb, binary.BigEndian, &bit32)
	n.RootDispersion = bit32
	binary.Read(rb, binary.BigEndian, &bit32)
	n.ReferenceIdentifier = bit32

	//以下几个字段都是64位时间戳(NTP都是64位的时间戳)
	binary.Read(rb, binary.BigEndian, &bit64)
	n.ReferenceTimestamp = bit64
	binary.Read(rb, binary.BigEndian, &bit64)
	n.OriginateTimestamp = bit64
	binary.Read(rb, binary.BigEndian, &bit64)
	n.ReceiveTimestamp = bit64
	binary.Read(rb, binary.BigEndian, &bit64)
	n.TransmitTimestamp = bit64
	//转换为unix时间戳,先左偏移32位拿到64位时间戳的整数部分，然后ntp的起始时间戳 1900年1月1日 0时0分0秒 2208988800
	if useUnixSec {
		n.ReferenceTimestamp = (n.ReceiveTimestamp >> 32) - UNIX_STA_TIMESTAMP
		if n.OriginateTimestamp > 0 {
			n.OriginateTimestamp = (n.OriginateTimestamp >> 32) - UNIX_STA_TIMESTAMP
		}
		n.ReceiveTimestamp = (n.ReceiveTimestamp >> 32) - UNIX_STA_TIMESTAMP
		n.TransmitTimestamp = (n.TransmitTimestamp >> 32) - UNIX_STA_TIMESTAMP
	}
}
