package metrics

import (
	"testing"
	"time"
	"fmt"
)

func TestCollector_ExporterMetrics(t *testing.T) {
	trigger := make(chan struct{}, 1)
	defer close(trigger)

	nr := NewRegistry()
	cl := NewCollector(nr)

	newMetrics(cl)

	go func() {
		for {
			select {
			case <- trigger:
				for metrics := range nr.All() {

					var count int
					for sample := range metrics.Points() {
						fmt.Println(sample)
						count ++
					}

					fmt.Println("------------------", "total", count)
				}
			}
		}
	}()


	for i := 0; i < 10; i ++ {
		generateMetrics(cl)
		trigger <- struct{}{}
		time.Sleep(2 * time.Second)
	}
}

func newMetrics(cl *Collector) {
	cl.NewMetrics("cpu", "idle")
	cl.NewMetrics("cpu", "sys")
	cl.NewMetrics("cpu", "usage")
}

func generateMetrics(cl *Collector) {
	cl.GetMetrics("cpu", "idle").Gauge(map[string]string{
		"tag": "test0",
		"name": "test",
	}).Set(10)

	cl.GetMetrics("cpu", "idle").Gauge(map[string]string{
		"tag": "test1",
		"name": "test",
	}).Set(20)
	cl.GetMetrics("cpu", "idle").Gauge(map[string]string{
		"tag": "test2",
		"name": "test",
	}).Set(20)
	cl.GetMetrics("cpu", "idle").Gauge(map[string]string{
		"tag": "test2",
		"name": "test",
	}).Set(20)

	cl.GetMetrics("cpu", "usage").Gauge(map[string]string{
		"tag": "test",
		"name": "test",
	}).Add(10)

	cl.GetMetrics("cpu", "usage").Gauge(map[string]string{
		"tag": "test",
		"name": "test",
	}).Add(20)

	cl.GetMetrics("cpu", "usage").Gauge(map[string]string{
		"tag": "test",
		"name": "test",
	}).Dec()

	cl.GetMetrics("cpu", "usage").Counter(map[string]string{
		"tag": "test1",
		"name": "test",
	}).Inc()

	cl.GetMetrics("cpu", "sys").Gauge(map[string]string{
		"tag": "test",
		"name": "test",
	}).Set(0)

	cl.GetMetrics("nerv_monitor_storage", "metrics_50ms_count").Counter(map[string]string{
		"direction": "in",
		"service_name": "monitor-storage",
	}).Add(10)

	cl.GetMetrics("nerv_monitor_storage", "metrics_total").Counter(map[string]string{
		"direction": "in",
		"service_name": "monitor-storage",
	}).Add(20)

}
