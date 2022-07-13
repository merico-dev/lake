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
const NullConnection = {
  id: null,
  ID: null,
  name: null,
  endpoint: null,
  proxy: null,
  token: null,
  username: null,
  password: null,
  basicAuthEncoded: null, // NOTE: we probably want to exclude/null this when exposing this object
  JIRA_ISSUE_TYPE_MAPPING: null,
  JIRA_ISSUE_EPIC_KEY_FIELD: null,
  JIRA_ISSUE_STORYPOINT_FIELD: null,
  JIRA_BOARD_GITLAB_PROJECTS: null,
  JIRA_ISSUE_INCIDENT_STATUS_MAPPING: null,
  JIRA_ISSUE_STORY_STATUS_MAPPING: null,
  createdAt: null,
  updatedAt: null,
}

export {
  NullConnection
}
