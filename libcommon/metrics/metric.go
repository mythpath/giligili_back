package metrics

//DataType defines data type of a metric field
type DataType int

const (
	STRING DataType = iota
	BOOL
	UINT64
	INT64
	FLOAT64
)

//MetricType defines type of a metric
type MetricType int

func (p MetricType) String() string {
	switch p {
	case GAUGE:
		return "GAUGE"
	case COUNTER:
		return "COUNTER"
	case DERIVE:
		return "DERIVE"
	}
	return "UNKOWN"
}

const (
	GAUGE MetricType = iota
	COUNTER
	DERIVE
)

//Metric defines an indicator of an application
type Metric interface {
	//Name
	Name() string

	//Type
	Type() MetricType

	//Fields
	Fields() <-chan *MetricField

	//Points taks all datapoints
	Points() <-chan *Sample
}

//MetricField defines the metadata of a field
type MetricField struct {
	Name     string
	DataType DataType
}

//MetricBuilder builds a metric
type MetricBuilder struct {
	name       string
	metricType MetricType
	fields     map[string]*MetricField
	value      SampleFunc
}

//NewMetric create a MetricBuilder
func NewMetric(name string, t MetricType, fn SampleFunc) *MetricBuilder {
	return &MetricBuilder{name: name, metricType: t, fields: map[string]*MetricField{}, value: fn}
}

func (p *MetricBuilder) Field(name string, dt DataType) *MetricBuilder {
	p.fields[name] = &MetricField{Name: name, DataType: dt}
	return p
}

func (p *MetricBuilder) Build() Metric {
	return &defaultMetric{name: p.name, fields: p.fields, value: p.value}
}

func NewMetrics(name string, fn SampleFunc) *MetricBuilder {
	return &MetricBuilder{name: name, value: fn}
}
