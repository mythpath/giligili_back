package golang

//Runtime returns golang runtime metric
//func NewRuntimeGauge() metric.Metric {
//	return metric.NewMetric("golang_runtime", metric.GAUGE, func() map[string]interface{} {
//		s := map[string]interface{}{}
//		s["goroutines"] = int64(runtime.NumGoroutine())
//		return s
//	}).
//		Field("goroutines", metric.INT64).
//		Build()
//}
