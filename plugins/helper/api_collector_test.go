package helper

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/merico-dev/lake/models/common"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// go test -gcflags=all=-l

type TestTable struct {
	Email string `gorm:"primaryKey;type:varchar(255)"`
	Name  string `gorm:"type:varchar(255)"`
	common.NoPKModel
}

var TestTableData *TestTable = &TestTable{
	Email: "test@test.com",
	Name:  "test",
}

type TestParam struct {
	Test string
}

func (TestTable) TableName() string {
	return "_tool_test"
}

var TestError error = fmt.Errorf("Error For Test")

var gt *gomonkey.Patches
var gc *gomonkey.Patches
var gd *gomonkey.Patches
var ga *gomonkey.Patches
var gs *gomonkey.Patches
var god *gomonkey.Patches
var gw *gomonkey.Patches
var gr *gomonkey.Patches

var TestUrlBefor string = "test1"
var TestUrlParam string = "test2"
var TestUrlAfter string = "test3"
var TestUrl string = "https://" + TestUrlBefor + TestUrlParam + TestUrlAfter

var TestRawMessage string = "{\"message\":\"TestRawMessage\"}"
var TestUrlValueKey string = "TestKey"
var TestUrlValueValue string = "TestValue"
var TestNoRunHere string = "should not run to this line of code"

var TestDataCount int = 100
var TestTotalPage int = 100
var TestDataCountNotFull int = 50
var TestPage int = 110
var TestSkip int = 100100
var TestSize int = 116102

var Cancel context.CancelFunc

var TestHttpResponse_Suc http.Response = http.Response{
	Status:     "200 OK",
	StatusCode: http.StatusOK,
	Proto:      "HTTP/1.0",
	ProtoMajor: 1,
	ProtoMinor: 0,
}

var TestHttpResponse_404 http.Response = http.Response{
	Status:     "404 Not Found",
	StatusCode: http.StatusNotFound,
	Proto:      "HTTP/1.0",
	ProtoMajor: 1,
	ProtoMinor: 0,
}

// Assert http.Response base test data
func AssertBaseResponse(t *testing.T, A *http.Response, B *http.Response) {
	assert.Equal(t, A.Status, B.Status)
	assert.Equal(t, A.StatusCode, B.StatusCode)
	assert.Equal(t, A.Proto, B.Proto)
	assert.Equal(t, A.ProtoMajor, B.ProtoMajor)
	assert.Equal(t, A.ProtoMinor, B.ProtoMinor)
}

func AddBodyData(res *http.Response, count int) {
	data := "["
	for i := 0; i < count; i++ {
		data += TestRawMessage
		if i != count-1 {
			data += ","
		}
	}
	data += "]"
	res.Body = ioutil.NopCloser(bytes.NewReader([]byte(data)))
}

func SetUrl(res *http.Response, rawURL string) {
	u, _ := url.Parse(rawURL)
	res.Request = &http.Request{
		URL: u,
	}
}

// Mock the DB api
// Need be released by UnMockDB
func MockDB(t *testing.T) {
	gt = gomonkey.ApplyMethod(reflect.TypeOf(&gorm.DB{}), "Table", func(db *gorm.DB, name string, args ...interface{}) *gorm.DB {
		assert.Equal(t, name, TestTableData.TableName())
		return db
	},
	)

	gc = gomonkey.ApplyMethod(reflect.TypeOf(&gorm.DB{}), "Create", func(db *gorm.DB, value interface{}) *gorm.DB {
		assert.Equal(t, TestTableData, value.(*TestTable))
		return db
	},
	)

	gd = gomonkey.ApplyMethod(reflect.TypeOf(&gorm.DB{}), "Delete", func(db *gorm.DB, value interface{}, conds ...interface{}) (tx *gorm.DB) {
		return db
	},
	)

	ga = gomonkey.ApplyMethod(reflect.TypeOf(&gorm.DB{}), "AutoMigrate", func(db *gorm.DB, dst ...interface{}) error {
		return nil
	},
	)

	god = gomonkey.ApplyMethod(reflect.TypeOf(&gorm.DB{}), "Order", func(db *gorm.DB, value interface{}) (tx *gorm.DB) {
		return db
	},
	)

	gw = gomonkey.ApplyMethod(reflect.TypeOf(&gorm.DB{}), "Where", func(db *gorm.DB, query interface{}, args ...interface{}) (tx *gorm.DB) {
		return db
	},
	)

	gr = gomonkey.ApplyMethod(reflect.TypeOf(&gorm.DB{}), "Rows", func(db *gorm.DB) (*sql.Rows, error) {
		return &sql.Rows{}, nil
	},
	)

	gs = gomonkey.ApplyMethod(reflect.TypeOf(&gorm.DB{}), "ScanRows", func(db *gorm.DB, rows *sql.Rows, dest interface{}) error {
		dest = TestRawMessage
		return nil
	},
	)

}

