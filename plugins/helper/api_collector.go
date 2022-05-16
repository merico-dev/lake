package helper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"sync"
	"text/template"

	"github.com/merico-dev/lake/plugins/core"
)

type Pager struct {
	Page int
	Skip int
	Size int
}

type RequestData struct {
	Pager  *Pager
	Params interface{}
	Input  interface{}
}

type AsyncResponseHandler func(res *http.Response) error

type ApiCollectorArgs struct {
	RawDataSubTaskArgs
	/*
		url may use arbitrary variables from different source in any order, we need GoTemplate to allow more
		flexible for all kinds of possibility.
		Pager contains information for a particular page, calculated by ApiCollector, and will be passed into
		GoTemplate to generate a url for that page.
		We want to do page-fetching in ApiCollector, because the logic are highly similar, by doing so, we can
		avoid duplicate logic for every tasks, and when we have a better idea like improving performance, we can
		do it in one place
	*/
	UrlTemplate string `comment:"GoTemplate for API url"`
	// (Optional) Return query string for request, or you can plug them into UrlTemplate directly
	Query func(reqData *RequestData) (url.Values, error) `comment:"Extra query string when requesting API, like 'Since' option for jira issues collection"`
	// Some api might do pagination by http headers
	Header      func(reqData *RequestData) (http.Header, error)
	PageSize    int
	Incremental bool `comment:"Indicate this is a incremental collection, so the existing data won't get flushed"`
	ApiClient   RateLimitedApiClient
	/*
		Sometimes, we need to collect data based on previous collected data, like jira changelog, it requires
		issue_id as part of the url.
		We can mimic `stdin` design, to accept a `Input` function which produces a `Iterator`, collector
		should iterate all records, and do data-fetching for each on, either in parallel or sequential order
		UrlTemplate: "api/3/issue/{{ Input.ID }}/changelog"
	*/
	Input          Iterator
	InputRateLimit int
	/*
		For api endpoint that returns number of total pages, ApiCollector can collect pages in parallel with ease,
		or other techniques are required if this information was missing.
	*/
	GetTotalPages  func(res *http.Response, args *ApiCollectorArgs) (int, error)
	Concurrency    int
	ResponseParser func(res *http.Response) ([]json.RawMessage, error)
}

type ApiCollector struct {
	*RawDataSubTask
	args        *ApiCollectorArgs
	urlTemplate *template.Template
}

// NewApiCollector allocates a new ApiCollector  with the given args.
// ApiCollector can help you collecting data from some api with ease, pass in a AsyncApiClient and tell it which part
// of response you want to save, ApiCollector will collect them from remote server and store them into database.
func NewApiCollector(args ApiCollectorArgs) (*ApiCollector, error) {
	// process args
	rawDataSubTask, err := newRawDataSubTask(args.RawDataSubTaskArgs)
	if err != nil {
		return nil, err
	}
	// TODO: check if args.Table is valid
	if args.UrlTemplate == "" {
		return nil, fmt.Errorf("UrlTemplate is required")
	}
	tpl, err := template.New(args.Table).Parse(args.UrlTemplate)
	if err != nil {
		return nil, fmt.Errorf("Failed to compile UrlTemplate: %w", err)
	}
	if args.ApiClient == nil {
		return nil, fmt.Errorf("ApiClient is required")
	}
	if args.ResponseParser == nil {
		return nil, fmt.Errorf("ResponseParser is required")
	}
	if args.InputRateLimit == 0 {
		args.InputRateLimit = 50
	}
	if args.Concurrency < 1 {
		args.Concurrency = 1
	}
	return &ApiCollector{
		RawDataSubTask: rawDataSubTask,
		args:           &args,
		urlTemplate:    tpl,
	}, nil
}

