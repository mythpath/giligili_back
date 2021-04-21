package serializer

// Video 视频序列化器
type Video struct {
	ID        uint   `json:"id"`
	Title     string `json:"title"`
	Info      string `json:"info"`
	Url       string `json:"url"`
	Avatar    string `json:"avatar"`
	View      uint64 `json:"view"`
	User      User   `json:"user"`
	CreatedAt int64  `json:"created_at"`
}
