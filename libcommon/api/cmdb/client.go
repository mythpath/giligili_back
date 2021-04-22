package cmdb

import (
	"context"
	"net/url"
	"net/http"
	"github.com/sirupsen/logrus"
	"time"
	"net"
	"io/ioutil"
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
)

const (
	UPCreate = "/api/objs/InstanceService/Create"
	UPUpdate = "/api/objs/InstanceService/Update"
	UPQuery = "/api/objs/InstanceService/Query"
	UPDelete = "/api/objs/InstanceService/Delete"
)

var (
	defaultClientOpt = &ClientOpt{
		Headers: map[string]string{
			"NERV-USER": "cmdb-client",
		},
		Timeout: 30 * time.Second,
		Keepalive: 30 * time.Second,
	}
)

// cmdb client
type Client struct {
	// http client
	hc 		*http.Client
	url 	*url.URL

	logger 	*logrus.Logger
	opt 	*ClientOpt
}

// options
type ClientOpt struct {
	// http header
	Headers 	map[string]string
	// http request timeout
	Timeout 	time.Duration
	// keepalive
	Keepalive 	time.Duration
}

// query result
type queryResult struct {
	Data 	Rows
}


func NewClient(urls string, lg *logrus.Logger, opt *ClientOpt) (*Client, error) {
	if lg == nil {
		lg = logrus.New()
	}

	if opt == nil {
		opt = defaultClientOpt
	}

	tr := &http.Transport{
		MaxIdleConns: 32,
		DialContext: (&net.Dialer{
			Timeout: opt.Timeout,
			KeepAlive: opt.Keepalive,
		}).DialContext,
	}

	cli := &http.Client{
		Transport: tr,
		Timeout: opt.Timeout,
	}

	uurl, err := url.Parse(urls)
	if err != nil {
		return nil, err
	}

	return &Client{
		hc: cli,
		url: uurl,
		logger: lg,
		opt: opt,
	}, nil
}

// query data from cmdb
func (c *Client) Query(ctx context.Context, sql string, args ...interface{}) (Rows, error) {
	body, err := c.jsonEnc([]interface{}{sql, args})
	if err != nil {
		return nil, err
	}

	c.url.Path = UPQuery
	rs, err := c.do(ctx, "POST", body)
	if err != nil {
		return nil, err
	}
	rsb, err := ioutil.ReadAll(rs.Body)
	if err != nil {
		return nil, err
	}
	defer rs.Body.Close()

	if rs.StatusCode != http.StatusOK {
		c.logger.WithField("url path", c.url.Path).Errorf("handle failed, alias code: %v, message: %s", rs.StatusCode, string(rsb))
		return nil, ErrHandleFailed
	}

	var result Rows
	// parse
	if err := c.parse(rsb, func(rsv []json.RawMessage) error {
		if len(rsv) < 2 {
			return ErrResultSize
		}
		err := string(rsv[1])
		if err != "null" {
			c.logger.WithField("url path", c.url.Path).Errorf("handle err: %s", err)
			return fmt.Errorf("%s", err)
		}

		qr := &queryResult{}
		if err := json.Unmarshal(rsv[0], qr); err != nil {
			c.logger.WithField("url path", c.url.Path).Errorf("parse err: %v", err)
			return err
		}

		result = qr.Data
		return nil
	}); err != nil {
		return nil, err
	}

	return result, nil
}

// create model to cmdb
// the first result is primary key
func (c *Client) Create(ctx context.Context, v interface{}) (uint, error) {
	body, err := c.jsonEnc(v)
	if err != nil {
		return 0, nil
	}

	c.logger.WithField("action", "create").Debugf("body: %s", string(body))

	c.url.Path = UPCreate
	rs, err := c.do(ctx, "POST", body)
	if err != nil {
		return 0, err
	}
	rsb, err := ioutil.ReadAll(rs.Body)
	if err != nil {
		return 0, err
	}
	defer rs.Body.Close()

	if rs.StatusCode != http.StatusOK {
		c.logger.WithField("url path", c.url.Path).Errorf("handle failed, alias code: %v, message: %s", rs.StatusCode, string(rsb))
		return 0, ErrHandleFailed
	}

	var id uint
	// parse
	if err := c.parse(rsb, func(rsv []json.RawMessage) error {
		if len(rsv) < 2 {
			return ErrResultSize
		}
		er := string(rsv[1])
		if er != "null" {
			c.logger.WithField("url path", c.url.Path).Errorf("handle err: %s", er)
			return fmt.Errorf("%s", er)
		}

		uid, err := strconv.ParseUint(string(rsv[0]), 10, 32)
		if err != nil {
			c.logger.WithField("url path", c.url.Path).Errorf("parse id error: %v", err)
			return ErrElementType
		}

		id = uint(uid)
		return nil
	}); err != nil {
		return 0, err
	}

	return id, nil
}