// Start collection
func (collector *ApiCollector) Execute() error {
	logger := collector.args.Ctx.GetLogger()
	logger.Info("start api collection")

	// make sure table is created
	db := collector.args.Ctx.GetDb()
	err := db.Table(collector.table).AutoMigrate(&RawData{})
	if err != nil {
		return err
	}

	// flush data if not incremental collection
	if !collector.args.Incremental {
		err = db.Table(collector.table).Delete(&RawData{}, "params = ?", collector.params).Error
		if err != nil {
			return err
		}
	}

	if collector.args.Input != nil {
		collector.args.Ctx.SetProgress(0, -1)
		// load all rows from iterator, and do multiple `exec` accordingly
		// TODO: this loads all records into memory, we need lazy-load
		iterator := collector.args.Input
		defer iterator.Close()
		// throttle input process speed so it can be canceled, create a channel to represent available slots
		slots := int(math.Ceil(collector.args.ApiClient.GetQps())) * 2
		if slots <= 0 {
			return fmt.Errorf("RateLimit can't use the 0 Qps")
		}
		slotsChan := make(chan bool, slots)
		defer close(slotsChan)
		for i := 0; i < slots; i++ {
			slotsChan <- true
		}

		errors := make(chan error, slots)
		defer close(errors)

		var wg sync.WaitGroup
		ctx := collector.args.Ctx.GetContext()

	out:
		for iterator.HasNext() {
			select {
			// canceled by user, stop
			case <-ctx.Done():
				err = ctx.Err()
				break out
			// obtain a slot
			case <-slotsChan:
				input, err := iterator.Fetch()
				if err != nil {
					break out
				}
				wg.Add(1)
				go func() {
					defer func() {
						wg.Done()
						recover()
					}()
					e := collector.exec(input)
					// propagate error
					if e != nil {
						errors <- e
					} else {
						// release 1 slot
						slotsChan <- true
					}
				}()
			case err = <-errors:
				break out
			}
		}
		if err == nil {
			wg.Wait()
		}
	} else {
		// or we just did it once
		err = collector.exec(nil)
	}

	if err != nil {
		return err
	}
	logger.Debug("wait for all async api to finished")
	err = collector.args.ApiClient.WaitAsync()
	logger.Info("end api collection")
	return err
}

func (collector *ApiCollector) exec(input interface{}) error {
	reqData := new(RequestData)
	reqData.Input = input
	if collector.args.PageSize <= 0 {
		// collect detail of a record
		return collector.fetchAsync(reqData, collector.handleResponse(reqData))
	}
	// collect multiple pages
	var err error
	if collector.args.GetTotalPages != nil {
		/* when total pages is available from api*/
		// fetch the very first page
		err = collector.fetchAsync(reqData, collector.handleResponseWithPages(reqData))
	} else {
		// if api doesn't return total number of pages, employ a step concurrent technique
		// when `Concurrency` was set to 3:
		// goroutine #1 fetches pages 1/4/7..
		// goroutine #2 fetches pages 2/5/8...
		// goroutine #3 fetches pages 3/6/9...
		errs := make(chan error, collector.args.Concurrency)
		var errCount int
		// cancel can only be called when error occurs, because we are doomed anyway.
		ctx, cancel := context.WithCancel(collector.args.Ctx.GetContext())
		defer cancel()
		for i := 0; i < collector.args.Concurrency; i++ {
			reqDataTemp := RequestData{
				Pager: &Pager{
					Page: i + 1,
					Size: collector.args.PageSize,
					Skip: collector.args.PageSize * (i),
				},
				Input: reqData.Input,
			}
			go func() {
				errs <- collector.stepFetch(ctx, cancel, reqDataTemp)
			}()
		}
		for e := range errs {
			errCount++
			if err != nil || errCount == collector.args.Concurrency {
				err = e
				break
			}
		}
	}
	if err != nil {
		return err
	}
	if collector.args.Input != nil {
		collector.args.Ctx.IncProgress(1)
	}
	return nil
}

func (collector *ApiCollector) generateUrl(pager *Pager, input interface{}) (string, error) {
	var buf bytes.Buffer
	err := collector.urlTemplate.Execute(&buf, &RequestData{
		Pager:  pager,
		Params: collector.args.Params,
		Input:  input,
	})
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// stepFetch collect pages synchronously. In practice, several stepFetch running concurrently, we could stop all of them by calling `cancel`.
func (collector *ApiCollector) stepFetch(ctx context.Context, cancel func(), reqData RequestData) error {
	// channel `c` is used to make sure fetchAsync is called serially
	c := make(chan struct{})
	var err1 error
	handler := func(res *http.Response, err error) error {
		select {
		case <-ctx.Done():
			err = ctx.Err()
		default:

		}
		if err != nil {
			err1 = err
			close(c)
			return err
		}
		count, err := collector.saveRawData(res, reqData.Input)
		if err != nil {
			err1 = err
			close(c)
			cancel()
			return err
		}
		if count < collector.args.PageSize {
			close(c)
			return nil
		}
		reqData.Pager.Skip += collector.args.PageSize
		reqData.Pager.Page += collector.args.Concurrency
		c <- struct{}{}
		return nil
	}
	// kick off
	go func() { c <- struct{}{} }()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case _, ok := <-c:
			if !ok || err1 != nil {
				return err1
			} else {
				err := collector.fetchAsync(&reqData, handler)
				if err != nil {
					close(c)
					cancel()
					return err
				}
			}
		}
	}
	return err1
}

