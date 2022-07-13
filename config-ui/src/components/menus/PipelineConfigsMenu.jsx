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
import { Menu } from '@blueprintjs/core'
import { defaultConfig as samplePipelineConfig } from '@/data/pipeline-config-samples/default'
import { refdiffConfig as sampleRefdiffPipelineConfig } from '@/data/pipeline-config-samples/refdiff'
import { gitextractorConfig as sampleGitextractorPipelineConfig } from '@/data/pipeline-config-samples/gitextractor'
import { githubConfig as sampleGithubPipelineConfig } from '@/data/pipeline-config-samples/github'
import { gitlabConfig as sampleGitlabPipelineConfig } from '@/data/pipeline-config-samples/gitlab'
import { jiraConfig as sampleJiraPipelineConfig } from '@/data/pipeline-config-samples/jira'
import { jenkinsConfig as sampleJenkinsPipelineConfig } from '@/data/pipeline-config-samples/jenkins'
import { feishuConfig as sampleFeishuPipelineConfig } from '@/data/pipeline-config-samples/feishu'
import { dbtConfig as sampleDbtPipelineConfig } from '@/data/pipeline-config-samples/dbt'
import { starRocksConfig as sampleStarRocksConfigPipelineConfig } from '@/data/pipeline-config-samples/starrocks'

const PipelineConfigsMenu = (props) => {
  const {
    setRawConfiguration = () => {},
    showTitle = false,
    // advancedMode = false
  } = props
  return (
    <Menu className='pipeline-configs-menu'>
      {showTitle && (<label style={{
        fontSize: '10px',
        fontWeight: 800,
        textTransform: 'uppercase',
        padding: '6px 8px',
        display: 'block'
      }}
                     >SAMPLE PIPELINE CONFIGURATIONS
      </label>)}
      <Menu.Item
        icon='code' text='Load General Configuration'
        onClick={() => setRawConfiguration(JSON.stringify(samplePipelineConfig, null, '  '))}
      />
      <Menu.Item
        icon='code'
        text='Load RefDiff Configuration'
        onClick={() => setRawConfiguration(JSON.stringify(sampleRefdiffPipelineConfig, null, '  '))}
      />
      <Menu.Item
        icon='code' text='Load GitExtractor Configuration'
        onClick={() => setRawConfiguration(JSON.stringify(sampleGitextractorPipelineConfig, null, '  '))}
      />
      <Menu.Item
        icon='code' text='Load GitHub Configuration'
        onClick={() => setRawConfiguration(JSON.stringify(sampleGithubPipelineConfig, null, '  '))}
      />
      <Menu.Item
        icon='code' text='Load GitLab Configuration'
        onClick={() => setRawConfiguration(JSON.stringify(sampleGitlabPipelineConfig, null, '  '))}
      />
      <Menu.Item
        icon='code' text='Load JIRA Configuration'
        onClick={() => setRawConfiguration(JSON.stringify(sampleJiraPipelineConfig, null, '  '))}
      />
      <Menu.Item
        icon='code' text='Load Jenkins Configuration'
        onClick={() => setRawConfiguration(JSON.stringify(sampleJenkinsPipelineConfig, null, '  '))}
      />
      <Menu.Item
        icon='code' text='Load Feishu Configuration'
        onClick={() => setRawConfiguration(JSON.stringify(sampleFeishuPipelineConfig, null, '  '))}
      />
      <Menu.Item
        icon='code' text='Load DBT Configuration'
        onClick={() => setRawConfiguration(JSON.stringify(sampleDbtPipelineConfig, null, '  '))}
      />
      <Menu.Item
        icon='code' text='Load StarRocks Configuration'
        onClick={() => setRawConfiguration(JSON.stringify(sampleStarRocksConfigPipelineConfig, null, '  '))}
      />
    </Menu>
  )
}

export default PipelineConfigsMenu
