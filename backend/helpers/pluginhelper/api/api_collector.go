/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"text/template"
	"time"

	"github.com/apache/incubator-devlake/core/dal"
	"github.com/apache/incubator-devlake/core/errors"
	plugin "github.com/apache/incubator-devlake/core/plugin"
	"github.com/apache/incubator-devlake/helpers/pluginhelper/common"
)

// Pager contains pagination information for a api request
type Pager struct {
	Page int
	Skip int
	Size int
}

// RequestData is the input of `UrlTemplate` `Query` and `Header`, so we can generate them dynamically
type RequestData struct {
	Pager     *Pager
	Params    interface{}
	Input     interface{}
	InputJSON []byte
	// equal to the return value from GetNextPageCustomData when PageSize>0 and not the first request
	CustomData interface{}
}

// AsyncResponseHandler FIXME ...
type AsyncResponseHandler func(res *http.Response) error

// ApiCollectorArgs FIXME ...
type ApiCollectorArgs struct {
	RawDataSubTaskArgs
	// UrlTemplate is used to generate the final URL for Api Collector to request
	// i.e. `api/3/issue/{{ .Input.IssueId }}/changelog`
	// For detail of what variables can be used, please check `RequestData`
	UrlTemplate string `comment:"GoTemplate for API url"`
	// Query would be sent out as part of the request URL
	Query func(reqData *RequestData) (url.Values, errors.Error)
	// Header would be sent out along with request
	Header func(reqData *RequestData) (http.Header, errors.Error)
	// GetTotalPages is to tell `ApiCollector` total number of pages based on response of the first page.
	// so `ApiCollector` could collect those pages in parallel for us
	GetTotalPages func(res *http.Response, args *ApiCollectorArgs) (int, errors.Error)
	// PageSize tells ApiCollector the page size
	PageSize int
	// GetNextPageCustomData indicate if this collection request each page in order and build query by the prev request
	GetNextPageCustomData func(prevReqData *RequestData, prevPageResponse *http.Response) (interface{}, errors.Error)
	// Incremental indicate if this is an incremental collection, the existing data won't get deleted if it was true
	Incremental bool `comment:"indicate if this collection is incremental update"`
	// ApiClient is a asynchronize api request client with qps
	ApiClient       RateLimitedApiClient
	MinTickInterval *time.Duration
	// Input helps us collect data based on previous collected data, like collecting changelogs based on jira
	// issue ids
	Input Iterator
	// Concurrency specify qps for api that doesn't return total number of pages/records
	// NORMALLY, DO NOT SPECIFY THIS PARAMETER, unless you know what it means
	Concurrency    int
	ResponseParser func(res *http.Response) ([]json.RawMessage, errors.Error)
	AfterResponse  common.ApiClientAfterResponse
	RequestBody    func(reqData *RequestData) map[string]interface{}
	Method         string
}

// ApiCollector FIXME ...
type ApiCollector struct {
	*RawDataSubTask
	args        *ApiCollectorArgs
	urlTemplate *template.Template
}

// NewApiCollector allocates a new ApiCollector with the given args.
// ApiCollector can help us collecting data from api with ease, pass in a AsyncApiClient and tell it which part
// of response we want to save, ApiCollector will collect them from remote server and store them into database.
func NewApiCollector(args ApiCollectorArgs) (*ApiCollector, errors.Error) {
	// process args
	rawDataSubTask, err := NewRawDataSubTask(args.RawDataSubTaskArgs)
	if err != nil {
		return nil, errors.Default.Wrap(err, "Couldn't resolve raw subtask args")
	}
	// TODO: check if args.Table is valid when this is a http GET request
	if args.UrlTemplate == "" && args.Method == "" {
		return nil, errors.Default.New("UrlTemplate is required")
	}
	tpl, err := errors.Convert01(template.New(args.Table).Parse(args.UrlTemplate))
	if err != nil {
		return nil, errors.Default.Wrap(err, "Failed to compile UrlTemplate")
	}
	if args.ApiClient == nil {
		return nil, errors.Default.New("ApiClient is required")
	}
	if args.ResponseParser == nil {
		return nil, errors.Default.New("ResponseParser is required")
	}
	apiCollector := &ApiCollector{
		RawDataSubTask: rawDataSubTask,
		args:           &args,
		urlTemplate:    tpl,
	}
	if args.AfterResponse != nil {
		apiCollector.SetAfterResponse(args.AfterResponse)
	} else {
		apiCollector.SetAfterResponse(func(res *http.Response) errors.Error {
			if res.StatusCode == http.StatusUnauthorized || res.StatusCode == http.StatusUnprocessableEntity {
				return errors.Unauthorized.New("authentication failed, please check your AccessToken")
			}
			return nil
		})
	}
	return apiCollector, nil
}

