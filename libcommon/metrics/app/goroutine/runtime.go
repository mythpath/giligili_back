package goroutine

import (
	"runtime"
	"selfText/giligili_back/libcommon/metrics/app/model"
)


func RuntimeGoRoutines(tags map[string]string) []*model.MetricsMeta {
	m := &model.MetricsMeta{
		Name: "runtime",
		Tags: make(map[string]string),
		Fields: make(map[string]interface{}),
	}
	m.SetTags(tags)
	m.SetTags(model.DefaultTags)


	m.Set("goroutines", uint64(runtime.NumGoroutine()))

	return []*model.MetricsMeta{m}
}

func RuntimeCgoCalls(tags map[string]string) []*model.MetricsMeta {

	m := &model.MetricsMeta{
		Name: "runtime",
		Tags: make(map[string]string),
		Fields: make(map[string]interface{}),
	}
	m.SetTags(tags)
	m.SetTags(model.DefaultTags)

	m.Set("cgo_calls", uint64(runtime.NumCgoCall()))

	return []*model.MetricsMeta{m}
}


