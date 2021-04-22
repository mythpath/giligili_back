package metrics

import (
	"bytes"
	"fmt"
	"math"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cespare/xxhash"
)

const (
	// max size is 16384
	hashSize = 1 << 14
	hashMark = hashSize - 1

	// tag hash size
	tagHashSize = 1 << 10
	tagHashMark = tagHashSize - 1
)

// separator
const (
	sep = "\xff"

	collectorTable = "____Collector_Table____"
)

type Collector struct {
	registry Registry

	t *table
}

func NewCollector(r Registry) *Collector {
	cl := &Collector{
		registry: r,
		t:        newTables(),
	}

	cl.register()

	return cl
}

func (p *Collector) init(r Registry) {
	p.registry = r
	p.t = newTables()
	// collector register into registry
	p.register()
}

func (p *Collector) NewMetrics(name, field string) {
	p.t.getOrCreate(name, field)
}

func (p *Collector) GetMetrics(name, field string) *tagDesc {
	return p.t.getOrCreate(name, field)
}

func (p *Collector) CleanTable() {
	for i, _ := range p.t.columns {
		p.t.locks[i].Lock()
		p.t.columns[i] = newHashMetrics()
		p.t.locks[i].Unlock()
	}
}

// export all samples
func (p *Collector) exporterMetrics() <-chan *Sample {
	descs := p.t.getAllDesc()
	capacity := len(descs)
	sampleCh := make(chan *Sample, capacity)
	var wg sync.WaitGroup
	wg.Add(capacity)

	go func() {
		wg.Wait()
		close(sampleCh)
	}()

	go func() {
		for _, dc := range descs {
			sample := &Sample{
				Name: dc.Desc().Name,
				Fields: map[string]interface{}{
					dc.Desc().Field: dc.Value(),
				},
				Tags:      dc.Desc().Tags,
				Type:      dc.Desc().Type,
				Timestamp: time.Now().UnixNano(),
			}

			sampleCh <- sample
			wg.Done()
		}
	}()

	return sampleCh
}

// collector registry into Registry
func (p *Collector) register() {
	p.registry.Put(NewMetrics(collectorTable, p.exporterMetrics).Build())
}

func newTables() *table {
	t := &table{}

	for i := range t.columns {
		t.columns[i] = newHashMetrics()
	}

	return t
}

// virtual table
type table struct {
	columns [hashSize]*hashDesc
	locks   [hashSize]lock
}

func (p *table) getOrCreate(name, field string) *tagDesc {
	hash := hashField(name, field)
	i := hashMark & hash

	p.locks[i].Lock()
	tm, ok := p.columns[i].hashes[hash]
	if !ok {
		tm = newTagMetrics(name, field)
		p.columns[i].hashes[hash] = tm
	}
	p.locks[i].Unlock()

	return tm
}

func (p *table) getAllDesc() []Desc {
	var ds []Desc

	for i := range p.locks {
		p.locks[i].RLock()

		ds = append(ds, p.columns[i].getAllDesc()...)

		p.locks[i].RUnlock()
	}

	return ds
}

// part lock
type lock struct {
	sync.RWMutex

	// Padding to avoid multiple locks being on the same cache line.
	_ [40]byte
}

func newHashMetrics() *hashDesc {
	return &hashDesc{
		hashes: make(map[uint64]*tagDesc),
	}
}

// key: hash(name,field)
type hashDesc struct {
	hashes map[uint64]*tagDesc
}

func (p *hashDesc) getAllDesc() []Desc {
	var ds []Desc

	for _, td := range p.hashes {
		ds = append(ds, td.getAllDesc()...)
	}

	return ds
}

// tags -> []metric
type tagDesc struct {
	sync.RWMutex
	name  string
	field string

	tms   [tagHashSize]*hashTagDesc
	locks [tagHashSize]lock
}

func newTagMetrics(name, field string) *tagDesc {
	tm := &tagDesc{
		name:  name,
		field: field,
	}

	for i := range tm.tms {
		tm.tms[i] = newHashTagDesc(name, field)
	}

	return tm
}

func (p *tagDesc) getAllDesc() []Desc {
	var ds []Desc

	for i := range p.locks {
		p.locks[i].RLock()

		ds = append(ds, p.tms[i].getAllDesc()...)

		p.locks[i].RUnlock()
	}

	return ds
}

// gauge metric
func (p *tagDesc) Gauge(tags map[string]string) Gauge {
	hash := hashTags(tags)
	i := tagHashMark & hash

	p.locks[i].Lock()
	dc := p.tms[i].get(hash, tags)
	// create
	if dc == nil {
		dc = &SampleGauge{
			desc: &desc{
				Name:  p.name,
				Field: p.field,
				Tags:  tags,
				Type:  "GAUGE",
			},
		}
		p.tms[i].set(hash, dc)
	}
	p.locks[i].Unlock()

	gauge, ok := dc.(Gauge)
	if !ok {
		panic(fmt.Sprintf("invalid metric type: %T", dc))
	}

	return gauge
}

