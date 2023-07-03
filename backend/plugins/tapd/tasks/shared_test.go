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

package tasks

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/plugin"
	mockdal "github.com/apache/incubator-devlake/mocks/core/dal"
	mockplugin "github.com/apache/incubator-devlake/mocks/core/plugin"
	"github.com/apache/incubator-devlake/plugins/tapd/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestParseIterationChangelog tests the parseIterationChangelog function
func TestParseIterationChangelog(t *testing.T) {
	data := &TapdTaskData{Options: &TapdOptions{WorkspaceId: 1, ConnectionId: 1}}

	// Set up the required data for testing
	mockCtx := new(mockplugin.SubTaskContext)
	mockDal := new(mockdal.Dal)

	mockCtx.On("GetData").Return(data)
	mockCtx.On("GetDal").Return(mockDal)
	// Set up the required data for testing
	mockIterationFrom := &models.TapdIteration{
		ConnectionId: 1,
		Id:           1,
		Name:         "",
	}
	mockIterationTo := &models.TapdIteration{
		ConnectionId: 1,
		Id:           2,
		Name:         "",
	}
	mockDal.On("First", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		dst := args.Get(0).(*models.TapdIteration)
		*dst = *mockIterationFrom
	}).Return(nil).Once()
	mockDal.On("First", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		dst := args.Get(0).(*models.TapdIteration)
		*dst = *mockIterationTo
	}).Return(nil).Once()
	// Test case 2: success scenario
	iterationFromId, iterationToId, err := parseIterationChangelog(mockCtx, "old", "new")
	assert.Equal(t, uint64(1), iterationFromId)
	assert.Equal(t, uint64(2), iterationToId)
	assert.Nil(t, err)
}

func TestGetRawMessageDirectFromResponse(t *testing.T) {
	// Create a mock HTTP response
	body := `{"data": {"count": 10}}`
	res := &http.Response{
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(strings.NewReader(body)),
	}

	// Call the function and check the result
	rawMessages, err := GetRawMessageDirectFromResponse(res)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(rawMessages) != 1 {
		t.Errorf("Expected 1 raw message, got %d", len(rawMessages))
	}
	var page Page
	err = errors.Convert(json.Unmarshal(rawMessages[0], &page))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if page.Data.Count != 10 {
		t.Errorf("Expected count to be 10, got %d", page.Data.Count)
	}
}

func TestGetTapdTypeMappings(t *testing.T) {
	// create a mock database connection
	db := new(mockdal.Dal)
	// create some test data
	data := &TapdTaskData{
		Options: &TapdOptions{
			ConnectionId: 1,
			WorkspaceId:  2,
		},
	}

	issueTypes := make([]models.TapdWorkitemType, 0)
	issueTypes = append(issueTypes, models.TapdWorkitemType{
		ConnectionId: 1,
		WorkspaceId:  2,
		Id:           1,
		Name:         "Story",
	})
	issueTypes = append(issueTypes, models.TapdWorkitemType{
		ConnectionId: 1,
		WorkspaceId:  2,
		Id:           2,
		Name:         "Bug",
	})
	db.On("All", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		dst := args.Get(0).(*[]models.TapdWorkitemType)
		*dst = issueTypes
	}).Return(nil).Once()
	// call the function being tested
	result, err := getTapdTypeMappings(data, db, "story")

	// check if the result is correct
	if err != nil {
		t.Errorf("getTapdTypeMappings returned an error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("getTapdTypeMappings returned %d items, expected 2", len(result))
	}
	if result[1] != "Story" {
		t.Errorf("getTapdTypeMappings returned incorrect value for ID 1: %s", result[1])
	}
	if result[2] != "Bug" {
		t.Errorf("getTapdTypeMappings returned incorrect value for ID 2: %s", result[2])
	}
}

// test case for when the status list is empty
func TestGetDefaultStdStatusMappingEmptyStatusList(t *testing.T) {
	data := &TapdTaskData{
		Options: &TapdOptions{
			ConnectionId: 123,
			WorkspaceId:  456,
		},
	}
	db := new(mockdal.Dal)
	statusList := []models.TapdStoryStatus{
		{
			ConnectionId: 123,
			WorkspaceId:  456,
			EnglishName:  "Done",
			ChineseName:  "已完成",
			IsLastStep:   true,
		},
		{
			ConnectionId: 123,
			WorkspaceId:  456,
			EnglishName:  "In Progress",
			ChineseName:  "进行中",
		},
	}
	db.On("All", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		dst := args.Get(0).(*[]models.TapdStoryStatus)
		*dst = statusList
	}).Return(nil).Once()
	statusLanguageMap, getStdStatus, err := getDefaultStdStatusMapping(data, db, statusList)
	if err != nil {
		t.Errorf("getDefaultStdStatusMapping returned an error: %v", err)
	}
	expectedStatusLanguageMap := map[string]string{
		"Done":        "已完成",
		"In Progress": "进行中",
	}
	if !reflect.DeepEqual(statusLanguageMap, expectedStatusLanguageMap) {
		t.Errorf("getDefaultStdStatusMapping returned unexpected statusLanguageMap: got %v, want %v", statusLanguageMap, expectedStatusLanguageMap)
	}
	expectedGetStdStatus := map[string]string{
		"已完成": "DONE",
		"进行中": "IN_PROGRESS",
	}
	for k, v := range expectedGetStdStatus {
		if getStdStatus(k) != v {
			t.Errorf("getDefaultStdStatusMapping returned unexpected getStdStatus for %v: got %v, want %v", k, getStdStatus(k), v)
		}
	}
}

