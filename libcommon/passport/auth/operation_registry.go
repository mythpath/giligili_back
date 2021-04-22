package auth

type ModulePermission struct {
	Name      string        `json:"name"` // module name
	Label     string        `json:"label"`
	LocalAddr string        `json:"ip"`
	Resource  ResourcesPerm `json:"resources"`
}

func NewModulePermission(name, label, addr string) *ModulePermission {
	p := &ModulePermission{}
	p.Name = name
	p.Label = label
	p.LocalAddr = addr
	return p
}

func (p *ModulePermission) register(rscName, rscLabel, opName, opLabel string, isSystem bool) {
	if isSystem {
		p.Resource.registerSystem(rscName, rscLabel, opName, opLabel)
	} else {
		p.Resource.register(rscName, rscLabel, opName, opLabel)
	}
}

type ResourcesPerm []ResourcePermission

func (p *ResourcesPerm) register(rscName, rscLabel, opName, opLabel string) {
	rsc, ok := p.ContainsSystem(rscName, false)
	if !ok {
		new := ResourcePermission{
			Name:      rscName,
			Label:     rscLabel,
			IsSystem:  false,
			Operation: []OperationPermission{},
		}
		new.Operation.register(opName, opLabel)
		p.Append(&new)
	} else {
		rsc.Operation.register(opName, opLabel)
	}
}

func (p *ResourcesPerm) registerSystem(rscName, rscLabel, opName, opLabel string) {
	rsc, ok := p.ContainsSystem(rscName, true)
	if !ok {
		new := ResourcePermission{
			Name:      rscName,
			Label:     rscLabel,
			IsSystem:  true,
			Operation: []OperationPermission{},
		}
		new.Operation.register(opName, opLabel)
		p.Append(&new)
	} else {
		rsc.Operation.register(opName, opLabel)
	}
}

func (p *ResourcesPerm) Contains(rscName string) (*ResourcePermission, bool) {
	all := (*[]ResourcePermission)(p)
	for i, _ := range *all {
		if (*all)[i].Name == rscName {
			return &(*all)[i], true
		}
	}
	return nil, false
}

func (p *ResourcesPerm) ContainsSystem(rscName string, isSystem bool) (*ResourcePermission, bool) {
	all := (*[]ResourcePermission)(p)
	for i, _ := range *all {
		if (*all)[i].Name == rscName && (*all)[i].IsSystem == isSystem {
			return &(*all)[i], true
		}
	}
	return nil, false
}

func (p *ResourcesPerm) Append(new *ResourcePermission) {
	all := (*[]ResourcePermission)(p)
	*all = append(*all, *new)
}

type ResourcePermission struct {
	Name      string         `json:"name"`
	Label     string         `json:"label"`
	IsSystem  bool           `json:"isSystem"`
	Operation OperationsPerm `json:"operations"`
}

type OperationsPerm []OperationPermission

func (p *OperationsPerm) register(name, label string) {
	if len(name) == 0 {
		return
	}
	ok := p.Contains(name)
	if !ok {
		new := OperationPermission{
			Name:  name,
			Label: label,
		}
		p.Append(&new)
	}
}

func (p *OperationsPerm) Contains(opName string) bool {
	all := (*[]OperationPermission)(p)
	for i, _ := range *all {
		if (*all)[i].Name == opName {
			return true
		}
	}
	return false
}

func (p *OperationsPerm) Append(new *OperationPermission) {
	all := (*[]OperationPermission)(p)
	*all = append(*all, *new)
}

type OperationPermission struct {
	Name  string `json:"name"`
	Label string `json:"label"`
}
