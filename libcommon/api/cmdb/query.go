package cmdb

import (
	"github.com/sirupsen/logrus"
	"context"
	"strings"
	"fmt"
	"reflect"
	"go/ast"
	"bytes"
	"sync"
	"encoding/json"
)

type Error interface {
	Error() error
}

// Query queries data from cmdb
type Querier interface {
	Error
	// select fields, by default, will select all fields
	Select(query interface{}, args ... interface{}) Querier
	// where condition
	Where(condition interface{}, args ... interface{}) Querier
	// select the first record
	First(out interface{}, where ...interface{}) Querier
	// select the all records
	Find(out interface{}, where ...interface{}) Querier
	// directly query
	Query(sql string, args ...interface{}) (Rows, error)
}

// Creator creates model to cmdb
type Creator interface {
	Create(ctx context.Context, v interface{}) (uint, error)
}

// Updater updates data of model to cmdb
type Updater interface {
	Update(ctx context.Context, v interface{}) error
}

type Deleter interface {
	Delete(ctx context.Context, v interface{}) error
}


// model
type Model struct {
	// primary key
	Id 		uint	`json:"id"`
	// type def id
	TypeName	string	`json:"type_name"`
}

// select column from A where a='b' and c='d' or e='f'
// select * from A where a='b'

type Cmdb struct {
	Err 			error

	EffectedRows	int64

	// search
	search 	*search

	// cmdb client
	cli 	*Client

	logger 	*logrus.Logger
}

func NewCmdb(url string, logger *logrus.Logger, opt *ClientOpt) (*Cmdb, error) {
	if logger == nil {
		logger = logrus.New()
	}

	cmdb := &Cmdb{
		logger: logger,
		search: &search{},
	}

	cli, err := NewClient(url, logger, opt)
	if err != nil {
		return nil, err
	}
	cmdb.cli = cli
	cmdb.search.db = cmdb

	return cmdb, nil
}

func (c *Cmdb) AddError(err error) error {

	if err != nil {
		if err != ErrRecordNotFound {

			errors := Errors(c.GetErrors())
			errors = errors.Add(err)
			if len(errors) >= 1 {
				err = errors
			}
		}

		c.Err = err
	}

	return err
}

func (c *Cmdb) GetErrors() []error {
	if errs, ok := c.Err.(Errors); ok {
		return errs
	} else if c.Err != nil {
		return []error{c.Err}
	}
	return []error{}
}

// get creator
func (c *Cmdb) Creator() Creator {
	return c.cli
}

// get updater
func (c *Cmdb) Updater() Updater {
	return c.cli
}

// get deleter
func (c *Cmdb) Deleter() Deleter {
	return c.cli
}

// get querier
func (c *Cmdb) Querier() Querier {
	return &query{db: c}
}

func (c *Cmdb) clone() *Cmdb {
	cmdb := &Cmdb{
		logger: c.logger,
		search: c.search.clone(),
		cli: c.cli,
	}

	return cmdb
}


type query struct {
	db	*Cmdb

	value interface{}
}

func (q *query) Error() error {
	return q.db.Err
}

// db.Where("name = ?", gongxtao).First(&user)
func (q *query) Where(query interface{}, args ... interface{}) Querier {
	qc := q.clone()
	qc.db.search.Where(query, args)
	return qc
}

// db.Select([]string{'name', 'age'}).Find(&users)
func (q *query) Select(query interface{}, args ... interface{}) Querier {
	qc := q.clone()
	qc.db.search.Select(query, args)
	return qc
}

func (q *query) First(out interface{}, where ...interface{}) Querier {
	ns := q.newScope(out)
	ns.Search.Limit(1)
	return ns.Exec().qr
}

func (q *query) Find(out interface{}, where ...interface{}) Querier {
	ns := q.newScope(out)
	return ns.Exec().qr
}

// directly query
func (q *query) Query(sql string, args ...interface{}) (Rows, error) {
	return q.db.cli.Query(context.TODO(), sql, args...)
}