// Execute will start collection
func (collector *ApiCollector) Execute() errors.Error {
	logger := collector.args.Ctx.GetLogger()
	logger.Info("start api collection")

	// make sure table is created
	db := collector.args.Ctx.GetDal()
	err := db.AutoMigrate(&RawData{}, dal.From(collector.table))
	if err != nil {
		return errors.Default.Wrap(err, "error auto-migrating collector")
	}

	// flush data if not incremental collection
	if !collector.args.Incremental {
		err = db.Delete(&RawData{}, dal.From(collector.table), dal.Where("params = ?", collector.params))
		if err != nil {
			return errors.Default.Wrap(err, "error deleting data from collector")
		}
	}

	// if MinTickInterval was specified
	if collector.args.MinTickInterval != nil {
		minTickInterval := *collector.args.MinTickInterval
		if minTickInterval <= time.Duration(0) {
			return errors.Default.Wrap(err, "MinTickInterval must be greater than 0")
		}
		oldTickInterval := collector.args.ApiClient.GetTickInterval()
		if oldTickInterval < minTickInterval {
			// reset the tick interval only if it exceeded the specified limit
			logger.Info("set tick interval to %v", minTickInterval.String())
			collector.args.ApiClient.Reset(minTickInterval)
			defer func() {
				logger.Info("restore tick interval to %v", oldTickInterval.String())
				collector.args.ApiClient.Reset(oldTickInterval)
			}()
		}
	}

	collector.args.Ctx.SetProgress(0, -1)
	if collector.args.Input != nil {
		iterator := collector.args.Input
		defer iterator.Close()
		apiClient := collector.args.ApiClient
		if apiClient == nil {
			return errors.Default.New("api_collector can not Execute with nil apiClient")
		}
		for {
			if !iterator.HasNext() || apiClient.HasError() {
				err = collector.args.ApiClient.WaitAsync()
				if err != nil {
					return err
				}
				if !iterator.HasNext() || apiClient.HasError() {
					break
				}
			}
			var input interface{}
			input, err = iterator.Fetch()
			if err != nil {
				break
			}
			collector.exec(input)
		}
	} else {
		// or we just did it once
		collector.exec(nil)
	}

	if err != nil {
		return errors.Default.Wrap(err, "error executing collector")
	}
	logger.Debug("wait for all async api to be finished")
	err = collector.args.ApiClient.WaitAsync()
	if err != nil {
		logger.Error(err, "end api collection error")
		err = errors.Default.Wrap(err, "Error waiting for async Collector execution")
	} else {
		logger.Info("end api collection without error")
	}

	return err
}

func (collector *ApiCollector) exec(input interface{}) {
	inputJson, err := json.Marshal(input)
	if err != nil {
		panic(err)
	}
	reqData := new(RequestData)
	reqData.Input = input
	reqData.InputJSON = inputJson
	reqData.Pager = &Pager{
		Page: 1,
		Size: collector.args.PageSize,
	}
	// featch the detail
	if collector.args.PageSize <= 0 {
		collector.fetchAsync(reqData, nil)
		// fetch pages sequentially
	} else if collector.args.GetNextPageCustomData != nil {
		collector.fetchPagesSequentially(reqData)
		// fetch pages in parallel with number of total pages can be determined from the first page
	} else if collector.args.GetTotalPages != nil {
		collector.fetchPagesDetermined(reqData)
		// fetch pages in parallel without number of total pages
	} else {
		collector.fetchPagesUndetermined(reqData)
	}
}

