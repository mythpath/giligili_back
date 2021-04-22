package cmdb

import (
	"selfText/giligili_back/libcommon/brick"
	"time"

	"github.com/sirupsen/logrus"
)

type CMDB struct {
	Config brick.Config `inject:"config"`

	*Cmdb
}

func (p *CMDB) Init() error {
	url := p.Config.GetMapString("cmdb", "url", "http://100.73.45.8:3404")
	if url == "" {
		panic("cmdb url is empty")
	}
	var err error

	timeout := p.Config.GetMapString("cmdb", "timeout", "30s")
	duration, err := time.ParseDuration(timeout)
	if err != nil {
		duration = 30 * time.Second
		logrus.Errorf("failed to parse cmdb.timeout: %s")
	}
	opt := &ClientOpt{
		Headers: map[string]string{
			"nerv-user": "NERV-AGENT",
		},
		Timeout:   duration,
		Keepalive: duration,
	}
	p.Cmdb, err = NewCmdb(url, logrus.StandardLogger(), opt)
	if err != nil {
		logrus.Errorf("failed to create cmdb client: %v", err)
		return err
	}

	return nil
}
