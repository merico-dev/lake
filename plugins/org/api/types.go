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
	"strings"
	"time"

	"github.com/apache/incubator-devlake/models/domainlayer"
	"github.com/apache/incubator-devlake/models/domainlayer/crossdomain"
)

const TimeFormat = "2006-01-02"

var fakeUsers = []user{{
	Id:      "1",
	Name:    "Tyrone K. Cummings",
	Email:   "TyroneKCummings@teleworm.us",
	TeamIds: "1,2",
}, {
	Id:      "2",
	Name:    "Dorothy R. Updegraff",
	Email:   "DorothyRUpdegraff@dayrep.com",
	TeamIds: "3",
}}

var fakeTeams = []team{{
	Id:           "1",
	Name:         "Maple Leafs",
	Alias:        "ML",
	ParentId:     "2",
	SortingIndex: 0,
}, {
	Id:           "2",
	Name:         "Friendly Confines",
	Alias:        "FC",
	ParentId:     "",
	SortingIndex: 1,
}, {
	Id:           "3",
	Name:         "Blue Jays",
	Alias:        "BJ",
	ParentId:     "",
	SortingIndex: 2,
}}

type user struct {
	Id      string
	Name    string
	Email   string
	TeamIds string
}

func (*user) fromDomainLayer(users []crossdomain.User, teamUsers []crossdomain.TeamUser) []user {
	var result []user
	teamUserMap := make(map[string][]string)
	for _, tu := range teamUsers {
		teamUserMap[tu.UserId] = append(teamUserMap[tu.UserId], tu.TeamId)
	}
	for _, u := range users {
		result = append(result, user{
			Id:      u.Id,
			Name:    u.Name,
			Email:   u.Email,
			TeamIds: strings.Join(teamUserMap[u.Id], ","),
		})
	}
	return result
}

func (*user) toDomainLayer(uu []user) (users []*crossdomain.User, teamUsers []*crossdomain.TeamUser) {
	for _, u := range uu {
		users = append(users, &crossdomain.User{
			DomainEntity: domainlayer.DomainEntity{Id: u.Id},
			Email:        u.Email,
			Name:         u.Name,
		})
		for _, teamId := range strings.Split(u.TeamIds, ",") {
			if u.Id == "" || teamId == "" {
				continue
			}
			teamUsers = append(teamUsers, &crossdomain.TeamUser{
				TeamId: teamId,
				UserId: u.Id,
			})
		}
	}
	return
}

func (*user) fakeData() []user {
	return fakeUsers
}

type account struct {
	Id           string
	Email        string
	FullName     string
	UserName     string
	AvatarUrl    string
	Organization string
	CreatedDate  string
	Status       int
	UserId       string
}

func (*account) fromDomainLayer(accounts []crossdomain.Account, userAccounts []crossdomain.UserAccount) []account {
	var result []account
	userAccountMap := make(map[string]string)
	for _, ua := range userAccounts {
		userAccountMap[ua.AccountId] = ua.UserId
	}
	for _, a := range accounts {
		var createdDate string
		if a.CreatedDate != nil {
			createdDate = a.CreatedDate.Format(TimeFormat)
		}
		result = append(result, account{
			Id:           a.Id,
			Email:        a.Email,
			FullName:     a.FullName,
			UserName:     a.UserName,
			AvatarUrl:    a.AvatarUrl,
			Organization: a.Organization,
			CreatedDate:  createdDate,
			Status:       a.Status,
			UserId:       userAccountMap[a.Id],
		})
	}
	return result
}

func (*account) toDomainLayer(aa []account) (accounts []*crossdomain.Account, userAccounts []*crossdomain.UserAccount) {
	for _, a := range aa {
		var createdDate *time.Time
		t, err := time.Parse(TimeFormat, a.CreatedDate)
		if err == nil {
			createdDate = &t
		}
		accounts = append(accounts, &crossdomain.Account{
			DomainEntity: domainlayer.DomainEntity{Id: a.Id},
			Email:        a.Email,
			FullName:     a.FullName,
			UserName:     a.UserName,
			AvatarUrl:    a.AvatarUrl,
			Organization: a.Organization,
			CreatedDate:  createdDate,
			Status:       a.Status,
		})
		if a.UserId == "" || a.Id == "" {
			continue
		}
		userAccounts = append(userAccounts, &crossdomain.UserAccount{
			UserId:    a.UserId,
			AccountId: a.Id,
		})
	}
	return
}

type accountUser struct {
	AccountId string
	UserId    string
}

func (au *accountUser) toDomainLayer(accountUsers []accountUser) []crossdomain.UserAccount {
	result := make([]crossdomain.UserAccount, len(accountUsers))
	for _, ac := range accountUsers {
		result = append(result, crossdomain.UserAccount{
			UserId:    ac.UserId,
			AccountId: ac.AccountId,
		})
	}
	return result
}

func (au *accountUser) fromDomainLayer(accountUsers []crossdomain.UserAccount) []accountUser {
	result := make([]accountUser, len(accountUsers))
	for _, ac := range accountUsers {
		result = append(result, accountUser{
			UserId:    ac.UserId,
			AccountId: ac.AccountId,
		})
	}
	return result
}

type team struct {
	Id           string
	Name         string
	Alias        string
	ParentId     string
	SortingIndex int
}

func (*team) fromDomainLayer(tt []crossdomain.Team) []team {
	var result []team
	for _, t := range tt {
		result = append(result, team{
			Id:           t.Id,
			Name:         t.Name,
			Alias:        t.Alias,
			ParentId:     t.ParentId,
			SortingIndex: t.SortingIndex,
		})
	}
	return result
}

func (*team) toDomainLayer(tt []team) []*crossdomain.Team {
	var result []*crossdomain.Team
	for _, t := range tt {
		result = append(result, &crossdomain.Team{
			DomainEntity: domainlayer.DomainEntity{Id: t.Id},
			Name:         t.Name,
			Alias:        t.Alias,
			ParentId:     t.ParentId,
			SortingIndex: t.SortingIndex,
		})
	}
	return result
}

func (*team) fakeData() []team {
	return fakeTeams
}
