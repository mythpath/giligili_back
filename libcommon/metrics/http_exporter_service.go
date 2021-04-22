package metrics

//HttpExporterService
type HttpExporterService struct{
	Registry Registry `inject:"MetricsRegistry"`
	httpExporter *HttpExporter 
}

func (p *HttpExporterService) Init() error {
	p.httpExporter = &HttpExporter{p.Registry}
	return nil
}

func (p HttpExporterService) Exporter() *HttpExporter {
	return p.httpExporter
}