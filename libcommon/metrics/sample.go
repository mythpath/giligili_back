package metrics


//Sample is a data point of a metric in sometime
type Sample struct {
	Name      string                 `json:"name"`   // metric name
	Fields    map[string]interface{} `json:"fields"` // metric field values
	Tags      map[string]string      `json:"tags"`   // tags
	Type      string                 `json:"type"`   // metric type
	Timestamp int64              `json:"timestamp"`
}

//SampleFunc returns a sample
type SampleFunc func() <-chan *Sample
