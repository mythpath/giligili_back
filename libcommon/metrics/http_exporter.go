package metrics

import (
	"net/http"
	"sync"

	"selfText/giligili_back/libcommon/net/http/rest/render"
)

// HttpExporter exports metric throught http
type HttpExporter struct {
	metrics Registry
}

func NewHttpExporter(metrics Registry) *HttpExporter {
	return &HttpExporter{
		metrics: metrics,
	}
}

func (p *HttpExporter) GetMetrics(w http.ResponseWriter, r *http.Request) {
	out := make(chan interface{}, 1000)

	var wg sync.WaitGroup
	for metric := range p.metrics.All() {
		wg.Add(1)
		go func(metric Metric, out chan<- interface{}) {
			defer wg.Done()

			for point := range metric.Points() {
				if point != nil {
					out <- point
				}
			}
		}(metric, out)
	}

	var ch chan interface{}
	ch = make(chan interface{})
	points := []interface{}{}
	go func(out chan interface{}, datas *[]interface{}, ch chan interface{}) {
		for data := range out {
			*datas = append(*datas, data)
		}
		ch <- "ok"
	}(out, &points, ch)
	wg.Wait()
	close(out)
	select {
	case <-ch:
	}

	if len(points) > 0 {
		render.Status(r, 200)
		render.JSON(w, r, points)
	}
	return

}

func (p *HttpExporter) TakeMetrics(w http.ResponseWriter, r *http.Request) {
	out := make(chan interface{}, 1000)

	var wg sync.WaitGroup
	for metric := range p.metrics.All() {
		wg.Add(1)
		go func(metric Metric, out chan<- interface{}) {
			defer wg.Done()
			for {
				select {
				case point := <-metric.Points():
					if point != nil {
						out <- point
					}
				default:
					return
				}
			}
		}(metric, out)
	}

	wg.Wait()

	points := []interface{}{}
	for {
		select {
		case point := <-out:
			points = append(points, point)
		default:
			if len(points) > 0 {
				render.Status(r, 200)
				render.JSON(w, r, points)
			}
			return
		}
	}
}
