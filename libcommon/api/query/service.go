package query

import (
	"time"
	"net/http"
	"bytes"
	"golang.org/x/net/context"
	"io/ioutil"
	"fmt"
	"encoding/json"
	"sync"
)

/*
@brief Query the local agent for the address information of services
*/

const (
	timeout	= 10 * time.Second
	url 	= "http://localhost:3335/api/objs/RegistryService/QueryServiceAddress"

	userAgent	= "nerv-agent"

	// cache
	defaultExpiration	= 60 * time.Second
	defaultForever		= 0
	flushInterval		= 10 * time.Second
)

var (
	localQuery Query
)

func init() {
	if localQuery == nil {
		localQuery = &LocalQuery{timeout: timeout, url: url}
	}
}

type Query interface {
	QueryAddress(ctx context.Context, name string) (string, error)
}

func QueryServiceAddress(ctx context.Context, name string) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if name == "" {
		return "", fmt.Errorf("name is empty")
	}

	return localQuery.QueryAddress(ctx, name)
}


type LocalQuery struct {
	sync.RWMutex
	items 	map[string]*item

	timeout	time.Duration
	url 	string
}

type item struct {
	rs 			*result
	expiration 	*time.Time	// expiration ntp
}

func (i *item) expire() bool {
	if i.expiration == nil {
		return false
	}

	return i.expiration.Before(time.Now())
}

type result struct {
	Address 	string	`json:"address"`
	Name 		string	`json:"name"`
}

func (p *LocalQuery) QueryAddress(ctx context.Context, name string) (string, error) {

	rs, ok := p.get(name)
	if ok {
		return rs.Address, nil
	}

	return p.query(ctx, name)
}

func (p *LocalQuery) set(k string, v *result, d time.Duration) {
	p.Lock()

	var et *time.Time

	if d > 0 {
		t := time.Now().Add(d)
		et = &t
	}

	if p.items == nil {
		p.items = make(map[string]*item)
	}

	p.items[k] = &item{
		rs: v,
		expiration: et,
	}

	p.Unlock()
}

func (p *LocalQuery) get(k string) (*result, bool) {
	p.RLock()
	item, ok := p.items[k]
	p.RUnlock()

	if !ok || item.expire() {
		return nil, false
	}

	return item.rs, true
}

func (p *LocalQuery) flushExpire() {
	p.Lock()

	for k, v := range p.items {
		if v.expire() {
			delete(p.items, k)
		}
	}

	p.Unlock()
}

func (p *LocalQuery) run() {
	for {
		select{
		case <- time.After(flushInterval):
			p.flushExpire()
		}
	}
}

func (p *LocalQuery) query(ctx context.Context, name string) (string, error) {
	body, err := json.Marshal([]string{name})
	if err != nil {
		return "", err
	}

	client := &http.Client{
		Timeout: p.timeout,
		Transport: http.DefaultTransport,
	}

	req, err := http.NewRequest("POST", p.url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/plain, text/*, */*")
	req.Header.Set("User-Agent", userAgent)
	req = req.WithContext(ctx)

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	if res.StatusCode != 200 {
		return "", fmt.Errorf("alias code:%d, message:%s", res.StatusCode, string(content))
	}

	// parse rpc response
	vs := []interface{}{}
	if err := json.Unmarshal(content, &vs); err != nil {
		return "", fmt.Errorf("failed parse content of rpc, err:%v", err)
	}
	if len(vs) < 2 {
		return "", fmt.Errorf("reponse content length is not enough")
	}
	// error type
	if vs[1] != nil {
		return "", fmt.Errorf("local agent handle error, err:%s", vs[1])
	}

	// result
	v, ok := vs[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("error reponse content of result")
	}
	b, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("failed encode content of result")
	}
	r := &result{}
	if err := json.Unmarshal(b, r); err != nil {
		return "", err
	}

	// set cache
	p.set(name, r, defaultExpiration)

	return r.Address, nil
}


