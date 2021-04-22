package cmdb

import (
	"testing"
	"context"
	"fmt"
)

func newClient(t *testing.T) *Client {
	url := "http://100.73.45.8:3404"
	cli, err := NewClient(url, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	return cli
}

func TestClient_Query(t *testing.T) {
	cli := newClient(t)

	sql := "SELECT * FROM `label` WHERE ((`name`='region' AND `value` = 'bj') OR (`name` = 'giligili' AND `value` = 'docker-agent-100.73.20')) AND `node-label` <> 0 LIMIT 1"
	values := []interface{}{}

	rows, err := cli.Query(context.Background(), sql, values)
	if err != nil {
		t.Fatal(err)
	}

	for _, row := range rows {
		fmt.Printf("%v\n", string(row))
	}
}