// released MockDB
func UnMockDB() {
	gt.Reset()
	gc.Reset()
	gd.Reset()
	ga.Reset()
	god.Reset()
	gw.Reset()
	gr.Reset()
	gs.Reset()
}

type TestIterator struct {
	data         TestTable
	count        int
	hasNextTimes int
	fetchTimes   int
	closeTimes   int
	unlimit      bool
}

func (it *TestIterator) HasNext() bool {
	it.hasNextTimes++
	return it.count > 0
}

func (it *TestIterator) Fetch() (interface{}, error) {
	it.fetchTimes++
	if it.count > 0 {
		if it.unlimit == false {
			it.count--
		}
		ret := it.data
		return &ret, nil
	}
	return nil, TestError
}

func (it *TestIterator) Close() error {
	it.closeTimes++
	return nil
}

func CreateTestApiCollector() (*ApiCollector, error) {
	db := &gorm.DB{}
	var ctx context.Context
	ctx, Cancel = context.WithCancel(context.Background())
	return NewApiCollector(ApiCollectorArgs{
		RawDataSubTaskArgs: RawDataSubTaskArgs{
			Ctx: &DefaultSubTaskContext{
				defaultExecContext: newDefaultExecContext(GetConfigForTest("../../"), NewDefaultLogger(logrus.New(), "Test", make(map[string]*logrus.Logger)), db, ctx, "Test", nil, nil),
			},
			Table: TestTable{}.TableName(),
			Params: &TestParam{
				Test: TestUrlParam,
			},
		},
		ApiClient:   &ApiAsyncClient{qps: 10},
		PageSize:    100,
		Incremental: false,
		UrlTemplate: TestUrlBefor + "{{ .Params.Test }}" + TestUrlAfter,
		Query: func(reqData *RequestData) (url.Values, error) {
			u := url.Values{}
			json, err := json.Marshal(reqData.Input)
			u.Add("Vjson", string(json))
			if err != nil {
				u.Add("Verr", err.Error())
			} else {
				u.Add("Verr", "")
			}
			return u, nil
		},
		Header: func(reqData *RequestData) (http.Header, error) {
			h := http.Header{}
			json, err := json.Marshal(reqData.Input)
			h.Add("Hjson", string(json))
			if err != nil {
				h.Add("Herr", err.Error())
			} else {
				h.Add("Herr", "")
			}
			return h, nil
		},
		GetTotalPages:  func(res *http.Response, args *ApiCollectorArgs) (int, error) { return TestTotalPage, nil },
		ResponseParser: GetRawMessageArrayFromResponse,
	})
}

func TestGormDB(t *testing.T) {
	ts := &TestTable{
		Email: "test@test.com",
		Name:  "test",
	}

	db := &gorm.DB{}
	MockDB(t)
	defer UnMockDB()

	db.Table(ts.TableName()).Order(ts).Where(ts).Create(ts).Delete(ts).AutoMigrate()
	db.Table(ts.TableName()).Order(ts).Where(ts).Create(ts).Delete(ts).Rows()
	db.Table(ts.TableName()).Order(ts).Where(ts).Create(ts).Delete(ts).ScanRows(nil, nil)
}

