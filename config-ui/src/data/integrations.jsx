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
import React from 'react'
import { Icon } from '@blueprintjs/core'
import { Providers, ProviderLabels, ProviderTypes } from '@/data/Providers'

import JiraSettings from '@/pages/configure/settings/jira'
import GitlabSettings from '@/pages/configure/settings/gitlab'
import JenkinsSettings from '@/pages/configure/settings/jenkins'
import GithubSettings from '@/pages/configure/settings/github'

import { ReactComponent as GitlabProvider } from '@/images/integrations/gitlab.svg'
import { ReactComponent as JenkinsProvider } from '@/images/integrations/jenkins.svg'
import { ReactComponent as JiraProvider } from '@/images/integrations/jira.svg'
import { ReactComponent as GitHubProvider } from '@/images/integrations/github.svg'
// import GitExtractorProvider from '@/images/git.png'
// import RefDiffProvider from '@/images/git-diff.png'
// import { ReactComponent as NullProvider } from '@/images/integrations/null.svg'

const integrationsData = [
  {
    id: Providers.GITLAB,
    type: ProviderTypes.INTEGRATION,
    enabled: true,
    multiConnection: true,
    name: ProviderLabels.GITLAB,
    icon: <GitlabProvider className='providerIconSvg' width='30' height='30' style={{ float: 'left', marginTop: '5px' }} />,
    iconDashboard: <GitlabProvider className='providerIconSvg' width='40' height='40' />,
    settings: ({ activeProvider, activeConnection, isSaving, isSavingConnection, setSettings }) => (
      <GitlabSettings
        provider={activeProvider}
        connection={activeConnection}
        isSaving={isSaving}
        isSavingConnection={isSavingConnection}
        onSettingsChange={setSettings}
      />
    )
  },
  {
    id: Providers.JENKINS,
    type: ProviderTypes.INTEGRATION,
    enabled: true,
    multiConnection: true,
    name: ProviderLabels.JENKINS,
    icon: <JenkinsProvider className='providerIconSvg' width='30' height='30' style={{ float: 'left', marginTop: '5px' }} />,
    iconDashboard: <JenkinsProvider className='providerIconSvg' width='40' height='40' />,
    settings: ({ activeProvider, activeConnection, isSaving, isSavingConnection, setSettings }) => (
      <JenkinsSettings
        provider={activeProvider}
        connection={activeConnection}
        isSaving={isSaving}
        isSavingConnection={isSavingConnection}
        onSettingsChange={setSettings}
      />
    )
  },
  {
    id: Providers.JIRA,
    type: ProviderTypes.INTEGRATION,
    enabled: true,
    multiConnection: true,
    name: ProviderLabels.JIRA,
    icon: <JiraProvider className='providerIconSvg' width='30' height='30' style={{ float: 'left', marginTop: '5px' }} />,
    iconDashboard: <JiraProvider className='providerIconSvg' width='40' height='40' />,
    settings: ({ activeProvider, activeConnection, isSaving, isSavingConnection, setSettings }) => (
      <JiraSettings
        provider={activeProvider}
        connection={activeConnection}
        isSaving={isSaving}
        isSavingConnection={isSavingConnection}
        onSettingsChange={setSettings}
      />
    )
  },
  {
    id: Providers.GITHUB,
    type: ProviderTypes.INTEGRATION,
    enabled: true,
    multiConnection: true,
    name: ProviderLabels.GITHUB,
    icon: <GitHubProvider className='providerIconSvg' width='30' height='30' style={{ float: 'left', marginTop: '5px' }} />,
    iconDashboard: <GitHubProvider className='providerIconSvg' width='40' height='40' />,
    settings: ({ activeProvider, activeConnection, isSaving, isSavingConnection, setSettings }) => (
      <GithubSettings
        provider={activeProvider}
        connection={activeConnection}
        isSaving={isSaving}
        isSavingConnection={isSavingConnection}
        onSettingsChange={setSettings}
      />
    )
  },
]

const pluginsData = [
  {
    id: Providers.GITEXTRACTOR,
    type: ProviderTypes.PIPELINE,
    enabled: true,
    multiConnection: false,
    name: ProviderLabels.GITEXTRACTOR,
    icon: <Icon icon='box' size={30} />,
    iconDashboard: <Icon icon='box' size={32} />,
    settings: ({ activeProvider, activeConnection, isSaving, setSettings }) => (
      null
    )
  },
  {
    id: Providers.REFDIFF,
    type: ProviderTypes.PIPELINE,
    enabled: true,
    multiConnection: false,
    name: ProviderLabels.REFDIFF,
    icon: <Icon icon='box' size={30} />,
    iconDashboard: <Icon icon='box' size={32} />,
    settings: ({ activeProvider, activeConnection, isSaving, setSettings }) => (
      null
    )
  },
]

export {
  integrationsData,
  pluginsData
}