func (q *query) clone() *query {
	return &query{ db: q.db.clone() }
}

func (q *query) newScope(out interface{}) *Scope {
	qc := q.clone()

	return &Scope{ Search: qc.db.search, Value: out, qr: qc}
}

type Scope struct {
	// search field
	Search          *search
	// output model
	Value           interface{}
	// sql grammar
	SQL             string
	// sql parameter
	SQLVars         []interface{}
	qr 				*query

	ctx 			context.Context

	selectAttrs     *[]string
}

func (s *Scope) Exec() *Scope {
	var (
		//*User -> User
		//&[]*User -> []*User
		results = indirectValue(reflect.ValueOf(s.Value))
		resultType 		reflect.Type

		isSlice, isPtr	bool
	)

	if kind := results.Kind(); kind == reflect.Slice {
		isSlice = true

		resultType = results.Type().Elem()
		results.Set(reflect.MakeSlice(results.Type(), 0, 0))

		if resultType.Kind() == reflect.Ptr {
			isPtr = true

			resultType = resultType.Elem()
		}
	} else if kind != reflect.Struct {
		s.AddErr(fmt.Errorf("unsupported destinational type, should slice or struct"))
		return s
	}

	s.PrepareQuerySql()
	if !s.HasError() {
		s.qr.db.EffectedRows = 0

		rows, err := s.qr.db.cli.Query(s.ctx, s.SQL, s.SQLVars...)
		if err != nil {
			s.AddErr(err)

			return s
		}

		iter := NewRowIterator(rows)
		for iter.Next() {
			if iter.Err() != nil {
				s.AddErr(iter.Err())
				return s
			}

			row := iter.At()
			s.qr.db.EffectedRows ++

			elem := results
			if isSlice {
				elem = reflect.New(resultType).Elem()
			}

			if err := json.Unmarshal(row, elem.Addr().Interface()); err != nil {
				s.AddErr(err)
				return s
			}
			s.fileDefaultValue(elem)

			if isSlice {
				if isPtr {
					results.Set(reflect.Append(results, elem.Addr()))
				} else {
					results.Set(reflect.Append(results, elem))
				}
			}
		}

		if iter.Err() != nil {
			s.AddErr(iter.Err())
		} else  if s.qr.db.EffectedRows == 0  && !isSlice {
			s.AddErr(ErrRecordNotFound)
		}
	}

	return s
}

func (s *Scope) HasError() bool {
	return s.qr.db.Err != nil
}

func (s *Scope) PrepareQuerySql() {
	s.SQL = s.selectSql() + s.tableSql() + s.combineConditionSql() + s.limitSql() + s.orderSql()

	s.qr.db.logger.Debugf("sql: %s, vars: %v", s.SQL, s.SQLVars)
	return
}

// select attrs
func (s Scope) SelectAttrs() []string {
	if s.selectAttrs == nil {
		attrs := []string{}

		qv := s.Search.selects["query"]
		if str, ok := qv.(string); ok {
			attrs = append(attrs, str)
		} else if strs, ok := qv.([]string); ok {
			attrs = append(attrs, strs...)
		} else if strs, ok := qv.([]interface{}); ok {
			for _, str := range strs {
				attrs = append(attrs, fmt.Sprintf("%v", str))
			}
		}

		s.selectAttrs = &attrs
	}

	return *s.selectAttrs
}

func (s *Scope) fileDefaultValue(value reflect.Value) {
	for value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	value.FieldByName("TypeName").Set(reflect.ValueOf(delQuoteValue(s.tableName())))
}

func (s *Scope) inlineCondition(where ...interface{}) *Scope {
	if len(where) > 0 {
		s.Search.Where(where[0], where[1:]...)
	}

	return s
}

func (s *Scope) selectSql() (sql string) {
	if len(s.Search.selects) == 0 {
		sql = "* "
	} else {
		// build sql
		sql = s.buildSelectSql(s.Search.selects)
	}

	sql = fmt.Sprintf("SELECT %s ", sql)

	return
}

