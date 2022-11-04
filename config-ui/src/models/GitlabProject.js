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
 * @typedef {object} GitlabProject
 * @property {number?} id
 * @property {number?} key
 * @property {number?} projectId
 * @property {string|number?} name
 * @property {string|number?} value
 * @property {string|number?} title
 * @property {string?} shortTitle
 * @property {string?} icon
 * @property {private|public?} visibility
 * @property {object?} owner
 * @property {number?} creator_id
 * @property {object?} _links
 * @property {object?} statistics
 * @property {string?} name_with_namespace
 * @property {string?} path_with_namespace
 * @property {string?} path
 * @property {string?} default_branch
 * @property {<Array<string>>} topics
 * @property {string?} ssh_url_to_repo
 * @property {string?} http_url_to_repo
 * @property {string?} web_url
 * @property {string?} readme_url
 * @property {string?} avatar_url
 * @property {string?} forks_count
 * @property {string?} star_count
 * @property {object?} namespace
 * @property {string|Date?} created_at
 * @property {boolean?} archived
 * @property {boolean?} useApi
 * @property {project|board?} variant
 * @property {string?} providerId
 */
class GitlabProject {
  constructor(data = {}) {
    this.id = data?.id || data?.projectId || null
    this.key = data?.key || this.id || null
    this.projectId = data?.projectId || this.id || null
    this.value = data?.value || this.id || this.projectId || null
    this.name = data?.name || this.projectId || this.value || null
    this.title = data?.title || this.name || this.id || null
    this.shortTitle = data?.shortTitle || null
    this.icon = data?.icon || null

    // @todo: GitLab API props to camelCase
    this.visibility = data?.visibility || 'private'
    this.description = data?.description || null
    this.owner = data?.owner || null
    this.creator_id = data?.creator_id || null
    this._links = data?._links || null
    this.statistics = data?.statistics || null
    this.name_with_namespace = data?.name_with_namespace || null
    this.path_with_namespace = data?.path_with_namespace || null
    this.path = data?.path || null
    this.default_branch = data?.default_branch || null
    this.topcs = data?.topics || null
    this.ssh_url_to_repo = data?.ssh_url_to_repo || null
    this.http_url_to_repo = data?.http_url_to_repo || null
    this.web_url = data?.web_url || null
    this.readme_url = data?.readme_url || null
    this.avatar_url = data?.avatar_url || null
    this.forks_count = data?.forks_count || null
    this.star_count = data?.star_count || null
    this.namespace = data?.namespace || null

    this.created_at = data?.created_at || null
    this.archived = data?.archived || false

    this.useApi = data?.useApi || false
    this.variant = data?.variant || 'project'
    this.providerId = 'gitlab'
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

export default GitlabProject
