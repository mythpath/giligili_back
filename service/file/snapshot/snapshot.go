package snapshot

import "fmt"

type Route struct {
	Url      string `json:"url"`
	DeployId uint   `json:"deploy_id"`
	BaseId   uint   `json:"base_id"`
	Copy     bool   `json:"copy"`
}

func (r *Route) ParseID(path string) error {
	url, err := NewUrl(path)
	if err != nil {
		return err
	}

	r.DeployId = url.deployId

	return nil
}

func (r *Route) String() string {
	return fmt.Sprintf("%s[%d]", r.Url, r.DeployId)
}
