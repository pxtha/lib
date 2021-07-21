package query

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ======================================================
// Time
// ======================================================
type Time struct {
	time.Time
}

// Json support
// ---------------------------------------------
func (t *Time) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	if s == "null" || s == "" {
		t.Time = time.Time{}
		return
	}
	t.Time, err = time.Parse(time.RFC3339, s)
	return
}

func (t *Time) MarshalJSON() ([]byte, error) {
	if t.Time.UnixNano() == (time.Time{}).UnixNano() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", t.Time.Format(time.RFC3339))), nil
}

// Gorm support
// ---------------------------------------------
func (t *Time) Scan(src interface{}) (err error) {
	var tt Time
	switch src.(type) {
	case time.Time:
		tt.Time = src.(time.Time)
	default:
		return errors.New("incompatible type for Skills")
	}
	*t = tt
	return
}

func (t Time) Value() (driver.Value, error) {
	if !t.IsSet() {
		return "null", nil
	}
	return t.Time, nil
}

func (t *Time) IsSet() bool {
	return t.UnixNano() != (time.Time{}).UnixNano()
}

func Jsonb(v string) postgres.Jsonb {
	return postgres.Jsonb{RawMessage: json.RawMessage(v)}
}

func FloorOf(a, b int) int {
	return int(math.Floor(float64(a) / float64(b)))
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func MustBeInt(v string) int {
	res, err := strconv.Atoi(v)
	if err != nil {
		return 0
	}
	return res
}

func CurrentUser(c *http.Request) (uuid.UUID, error) {
	userIdStr := c.Header.Get("x-user-id")
	if strings.Contains(userIdStr, "|") {
		userIdStr = strings.Split(userIdStr, "|")[0]
	}
	res, err :=  uuid.Parse(userIdStr)
	if err != nil {
		userIdStr = c.Header.Get("x-user-extra")
		if strings.Contains(userIdStr, "|") {
			userIdStr = strings.Split(userIdStr, "|")[0]
		}
	} else {
		return res, nil
	}

	return uuid.Parse(userIdStr)
}

func PaginationQuery(c *http.Request) (int, int) {
	currentPage := 1
	perPage := 30
	if c.Method == http.MethodGet || c.Method == http.MethodDelete {
		currentPage = MustBeInt(strings.Trim(c.URL.Query().Get("page"), " "))
		perPage = MustBeInt(strings.Trim(c.URL.Query().Get("size"), " "))
	} else if c.Method == http.MethodPost ||
		c.Method == http.MethodPatch ||
		c.Method == http.MethodPut {
		currentPage = MustBeInt(strings.Trim(c.FormValue("page"), " "))
		perPage = MustBeInt(strings.Trim(c.FormValue("size"), " "))
	}
	// validate
	if currentPage < 1 {
		currentPage = 1
	}
	if perPage < 1 {
		perPage = 30
	}
	return currentPage, perPage
}

//=========================================================================
// Minimal Chain Data Getter For  Framework
//=========================================================================
type chain struct {
	req           *http.Request
	ResPagination Pagination
	q             *gorm.DB
	err           error
	checkOn       string
	updater       map[string]interface{}
}

type Pagination struct {
	Page     int
	Size     int
	NextPage int
	LastPage int
	Total    int
}

func With(w *http.Request, db *gorm.DB) *chain {
	return &chain{req: w,q: db}
}

func (c *chain) On(checkOn string) *chain {
	c.checkOn = checkOn
	return c
}

func (c *chain) Get(name string) string {
	if c.checkOn == "query" {
		return strings.Trim(c.req.URL.Query().Get(name), " ")
	} else if c.checkOn == "form" {
		return strings.Trim(c.req.FormValue(name), " ")
	} else {
		if c.req.Method == http.MethodGet || c.req.Method == http.MethodDelete {
			return strings.Trim(c.req.URL.Query().Get(name), " ")
		} else if c.req.Method == http.MethodPost ||
			c.req.Method == http.MethodPatch ||
			c.req.Method == http.MethodPut {
			return strings.Trim(c.req.FormValue(name), " ")
		}
	}
	return ""
}

func (c *chain) Uuid(name string) (uuid.UUID, error) {
	v := c.Get(name)
	return uuid.Parse(v)
}

func (c *chain) Int(name string) (int, error) {
	v := c.Get(name)
	return strconv.Atoi(v)
}
func (c *chain) Date(name string) (time.Time, error) {
	v := c.Get(name)
	return time.Parse(time.RFC3339, v)
}
func (c *chain) Time(name string) (time.Time, error) {
	v := c.Get(name)
	return time.Parse("15:04", v)
}
func (c *chain) Bool(name string) bool {
	v := c.Get(name)
	return strings.ToLower(v) == "true" || v == "1"
}

//---------------------------------------------------------------
// Special methods apply directly to query
//---------------------------------------------------------------
func (c *chain) DB(query *gorm.DB) *chain {
	c.q = query
	return c
}

// ------------------------------------------------

func (c *chain) Query() *gorm.DB {
	return c.q
}

// return DB error and Chain error if have
func (c *chain) Preload(table string) *chain {
	c.q = c.q.Preload(table)
	return c
}

// return DB error and Chain error if have
func (c *chain) Find(output interface{}) (error, error) {
	return c.q.Find(output).Error, c.err
}

// return DB error and Chain error if have
func (c *chain) Update(m interface{}) (error, error) {
	return c.q.Model(m).Update(c.updater).Error, c.err
}

// ------------------------------------------------

func (c *chain) Order(name string) *chain {
	v := c.Get(name)
	if v != "" {
		c.q = c.q.Order(v, true)
	}
	return c
}

func (c *chain) Pagination(model interface{}) *Pagination {
	if c.err != nil {
		return nil
	}
	currentPage, perPage := PaginationQuery(c.req)

	limit := perPage
	page := currentPage

	totalRecords := 0
	c.q.Model(model).Count(&totalRecords)
	lastPage := FloorOf(totalRecords, perPage)
	if lastPage*perPage < totalRecords {
		lastPage++
	}
	lastPage = Max(1, lastPage)
	currentPage = Min(currentPage, lastPage)
	nextPage := currentPage + 1
	if nextPage > lastPage {
		nextPage = 0
	}
	// add to query
	c.q = c.q.Limit(limit).Offset((page - 1) * limit)
	// set header
	return &Pagination{
		Page:    currentPage,
		Size:    perPage,
		NextPage:nextPage,
		LastPage:lastPage,
		Total:   totalRecords,
	}
}

//---------------------------------------------------------------
// Almighty combine methods
//---------------------------------------------------------------
func (c *chain) ValidateChain() *chain {
	if c.err != nil {
		return c
	}
	return nil
}

// Example: searchUUID("product_id", []string{"product_id", "=", "?"})
func (c *chain) WhereUUID(name string, domain []string) *chain {
	if c := c.ValidateChain(); c != nil {
		return c
	}
	if c.Get(name) != "" {
		if v, err := c.Uuid(name); err != nil {
			c.err = err
		} else {
			c.q = c.q.Where(strings.Join(domain, " "), v)
		}
	}
	return c
}

func (c *chain) WhereString(name string, domain []string) *chain {
	if c := c.ValidateChain(); c != nil {
		return c
	}
	if v := c.Get(name); v != "" {
		c.q = c.q.Where(strings.Join(domain, " "), v)
	}
	return c
}

func (c *chain) WhereListString(name string, domain []string) *chain {
	if c := c.ValidateChain(); c != nil {
		return c
	}
	if v := c.Get(name); v != "" {
		tmp := strings.Split(v, ",")
		c.q = c.q.Where(strings.Join(domain, " "), tmp)
	}
	return c
}

func (c *chain) WhereStringAdv(name string, domain []string, pre string, sub string) *chain {
	if c := c.ValidateChain(); c != nil {
		return c
	}
	if v := c.Get(name); v != "" {
		c.q = c.q.Where(strings.Join(domain, " "), pre+v+sub)
	}
	return c
}

func (c *chain) WhereDate(name string, domain []string) *chain {
	if c := c.ValidateChain(); c != nil {
		return c
	}
	if c.Get(name) != "" {
		if v, err := c.Date(name); err != nil {
			c.err = err
		} else {
			c.q = c.q.Where(strings.Join(domain, " "), v)
		}
	}
	return c
}

func (c *chain) WhereTime(name string, domain []string) *chain {
	if c := c.ValidateChain(); c != nil {
		return c
	}
	if c.Get(name) != "" {
		if _, err := c.Time(name); err != nil {
			c.err = err
		} else {
			c.q = c.q.Where(strings.Join(domain, " "), c.Get(name))
		}
	}
	return c
}

func (c *chain) WhereInt(name string, domain []string) *chain {
	if c := c.ValidateChain(); c != nil {
		return c
	}
	if c.Get(name) != "" {
		if v, err := c.Int(name); err != nil {
			c.err = err
		} else {
			c.q = c.q.Where(strings.Join(domain, " "), v)
		}
	}
	return c
}

func (c *chain) WhereBool(name string, domain []string) *chain {
	if c := c.ValidateChain(); c != nil {
		return c
	}
	if c.Get(name) != "" {
		c.q = c.q.Where(strings.Join(domain, " "), c.Bool(name))
	}
	return c
}

//---------------------------------------------------------------
// Almighty combine methods
//---------------------------------------------------------------

func (c *chain) UpdateUUIDDirect(name string, v uuid.UUID) *chain {
	if c := c.ValidateChain(); c != nil {
		return c
	}
	c.updater[name] = v
	return c
}

func (c *chain) UpdateUUID(name string) *chain {
	if c := c.ValidateChain(); c != nil {
		return c
	}
	if c.Get(name) != "" {
		if v, err := c.Uuid(name); err != nil {
			c.err = err
		} else {
			c.updater[name] = v
		}
	}
	return c
}

func (c *chain) UpdateString(name string) *chain {
	if c := c.ValidateChain(); c != nil {
		return c
	}
	if c.updater == nil {
		c.updater = make(map[string]interface{})
	}
	if c.Get(name) != "" {
		if v := c.Get(name); v != "" {
			c.updater[name] = v
		}
	}
	return c
}

func (c *chain) UpdateDate(name string) *chain {
	if c := c.ValidateChain(); c != nil {
		return c
	}
	if c.updater == nil {
		c.updater = make(map[string]interface{})
	}
	if c.Get(name) != "" {
		if v, err := c.Date(name); err != nil {
			c.err = err
		} else {
			c.updater[name] = v
		}
	}
	return c
}

func (c *chain) UpdateTime(name string) *chain {
	if c := c.ValidateChain(); c != nil {
		return c
	}
	if c.updater == nil {
		c.updater = make(map[string]interface{})
	}
	if c.Get(name) != "" {
		if _, err := c.Time(name); err == nil {
			c.updater[name] = c.Get(name)
		} else {
			c.err = err
		}
	}
	return c
}

func (c *chain) UpdateInt(name string) *chain {
	if c := c.ValidateChain(); c != nil {
		return c
	}
	if c.updater == nil {
		c.updater = make(map[string]interface{})
	}
	if c.Get(name) != "" {
		if v, err := c.Int(name); err != nil {
			c.err = err
		} else {
			c.updater[name] = v
		}
	}
	return c
}

func (c *chain) UpdateBool(name string) *chain {
	if c := c.ValidateChain(); c != nil {
		return c
	}
	if c.updater == nil {
		c.updater = make(map[string]interface{})
	}
	if c.Get(name) != "" {
		c.updater[name] = c.Bool(name)
	}
	return c
}

func (c *chain) UpdateJsonB(name string) *chain {
	if c := c.ValidateChain(); c != nil {
		return c
	}
	if c.updater == nil {
		c.updater = make(map[string]interface{})
	}
	if v := c.Get(name); v != "" {
		c.updater[name] = Jsonb(v)
	}
	return c
}

type TimeFromTo struct {
	TimeFrom time.Time `json:"time_from"`
	TimeTo   time.Time `json:"time_to"`
}

func SlideTime(tFrom time.Time, tTo time.Time, period time.Duration) (res []TimeFromTo) {
	for {
		if tFrom.Add(period).Before(tTo) {
			res = append(res, TimeFromTo{
				TimeFrom: tFrom,
				TimeTo:   tFrom.Add(period),
			})
			tFrom = tFrom.Add(period)
		} else {
			res = append(res, TimeFromTo{
				TimeFrom: tFrom,
				TimeTo:   tTo,
			})
			break
		}
	}
	return
}