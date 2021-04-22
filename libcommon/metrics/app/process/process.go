package process

import (
	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
	"os"
	"selfText/giligili_back/libcommon/metrics/app/model"
	"selfText/giligili_back/libcommon/metrics/app/util"
	"syscall"
)

var (
	app *AppPS
)

type AppPS struct {
	ps *process.Process

	tags map[string]string
	pid  int
}

func NewAppPS(tags map[string]string) (*AppPS, error) {

	if app == nil {
		id := os.Getpid()

		ps, err := process.NewProcess(int32(id))
		if err != nil {
			return nil, err
		}

		app = &AppPS{
			pid:  id,
			ps:   ps,
			tags: tags,
		}
	}

	return app, nil
}

func (p *AppPS) CpuUsage() (*model.MetricsMeta, error) {
	cpu, err := p.ps.Percent(0)
	if err != nil {
		return nil, err
	}

	metrics := &model.MetricsMeta{
		Name:   "cpu",
		Tags:   make(map[string]string),
		Fields: make(map[string]interface{}),
	}
	metrics.SetTags(model.DefaultTags)
	metrics.SetTags(p.tags)

	metrics.Set("usage", util.AccurancyFloat64(cpu, 2, false))

	return metrics, nil
}

func (p *AppPS) MemoryUsage() (*model.MetricsMeta, error) {
	mem, err := p.ps.MemoryPercent()
	if err != nil {
		return nil, err
	}

	metrics := &model.MetricsMeta{
		Name:   "memory",
		Tags:   make(map[string]string),
		Fields: make(map[string]interface{}),
	}
	metrics.SetTags(p.tags)
	metrics.SetTags(model.DefaultTags)

	metrics.Set("usage", util.AccurancyFloat64(float64(mem), 2, false))

	return metrics, nil
}

func (p *AppPS) OutputMemStats() (*model.MetricsMeta, error) {
	ms, err := p.ps.MemoryInfo()
	if err != nil {
		return nil, err
	}

	metrics := &model.MetricsMeta{
		Name:   "memory",
		Tags:   make(map[string]string),
		Fields: make(map[string]interface{}),
	}
	metrics.SetTags(p.tags)
	metrics.SetTags(model.DefaultTags)

	metrics.Set("rss", ms.RSS)
	metrics.Set("vms", ms.VMS)
	metrics.Set("swap", ms.Swap)
	metrics.Set("locked", ms.Locked)
	metrics.Set("data", ms.Data)

	return metrics, nil
}

func (p *AppPS) OutputNetIOStats() ([]*model.MetricsMeta, error) {
	netios, err := p.ps.NetIOCounters(true)
	if err != nil {
		return nil, err
	}

	metas := []*model.MetricsMeta{}

	for _, netio := range netios {

		if netio.Name == "lo" {
			continue
		}

		metas = append(metas, p.combineNetIO(&netio))
	}

	return metas, nil
}

func (p *AppPS) combineNetIO(netio *net.IOCountersStat) *model.MetricsMeta {
	m := &model.MetricsMeta{
		Name:   "net",
		Tags:   make(map[string]string),
		Fields: make(map[string]interface{}),
	}
	m.Tags["interface"] = netio.Name
	m.SetTags(p.tags)
	m.SetTags(model.DefaultTags)

	m.Set("bytes_sent", netio.BytesSent)
	m.Set("bytes_recv", netio.BytesRecv)
	m.Set("packets_sent", netio.PacketsSent)
	m.Set("packets_recv", netio.PacketsRecv)
	m.Set("err_in", netio.Errin)
	m.Set("err_out", netio.Errout)
	m.Set("drop_in", netio.Dropin)
	m.Set("drop_out", netio.Dropout)
	m.Set("fifo_in", netio.Fifoin)
	m.Set("fifo_out", netio.Fifoout)

	return m
}

func (p *AppPS) OutputConnectionsStats() ([]*model.MetricsMeta, error) {
	conns, err := p.ps.Connections()
	if err != nil {
		return nil, err
	}

	metas := []*model.MetricsMeta{}

	metas = append(metas,
		p.combineConn(conns, syscall.AF_INET),
		p.combineConn(conns, syscall.AF_INET6),
		p.combineNetTypeConn(conns, syscall.SOCK_DGRAM),
		p.combineNetTypeConn(conns, syscall.SOCK_STREAM),
	)

	return metas, nil
}

func (p *AppPS) combineNetTypeConn(conns []net.ConnectionStat, net uint32) *model.MetricsMeta {
	var (
		upProtoMap = map[uint32]string{
			syscall.SOCK_STREAM: "TCP",
			syscall.SOCK_DGRAM:  "UDP",
		}
	)

	m := &model.MetricsMeta{
		Name:   "connection",
		Tags:   make(map[string]string),
		Fields: make(map[string]interface{}),
	}

	m.Tags["net_type"] = upProtoMap[net]
	m.SetTags(p.tags)
	m.SetTags(model.DefaultTags)

	count := 0
	for _, conn := range conns {
		if conn.Type == net {
			count += 1
		}
	}

	m.Set("count", count)

	return m
}

func (p *AppPS) combineConn(conns []net.ConnectionStat, family uint32) *model.MetricsMeta {

	var (
		downProtoMap = map[uint32]string{
			syscall.AF_INET:  "IPv4",
			syscall.AF_INET6: "IPv6",
		}
	)

	m := &model.MetricsMeta{
		Name:   "connection",
		Tags:   make(map[string]string),
		Fields: make(map[string]interface{}),
	}

	m.Tags["family"] = downProtoMap[family]
	m.SetTags(p.tags)
	m.SetTags(model.DefaultTags)

	count := 0
	for _, conn := range conns {
		if conn.Family == family {
			count += 1
		}
	}

	m.Set("count", count)

	return m
}