// fetchPagesSequentially fetches data of all pages in order to build RequestData by prev response
func (collector *ApiCollector) fetchPagesSequentially(reqData *RequestData) {
	var collect func() errors.Error
	collect = func() errors.Error {
		collector.fetchAsync(reqData, func(count int, body []byte, res *http.Response) errors.Error {
			if count < collector.args.PageSize {
				return nil
			}
			customData, err := collector.args.GetNextPageCustomData(reqData, res)
			if err != nil {
				if errors.Is(err, ErrFinishCollect) {
					return nil
				} else {
					panic(err)
				}
			}
			reqData.CustomData = customData
			reqData.Pager.Skip += collector.args.PageSize
			reqData.Pager.Page += 1
			return collect()
		})
		return nil
	}
	collector.args.ApiClient.NextTick(collect)
}

// fetchPagesDetermined fetches data of all pages for APIs that return paging information
func (collector *ApiCollector) fetchPagesDetermined(reqData *RequestData) {
	// fetch first page
	collector.fetchAsync(reqData, func(count int, body []byte, res *http.Response) errors.Error {
		totalPages, err := collector.args.GetTotalPages(res, collector.args)
		if err != nil {
			return errors.Default.Wrap(err, "fetchPagesDetermined get totalPages failed")
		}
		// spawn a none blocking go routine to fetch other pages
		collector.args.ApiClient.NextTick(func() errors.Error {
			for page := 2; page <= totalPages; page++ {
				reqDataTemp := &RequestData{
					Pager: &Pager{
						Page: page,
						Skip: collector.args.PageSize * (page - 1),
						Size: collector.args.PageSize,
					},
					Input:     reqData.Input,
					InputJSON: reqData.InputJSON,
				}
				collector.fetchAsync(reqDataTemp, nil)
			}
			return nil
		})
		return nil
	})
}

// fetchPagesUndetermined fetches data of all pages for APIs that do NOT return paging information
func (collector *ApiCollector) fetchPagesUndetermined(reqData *RequestData) {
	//logger := collector.args.Ctx.GetLogger()
	//logger.Debug("fetch all pages in parallel with specified concurrency: %d", collector.args.Concurrency)
	// if api doesn't return total number of pages, employ a step concurrent technique
	// when `Concurrency` was set to 3:
	// goroutine #1 fetches pages 1/4/7..
	// goroutine #2 fetches pages 2/5/8...
	// goroutine #3 fetches pages 3/6/9...
	apiClient := collector.args.ApiClient
	concurrency := collector.args.Concurrency
	if concurrency == 0 {
		// normally when a multi-pages api depends on a another resource, like jira changelogs depend on issue ids
		// it tend to have less page, like 1 or 2 pages in total
		if collector.args.Input != nil {
			concurrency = 2
		} else {
			concurrency = apiClient.GetNumOfWorkers() / 10
			if concurrency < 10 {
				concurrency = 10
			}
		}
	}
	for i := 0; i < concurrency; i++ {
		reqDataCopy := RequestData{
			Pager: &Pager{
				Page: i + 1,
				Size: collector.args.PageSize,
				Skip: collector.args.PageSize * (i),
			},
			Input:     reqData.Input,
			InputJSON: reqData.InputJSON,
		}
		var collect func() errors.Error
		collect = func() errors.Error {
			collector.fetchAsync(&reqDataCopy, func(count int, body []byte, res *http.Response) errors.Error {
				if count < collector.args.PageSize {
					return nil
				}
				apiClient.NextTick(func() errors.Error {
					reqDataCopy.Pager.Skip += collector.args.PageSize * concurrency
					reqDataCopy.Pager.Page += concurrency
					return collect()
				})
				return nil
			})
			return nil
		}
		apiClient.NextTick(collect)
	}
}

func (collector *ApiCollector) generateUrl(pager *Pager, input interface{}) (string, errors.Error) {
	params := collector.args.Params
	if collector.args.Options != nil {
		params = collector.args.Options.GetParams()
	}
	var buf bytes.Buffer
	err := collector.urlTemplate.Execute(&buf, &RequestData{
		Pager:  pager,
		Params: params,
		Input:  input,
	})
	if err != nil {
		return "", errors.Convert(err)
	}
	return buf.String(), nil
}