// counter metric
func (p *tagDesc) Counter(tags map[string]string) Counter {
	hash := hashTags(tags)
	i := tagHashMark & hash

	p.locks[i].Lock()
	dc := p.tms[i].get(hash, tags)
	// create
	if dc == nil {
		dc = &SampleCounter{
			desc: &desc{
				Name:  p.name,
				Field: p.field,
				Tags:  tags,
				Type:  "COUNTER",
			},
		}
		p.tms[i].set(hash, dc)
	}
	p.locks[i].Unlock()

	counter, ok := dc.(Counter)
	if !ok {
		panic(fmt.Sprintf("invalid metric type: %T", dc))
	}

	return counter
}

func newHashTagDesc(name, field string) *hashTagDesc {
	return &hashTagDesc{
		name:  name,
		field: field,
		htms:  make(map[uint64][]Desc),
	}
}

// tags -> metric
type hashTagDesc struct {
	name  string
	field string
	htms  map[uint64][]Desc
}

func (p *hashTagDesc) getAllDesc() []Desc {
	var ds []Desc
	for _, dc := range p.htms {
		ds = append(ds, dc...)
	}

	return ds
}

func (p *hashTagDesc) get(hash uint64, tags map[string]string) Desc {

	for _, desc := range p.htms[hash] {
		if equalTags(desc.Desc().Tags, tags) {
			return desc
		}
	}

	return nil
}

func (p *hashTagDesc) set(hash uint64, desc Desc) {
	ds := p.htms[hash]
	for i, dc := range ds {
		if equalTags(dc.Desc().Tags, desc.Desc().Tags) {
			ds[i] = desc
			return
		}
	}

	if ds == nil {
		p.htms[hash] = make([]Desc, 0, 1)
	}

	p.htms[hash] = append(p.htms[hash], desc)
}

func (p *hashTagDesc) del(hash uint64, desc Desc) {
	var sdecs []Desc

	for _, dc := range p.htms[hash] {
		if !equalTags(dc.Desc().Tags, desc.Desc().Tags) {
			sdecs = append(sdecs, dc)
		}
	}

	if len(sdecs) == 0 {
		delete(p.htms, hash)
	} else {
		p.htms[hash] = sdecs
	}
}

type desc struct {
	Name      string            `json:"name"`
	Field     string            `json:"field"`
	Tags      map[string]string `json:"tags"`
	Type      string            `json:"type"`
	Timestamp int64             `json:"timestamp"`
}

func (d *desc) TagValues() map[string]string {
	return d.Tags
}

func (d *desc) String() string {
	return fmt.Sprintf("name:%s, field:%s, tags:%v, type:%s", d.Name, d.Field, d.Tags, d.Type)
}

// different type metric
type Desc interface {
	Desc() *desc
	Value() interface{}
}

type Gauge interface {
	Desc
	Set(v interface{})
	Dec()
	Inc()
	Add(float64)
	Sub(float64)
}

type SampleGauge struct {
	desc *desc

	// value type is int64/uint64/string/bool/float64
	value interface{}
	sync.RWMutex
}

func (p *SampleGauge) Desc() *desc {
	return p.desc
}

// int64/uint64/string/bool/float64
func (p *SampleGauge) Set(v interface{}) {
	p.Lock()
	p.value = v
	p.Unlock()
}

func (p *SampleGauge) Value() interface{} {
	p.RLock()
	v := p.value
	p.RUnlock()

	return v
}

func (p *SampleGauge) Dec()          { p.Add(-1) }
func (p *SampleGauge) Inc()          { p.Add(1) }
func (p *SampleGauge) Sub(v float64) { p.Add(-1 * v) }
func (p *SampleGauge) Add(v float64) {
	p.Lock()
	defer p.Unlock()

	if v == 0 {
		return
	}

	if p.value == nil {
		p.value = float64(0)
	}

	switch ov := p.value.(type) {
	case float64:
		p.value = ov + v
	}
}

type Counter interface {
	Desc
	Clear()
	Inc()
	Add(float64)
}

type SampleCounter struct {
	desc  *desc
	value uint64
}

func (p *SampleCounter) Desc() *desc {
	return p.desc
}

func (p *SampleCounter) Value() interface{} {
	old := atomic.LoadUint64(&p.value)

	return math.Float64frombits(old)
}

func (p *SampleCounter) Clear() { atomic.StoreUint64(&p.value, 0) }
func (p *SampleCounter) Inc()   { p.Add(1) }
func (p *SampleCounter) Add(v float64) {
	if v == 0 {
		return
	}

	for {
		old := atomic.LoadUint64(&p.value)
		new := math.Float64bits(math.Float64frombits(old) + v)
		if atomic.CompareAndSwapUint64(&p.value, old, new) {
			return
		}
	}
}

func hashTags(tags map[string]string) uint64 {
	tagKs := make(sort.StringSlice, 0, len(tags))
	for k := range tags {
		tagKs = append(tagKs, k)
	}
	sort.Sort(tagKs)

	var b bytes.Buffer
	for _, k := range tagKs {
		b.WriteString(k)
		b.WriteString(sep)
		b.WriteString(tags[k])
		b.WriteString(sep)
	}

	return xxhash.Sum64String(b.String())
}

func hashField(name string, field string) uint64 {
	return xxhash.Sum64String(fmt.Sprintf("%s%s%s", name, sep, field))
}

func equalTags(left, right map[string]string) bool {
	if len(left) != len(right) {
		return false
	}

	for k, v := range left {
		if rv, ok := right[k]; !ok || rv != v {
			return false
		}
	}

	return true
}
