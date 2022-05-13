package helper

import (
	"testing"

	"github.com/merico-dev/lake/models/domainlayer"
	"github.com/merico-dev/lake/models/domainlayer/ticket"
	"github.com/stretchr/testify/assert"
)

func Test_getPrimaryKeyValue(t *testing.T) {
	type args struct {
		iface interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"",
			args{&ticket.Sprint{
				DomainEntity: domainlayer.DomainEntity{Id: "abc"},
			},
			},
			"abc",
		},
		{
			"",
			args{ticket.Sprint{
				DomainEntity: domainlayer.DomainEntity{Id: "abc"},
			},
			},
			"abc",
		},
		{
			"",
			args{ticket.SprintIssue{
				SprintId: "abc",
				IssueId:  "123",
			},
			},
			"abc:123",
		},
		{
			"",
			args{&ticket.SprintIssue{
				SprintId: "abc",
				IssueId:  "123",
			},
			},
			"abc:123",
		},
		{
			"",
			args{ticket.Issue{}},
			"",
		},
		{
			"",
			args{&ticket.Issue{}},
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, getPrimaryKeyValue(tt.args.iface), "getPrimaryKeyValue(%v)", tt.args.iface)
		})
	}
}

func Test_hasPrimaryKey(t *testing.T) {
	type args struct {
		iface interface{}
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"",
			args{ticket.SprintIssue{
				SprintId: "abc",
				IssueId:  "123",
			},
			},
			true,
		},
		{
			"",
			args{&ticket.SprintIssue{
				SprintId: "abc",
				IssueId:  "123",
			},
			},
			true,
		},
		{
			"",
			args{ticket.Issue{}},
			true,
		},
		{
			"",
			args{&ticket.Issue{}},
			true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, hasPrimaryKey(tt.args.iface), "hasPrimaryKey(%v)", tt.args.iface)
		})
	}
}
