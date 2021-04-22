package model

import (
	"bytes"
	"fmt"
)

var (
	DefaultTags = map[string]string{
		"type" : "app",
	}
)


type MetricsMeta struct {
	Name 		string	`json:"name"`	// metrics name
	Fields		map[string]interface{}	`json:"fields"`	// metrics sub name map the value. example: `key:value`. only support uint64/float64/string/bool
	Tags 		map[string]string 		`json:"tags"`
	Type 		string	`json:"type"`					// COUNTER | GAUGE | DERIVE
	Timestamp 	int64	`json:"timestamp"`				// timestamp	default: unix nano(ns)
	DStore 		uint8 	`json:"dStore"`
}

func (m *MetricsMeta) Set(field string, v interface{}) *MetricsMeta {
	m.Fields[field] = v

	return m
}

func (m *MetricsMeta) SetDStore(ds uint8) *MetricsMeta {
	m.DStore = ds

	return m
}

func (m *MetricsMeta) SetTimestamp(timestamp int64) *MetricsMeta {
	m.Timestamp = timestamp

	return m
}

func (m *MetricsMeta) SetTags(tags map[string]string) *MetricsMeta {
	for k, v := range tags {
		m.Tags[k] = v
	}

	return m
}

func (m *MetricsMeta) SetTag(key, value string) *MetricsMeta {
	m.Tags[key] = value

	return m
}

func (m *MetricsMeta) String() string {
	var buf bytes.Buffer

	buf.WriteString("metrics:"+m.Name+",")
	buf.WriteString(fmt.Sprintf("fields:%v,", m.Fields))
	buf.WriteString(fmt.Sprintf("tags:%v,", m.Tags))
	buf.WriteString(fmt.Sprintf("timestamp:%v", m.Timestamp))
	buf.WriteString(fmt.Sprintf("type:%v", m.Type))
	buf.WriteString(fmt.Sprintf("dstore:%v", m.DStore))

	return buf.String()
}
