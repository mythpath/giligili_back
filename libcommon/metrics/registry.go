package metrics

//Registry is the container for all your application’s metric.
//You’ll probably want to integrate this into your application’s lifecycle
type Registry interface {
	//Register a metric. The previous metric which has the same name will be override by the later metric
	Put(metric Metric)

	All() <-chan Metric
}

//NewRegistry create an instance of a Registry
func NewRegistry() Registry {
	return &registry{
		ms: map[string]Metric{},
	}
}

type registry struct {
	ms map[string]Metric
}

func (p *registry) Put(metric Metric) {
	p.ms[metric.Name()] = metric
}

func (p *registry) All() <-chan Metric {
	ms := make(chan Metric)
	go func() {
		for _, m := range p.ms {
			ms <- m
		}
		close(ms)
	}()
	return ms
}
