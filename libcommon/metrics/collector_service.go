package metrics

type CollectorService struct {
	Registry 	Registry `inject:"MetricsRegistry"`

	*Collector
}

func (p *CollectorService) Init() error {

	p.Collector = NewCollector(p.Registry)

	return nil
}

