package metrics

//RegistryService XXX
type RegistryService struct {
	registry
}

func (p *RegistryService) Init () error {
	p.registry  = registry{
		ms: map[string]Metric{},
	}
	return nil
}