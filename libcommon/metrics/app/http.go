package service

import (
	"log"
	"net/http"
	"selfText/giligili_back/libcommon/metrics/app/goroutine"
	"selfText/giligili_back/libcommon/metrics/app/model"
	"selfText/giligili_back/libcommon/metrics/app/process"
	"selfText/giligili_back/libcommon/net/http/rest/render"
)

func RuntimeCounterMetrics(w http.ResponseWriter, r *http.Request, defaultTags map[string]string) {
	metas := []*model.MetricsMeta{}

	ps, err := process.NewAppPS(defaultTags)
	if err != nil {
		log.Printf("Failed to new app process, err:%v", err)
		render.JSON(w, r, metas)
		return
	}

	metasCh := make(chan []*model.MetricsMeta, 0)
	defer close(metasCh)
	count := 0

	go func(ps *process.AppPS) {

		netios, err := ps.OutputNetIOStats()
		if err != nil {
			log.Printf("Failed to collect net io, err:%v", err)
		}

		metasCh <- netios
	}(ps)
	count++

	for i := count; i > 0; i-- {
		select {
		case ms := <-metasCh:
			metas = append(metas, ms...)
		}
	}

	render.JSON(w, r, metas)
}

func RuntimeGaugeMetrics(w http.ResponseWriter, r *http.Request, defaultTags map[string]string) {
	metas := []*model.MetricsMeta{}

	metas = append(metas, goroutine.RuntimeGoRoutines(defaultTags)...)
	metas = append(metas, goroutine.RuntimeCgoCalls(defaultTags)...)
	metas = append(metas, goroutine.RuntimeMemory(defaultTags)...)

	ps, err := process.NewAppPS(defaultTags)
	if err != nil {
		log.Printf("Failed to new app process, err:%v", err)
		render.JSON(w, r, metas)
		return
	}

	metasCh := make(chan []*model.MetricsMeta, 0)
	defer close(metasCh)
	count := 0

	go func(ps *process.AppPS) {
		cpuUsage, err := ps.CpuUsage()
		if err != nil {
			log.Printf("Failed to collect cpu usage, err:%v", err)
			metasCh <- nil
			return
		}

		metasCh <- []*model.MetricsMeta{cpuUsage}
	}(ps)
	count++

	go func(ps *process.AppPS) {

		memUsage, err := ps.MemoryUsage()
		if err != nil {
			log.Printf("Failed to collect memory usage, err:%v", err)
			metasCh <- nil
			return
		}

		metasCh <- []*model.MetricsMeta{memUsage}
	}(ps)
	count++

	go func(ps *process.AppPS) {
		count++

		conns, err := ps.OutputConnectionsStats()
		if err != nil {
			log.Printf("Failed to collect connection stats, err:%v", err)
		}

		metasCh <- conns
	}(ps)
	count++

	go func(ps *process.AppPS) {

		memStats, err := ps.OutputMemStats()
		if err != nil {
			log.Printf("Failed to collect memory stats, err:%v", err)
			metasCh <- nil
			return
		}

		metasCh <- []*model.MetricsMeta{memStats}
	}(ps)
	count++

	for i := count; i > 0; i-- {
		select {
		case ms := <-metasCh:
			metas = append(metas, ms...)
		}
	}

	render.JSON(w, r, metas)
}