func TestSaveRawData(t *testing.T) {
	gt = gomonkey.ApplyMethod(reflect.TypeOf(&gorm.DB{}), "Table", func(db *gorm.DB, name string, args ...interface{}) *gorm.DB {
		assert.Equal(t, name, "_raw_"+TestTableData.TableName())
		return db
	},
	)
	defer gt.Reset()

	gc = gomonkey.ApplyMethod(reflect.TypeOf(&gorm.DB{}), "Create", func(db *gorm.DB, value interface{}) *gorm.DB {
		rd := value.([]*RawData)
		params, _ := json.Marshal(&TestParam{
			Test: TestUrlParam,
		})
		input, _ := json.Marshal(TestTableData)
		for _, v := range rd {
			// check data and url
			assert.Equal(t, v.Params, string(params))
			assert.Equal(t, string(v.Data), TestRawMessage)
			assert.Equal(t, v.Url, TestUrl)
			assert.Equal(t, v.Input.String(), string(input))
		}
		return db
	},
	)
	defer gc.Reset()

	apiCollector, _ := CreateTestApiCollector()

	resBase := TestHttpResponse_Suc
	res := &resBase

	// build data and url
	AddBodyData(res, TestDataCount)
	SetUrl(res, TestUrl)

	i, err := apiCollector.saveRawData(res, TestTableData)
	assert.Equal(t, i, TestDataCount)
	assert.Equal(t, err, nil)
}

func TestSaveRawData_Fail(t *testing.T) {
	gt = gomonkey.ApplyMethod(reflect.TypeOf(&gorm.DB{}), "Table", func(db *gorm.DB, name string, args ...interface{}) *gorm.DB {
		assert.Empty(t, TestNoRunHere)
		return db
	},
	)
	defer gt.Reset()

	gc = gomonkey.ApplyMethod(reflect.TypeOf(&gorm.DB{}), "Create", func(db *gorm.DB, value interface{}) *gorm.DB {
		assert.Empty(t, TestNoRunHere)
		return db
	},
	)
	defer gc.Reset()

	apiCollector, _ := CreateTestApiCollector()

	// build data and url
	resBase := TestHttpResponse_404
	res := &resBase

	AddBodyData(res, 0)
	SetUrl(res, TestUrl)

	//run testing
	i, err := apiCollector.saveRawData(res, TestTableData)
	assert.Equal(t, i, 0)
	assert.Equal(t, err, nil)
}

func TestHandleResponse(t *testing.T) {
	apiCollector, _ := CreateTestApiCollector()

	gs := gomonkey.ApplyPrivateMethod(reflect.TypeOf(apiCollector), "saveRawData", func(collector *ApiCollector, res *http.Response, input interface{}) (int, error) {
		items, err := collector.args.ResponseParser(res)
		assert.Equal(t, err, nil)
		// check items data
		for _, v := range items {
			jsondata, err := json.Marshal(v)
			assert.Equal(t, err, nil)
			assert.Equal(t, string(jsondata), TestRawMessage)
		}
		assert.Equal(t, input, TestTableData)
		AssertBaseResponse(t, res, &TestHttpResponse_Suc)
		return len(items), nil
	})
	defer gs.Reset()

	// build requeset input
	reqData := new(RequestData)
	reqData.Input = TestTableData
	handle := apiCollector.handleResponse(reqData)

	resBase := TestHttpResponse_Suc
	res := &resBase

	// build data and url
	AddBodyData(res, TestDataCount)
	SetUrl(res, TestUrl)

	// run testing
	err := handle(res, nil)
	assert.Equal(t, err, nil)
}

func TestHandleResponse_Fail(t *testing.T) {
	apiCollector, _ := CreateTestApiCollector()

	gs := gomonkey.ApplyPrivateMethod(reflect.TypeOf(apiCollector), "saveRawData", func(collector *ApiCollector, res *http.Response, input interface{}) (int, error) {
		items, err := collector.args.ResponseParser(res)
		assert.Equal(t, err, nil)
		for _, v := range items {
			jsondata, err := json.Marshal(v)
			assert.Equal(t, err, nil)
			assert.Equal(t, string(jsondata), TestRawMessage)
		}
		assert.Equal(t, input, TestTableData)
		AssertBaseResponse(t, res, &TestHttpResponse_404)
		return len(items), TestError
	})
	defer gs.Reset()

	// build requeset input
	reqData := new(RequestData)
	reqData.Input = TestTableData
	handle := apiCollector.handleResponse(reqData)

	// build data and url
	resBase := TestHttpResponse_404
	res := &resBase

	AddBodyData(res, 0)
	SetUrl(res, TestUrl)

	//  run testing
	err := handle(res, nil)
	assert.Equal(t, err, TestError)
}

