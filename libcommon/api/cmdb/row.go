package cmdb

import (
	"encoding/json"
	"errors"
)

type Row json.RawMessage

type Rows []Row

// MarshalJSON returns m as the JSON encoding of m.
func (m Row) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}
	return m, nil
}

// UnmarshalJSON sets *m to a copy of data.
func (m *Row) UnmarshalJSON(data []byte) error {
	if m == nil {
		return errors.New("json.RawMessage: UnmarshalJSON on nil pointer")
	}
	*m = append((*m)[0:0], data...)
	return nil
}

type RowIterator struct {
	rows 	[]Row

	cur 	Row

	total 	int
	i 		int
}

func NewRowIterator(rows Rows) *RowIterator {
	return &RowIterator{
		rows: rows,
		total: len(rows),
		i: 0,
		cur: nil,
	}
}

func (r *RowIterator) Next() bool {
	if r.i >= r.total {
		return false
	}

	r.cur = r.rows[r.i]
	r.i ++

	return true
}

func (r *RowIterator) At() Row {
	return r.cur
}

func (r *RowIterator) Err() error {
	return nil
}

