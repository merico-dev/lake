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

package singer

import (
	"github.com/apache/incubator-devlake/config"
	"github.com/apache/incubator-devlake/plugins/core/singer"
	"github.com/apache/incubator-devlake/plugins/github/models"
)

type GithubPullRequestTap struct {
	tap *singer.Tap[models.GithubPullRequestSchema]
}

func NewGithubPullRequestTap(cfg *models.GithubConfig) *GithubPullRequestTap {
	env := config.GetConfig()
	return &GithubPullRequestTap{
		tap: singer.NewTap[models.GithubPullRequestSchema](&singer.Config{
			Mappings: cfg,
			Cmd:      env.GetString("TAP_GITHUB"),
			TapType:  "github_pull_request",
		}),
	}
}

func (t *GithubPullRequestTap) Setup() {
	t.tap.WriteConfig()
	t.tap.DiscoverProperties()
	t.tap.SetProperties(func(stream *singer.Stream[models.GithubPullRequestSchema]) bool {
		if stream.Stream == "pull_requests" {
			stream.Schema.Selected = true
		}
		return stream.Schema.Selected
	})
}

func (t *GithubPullRequestTap) Run(initialState *models.GithubPullRequestState) ([]*singer.TapRecord[models.GithubPullRequestRecord], *singer.TapState[models.GithubPullRequestState]) {
	if initialState != nil {
		t.tap.WriteState(initialState)
	}
	var records []*singer.TapRecord[models.GithubPullRequestRecord]
	var state = &singer.TapState[models.GithubPullRequestState]{}
	stream := t.tap.Run()
	for d := range stream {
		if d.Err != nil {
			panic(d.Err)
		}
		if record, ok := singer.AsTapRecord[models.GithubPullRequestRecord](d.Data); ok {
			records = append(records, record)
			continue
		} else if state, ok = singer.AsTapState[models.GithubPullRequestState](d.Data); ok {
			continue
		}
	}
	return records, state
}
