package metric

import (
	"selfText/giligili_back/libcommon/brick"
	"selfText/giligili_back/libcommon/metrics"
	"time"
)

type InternalCollector struct {
	Collector *metrics.CollectorService `inject:"MetricsCollector"`
	Config    brick.Config              `inject:"config"`

	defaultTags map[string]string
}

func (p *InternalCollector) Init() error {

	p.defaultTags = make(map[string]string)
	tagM := p.Config.GetMap("defaultTags")
	for k, v := range tagM {
		tv, ok := v.(string)
		if ok {
			p.defaultTags[k] = tv
		}
	}

	p.newMetrics()

	return nil
}

func (p *InternalCollector) GetField(field string) *entry {
	return newEntry(p, KDefaultTable, field)
}

func (p *InternalCollector) GetFinchField(field string) *entry {
	return newEntry(p, NervFinch, field)
}

func (p *InternalCollector) newMetrics() {
	// GAUGE
	p.Collector.NewMetrics(KDefaultTable, ReceiveRequest)
	p.Collector.NewMetrics(KDefaultTable, ReceiveWebconsoleNum)
	p.Collector.NewMetrics(KDefaultTable, SendWebconsoleSuccessNum)
	p.Collector.NewMetrics(KDefaultTable, SendWebconsoleRetryNum)
	p.Collector.NewMetrics(KDefaultTable, SendWebconsoleThrowNum)
	p.Collector.NewMetrics(KDefaultTable, ReceiveEmailNum)
	p.Collector.NewMetrics(KDefaultTable, SendEmailSuccessNum)
	p.Collector.NewMetrics(KDefaultTable, SendEmailRetryNum)
	p.Collector.NewMetrics(KDefaultTable, SendEmailThrowNum)
	p.Collector.NewMetrics(KDefaultTable, SendEmailFail)
	p.Collector.NewMetrics(KDefaultTable, SendWebConsoleFail)
	p.Collector.NewMetrics(KDefaultTable, SendWeChatFail)

	p.Collector.NewMetrics(NervFinch, SendFail)
	p.Collector.NewMetrics(NervFinch, SendWeChatFail)
	p.Collector.NewMetrics(NervFinch, F100msRCount)
	p.Collector.NewMetrics(NervFinch, F150msHCount)
	p.Collector.NewMetrics(NervFinch, F200msSCount)
}

const (
	// nerv-notify
	KDefaultTable            = "nerv_notify"
	ReceiveRequest           = "receive_request"
	ReceiveWebconsoleNum     = "receive_webconsole_num"
	SendWebconsoleSuccessNum = "send_webconsole_success_num"
	SendWebconsoleRetryNum   = "send_webconsole_retry_num"
	SendWebconsoleThrowNum   = "send_webconsole_throw_num"
	ReceiveEmailNum          = "receive_email_num"
	SendEmailSuccessNum      = "send_email_success_num"
	SendEmailRetryNum        = "send_email_retry_num"
	SendEmailThrowNum        = "send_email_throw_num"
	Receive                  = "receive"
	Success                  = "success"
	Retry                    = "retry"
	Throw                    = "throw"

	// nerv-finch
	NervFinch = "nerv_finch"
	// response count per 100ms
	F100msRCount = "f100ms_rcount"
	// handle count per 150ms
	F150msHCount = "f150ms_hcount"
	// send count per 200ms
	F200msSCount = "f200ms_scount"
	// send message to notify failed
	SendFail = "send_fail"

	// common value
	SendEmailFail      = "send_email_fail"
	SendWebConsoleFail = "send_webconsole_fail"
	SendWeChatFail     = "send_wechat_fail"
)

var (
	R100ms = 100 * time.Millisecond
	R200ms = 200 * time.Millisecond
	R150ms = 150 * time.Millisecond
	R1s    = time.Second
)

func (p *InternalCollector) IndexOP(op string, msgType string, tags *map[string]string) {
	//todo 空置，后续有需求再修改填充
	//if op == ReceiveRequest {
	//	p.GetField(ReceiveRequest).Counter(*tags).Inc()
	//} else if op == Receive {
	//	switch msgType {
	//	case request.STWebConsole:
	//		{
	//			p.GetField(ReceiveWebconsoleNum).Counter(*tags).Inc()
	//		}
	//	case request.STEmail:
	//		{
	//			p.GetField(ReceiveEmailNum).Counter(*tags).Inc()
	//		}
	//	}
	//} else if op == Success {
	//	switch msgType {
	//	case request.STWebConsole:
	//		{
	//			p.GetField(SendWebconsoleSuccessNum).Gauge(*tags).Inc()
	//		}
	//	case request.STEmail:
	//		{
	//			p.GetField(SendEmailSuccessNum).Gauge(*tags).Inc()
	//		}
	//	}
	//} else if op == Retry {
	//	switch msgType {
	//	case request.STWebConsole:
	//		{
	//			p.GetField(SendWebconsoleRetryNum).Gauge(*tags).Inc()
	//		}
	//	case request.STEmail:
	//		{
	//			p.GetField(SendEmailRetryNum).Gauge(*tags).Inc()
	//		}
	//	}
	//} else if op == Throw {
	//	switch msgType {
	//	case request.STWebConsole:
	//		{
	//			p.GetField(SendWebconsoleThrowNum).Gauge(*tags).Inc()
	//		}
	//	case request.STEmail:
	//		{
	//			p.GetField(SendEmailThrowNum).Gauge(*tags).Inc()
	//		}
	//	}
	//}
}

func (p *InternalCollector) SendStatusOP(component string, field string, success bool, tags map[string]string) {
	//todo 空置，后续有需求再修改填充
	//var flag int
	//if success {
	//	flag = 0
	//} else {
	//	flag = 1
	//}
	//switch component {
	//case _type.ComponentFinch:
	//	p.GetFinchField(field).Gauge(tags).Set(flag)
	//case _type.ComponentNotify:
	//	p.GetField(field).Gauge(tags).Set(flag)
	//}
}

type entry struct {
	cl    *InternalCollector
	name  string
	field string
}

func newEntry(cl *InternalCollector, name, field string) *entry {
	return &entry{
		cl:    cl,
		name:  name,
		field: field,
	}
}

func (p *entry) Counter(tags map[string]string) metrics.Counter {

	return p.cl.Collector.GetMetrics(p.name, p.field).Counter(p.replaceTags(tags))
}

func (p *entry) Gauge(tags map[string]string) metrics.Gauge {

	return p.cl.Collector.GetMetrics(p.name, p.field).Gauge(p.replaceTags(tags))
}

func (p *entry) replaceTags(tags map[string]string) map[string]string {
	if tags == nil {
		tags = make(map[string]string)
	}
	cloneTags := copyTags(tags)
	for k, v := range p.cl.defaultTags {
		if _, ok := cloneTags[k]; !ok {
			cloneTags[k] = v
		}
	}

	return cloneTags
}

func copyTags(tags map[string]string) map[string]string {
	cloneTags := make(map[string]string)
	for k, v := range tags {
		cloneTags[k] = v
	}

	return cloneTags
}
