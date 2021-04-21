package protocal

// ListInput 通用列表查询输入结构
type ListInput struct {
	Where     string   `invoke-query:"where"`
	Values    []string `invoke-query:"values"`
	Order     string   `invoke-query:"order"`
	Page      int      `invoke-query:"page"`
	PageSize  int      `invoke-query:"pageSize"`
	Search    string   `invoke-query:"search"`
	ProjectId uint     `invoke-query:"projectId"`
}

func (l *ListInput) Validate() error {

	return nil
}