func TestFetchAsync(t *testing.T) {
	apiCollector, _ := CreateTestApiCollector()

	gg := gomonkey.ApplyMethod(reflect.TypeOf(&ApiAsyncClient{}), "GetAsync", func(apiAsyncClient *ApiAsyncClient, path string, query url.Values, header http.Header, handler ApiAsyncCallback) error {
		assert.Equal(t, path, TestUrlBefor+TestUrlParam+TestUrlAfter)

		json, err := json.Marshal(TestTableData)
		assert.Equal(t, query.Get("Vjson"), string(json))
		assert.Equal(t, header.Get("Hjson"), string(json))
		if err != nil {
			assert.Equal(t, query.Get("Verr"), err.Error())
			assert.Equal(t, header.Get("Herr"), err.Error())
		} else {
			assert.Equal(t, query.Get("Verr"), "")
			assert.Equal(t, header.Get("Herr"), "")
		}

		res := TestHttpResponse_Suc
		handler(&res, TestError)
		return nil
	})
	defer gg.Reset()

	// build request Input
	reqData := new(RequestData)
	reqData.Input = TestTableData

	// run testing
	err := apiCollector.fetchAsync(reqData, func(r *http.Response, err error) error {
		AssertBaseResponse(t, r, &TestHttpResponse_Suc)
		assert.Equal(t, err, TestError)
		return err
	})

	assert.Equal(t, err, nil)
}

func TestFetchAsync_Fail(t *testing.T) {
	apiCollector, _ := CreateTestApiCollector()

	gg := gomonkey.ApplyMethod(reflect.TypeOf(&ApiAsyncClient{}), "GetAsync", func(apiAsyncClient *ApiAsyncClient, path string, query url.Values, header http.Header, handler ApiAsyncCallback) error {
		assert.Equal(t, path, TestUrlBefor+TestUrlParam+TestUrlAfter)

		json, err := json.Marshal(TestTableData)
		assert.Equal(t, query.Get("Vjson"), string(json))
		assert.Equal(t, header.Get("Hjson"), string(json))
		if err != nil {
			assert.Equal(t, query.Get("Verr"), err.Error())
			assert.Equal(t, header.Get("Herr"), err.Error())
		} else {
			assert.Equal(t, query.Get("Verr"), "")
			assert.Equal(t, header.Get("Herr"), "")
		}

		res := TestHttpResponse_404
		handler(&res, TestError)
		return TestError
	})
	defer gg.Reset()

	// build request Input
	reqData := new(RequestData)
	reqData.Input = TestTableData

	// run testing
	err := apiCollector.fetchAsync(reqData, func(r *http.Response, err error) error {
		AssertBaseResponse(t, r, &TestHttpResponse_404)
		assert.Equal(t, err, TestError)
		return err
	})

	assert.Equal(t, err, TestError)
}

func TestHandleResponseWithPages(t *testing.T) {
	apiCollector, _ := CreateTestApiCollector()
	pages := make([]bool, TestTotalPage+1)
	for i := 1; i <= TestTotalPage; i++ {
		pages[i] = false
	}

	gf := gomonkey.ApplyPrivateMethod(reflect.TypeOf(apiCollector), "fetchAsync", func(collector *ApiCollector, reqData *RequestData, handler ApiAsyncCallback) error {
		page := reqData.Pager.Page
		pages[page] = true
		return nil
	})
	defer gf.Reset()

	gs := gomonkey.ApplyPrivateMethod(reflect.TypeOf(apiCollector), "saveRawData", func(collector *ApiCollector, res *http.Response, input interface{}) (int, error) {
		items, err := collector.args.ResponseParser(res)
		assert.Equal(t, err, nil)
		for _, v := range items {
			jsondata, err := json.Marshal(v)
			assert.Equal(t, err, nil)
			assert.Equal(t, string(jsondata), TestRawMessage)
		}
		assert.Equal(t, input, TestTableData)
		AssertBaseResponse(t, res, &TestHttpResponse_Suc)
		return len(items), nil
	})
	defer gs.Reset()

	// build request Input
	reqData := new(RequestData)
	reqData.Input = TestTableData
	handle := apiCollector.handleResponseWithPages(reqData)

	// build data and url
	resBase := TestHttpResponse_Suc
	res := &resBase

	AddBodyData(res, TestDataCount)
	SetUrl(res, TestUrl)

	// run testing
	err := handle(res, nil)
	assert.Equal(t, err, nil)
	for i := 2; i <= TestTotalPage; i++ {
		assert.True(t, pages[i], i)
	}
}

