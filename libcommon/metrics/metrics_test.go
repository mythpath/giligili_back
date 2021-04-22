package metrics

import (
	"sync"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMetrics(t *testing.T) {
	Convey("New Metric", t, func() {
		ms := NewRegistry()
		m := newMetric()
		ms.Put(m)

		Convey("field f1", func() {
			all := ms.All()
			So(all,ShouldNotBeNil)
			var wg sync.WaitGroup
			for m := range all {
				wg.Add(1)
				points:=m.Points()
				So(points,ShouldNotBeNil)
				go func(){
					defer wg.Done()
					for {
						select {
						case point:=<-points:
							So(point.Fields["f1"], ShouldEqual, 1)
						default:
							return 
						}
					}
				}()								
			}
			wg.Wait()
		})
	})
}

func newMetric() Metric {
	return NewMetric("test_metric", GAUGE, func() <-chan *Sample {
		ss := make(chan *Sample)
		go func() {
			s := &Sample{Fields: map[string]interface{}{}}
			s.Fields["f1"] = 1
			s.Timestamp = time.Now().UnixNano()
			ss <- s
		}()

		return ss
	}).
		Field("f1", INT64).
		Build()
}
