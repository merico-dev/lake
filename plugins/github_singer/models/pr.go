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

package models

import "time"

// GithubPullRequestState models corresponds to docs here https://github.com/singer-io/tap-github
type GithubPullRequestState struct {
	Bookmarks map[string]struct {
		PullRequests struct {
			Since time.Time `json:"since"`
		} `json:"pull_requests"`
	} `json:"bookmarks"`
}

// GithubPullRequestRecord models corresponds to docs here https://github.com/singer-io/tap-github
type GithubPullRequestRecord struct {
	Url    string `json:"url"`
	Id     string `json:"id"`
	Number int    `json:"number"`
	State  string `json:"state"`
	Title  string `json:"title"`
	User   struct {
		Login string `json:"login"`
		Id    int    `json:"id"`
	} `json:"user"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	ClosedAt  time.Time `json:"closed_at"`
	MergedAt  time.Time `json:"merged_at"`
	Labels    []struct {
		Id          int64  `json:"id"`
		NodeId      string `json:"node_id"`
		Url         string `json:"url"`
		Name        string `json:"name"`
		Color       string `json:"color"`
		Default     bool   `json:"default"`
		Description string `json:"description"`
	} `json:"labels"`
	Base struct {
		Label string `json:"label"`
		Ref   string `json:"ref"`
		Sha   string `json:"sha"`
		Repo  struct {
			Id   int    `json:"id"`
			Name string `json:"name"`
			Url  string `json:"url"`
		} `json:"repo"`
	} `json:"base"`
	SdcRepository string `json:"_sdc_repository"`
}

// GithubPullRequestSchema models corresponds to docs here https://github.com/singer-io/tap-github
type GithubPullRequestSchema struct {
	Selected             bool        `json:"selected"`
	Type                 interface{} `json:"type"`
	AdditionalProperties bool        `json:"additionalProperties"`
	Properties           interface{} `json:"properties"`
}
