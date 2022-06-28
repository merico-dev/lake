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
import {
  Popover,
  Tag,
  Icon,
  Intent,
  Colors,
} from '@blueprintjs/core'
import { ProviderLabels, ProviderIcons } from '@/data/Providers'
const PipelineTasks = (props) => {
  const {
    tasks = []
  } = props

  const renderPluginTag = (providerTask, pIdx) => {
    const provider = providerTask.Plugin || providerTask.plugin
    return provider && (
      <Popover key={`provider-popover-key-${pIdx}`} usePortal={true}>
        <Tag
          key={`provider-icon-key-${pIdx}`} intent={Intent.NONE} round='true' style={{
            backgroundColor: '#fff',
            border: '1px solid #aaa',
            margin: '0 5px 5px 0',
            boxShadow: '0px 0px 2px rgba(0, 0, 0, 0.45)',
            color: Colors.DARK_GRAY1
          }}
        >
          <span className='detected-provider-icon' style={{ margin: '2px 3px 0 0px', float: 'left' }}>
            {ProviderIcons[provider] ? ProviderIcons[provider](20, 20) : <></>}
          </span>
          <span style={{ display: 'flex', marginTop: '3px', fontWeight: 800 }}>
            {ProviderLabels[provider?.toUpperCase()] ? ProviderLabels[provider?.toUpperCase()] : 'Data Provider'}
          </span>
        </Tag>
        <div style={{ padding: '10px', maxWidth: '340px', overflow: 'hidden', overflowX: 'auto' }}>
          <div style={{ marginBottom: '10px', fontWeight: 700, fontSize: '14px' }}>
            <Icon icon='layers' size={16} /> {ProviderLabels[provider?.toUpperCase()] ? ProviderLabels[provider?.toUpperCase()] : 'Plugin'}
          </div>
          <code>
            {JSON.stringify(tasks.flat().find(t => t.Plugin === provider || t.plugin === provider))}
          </code>
        </div>
      </Popover>
    )
  }

  return (
    <>
      {tasks.map((providerTask, pIdx) => (
        renderPluginTag(providerTask, pIdx)
      ))}
    </>
  )
}

export default PipelineTasks