func TestHandleResponseWithPages_Fail(t *testing.T) {
	apiCollector, _ := CreateTestApiCollector()

	gf := gomonkey.ApplyPrivateMethod(reflect.TypeOf(apiCollector), "fetchAsync", func(collector *ApiCollector, reqData *RequestData, handler ApiAsyncCallback) error {
		return TestError
	})
	defer gf.Reset()

	gs := gomonkey.ApplyPrivateMethod(reflect.TypeOf(apiCollector), "saveRawData", func(collector *ApiCollector, res *http.Response, input interface{}) (int, error) {
		items, err := collector.args.ResponseParser(res)
		assert.Equal(t, err, nil)
		for _, v := range items {
			jsondata, err := json.Marshal(v)
			assert.Equal(t, err, nil)
			assert.Equal(t, string(jsondata), TestRawMessage)
		}
		assert.Equal(t, input, TestTableData)
		AssertBaseResponse(t, res, &TestHttpResponse_404)
		return len(items), TestError
	})
	defer gs.Reset()

	// build request Input
	reqData := new(RequestData)
	reqData.Input = TestTableData
	handle := apiCollector.handleResponseWithPages(reqData)

	// build data and url
	resBase := TestHttpResponse_404
	res := &resBase

	AddBodyData(res, 0)
	SetUrl(res, TestUrl)

	// run testing
	err := handle(res, nil)
	assert.Equal(t, err, TestError)
}

func TestStepFetch(t *testing.T) {
	apiCollector, _ := CreateTestApiCollector()

	init := false
	noFullTimes := 0

	gs := gomonkey.ApplyPrivateMethod(reflect.TypeOf(apiCollector), "saveRawData", func(collector *ApiCollector, res *http.Response, input interface{}) (int, error) {
		items, err := collector.args.ResponseParser(res)
		assert.Equal(t, err, nil)
		for _, v := range items {
			jsondata, err := json.Marshal(v)
			assert.Equal(t, err, nil)
			assert.Equal(t, string(jsondata), TestRawMessage)
		}
		// full page
		assert.Equal(t, input, TestTableData)
		AssertBaseResponse(t, res, &TestHttpResponse_Suc)
		return len(items), nil
	})
	defer gs.Reset()

	gf := gomonkey.ApplyPrivateMethod(reflect.TypeOf(apiCollector), "fetchAsync", func(collector *ApiCollector, reqData *RequestData, handler ApiAsyncCallback) error {
		resBase := TestHttpResponse_Suc
		res := &resBase
		SetUrl(res, TestUrl)

		// full page for continue
		if reqData.Pager.Page == TestPage {
			init = true
			assert.Equal(t, reqData.Pager.Skip, TestSkip)
			assert.Equal(t, reqData.Pager.Size, TestSize)
			AddBodyData(res, TestDataCount)
		} else {
			// not full page for stop
			AddBodyData(res, TestDataCountNotFull)
			noFullTimes++
		}

		go handler(res, nil)
		return nil
	})
	defer gf.Reset()

	// build request Input
	reqData := new(RequestData)
	reqData.Input = TestTableData
	reqData.Pager = &Pager{
		Page: TestPage,
		Skip: TestSkip,
		Size: TestSize,
	}

	// cancel can only be called when error occurs, because we are doomed anyway.
	ctx, cancel := context.WithCancel(apiCollector.args.Ctx.GetContext())

	// run testing
	err := apiCollector.stepFetch(ctx, cancel, *reqData)

	assert.Equal(t, noFullTimes, 1)
	assert.Equal(t, init, true)
	assert.Equal(t, err, nil)
}

