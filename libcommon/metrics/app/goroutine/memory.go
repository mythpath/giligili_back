package goroutine

import (
	"runtime"
	"time"
	"selfText/giligili_back/libcommon/metrics/app/model"
)

func RuntimeMemory(tags map[string]string) []*model.MetricsMeta {
	m := &runtime.MemStats{}
	runtime.ReadMemStats(m)

	metricses := []*model.MetricsMeta{}

	metricses = append(metricses,
		outputMemStats(m, tags),
		outputMemGCStats(m, tags),
		outputMemHeapStats(m, tags),
		outputMemStackStats(m, tags))

	return metricses
}

func outputMemStats(m *runtime.MemStats, tags map[string]string) *model.MetricsMeta {

	metrics := &model.MetricsMeta{
		Name: "memory",
		Tags: make(map[string]string),
		Fields: make(map[string]interface{}),
	}
	metrics.SetTags(tags)
	metrics.SetTags(model.DefaultTags)

	// General
	metrics.Set("alloc", m.Alloc)
	metrics.Set("total", m.TotalAlloc)
	metrics.Set("sys", m.Sys)
	metrics.Set("malloc", m.Mallocs)
	metrics.Set("frees", m.Frees)
	metrics.Set("othersys", m.OtherSys)

	metrics.Set("lookups", m.Lookups)

	return metrics
}

func outputMemStackStats(m *runtime.MemStats, tags map[string]string) *model.MetricsMeta {
	metrics := &model.MetricsMeta{
		Name: "memory_stack",
		Tags: make(map[string]string),
		Fields: make(map[string]interface{}),
	}
	metrics.SetTags(tags)
	metrics.SetTags(model.DefaultTags)

	metrics.Set("inuse", m.StackInuse)
	metrics.Set("sys", m.StackSys)
	metrics.Set("mspan_inuse", m.MSpanInuse)
	metrics.Set("mspan_sys", m.MSpanSys)
	metrics.Set("mcache_inuse", m.MCacheInuse)
	metrics.Set("mcache_sys", m.MCacheSys)

	return metrics
}

func outputMemHeapStats(m *runtime.MemStats, tags map[string]string) *model.MetricsMeta {
	metrics := &model.MetricsMeta{
		Name: "memory_heap",
		Tags: make(map[string]string),
		Fields: make(map[string]interface{}),
	}
	metrics.SetTags(tags)
	metrics.SetTags(model.DefaultTags)

	metrics.Set("alloc", m.HeapAlloc)
	metrics.Set("sys", m.HeapSys)
	metrics.Set("idle", m.HeapIdle)
	metrics.Set("inuse", m.HeapInuse)
	metrics.Set("released", m.HeapReleased)
	metrics.Set("objects", m.HeapObjects)

	return metrics
}

func outputMemGCStats(m *runtime.MemStats, tags map[string]string) *model.MetricsMeta {
	metrics := &model.MetricsMeta{
		Name: "memory_gc",
		Tags: make(map[string]string),
		Fields: make(map[string]interface{}),
	}
	metrics.SetTags(tags)
	metrics.SetTags(model.DefaultTags)

	// bytes
	metrics.Set("sys", m.GCSys)
	metrics.Set("next", m.NextGC)
//	metric.Set("gc.last", fmt.Sprintf("%v", ntp.Unix(model.TimeFromUnixNano(int64(m.LastGC)).Unix(), 0)))

	// us
	metrics.Set("pause_total", uint64(m.PauseTotalNs) / uint64(time.Microsecond))
	metrics.Set("pause", uint64(m.PauseNs[(m.NumGC+255)%256]) / uint64(time.Microsecond))

	// gauge
	metrics.Set("count", uint64(m.NumGC))

	return metrics
}
