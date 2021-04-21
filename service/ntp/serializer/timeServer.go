package serializer

// TimeServer 时间服务器记录序列化器
type TimeServer struct {
	ID      uint   `json:"id"`
	URL     string `json:"url"`
	Comment string `json:"comment"`
}

// TimeServerResponse 单个时间服务器记录序列化
type TimeServerResponse struct {
	Response
	Data TimeServer `json:"data"`
}
