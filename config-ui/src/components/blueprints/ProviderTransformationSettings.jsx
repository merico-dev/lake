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
import React, { Fragment, useEffect, useState, useCallback } from 'react'
import {
  Providers,
  ProviderTypes,
  ProviderIcons,
  ConnectionStatus,
  ConnectionStatusLabels,
} from '@/data/Providers'
import { DataEntities, DataEntityTypes } from '@/data/DataEntities'
import JiraSettings from '@/pages/configure/settings/jira'
import GitlabSettings from '@/pages/configure/settings/gitlab'
import JenkinsSettings from '@/pages/configure/settings/jenkins'
import GithubSettings from '@/pages/configure/settings/github'

const ProviderTransformationSettings = (props) => {
  const {
    provider,
    configuredConnection,
    configuredProject,
    configuredBoard,
    transformation,
    boards = [],
    issueTypes = [],
    fields = [],
    onSettingsChange = () => {},
    isSaving = false,
    isSavingConnection = false,
    jiraProxyError,
    isFetchingJIRA = false
  } = props

  useEffect(() => {
    console.log('>>> activeTransformation?', transformation)
  }, [transformation, boards, issueTypes, fields, configuredBoard])

  return (
    <div className='transformation-settings' data-provider={provider?.id}>
      {provider?.id === Providers.GITHUB && (
        <GithubSettings
          provider={provider}
          connection={configuredConnection}
          configuredProject={configuredProject}
          transformation={transformation}
          onSettingsChange={onSettingsChange}
          entity={DataEntityTypes.TICKET}
          isSaving={isSaving}
          isSavingConnection={isSavingConnection}
      />
      )}

      {provider?.id === Providers.GITLAB && (
        <GitlabSettings
          provider={provider}
          connection={configuredConnection}
          configuredProject={configuredProject}
          transformation={transformation}
          onSettingsChange={onSettingsChange}
          entity={DataEntityTypes.TICKET}
          isSaving={isSaving}
          isSavingConnection={isSavingConnection}
      />
      )}

      {provider?.id === Providers.JIRA && (
        <JiraSettings
          provider={provider}
          connection={configuredConnection}
          configuredBoard={configuredBoard}
          boards={boards}
          issueTypes={issueTypes}
          fields={fields}
          transformation={transformation}
          onSettingsChange={onSettingsChange}
          entity={DataEntityTypes.TICKET}
          isSaving={isSaving}
          isSavingConnection={isSavingConnection}
          jiraProxyError={jiraProxyError}
          isFetchingJIRA={isFetchingJIRA}
      />
      )}

      {provider?.id === Providers.JENKINS && (
        <JenkinsSettings
          provider={provider}
          connection={configuredConnection}
          transformation={transformation}
          onSettingsChange={onSettingsChange}
          entity={DataEntityTypes.TICKET}
          isSaving={isSaving}
          isSavingConnection={isSavingConnection}
      />
      )}
    </div>
  )
}

export default ProviderTransformationSettings
