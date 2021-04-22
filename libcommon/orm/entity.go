package orm

// Entity base mode definition
type Entity struct {
	SelfGormModel
	Tenant      string `json:"tenant"`     //tenant name
	CreatedUser string `json:"createUser"` //created user name
	UpdatedUser string `json:"updateUser"` //updated user name
}