// build select fields
func (s *Scope) buildSelectSql(selects map[string]interface{}) (sql string) {
	switch value := s.Search.selects["query"].(type) {
	case string:
		sql = quoteValue(value)
	case []string:
		sql = quoteValue(strings.Join(value, ", "))
	}

	s.sqlVars(s.Search.selects["args"])
	return
}

func (s *Scope) tableSql() (sql string) {
	sql = "FROM " + s.tableName()

	return
}

func (s *Scope) tableName() string {
	if s.Search.tableName != "" {
		return s.Search.tableName
	}

	return s.getModelStruct().TableName()
}

func (s *Scope) getModelStruct() *ModelStruct {
	var ms ModelStruct

	if s.Value == nil {
		return &ms
	}

	rvt := reflect.ValueOf(s.Value).Type()
	for rvt.Kind() == reflect.Ptr || rvt.Kind() == reflect.Slice {
		rvt = rvt.Elem()
	}

	if rvt.Kind() != reflect.Struct {
		return &ms
	}

	if value := modelStructsMap.Get(rvt); value != nil {
		return value
	}

	ms.ModelType = rvt

	// all fields
	for i := 0; i < rvt.NumField(); i ++ {
		if rvtf := rvt.Field(i); ast.IsExported(rvtf.Name) {

			sf := &StructField {
				Name: rvtf.Name,
				Struct: rvtf,
				Tag: rvtf.Tag,
				TagSettings: parseTagSetting(rvtf.Tag),
			}

			if _, ok := sf.TagSettings["-"]; ok {
				sf.IsIgnored = true
			}

			ms.StructFields = append(ms.StructFields, sf)
		}
	}

	modelStructsMap.Set(rvt, &ms)
	return &ms
}

func (s *Scope) orderSql() (sql string) {
	if len(s.Search.orders) == 0 {
		return ""
	}

	orders := []string{}
	for _, order := range s.Search.orders {
		if v, ok := order.(string); ok {
			orders = append(orders, v)
		}
	}

	sql = " ORDER BY " + strings.Join(orders, ", ")
	return
}

func (s *Scope) limitSql() (sql string) {
	if s.Search.limit != nil {
		kind := reflect.TypeOf(s.Search.limit).Kind()
		switch {
		case kind == reflect.String:
			sql = fmt.Sprintf(" LIMIT  %v", s.Search.limit)
		case kind >= reflect.Int && kind <= reflect.Uint64:
			sql = fmt.Sprintf(" LIMIT  %v", s.Search.limit)
		default:
			s.AddErr(fmt.Errorf("invalid limit type: %v", s.Search.limit))
		}
	}

	return
}

func (s *Scope) combineConditionSql() (sql string) {
	var (
		andConditions, orConditions []string
	)

	for _, condition := range s.Search.whereConditions {
		andConditions = append(andConditions, s.buildConditionSql(condition))
	}

	for _, condition := range s.Search.orConditions {
		orConditions = append(orConditions, s.buildConditionSql(condition))
	}

	orSql := strings.Join(orConditions, " OR ")
	andSql := strings.Join(andConditions, " AND ")
	if len(andSql) > 0 {
		if len(orSql) > 0 {
			sql = andSql + " OR " + orSql
		} else {
			sql = andSql
		}
	} else {
		sql = orSql
	}

	if sql != "" {
		sql = fmt.Sprintf(" WHERE %s ", sql)
	}
	return
}

func (s *Scope) buildConditionSql(clause map[string]interface{}) (sql string) {
	switch value := clause["query"].(type) {
	case string:
		sql = fmt.Sprintf("( %s )", value)
	default:
		s.qr.db.logger.Errorf("invalid condition type: %T", value)
	}

	// args
	s.sqlVars(clause["args"])
	return
}

func (s *Scope) sqlVars(v interface{}) {
	switch value := v.(type) {
	case []interface{}:
		for _, vi := range value {
			s.sqlVars(vi)
		}
	default:
		s.SQLVars = append(s.SQLVars, value)
	}
}