func TestStepFetch_Fail(t *testing.T) {
	apiCollector, _ := CreateTestApiCollector()

	init := false
	noFullTimes := 0

	gs := gomonkey.ApplyPrivateMethod(reflect.TypeOf(apiCollector), "saveRawData", func(collector *ApiCollector, res *http.Response, input interface{}) (int, error) {
		items, err := collector.args.ResponseParser(res)
		assert.Equal(t, err, nil)
		for _, v := range items {
			jsondata, err := json.Marshal(v)
			assert.Equal(t, err, nil)
			assert.Equal(t, string(jsondata), TestRawMessage)
		}
		// full page
		assert.Equal(t, input, TestTableData)
		AssertBaseResponse(t, res, &TestHttpResponse_404)
		return len(items), nil
	})
	defer gs.Reset()

	gf := gomonkey.ApplyPrivateMethod(reflect.TypeOf(apiCollector), "fetchAsync", func(collector *ApiCollector, reqData *RequestData, handler ApiAsyncCallback) error {
		resBase := TestHttpResponse_404
		res := &resBase
		SetUrl(res, TestUrl)
		// full page for continue
		if reqData.Pager.Page == TestPage {
			init = true
			assert.Equal(t, reqData.Pager.Skip, TestSkip)
			assert.Equal(t, reqData.Pager.Size, TestSize)
			AddBodyData(res, TestDataCount)
			go handler(res, nil)
		} else {
			// not full page for stop
			AddBodyData(res, TestDataCountNotFull)
			noFullTimes++
			return TestError
		}
		return nil
	})
	defer gf.Reset()

	// build request Input
	reqData := new(RequestData)
	reqData.Input = TestTableData
	reqData.Pager = &Pager{
		Page: TestPage,
		Skip: TestSkip,
		Size: TestSize,
	}

	// cancel can only be called when error occurs, because we are doomed anyway.
	ctx, cancel := context.WithCancel(apiCollector.args.Ctx.GetContext())

	err := apiCollector.stepFetch(ctx, cancel, *reqData)

	assert.Equal(t, noFullTimes, 1)
	assert.Equal(t, init, true)
	assert.Equal(t, err, TestError)
}

func TestStepFetch_Cancel(t *testing.T) {
	apiCollector, _ := CreateTestApiCollector()

	gs := gomonkey.ApplyPrivateMethod(reflect.TypeOf(apiCollector), "saveRawData", func(collector *ApiCollector, res *http.Response, input interface{}) (int, error) {
		items, err := collector.args.ResponseParser(res)
		assert.Equal(t, err, nil)
		for _, v := range items {
			jsondata, err := json.Marshal(v)
			assert.Equal(t, err, nil)
			assert.Equal(t, string(jsondata), TestRawMessage)
		}
		// full page
		assert.Equal(t, input, TestTableData)
		AssertBaseResponse(t, res, &TestHttpResponse_Suc)
		return len(items), nil
	})
	defer gs.Reset()

	gf := gomonkey.ApplyPrivateMethod(reflect.TypeOf(apiCollector), "fetchAsync", func(collector *ApiCollector, reqData *RequestData, handler ApiAsyncCallback) error {
		resBase := TestHttpResponse_Suc
		res := &resBase
		SetUrl(res, TestUrl)
		// always to continue
		assert.Equal(t, reqData.Pager.Size, TestSize)
		AddBodyData(res, TestDataCount)
		go handler(res, nil)

		return nil
	})
	defer gf.Reset()

	// build request Input
	reqData := new(RequestData)
	reqData.Input = TestTableData
	reqData.Pager = &Pager{
		Page: TestPage,
		Skip: TestSkip,
		Size: TestSize,
	}

	// cancel can only be called when error occurs, because we are doomed anyway.

	ctx, cancel := context.WithCancel(apiCollector.args.Ctx.GetContext())

	go func() {
		time.Sleep(time.Duration(500) * time.Microsecond)
		Cancel()
	}()

	err := apiCollector.stepFetch(ctx, cancel, *reqData)

	assert.Equal(t, err, fmt.Errorf("context canceled"))
}