func TestUnicodeToZh(t *testing.T) {
	input := "\\u4e2d\\u6587"
	expected := "中文"
	output, err := unicodeToZh(input)
	if err != nil {
		t.Errorf("unicodeToZh(%q) returned error %v", input, err)
	}
	if output != expected {
		t.Errorf("unicodeToZh(%q) = %q, want %q", input, output, expected)
	}
}

func TestConvertUnicode(t *testing.T) {
	testStruct := struct {
		ValueBeforeParsed string
		ValueAfterParsed  string
	}{
		ValueBeforeParsed: "Hello, \\u4e16\\u754c!",
		ValueAfterParsed:  "--",
	}

	err := convertUnicode(&testStruct)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expectedBefore := "Hello, 世界!"
	if testStruct.ValueBeforeParsed != expectedBefore {
		t.Errorf("Expected ValueBeforeParsed to be %q, but got %q", expectedBefore, testStruct.ValueBeforeParsed)
	}

	expectedAfter := ""
	if testStruct.ValueAfterParsed != expectedAfter {
		t.Errorf("Expected ValueAfterParsed to be %q, but got %q", expectedAfter, testStruct.ValueAfterParsed)
	}
}

func TestGenerateDomainAccountIdForUsers(t *testing.T) {
	connectionId := uint64(123)
	testCases := []struct {
		param    string
		expected string
	}{
		{"user1,user2,user3", "tapd:TapdAccount:123:user1,tapd:TapdAccount:123:user2,tapd:TapdAccount:123:user3"},
		{"user4;user5;user6", "tapd:TapdAccount:123:user4,tapd:TapdAccount:123:user5,tapd:TapdAccount:123:user6"},
		{"", ""},
	}
	mockMeta := mockplugin.NewPluginMeta(t)
	mockMeta.On("RootPkgPath").Return("github.com/apache/incubator-devlake/plugins/tapd")
	mockMeta.On("Name").Return("tapd").Maybe()
	err := plugin.RegisterPlugin("tapd", mockMeta)
	assert.Nil(t, err)
	for _, testCase := range testCases {
		result := generateDomainAccountIdForUsers(testCase.param, connectionId)
		if result != testCase.expected {
			t.Errorf("generateDomainAccountIdForUsers(%s, %d) = %s; expected %s", testCase.param, connectionId, result, testCase.expected)
		}
	}
}

func Test_extractStatus(t *testing.T) {
	type args struct {
		blob []byte
	}
	tests := []struct {
		name  string
		args  args
		want  map[string]string
		want1 errors.Error
	}{
		{
			"non-empty array",
			args{
				blob: []byte(`{"data":[{"id":1,"name":"已完成","is_last_step":true},{"id":2,"name":"进行中"}],"status":1,"message":"success"}`),
			},
			nil,
			nil,
		},
		{
			"empty array",
			args{
				blob: []byte(`{"status":1,"data":[],"info":"success"}`),
			},
			nil,
			nil,
		},
		{
			"object",
			args{
				blob: []byte(`{"status":1,"data":{"new":"新建","in_progress":"开发处理"},"info":"success"}`),
			},
			map[string]string{"new": "新建", "in_progress": "开发处理"},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := extractStatus(tt.args.blob)
			assert.Equalf(t, tt.want, got, "extractStatus(%v)", tt.args.blob)
			assert.Equalf(t, tt.want1, got1, "extractStatus(%v)", tt.args.blob)
		})
	}
}

func TestUnicodeToZh2(t *testing.T) {
	input := "app\\\\user\"\\u4e2d\\u6587"
	expected := "app\\user\"中文"
	output, err := unicodeToZh(input)
	if err != nil {
		t.Errorf("unicodeToZh(%q) returned error %v", input, err)
	}
	if output != expected {
		t.Errorf("unicodeToZh(%q) = %q, want %q", input, output, expected)
	}
}
