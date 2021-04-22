package secret_client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	DefaultTimeout   = 3 * time.Second
	ClientAuthHeader = "Secret-Client"
)

type SecrectClient struct {
	auth       string
	endpoints  []string
	pool       chan string
	httpClient *http.Client
}

func New(auth string, endpoints []string) *SecrectClient {
	pool := make(chan string, len(endpoints))
	for _, e := range endpoints {
		pool <- e
	}
	return &SecrectClient{
		auth:       auth,
		endpoints:  endpoints,
		pool:       pool,
		httpClient: &http.Client{Timeout: DefaultTimeout},
	}
}

func (c *SecrectClient) Auth() string {
	return c.auth
}
func (c *SecrectClient) Endpoints() []string {
	return c.endpoints
}

func (c *SecrectClient) GetEndpoint() string {
	e := <-c.pool
	c.pool <- e
	return e
}

func (c *SecrectClient) prepareRequest(method, path string, data interface{}) (*http.Request, error) {
	body, err := json.Marshal([]interface{}{data})
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s%s", c.GetEndpoint(), path)
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set(ClientAuthHeader, c.auth)
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func (c *SecrectClient) Do(method, path string, data interface{}) (*http.Response, error) {
	req, err := c.prepareRequest(method, path, data)
	if err != nil {
		return nil, err
	}
	return c.httpClient.Do(req)
}

func (c *SecrectClient) parseResponse(resp *http.Response) (map[string]interface{}, error) {
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("alias code: %d, message: %s", resp.StatusCode, string(content))
	}

	var ret []interface{}
	if err := json.Unmarshal(content, &ret); err != nil {
		return nil, err
	}
	r, ok := ret[0].(map[string]interface{})
	if !ok {
		return nil, errors.New("response result not type map[string]interface{}")
	}
	return r, nil
}

func (c *SecrectClient) Encrypt(data map[string]interface{}) (map[string]interface{}, error) {
	resp, err := c.Do("POST", "/SecretKey/Encrypt", data)
	if err != nil {
		return nil, err
	}
	return c.parseResponse(resp)
}

func (c *SecrectClient) Decrypt(data map[string]interface{}) (map[string]interface{}, error) {
	resp, err := c.Do("POST", "/SecretKey/Decrypt", data)
	if err != nil {
		return nil, err
	}
	return c.parseResponse(resp)
}
