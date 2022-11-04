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

/**
 * @typedef {object} JiraBoard
 * @property {number?} id
 * @property {number?} key
 * @property {string?} self
 * @property {string?} name
 * @property {number|string?} value
 * @property {number|string?} title
 * @property {string?} shortTitle
 * @property {string?} icon
 * @property {kanban|scrum?} type
 * @property {object?} location
 * @property {boolean?} useApi
 * @property {project|board?} variant
 * @property {string?} providerId
 */
class JiraBoard {
  constructor(data = {}) {
    this.id = data?.id || null
    this.key = data?.key || this.id || null
    this.self = data?.self || null
    this.name = data?.name || `Board ${this.id}` || null
    this.value = data?.value || this.name || `Board ${this.id}` || null
    this.title = data?.title || this.title || `Board ${this.id}` || null
    this.shortTitle = data?.shortTitle || null
    this.icon = data?.icon || null
    this.type = data?.type || 'kanban'
    this.location = data?.location
      ? { ...data?.location }
      : {
          projectId: null,
          displayName: null,
          projectName: null,
          projectKey: null,
          projectTypeKey: null,
          avatarURI: null,
          name: null
        }

    this.useApi = data?.useApi || true
    this.variant = data?.variant || 'board'
    this.providerId = 'jira'
  }

  get(property) {
    return this[property]
  }

  set(property, value) {
    this[property] = value
    return this.property
  }

  getConfiguredEntityId() {
    return this.id
  }
}

export default JiraBoard