// update model to cmdb
func (c *Client) Update(ctx context.Context, v interface{}) error {
	body, err := c.jsonEnc(v)
	if err != nil {
		return nil
	}

	c.logger.WithField("action", "update").Debugf("body: %s", string(body))

	c.url.Path = UPUpdate
	rs, err := c.do(ctx, "POST", body)
	if err != nil {
		return err
	}
	rsb, err := ioutil.ReadAll(rs.Body)
	if err != nil {
		return err
	}
	defer rs.Body.Close()

	if rs.StatusCode != http.StatusOK {
		c.logger.WithField("url path", c.url.Path).Errorf("handle failed, alias code: %v, message: %s", rs.StatusCode, string(rsb))
		return ErrHandleFailed
	}

	// parse
	if err := c.parse(rsb, func(rsv []json.RawMessage) error {
		if len(rsv) != 1 {
			return ErrResultSize
		}
		err := string(rsv[0])
		if err != "null" {
			c.logger.WithField("url path", c.url.Path).Errorf("handle err: %s", err)
			return fmt.Errorf("%s", err)
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

// delete model to cmdb
// v is array(typeId, primaryId)
func (c *Client) Delete(ctx context.Context, v interface{}) error {
	body, err := c.jsonEnc(v)
	if err != nil {
		return nil
	}

	c.url.Path = UPDelete
	rs, err := c.do(ctx, "POST", body)
	if err != nil {
		return err
	}
	rsb, err := ioutil.ReadAll(rs.Body)
	if err != nil {
		return err
	}
	defer rs.Body.Close()

	if rs.StatusCode != http.StatusOK {
		c.logger.WithField("url path", c.url.Path).Errorf("handle failed, alias code: %v, message: %s", rs.StatusCode, string(rsb))
		return ErrHandleFailed
	}

	// parse
	if err := c.parse(rsb, func(rsv []json.RawMessage) error {
		if len(rsv) != 1 {
			return ErrResultSize
		}
		err := string(rsv[0])
		if err != "null" {
			c.logger.WithField("url path", c.url.Path).Errorf("handle err: %", err)
			return fmt.Errorf("%s", err)
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

// request http server
func (c *Client) do(ctx context.Context, method string, body []byte) (*http.Response, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	return c.hc.Do(c.buildRequest(method, body).WithContext(ctx))
}

// build request
func (c *Client) buildRequest(method string, body []byte) *http.Request {
	if method == "" {
		method = "GET"
	}

	rq := &http.Request{
		Method: 	method,
		URL: 		c.url,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Host:       c.url.Host,
	}

	// header
	for k, v := range c.opt.Headers {
		rq.Header.Set(k, v)
	}
	if rq.Header.Get("Accept") == "" {
		rq.Header.Set("Accept", "text/plain, text/*, */*")
	}

	lenBody := int64(len(body))
	if lenBody > 0 {
		rq.ContentLength = lenBody
		rq.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	}

	return rq
}

// json encode, if v is not []byte
func (c *Client) jsonEnc(v interface{}) ([]byte, error) {
	switch vt := v.(type) {
	case []byte:
		return vt, nil
	default:
		return json.Marshal(v)
	}
}

// parse result
func (c *Client) parse(rsb []byte, f func(rsv []json.RawMessage) error) error {
	rvs := []json.RawMessage{}
	if err := json.Unmarshal(rsb, &rvs); err != nil {
		return err
	}

	return f(rvs)
}

