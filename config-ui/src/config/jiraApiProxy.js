/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
const API_VERSION = 2
// @todo: add string replacer for [:connectionId] or refactor this const
const API_PROXY_ENDPOINT = '/api/plugins/jira/connections/[:connectionId:]/proxy/rest'
const ISSUE_TYPES_ENDPOINT = `${API_PROXY_ENDPOINT}/api/${API_VERSION}/issuetype`
const ISSUE_FIELDS_ENDPOINT = `${API_PROXY_ENDPOINT}/api/${API_VERSION}/field`
const BOARDS_ENDPOINT = `${API_PROXY_ENDPOINT}/api/${API_VERSION}/board`

export {
  API_VERSION,
  API_PROXY_ENDPOINT,
  ISSUE_TYPES_ENDPOINT,
  ISSUE_FIELDS_ENDPOINT,
  BOARDS_ENDPOINT,
}