func TestExecWithOutPageSize(t *testing.T) {
	apiCollector, _ := CreateTestApiCollector()
	apiCollector.args.PageSize = 0

	gf := gomonkey.ApplyPrivateMethod(reflect.TypeOf(apiCollector), "fetchAsync", func(collector *ApiCollector, reqData *RequestData, handler ApiAsyncCallback) error {
		assert.Equal(t, reqData.Input, TestTableData)
		return nil
	})
	defer gf.Reset()

	gs := gomonkey.ApplyPrivateMethod(reflect.TypeOf(apiCollector), "stepFetch", func(collector *ApiCollector, ctx context.Context, cancel func(), reqData RequestData) error {
		assert.Empty(t, TestNoRunHere)
		return TestError
	})
	defer gs.Reset()

	// run testing
	err := apiCollector.exec(TestTableData)
	assert.Equal(t, err, nil)
}

func TestExecWithGetTotalPages(t *testing.T) {
	apiCollector, _ := CreateTestApiCollector()

	gf := gomonkey.ApplyPrivateMethod(reflect.TypeOf(apiCollector), "fetchAsync", func(collector *ApiCollector, reqData *RequestData, handler ApiAsyncCallback) error {
		assert.Equal(t, reqData.Input, TestTableData)
		return nil
	})
	defer gf.Reset()

	gs := gomonkey.ApplyPrivateMethod(reflect.TypeOf(apiCollector), "stepFetch", func(collector *ApiCollector, ctx context.Context, cancel func(), reqData RequestData) error {
		assert.Empty(t, TestNoRunHere)
		return TestError
	})
	defer gs.Reset()

	// run testing
	err := apiCollector.exec(TestTableData)
	assert.Equal(t, err, nil)
}

func TestExecWithOutGetTotalPages(t *testing.T) {
	apiCollector, _ := CreateTestApiCollector()
	apiCollector.args.GetTotalPages = nil
	apiCollector.args.Concurrency = TestTotalPage

	pages := make([]bool, TestTotalPage+1)
	for i := 1; i <= TestTotalPage; i++ {
		pages[i] = false
	}

	gf := gomonkey.ApplyPrivateMethod(reflect.TypeOf(apiCollector), "fetchAsync", func(collector *ApiCollector, reqData *RequestData, handler ApiAsyncCallback) error {
		assert.Equal(t, reqData.Input, TestTableData)
		return nil
	})
	defer gf.Reset()

	gs := gomonkey.ApplyPrivateMethod(reflect.TypeOf(apiCollector), "stepFetch", func(collector *ApiCollector, ctx context.Context, cancel func(), reqData RequestData) error {
		assert.Equal(t, reqData.Input, TestTableData)
		page := reqData.Pager.Page
		pages[page] = true
		assert.Equal(t, reqData.Pager.Size, apiCollector.args.PageSize)
		assert.Equal(t, reqData.Pager.Skip, apiCollector.args.PageSize*(page-1))
		return nil
	})
	defer gs.Reset()

	// run testing
	err := apiCollector.exec(TestTableData)
	assert.Equal(t, err, nil)

	for i := 2; i <= TestTotalPage; i++ {
		assert.True(t, pages[i], i)
	}
}

func TestExec_Cancel(t *testing.T) {
	apiCollector, _ := CreateTestApiCollector()
	apiCollector.args.GetTotalPages = nil
	apiCollector.args.Concurrency = TestTotalPage

	pages := make([]bool, TestTotalPage+1)
	for i := 1; i <= TestTotalPage; i++ {
		pages[i] = false
	}

	gf := gomonkey.ApplyPrivateMethod(reflect.TypeOf(apiCollector), "fetchAsync", func(collector *ApiCollector, reqData *RequestData, handler ApiAsyncCallback) error {
		assert.Equal(t, reqData.Input, TestTableData)
		return nil
	})
	defer gf.Reset()

	gs := gomonkey.ApplyPrivateMethod(reflect.TypeOf(apiCollector), "stepFetch", func(collector *ApiCollector, ctx context.Context, cancel func(), reqData RequestData) error {
		assert.Equal(t, reqData.Input, TestTableData)
		page := reqData.Pager.Page
		pages[page] = true
		assert.Equal(t, reqData.Pager.Size, apiCollector.args.PageSize)
		assert.Equal(t, reqData.Pager.Skip, apiCollector.args.PageSize*(page-1))

		// check if it can get cancel command
		for range ctx.Done() {
		}

		return nil
	})
	defer gs.Reset()

	go func() {
		time.Sleep(time.Duration(500) * time.Microsecond)
		Cancel()
	}()

	// run testing
	err := apiCollector.exec(TestTableData)
	assert.Equal(t, err, nil)

	for i := 2; i <= TestTotalPage; i++ {
		assert.True(t, pages[i], i)
	}
}

