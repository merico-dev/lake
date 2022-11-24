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
	"github.com/apache/incubator-devlake/errors"
	"github.com/apache/incubator-devlake/models"
	"github.com/apache/incubator-devlake/plugins/core/dal"
)

// UnwrapObject if the actual object is wrapped in some proxy, it unwinds and returns it, otherwise this is idempotent
func UnwrapObject(ifc any) any {
	if dynamic, ok := ifc.(*models.DynamicTabler); ok {
		return dynamic.Unwrap()
	}
	return ifc
}

// CallDB wraps DB calls with this signature, and handles the case if the struct is wrapped in a models.DynamicTabler.
func CallDB(f func(any, ...dal.Clause) errors.Error, x any, clauses ...dal.Clause) errors.Error {
	if dynamic, ok := x.(*models.DynamicTabler); ok {
		clauses = append(clauses, dal.From(dynamic.TableName()))
		x = dynamic.Unwrap()
	}
	return f(x, clauses...)
}
