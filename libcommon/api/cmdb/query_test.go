package cmdb

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"fmt"
	"github.com/sirupsen/logrus"
)

type Node struct {
	Model
	Address string 	`json:"address"`
}

type Label struct {
	Model
	Name 	string
	Value 	string
}



func TestCmdb_Querier(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	url := "http://100.73.45.8:3404"
	db, err := NewCmdb(url, logrus.StandardLogger(), nil)
	assert.NoError(t, err)

	user := &Node{}
	db.Querier().First(user)
	assert.NoError(t, db.Err)

	fmt.Println(user)

	label := &Label{}
	err = db.Querier().Where("`node-label` = ?", 1).First(label).Error()
	assert.NoError(t, err)
	fmt.Println(label)
}

func TestCmdb_QuerierArray(t *testing.T) {
	url := "http://100.73.45.8:3404"
	db, err := NewCmdb(url, nil, nil)
	assert.NoError(t, err)

	user := &[]*Node{}
	db.Querier().Find(user)
	assert.NoError(t, db.Err)

	for _, v := range *user {
		fmt.Println(v)
	}
}