func (collector *ApiCollector) fetchAsync(reqData *RequestData, handler ApiAsyncCallback) error {
	if reqData.Pager == nil {
		reqData.Pager = &Pager{
			Page: 1,
			Size: 100,
			Skip: 0,
		}
	}
	ctx := collector.args.Ctx.GetContext()
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:

	}
	apiUrl, err := collector.generateUrl(reqData.Pager, reqData.Input)
	if err != nil {
		return err
	}
	var apiQuery url.Values
	if collector.args.Query != nil {
		apiQuery, err = collector.args.Query(reqData)
		if err != nil {
			return err
		}
	}

	apiHeader := (http.Header)(nil)
	if collector.args.Header != nil {
		apiHeader, err = collector.args.Header(reqData)
		if err != nil {
			return err
		}
	}
	return collector.args.ApiClient.GetAsync(apiUrl, apiQuery, apiHeader, handler)
}

func (collector *ApiCollector) handleResponse(reqData *RequestData) ApiAsyncCallback {
	return func(res *http.Response, err error) error {
		if err != nil {
			return err
		}
		_, err = collector.saveRawData(res, reqData.Input)
		collector.args.Ctx.IncProgress(1)
		return err
	}
}

func (collector *ApiCollector) handleResponseWithPages(reqData *RequestData) ApiAsyncCallback {
	return func(res *http.Response, e error) error {
		if e != nil {
			return e
		}
		// gather total pages
		body, e := ioutil.ReadAll(res.Body)
		if e != nil {
			return e
		}
		res.Body.Close()
		res.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		totalPages, e := collector.args.GetTotalPages(res, collector.args)
		if e != nil {
			return e
		}
		// save response body of first page
		res.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		_, e = collector.saveRawData(res, reqData.Input)
		if e != nil {
			return e
		}
		if collector.args.Input == nil {
			collector.args.Ctx.SetProgress(1, totalPages)
		}
		// fetch other pages in parallel
		for page := 2; page <= totalPages; page++ {
			reqDataTemp := &RequestData{
				Pager: &Pager{
					Page: page,
					Size: collector.args.PageSize,
					Skip: collector.args.PageSize * (page - 1),
				},
				Input: reqData.Input,
			}
			e = collector.fetchAsync(reqDataTemp, collector.handleResponse(reqDataTemp))
			if e != nil {
				return e
			}
		}
		return nil
	}
}

func (collector *ApiCollector) saveRawData(res *http.Response, input interface{}) (int, error) {
	items, err := collector.args.ResponseParser(res)
	logger := collector.args.Ctx.GetLogger()
	if err != nil {
		return 0, err
	}
	res.Body.Close()

	inputJson, _ := json.Marshal(input)

	if len(items) == 0 {
		return 0, nil
	}
	db := collector.args.Ctx.GetDb()
	u := res.Request.URL.String()
	dd := make([]*RawData, len(items))
	for i, msg := range items {
		dd[i] = &RawData{
			Params: collector.params,
			Data:   msg,
			Url:    u,
			Input:  inputJson,
		}
	}
	err = db.Table(collector.table).Create(dd).Error
	if err != nil {
		logger.Error("failed to save raw data: %s", err)
	}
	return len(dd), err
}

func GetRawMessageDirectFromResponse(res *http.Response) ([]json.RawMessage, error) {
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, err
	}
	return []json.RawMessage{body}, nil
}

func GetRawMessageArrayFromResponse(res *http.Response) ([]json.RawMessage, error) {
	rawMessages := []json.RawMessage{}

	if res == nil {
		return nil, fmt.Errorf("res is nil")
	}
	defer res.Body.Close()
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %s", err, res.Request.URL.String())
	}

	err = json.Unmarshal(resBody, &rawMessages)
	if err != nil {
		return nil, fmt.Errorf("%w %s %s", err, res.Request.URL.String(), string(resBody))
	}

	return rawMessages, nil
}

var _ core.SubTask = (*ApiCollector)(nil)
