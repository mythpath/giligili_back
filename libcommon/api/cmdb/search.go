package cmdb

type search struct {
	db             	 *Cmdb
	whereConditions  []map[string]interface{}
	orConditions     []map[string]interface{}
	selects          map[string]interface{}
	omits            []string
	orders           []interface{}
	preload          []searchPreload
	offset           interface{}
	limit            interface{}

	tableName 		 string
}

type searchPreload struct {
	schema     string
	conditions []interface{}
}

func (s *search) clone() *search {
	clone := *s
	return &clone
}

func (s *search) Where(query interface{}, values ...interface{}) *search {
	s.whereConditions = append(s.whereConditions, map[string]interface{}{"query": query, "args": values})
	return s
}

func (s *search) Or(query interface{}, values ...interface{}) *search {
	s.orConditions = append(s.orConditions, map[string]interface{}{"query": query, "args": values})
	return s
}


func (s *search) Order(value interface{}, reorder ...bool) *search {
	if len(reorder) > 0 && reorder[0] {
		s.orders = []interface{}{}
	}

	if value != nil && value != "" {
		s.orders = append(s.orders, value)
	}
	return s
}

func (s *search) Select(query interface{}, args ...interface{}) *search {
	s.selects = map[string]interface{}{"query": query, "args": args}
	return s
}

func (s *search) Omit(columns ...string) *search {
	s.omits = columns
	return s
}

func (s *search) Limit(limit interface{}) *search {
	s.limit = limit
	return s
}

func (s *search) Offset(offset interface{}) *search {
	s.offset = offset
	return s
}

func (s *search) TableName(name string) *search {
	s.tableName = name
	return s
}

func (s *search) Preload(schema string, values ...interface{}) *search {
	var preloads []searchPreload
	for _, preload := range s.preload {
		if preload.schema != schema {
			preloads = append(preloads, preload)
		}
	}
	preloads = append(preloads, searchPreload{schema, values})
	s.preload = preloads
	return s
}

