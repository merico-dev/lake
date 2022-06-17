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

package helper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"text/template"

	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/core/dal"
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
}

type AsyncResponseHandler func(res *http.Response) error

type ApiCollectorArgs struct {
	RawDataSubTaskArgs
	// UrlTemplate is used to generate the final URL for Api Collector to request
	// i.e. `api/3/issue/{{ .Input.IssueId }}/changelog`
	// For detail of what variables can be used, please check `RequestData`
	UrlTemplate string `comment:"GoTemplate for API url"`
	// Query would be sent out as part of the request URL
	Query func(reqData *RequestData) (url.Values, error) ``
	// Header would be sent out along with request
	Header func(reqData *RequestData) (http.Header, error)
	// PageSize tells ApiCollector the page size
	PageSize int
	// Incremental indicate if this is a incremental collection, the existing data won't get deleted if it was true
	Incremental bool `comment:""`
	// ApiClient is a asynchronize api request client with qps
	ApiClient RateLimitedApiClient
	// Input helps us collect data based on previous collected data, like collecting changelogs based on jira
	// issue ids
	Input Iterator
	// GetTotalPages is to tell `ApiCollector` total number of pages based on response of the first page.
	// so `ApiCollector` could collect those pages in parallel for us
	GetTotalPages func(res *http.Response, args *ApiCollectorArgs) (int, error)
	// Concurrency specify qps for api that doesn't return total number of pages/records
	// NORMALLY, DO NOT SPECIFY THIS PARAMETER, unless you know what it means
	Concurrency    int
	ResponseParser func(res *http.Response) ([]json.RawMessage, error)
}

type ApiCollector struct {
	*RawDataSubTask
	args        *ApiCollectorArgs
	urlTemplate *template.Template
}

// NewApiCollector allocates a new ApiCollector with the given args.
// ApiCollector can help us collecting data from api with ease, pass in a AsyncApiClient and tell it which part
// of response we want to save, ApiCollector will collect them from remote server and store them into database.
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
	db := collector.args.Ctx.GetDal()
	err := db.AutoMigrate(&RawData{}, dal.From(collector.table))
	if err != nil {
		return err
	}

	// flush data if not incremental collection
	if !collector.args.Incremental {
		err = db.Delete(&RawData{}, dal.From(collector.table), dal.Where("params = ?", collector.params))
		if err != nil {
			return err
		}
	}

	collector.args.Ctx.SetProgress(0, -1)
	if collector.args.Input != nil {
		iterator := collector.args.Input
		apiClient := collector.args.ApiClient
		defer iterator.Close()
		for iterator.HasNext() && !apiClient.HasError() {
			input, err := iterator.Fetch()
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
		return err
	}
	logger.Debug("wait for all async api to finished")
	err = collector.args.ApiClient.WaitAsync()
	logger.Info("end api collection error: %w", err)
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
	if collector.args.PageSize <= 0 {
		collector.fetchAsync(reqData, nil)
	} else if collector.args.GetTotalPages != nil {
		collector.fetchPagesDetermined(reqData)
	} else {
		collector.fetchPagesUndetermined(reqData)
	}
}

// fetchPagesDetermined fetches data of all pages for APIs that return paging information
func (collector *ApiCollector) fetchPagesDetermined(reqData *RequestData) {
	// fetch first page
	collector.fetchAsync(reqData, func(count int, body []byte, res *http.Response) error {
		totalPages, err := collector.args.GetTotalPages(res, collector.args)
		if err != nil {
			return fmt.Errorf("fetchPagesDetermined get totalPages faileds: %s", err.Error())
		}
		// spawn a none blocking go routine to fetch other pages
		collector.args.ApiClient.NextTick(func() error {
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
			cap := apiClient.GetNumOfWorkers() / 10
			if cap < 10 {
				concurrency = 10
			} else {
				concurrency = cap
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
		var collect func() error
		collect = func() error {
			collector.fetchAsync(&reqDataCopy, func(count int, body []byte, res *http.Response) error {
				if count < collector.args.PageSize {
					return nil
				}
				apiClient.NextTick(func() error {
					reqDataCopy.Pager.Skip += collector.args.PageSize
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

func (collector *ApiCollector) fetchAsync(reqData *RequestData, handler func(int, []byte, *http.Response) error) {
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

	apiHeader := (http.Header)(nil)
	if collector.args.Header != nil {
		apiHeader, err = collector.args.Header(reqData)
		if err != nil {
			panic(err)
		}
	}
	logger := collector.args.Ctx.GetLogger()
	logger.Debug("fetchAsync <<< enqueueing for %s %v", apiUrl, apiQuery)
	collector.args.ApiClient.GetAsync(apiUrl, apiQuery, apiHeader, func(res *http.Response) error {
		defer logger.Debug("fetchAsync >>> done for %s %v", apiUrl, apiQuery)
		logger := collector.args.Ctx.GetLogger()
		// read body to buffer
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		res.Body.Close()
		res.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		// convert body to array of RawJSON
		items, err := collector.args.ResponseParser(res)
		if err != nil {
			return err
		}
		// save to db
		count := len(items)
		if count == 0 {
			return nil
		}
		db := collector.args.Ctx.GetDal()
		url := res.Request.URL.String()
		rows := make([]*RawData, count)
		for i, msg := range items {
			rows[i] = &RawData{
				Params: collector.params,
				Data:   msg,
				Url:    url,
				Input:  reqData.InputJSON,
			}
		}
		err = db.Create(rows, dal.From(collector.table))
		if err != nil {
			return err
		}
		logger.Debug("fetchAsync === total %d rows were saved into database", count)
		// increase progress only when it was not nested
		collector.args.Ctx.IncProgress(1)
		if handler != nil {
			res.Body = ioutil.NopCloser(bytes.NewBuffer(body))
			return handler(count, body, res)
		}
		return nil
	})
	logger.Debug("fetchAsync === enqueued for %s %v", apiUrl, apiQuery)
}

var _ core.SubTask = (*ApiCollector)(nil)
