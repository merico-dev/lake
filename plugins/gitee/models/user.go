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

import "github.com/apache/incubator-devlake/models/common"

type GiteeUser struct {
	Id                int    `json:"id" gorm:"primaryKey"`
	Login             string `json:"login" gorm:"type:varchar(255)"`
	Name              string `json:"name" gorm:"type:varchar(255)"`
	AvatarUrl         string `json:"avatar_url" gorm:"type:varchar(255)"`
	EventsUrl         string `json:"events_url" gorm:"type:varchar(255)"`
	FollowersUrl      string `json:"followers_url" gorm:"type:varchar(255)"`
	FollowingUrl      string `json:"following_url" gorm:"type:varchar(255)"`
	GistsUrl          string `json:"gists_url" gorm:"type:varchar(255)"`
	HtmlUrl           string `json:"html_url" gorm:"type:varchar(255)"`
	OrganizationsUrl  string `json:"organizations_url" gorm:"type:varchar(255)"`
	ReceivedEventsUrl string `json:"received_events_url" gorm:"type:varchar(255)"`
	Remark            string `json:"remark" gorm:"type:varchar(255)"`
	ReposUrl          string `json:"repos_url" gorm:"type:varchar(255)"`
	StarredUrl        string `json:"starred_url" gorm:"type:varchar(255)"`
	SubscriptionsUrl  string `json:"subscriptions_url" gorm:"type:varchar(255)"`
	Url               string `json:"url" gorm:"type:varchar(255)"`
	Type              string `json:"type" gorm:"type:varchar(255)"`
	common.NoPKModel
}

func (GiteeUser) TableName() string {
	return "_tool_gitee_users"
}
