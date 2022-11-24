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
	"encoding/base64"
	"fmt"
	"github.com/apache/incubator-devlake/errors"
	"reflect"
	"strconv"
	"strings"

	"github.com/apache/incubator-devlake/models/common"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/core/dal"
	"github.com/go-playground/validator/v10"
)

// BaseConnection FIXME ...
type BaseConnection struct {
	Name string `gorm:"type:varchar(100);uniqueIndex" json:"name" validate:"required"`
	common.Model
}

// BasicAuth FIXME ...
type BasicAuth struct {
	Username string `mapstructure:"username" validate:"required" json:"username"`
	Password string `mapstructure:"password" validate:"required" json:"password" encrypt:"yes"`
}

// GetEncodedToken FIXME ...
func (ba BasicAuth) GetEncodedToken() string {
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%v:%v", ba.Username, ba.Password)))
}

// AccessToken FIXME ...
type AccessToken struct {
	Token string `mapstructure:"token" validate:"required" json:"token" encrypt:"yes"`
}

// AppKey FIXME ...
type AppKey struct {
	AppId     string `mapstructure:"app_id" validate:"required" json:"appId"`
	SecretKey string `mapstructure:"secret_key" validate:"required" json:"secretKey" encrypt:"yes"`
}

// RestConnection FIXME ...
type RestConnection struct {
	BaseConnection   `mapstructure:",squash"`
	Endpoint         string `mapstructure:"endpoint" validate:"required" json:"endpoint"`
	Proxy            string `mapstructure:"proxy" json:"proxy"`
	RateLimitPerHour int    `comment:"api request rate limit per hour" json:"rateLimitPerHour"`
}

// ConnectionApiHelper is used to write the CURD of connection
type ConnectionApiHelper struct {
	encKey    string
	log       core.Logger
	db        dal.Dal
	validator *validator.Validate
}

// NewConnectionHelper FIXME ...
func NewConnectionHelper(
	basicRes core.BasicRes,
	vld *validator.Validate,
) *ConnectionApiHelper {
	if vld == nil {
		vld = validator.New()
	}
	return &ConnectionApiHelper{
		encKey:    basicRes.GetConfig(core.EncodeKeyEnvStr),
		log:       basicRes.GetLogger(),
		db:        basicRes.GetDal(),
		validator: vld,
	}
}

// Create a connection record based on request body
func (c *ConnectionApiHelper) Create(connection interface{}, input *core.ApiResourceInput) errors.Error {
	// update fields from request body
	err := c.merge(connection, input.Body)
	if err != nil {
		return err
	}
	return c.save(connection)
}

// Patch (Modify) a connection record based on request body
func (c *ConnectionApiHelper) Patch(connection interface{}, input *core.ApiResourceInput) errors.Error {
	err := c.First(connection, input.Params)
	if err != nil {
		return err
	}

	err = c.merge(connection, input.Body)
	if err != nil {
		return err
	}
	return c.save(connection)
}

// First finds connection from db  by parsing request input and decrypt it
func (c *ConnectionApiHelper) First(connection interface{}, params map[string]string) errors.Error {
	connectionId := params["connectionId"]
	if connectionId == "" {
		return errors.BadInput.New("missing connectionId")
	}
	id, err := strconv.ParseUint(connectionId, 10, 64)
	if err != nil || id < 1 {
		return errors.BadInput.New("invalid connectionId")
	}
	return c.FirstById(connection, id)
}

// FirstById finds connection from db by id and decrypt it
func (c *ConnectionApiHelper) FirstById(connection interface{}, id uint64) errors.Error {
	err := CallDB(c.db.First, connection, dal.Where("id = ?", id))
	if err != nil {
		return err
	}
	c.decrypt(connection)
	return nil
}

// List returns all connections with password/token decrypted
func (c *ConnectionApiHelper) List(connections interface{}) errors.Error {
	err := CallDB(c.db.All, connections)
	if err != nil {
		return err
	}
	conns := reflect.ValueOf(UnwrapObject(connections)).Elem()
	for i := 0; i < conns.Len(); i++ {
		c.decrypt(conns.Index(i).Addr().Interface())
	}
	return nil
}

// Delete connection
func (c *ConnectionApiHelper) Delete(connection interface{}) errors.Error {
	return CallDB(c.db.Delete, connection)
}

func (c *ConnectionApiHelper) merge(connection interface{}, body map[string]interface{}) errors.Error {
	return Decode(body, connection, c.validator)
}

func (c *ConnectionApiHelper) save(connection interface{}) errors.Error {
	c.encrypt(connection)
	err := CallDB(c.db.CreateOrUpdate, connection)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			return errors.BadInput.Wrap(err, "duplicated Connection Name")
		}
		return err
	}
	c.decrypt(connection)
	return nil
}

func (c *ConnectionApiHelper) decrypt(connection interface{}) {
	err := UpdateEncryptFields(connection, func(encrypted string) (string, errors.Error) {
		return core.Decrypt(c.encKey, encrypted)
	})
	if err != nil {
		c.log.Error(err, "failed to decrypt")
	}
}

func (c *ConnectionApiHelper) encrypt(connection interface{}) {
	err := UpdateEncryptFields(connection, func(plaintext string) (string, errors.Error) {
		return core.Encrypt(c.encKey, plaintext)
	})
	if err != nil {
		c.log.Error(err, "failed to encrypt")
	}
}

// UpdateEncryptFields update fields of val with tag `encrypt:"yes|true"`
func UpdateEncryptFields(val interface{}, update func(in string) (string, errors.Error)) errors.Error {
	val = UnwrapObject(val)
	v := reflect.ValueOf(val)
	if v.Kind() != reflect.Ptr {
		panic(errors.Default.New(fmt.Sprintf("val is not a pointer: %v", val)))
	}
	e := v.Elem()
	if e.Kind() != reflect.Struct {
		panic(errors.Default.New(fmt.Sprintf("*val is not a struct: %v", val)))
	}
	t := e.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		if field.Type.Kind() == reflect.Struct {
			err := UpdateEncryptFields(e.Field(i).Addr().Interface(), update)
			if err != nil {
				return err
			}
		} else if field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct {
			fmt.Printf("field : %v\n", e.Field(i).Interface())
			err := UpdateEncryptFields(e.Field(i).Interface(), update)
			if err != nil {
				return err
			}
		} else if field.Type.Kind() == reflect.String {
			tagValue := field.Tag.Get("encrypt")
			if tagValue == "yes" || tagValue == "true" {
				out, err := update(e.Field(i).String())
				if err != nil {
					return err
				}
				e.Field(i).Set(reflect.ValueOf(out))
			}
		}
	}
	return nil
}
