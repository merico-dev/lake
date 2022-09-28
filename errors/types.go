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

package errors

import (
	"fmt"
	"net/http"
)

// Supported error types
var (
	// Default special error type. If it's wrapping another error, then it will take the type of that error if it's an Error. Otherwise, it equates to Internal.
	Default = register(nil)

	SubtaskErr   = register(&Type{meta: "subtask"})
	NotFound     = register(&Type{httpCode: http.StatusNotFound, meta: "not-found"})
	BadInput     = register(&Type{httpCode: http.StatusBadRequest, meta: "bad-input"})
	Unauthorized = register(&Type{httpCode: http.StatusUnauthorized, meta: "unauthorized"})
	Forbidden    = register(&Type{httpCode: http.StatusForbidden, meta: "forbidden"})
	Internal     = register(&Type{httpCode: http.StatusInternalServerError, meta: "internal"})
	Timeout      = register(&Type{httpCode: http.StatusGatewayTimeout, meta: "timeout"})

	//cached values
	typesByHttpCode = newSyncMap[int, *Type]()
)

type (
	// Type error are constructed from these, and they contain metadata about them.
	Type struct {
		meta string
		// below are optional fields
		httpCode int
	}

	// Option add customized properties to the Error
	Option func(*options)

	options struct {
		data        interface{}
		stackOffset uint
	}
)

func HttpStatus(code int) *Type {
	t, ok := typesByHttpCode.Load(code)
	if !ok { // lazily cache any missing codes
		t = &Type{httpCode: code, meta: fmt.Sprintf("type_http_%d", code)}
		typesByHttpCode.Store(code, t)
	}
	return t
}

// New constructs a new Error instance with this message
func (t *Type) New(message string, opts ...Option) Error {
	return newSingleCrdbError(t, nil, message, opts...)
}

// Wrap constructs a new Error instance with this message and wraps the passed in error. A nil 'err' will return a nil Error.
func (t *Type) Wrap(err error, message string, opts ...Option) Error {
	if err == nil {
		return nil
	}
	return newSingleCrdbError(t, err, message, opts...)
}

// WrapRaw constructs a new Error instance that directly wraps this error with no additional context. A nil 'err' will return a nil Error.
// This additional wrapping will create an additional nested stacktrace on this line if the setting is enabled.
func (t *Type) WrapRaw(err error) Error {
	return t.wrapRaw(err, true, withStackOffset(1))
}

func (t *Type) wrapRaw(err error, forceWrap bool, opts ...Option) Error {
	if err == nil {
		return nil
	}
	if !forceWrap {
		if lakeErr, ok := err.(Error); ok {
			return lakeErr
		}
	}
	msg := ""
	lakeErr := AsLakeErrorType(err)
	if lakeErr != nil {
		if !forceWrap {
			return lakeErr
		}
		msg = "" // there's nothing new to add
	} else {
		msg = err.Error()
	}
	return newSingleCrdbError(t, err, msg, opts...)
}

// Combine constructs a new Error from combining multiple errors. Stacktrace info for each of the errors will not be present in the result, so it's
// best to log the errors before combining them.
func (t *Type) Combine(errs []error) Error {
	return newCombinedCrdbError(t, errs)
}

// GetHttpCode gets the associated Http code with this Type, if explicitly set, otherwise http.StatusInternalServerError
func (t *Type) GetHttpCode() int {
	if t.httpCode == 0 {
		return http.StatusInternalServerError
	}
	return t.httpCode
}

// WithData associate data with this Error
func WithData(data interface{}) Option {
	return func(opts *options) {
		opts.data = data
	}
}

// withStackOffset the number of indirections in function calls before the 'new' error function is called (e.g. newCrdbError).
// this must remain internal to the errors package
func withStackOffset(offset uint) Option {
	return func(opts *options) {
		opts.stackOffset = offset
	}
}

func register(t *Type) *Type {
	if t == nil {
		t = &Type{meta: "default"}
		typesByHttpCode.Store(t.httpCode, t)
	} else if t.httpCode != 0 {
		typesByHttpCode.Store(t.httpCode, t)
	}
	return t
}
