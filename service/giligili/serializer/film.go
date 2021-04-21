package serializer

// Film 影片序列化器
type Film struct {
	Name        string
	Code        string
	PerformerID uint
	Post        string
	Rank        string
	Comment     string
}

// FilmResponse 单个影片序列化
type FilmResponse struct {
	Response
	Data Film `json:"data"`
}


