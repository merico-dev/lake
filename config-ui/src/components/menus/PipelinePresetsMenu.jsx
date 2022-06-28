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
import PipelineConfigsMenu from '@/components/menus/PipelineConfigsMenu'

const PipelinePresetsMenu = (props) => {
  const {
    namePrefix,
    pipelineSuffixes,
    setNamePrefix = () => {},
    setNameSuffix = () => {},
    setRawConfiguration = () => {},
    advancedMode = false
  } = props
  return (
    <Menu className='pipeline-presets-menu'>
      <label style={{
        fontSize: '10px',
        fontWeight: 800,
        textTransform: 'uppercase',
        padding: '6px 8px',
        display: 'block'
      }}
      >Preset Naming Options
      </label>
      <Menu.Item text='COLLECTION ...' active={namePrefix === 'COLLECT'}>
        <Menu.Item
          icon='key-option'
          text='COLLECT [UNIXTIME]'
          onClick={() => setNamePrefix('COLLECT') | setNameSuffix(pipelineSuffixes[0])}
        />
        <Menu.Item
          icon='key-option'
          text='COLLECT [YYYYMMDDHHMMSS]' onClick={() => setNamePrefix('COLLECT') | setNameSuffix(pipelineSuffixes[3])}
        />
        <Menu.Item
          icon='key-option' text='COLLECT [ISO]'
          onClick={() => setNamePrefix('COLLECT') | setNameSuffix(pipelineSuffixes[2])}
        />
        <Menu.Item icon='key-option' text='COLLECT [UTC]' onClick={() => setNamePrefix('COLLECT') | setNameSuffix(pipelineSuffixes[4])} />
      </Menu.Item>
      <Menu.Item text='SYNCHRONIZE ...' active={namePrefix === 'SYNC'}>
        <Menu.Item
          icon='key-option' text='SYNC [UNIXTIME]'
          onClick={() => setNamePrefix('SYNC') | setNameSuffix(pipelineSuffixes[0])}
        />
        <Menu.Item
          icon='key-option' text='SYNC [YYYYMMDDHHMMSS]'
          onClick={() => setNamePrefix('SYNC') | setNameSuffix(pipelineSuffixes[3])}
        />
        <Menu.Item
          icon='key-option' text='SYNC [ISO]'
          onClick={() => setNamePrefix('SYNC') | setNameSuffix(pipelineSuffixes[2])}
        />
        <Menu.Item
          icon='key-option' text='SYNC [UTC]'
          onClick={() => setNamePrefix('SYNC') | setNameSuffix(pipelineSuffixes[4])}
        />
      </Menu.Item>
      <Menu.Item text='RUN ...' active={namePrefix === 'RUN'}>
        <Menu.Item
          icon='key-option'
          text='RUN [UNIXTIME]'
          onClick={() => setNamePrefix('RUN') | setNameSuffix(pipelineSuffixes[0])}
        />
        <Menu.Item
          icon='key-option' text='RUN [YYYYMMDDHHMMSS]'
          onClick={() => setNamePrefix('RUN') | setNameSuffix(pipelineSuffixes[3])}
        />
        <Menu.Item
          icon='key-option'
          text='RUN [ISO]'
          onClick={() => setNamePrefix('RUN') | setNameSuffix(pipelineSuffixes[2])}
        />
        <Menu.Item
          icon='key-option'
          text='RUN [UTC]'
          onClick={() => setNamePrefix('RUN') | setNameSuffix(pipelineSuffixes[4])}
        />
      </Menu.Item>
      <Menu.Divider />
      <Menu.Item text='Advanced Options' icon='cog'>
        <Menu.Item icon='new-object' text='Save Pipeline Blueprint' disabled />
        {advancedMode && (
          <PipelineConfigsMenu
            setRawConfiguration={setRawConfiguration}
            advancedMode={advancedMode}
          />
        )}
      </Menu.Item>
    </Menu>
  )
}

export default PipelinePresetsMenu
