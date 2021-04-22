package process

import (
	"testing"
	"log"
	"time"
)

func TestProcessCpu(t *testing.T) {
	ps, err := NewAppPS(nil)
	if err != nil {
		t.Fatal(err)
	}

	meta, err := ps.MemoryUsage()
	if err != nil {
		t.Fatal(err)
	}
	log.Println(meta)

	meta, err = ps.CpuUsage()
	if err != nil {
		t.Fatal(err)
	}
	log.Println(meta)

	meta, _ = ps.OutputMemStats()
	log.Println(meta)

	metas, _ := ps.OutputConnectionsStats()
	log.Println(metas)

	metas, _ = ps.OutputNetIOStats()
	log.Println(metas)

	time.Sleep(5 * time.Second)

	meta, err = ps.CpuUsage()
	if err != nil {
		t.Fatal(err)
	}
	log.Println(meta)

}