func (s *Scope) AddErr(err error) {
	if err != nil {
		s.qr.db.AddError(err)
	}
}

// model struct
type ModelStruct struct {
	StructFields 		[]*StructField
	ModelType 			reflect.Type
	defaultTableName	string
}

func (m *ModelStruct) TableName() string {
	if m.defaultTableName == "" && m.ModelType != nil {

		m.defaultTableName = toSwapName(m.ModelType.Name())
	}

	return quoteValue(m.defaultTableName)
}

// StructField model field's struct definition
type StructField struct {
	Name            string
	IsPrimaryKey    bool
	IsNormal        bool
	IsIgnored       bool
	HasDefaultValue bool
	Tag             reflect.StructTag
	TagSettings     map[string]string
	Struct          reflect.StructField
}

func parseTagSetting(tags reflect.StructTag) map[string]string {
	setting := map[string]string{}
	for _, str := range []string{tags.Get("sql"), tags.Get("gcmdb")} {
		tags := strings.Split(str, ";")
		for _, value := range tags {
			v := strings.Split(value, ":")
			k := strings.TrimSpace(strings.ToUpper(v[0]))
			if len(v) >= 2 {
				setting[k] = strings.Join(v[1:], ":")
			} else {
				setting[k] = k
			}
		}
	}
	return setting
}

type safeMap struct {
	m map[string]string
	l *sync.RWMutex
}

func (s *safeMap) Set(key string, value string) {
	s.l.Lock()
	defer s.l.Unlock()
	s.m[key] = value
}

func (s *safeMap) Get(key string) string {
	s.l.RLock()
	defer s.l.RUnlock()
	return s.m[key]
}

func newSafeMap() *safeMap {
	return &safeMap{l: new(sync.RWMutex), m: make(map[string]string)}
}

var smap = newSafeMap()

type strCase bool

const (
	lower strCase = false
	upper strCase = true
)

// toSwapName convert string to field name
func toSwapName(name string) string {
	if v := smap.Get(name); v != "" {
		return v
	}

	if name == "" {
		return ""
	}

	var (
		value                        = name
		buf                          = bytes.NewBufferString("")
		lastCase, currCase, nextCase strCase
	)

	for i, v := range value[:len(value)-1] {
		nextCase = strCase(value[i+1] >= 'A' && value[i+1] <= 'Z')
		if i > 0 {
			if currCase == upper {
				if lastCase == upper && nextCase == upper {
					buf.WriteRune(v)
				} else {
					if value[i-1] != '_' && value[i+1] != '_' {
						buf.WriteRune('_')
					}
					buf.WriteRune(v)
				}
			} else {
				buf.WriteRune(v)
				if i == len(value)-2 && nextCase == upper {
					buf.WriteRune('_')
				}
			}
		} else {
			currCase = upper
			buf.WriteRune(v)
		}
		lastCase = currCase
		currCase = nextCase
	}

	buf.WriteByte(value[len(value)-1])

	s := strings.ToLower(buf.String())
	smap.Set(name, s)
	return s
}

func indirectValue(value reflect.Value) reflect.Value {
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	return value
}

type safeModelStructsMap struct {
	m map[reflect.Type]*ModelStruct
	l *sync.RWMutex
}

func (s *safeModelStructsMap) Set(key reflect.Type, value *ModelStruct) {
	s.l.Lock()
	defer s.l.Unlock()
	s.m[key] = value
}

func (s *safeModelStructsMap) Get(key reflect.Type) *ModelStruct {
	s.l.RLock()
	defer s.l.RUnlock()
	return s.m[key]
}

func newModelStructsMap() *safeModelStructsMap {
	return &safeModelStructsMap{l: new(sync.RWMutex), m: make(map[reflect.Type]*ModelStruct)}
}

var modelStructsMap = newModelStructsMap()

func quoteValue(v string) string {
	return "`"+v+"`"
}

func delQuoteValue(v string) string {
	return strings.Trim(v, "`")
}