func TestExecute(t *testing.T) {
	MockDB(t)
	defer UnMockDB()
	apiCollector, _ := CreateTestApiCollector()

	apiCollector.args.Input = &TestIterator{
		data:         *TestTableData,
		count:        TestDataCount,
		hasNextTimes: 0,
		fetchTimes:   0,
		closeTimes:   0,
		unlimit:      false,
	}

	gt.Reset()
	gt = gomonkey.ApplyMethod(reflect.TypeOf(&gorm.DB{}), "Table", func(db *gorm.DB, name string, args ...interface{}) *gorm.DB {
		assert.Equal(t, name, "_raw_"+TestTableData.TableName())
		return db
	},
	)

	NeedWait := int64(0)
	execTimes := 0

	ge := gomonkey.ApplyPrivateMethod(reflect.TypeOf(apiCollector), "exec", func(collector *ApiCollector, input interface{}) error {
		atomic.AddInt64(&NeedWait, 1)
		execTimes++
		assert.Equal(t, input.(*TestTable).Email, TestTableData.Email)
		assert.Equal(t, input.(*TestTable).Name, TestTableData.Name)
		atomic.AddInt64(&NeedWait, -1)
		return nil
	})
	defer ge.Reset()

	gw := gomonkey.ApplyMethod(reflect.TypeOf(&ApiAsyncClient{}), "WaitAsync", func(apiClient *ApiAsyncClient) error {
		for atomic.LoadInt64(&NeedWait) > 0 {
			time.Sleep(time.Millisecond)
		}
		return nil
	})
	defer gw.Reset()

	// run testing
	err := apiCollector.Execute()
	assert.Equal(t, err, nil)
	assert.Equal(t, execTimes, TestDataCount)

	input := apiCollector.args.Input.(*TestIterator)
	assert.Equal(t, input.fetchTimes, TestDataCount)
	assert.Equal(t, input.hasNextTimes >= input.fetchTimes, true)
	assert.Equal(t, input.closeTimes > 0, true)
}

func TestExecute_Cancel(t *testing.T) {
	MockDB(t)
	defer UnMockDB()
	apiCollector, _ := CreateTestApiCollector()

	apiCollector.args.Input = &TestIterator{
		data:         *TestTableData,
		count:        TestDataCount,
		hasNextTimes: 0,
		fetchTimes:   0,
		closeTimes:   0,
		unlimit:      true,
	}

	apiCollector.args.Input.HasNext()

	gt.Reset()
	gt = gomonkey.ApplyMethod(reflect.TypeOf(&gorm.DB{}), "Table", func(db *gorm.DB, name string, args ...interface{}) *gorm.DB {
		assert.Equal(t, name, "_raw_"+TestTableData.TableName())
		return db
	},
	)

	NeedWait := int64(0)
	execTimes := 0

	ge := gomonkey.ApplyPrivateMethod(reflect.TypeOf(apiCollector), "exec", func(collector *ApiCollector, input interface{}) error {
		atomic.AddInt64(&NeedWait, 1)
		execTimes++
		assert.Equal(t, input.(*TestTable).Email, TestTableData.Email)
		assert.Equal(t, input.(*TestTable).Name, TestTableData.Name)
		atomic.AddInt64(&NeedWait, -1)
		return nil
	})
	defer ge.Reset()

	gw := gomonkey.ApplyMethod(reflect.TypeOf(&ApiAsyncClient{}), "WaitAsync", func(apiClient *ApiAsyncClient) error {
		for atomic.LoadInt64(&NeedWait) > 0 {
			time.Sleep(time.Millisecond)
		}
		return nil
	})
	defer gw.Reset()

	go func() {
		time.Sleep(time.Duration(500) * time.Microsecond)
		Cancel()
	}()

	// run testing
	err := apiCollector.Execute()
	assert.Equal(t, err, fmt.Errorf("context canceled"))

	input := apiCollector.args.Input.(*TestIterator)
	assert.Equal(t, input.hasNextTimes >= input.fetchTimes, true)
	assert.Equal(t, input.closeTimes > 0, true)
}
