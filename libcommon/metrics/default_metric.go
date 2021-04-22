package metrics

type defaultMetric struct {
	name       string
	metricType MetricType
	fields     map[string]*MetricField
	value      SampleFunc
}

func (p *defaultMetric) Name() string {
	return p.name
}

func (p *defaultMetric) Type() MetricType {
	return p.metricType
}

func (p *defaultMetric) Fields() <-chan *MetricField {
	fields := make(chan *MetricField)
	go func() {
		for _, field := range p.fields {
			fields <- field
		}
		close(fields)
	}()
	return fields
}

func (p *defaultMetric) Points() <-chan *Sample {
	return p.value()
}
