package service

import (
	"testing"
	"net/http"
	"time"
)

func TestRuntimeMetrics(t *testing.T) {
	r := &http.Request{}

	RuntimeGaugeMetrics(nil, r, nil)

	time.Sleep(2 * time.Second)
}