// GetAfterResponse return apiClient's afterResponseFunction
func (collector *ApiCollector) GetAfterResponse() common.ApiClientAfterResponse {
	return collector.args.ApiClient.GetAfterFunction()
}

// SetAfterResponse set apiClient's afterResponseFunction
func (collector *ApiCollector) SetAfterResponse(f common.ApiClientAfterResponse) {
	collector.args.ApiClient.SetAfterFunction(f)
}

func (collector *ApiCollector) fetchAsync(reqData *RequestData, handler func(int, []byte, *http.Response) errors.Error) {
	if reqData.Pager == nil {
		reqData.Pager = &Pager{
			Page: 1,
			Size: 100,
			Skip: 0,
		}
	}
	apiUrl, err := collector.generateUrl(reqData.Pager, reqData.Input)
	if err != nil {
		panic(err)
	}
	var apiQuery url.Values
	if collector.args.Query != nil {
		apiQuery, err = collector.args.Query(reqData)
		if err != nil {
			panic(err)
		}
	}
	var reqBody interface{}
	if collector.args.RequestBody != nil {
		reqBody = collector.args.RequestBody(reqData)
		if err != nil {
			panic(err)
		}
	}

	apiHeader := (http.Header)(nil)
	if collector.args.Header != nil {
		apiHeader, err = collector.args.Header(reqData)
		if err != nil {
			panic(err)
		}
	}
	logger := collector.args.Ctx.GetLogger()
	logger.Debug("fetchAsync <<< enqueueing for %s %v", apiUrl, apiQuery)
	responseHandler := func(res *http.Response) errors.Error {
		defer logger.Debug("fetchAsync >>> done for %s %v %v", apiUrl, apiQuery, collector.args.RequestBody)
		logger := collector.args.Ctx.GetLogger()
		// read body to buffer
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return errors.Default.Wrap(err, fmt.Sprintf("error reading response from %s", apiUrl))
		}
		res.Body.Close()
		res.Body = io.NopCloser(bytes.NewBuffer(body))
		// convert body to array of RawJSON
		items, err := collector.args.ResponseParser(res)
		if err != nil {
			if errors.Is(err, ErrFinishCollect) {
				logger.Info("a fetch stop by parser, reqInput: #%s", reqData.Params)
				handler = nil
			} else {
				return errors.Default.Wrap(err, fmt.Sprintf("error parsing response from %s", apiUrl))
			}
		}
		// save to db
		count := len(items)
		if count == 0 {
			collector.args.Ctx.IncProgress(1)
			return nil
		}
		db := collector.args.Ctx.GetDal()
		urlString := res.Request.URL.String()
		rows := make([]*RawData, count)
		for i, msg := range items {
			rows[i] = &RawData{
				Params: collector.params,
				Data:   msg,
				Url:    urlString,
				Input:  reqData.InputJSON,
			}
		}
		err = db.Create(rows, dal.From(collector.table))
		if err != nil {
			return errors.Default.Wrap(err, fmt.Sprintf("error inserting raw rows into %s", collector.table))
		}
		logger.Debug("fetchAsync === total %d rows were saved into database", count)
		// increase progress only when it was not nested
		collector.args.Ctx.IncProgress(1)
		if handler != nil {
			// trigger next fetch, but return if ErrFinishCollect got from ResponseParser
			res.Body = io.NopCloser(bytes.NewBuffer(body))
			return handler(count, body, res)
		}
		return nil
	}
	if collector.args.Method == http.MethodPost {
		collector.args.ApiClient.DoPostAsync(apiUrl, apiQuery, reqBody, apiHeader, responseHandler)
	} else {
		collector.args.ApiClient.DoGetAsync(apiUrl, apiQuery, apiHeader, responseHandler)
	}
	logger.Debug("fetchAsync === enqueued for %s %v", apiUrl, apiQuery)
}

var _ plugin.SubTask = (*ApiCollector)(nil)
