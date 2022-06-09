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
// import { CSSTransition } from 'react-transition-group'
import {
  Classes,
  Drawer,
  DrawerSize,
  Card,
  Elevation,
  Position,
  Colors,
  Icon
} from '@blueprintjs/core'

const CodeInspector = (props) => {
  const { activePipeline, isOpen, onClose, titleIcon = 'code', title = `Inspect RUN #${activePipeline.ID}`, subtitle = 'JSON RESPONSE', hasBackdrop = true } = props

  return (
    <Drawer
      className='drawer-json-inspector'
      icon={titleIcon}
      onClose={() => onClose(false)}
      title={title}
      position={Position.RIGHT}
      size={DrawerSize.SMALL}
      autoFocus
      canEscapeKeyClose
      canOutsideClickClose
      enforceFocus
      hasBackdrop={hasBackdrop}
      isOpen={isOpen}
      usePortal
    >
      <div className={Classes.DRAWER_BODY}>
        <div className={Classes.DIALOG_BODY}>
          <h3 className='no-user-select' style={{ margin: 0, padding: '8px 0' }}>
            <span style={{ float: 'right', fontSize: '9px', color: '#aaaaaa' }}>application/json</span> {subtitle}
          </h3>
          <p className='no-user-select'>
            If you are submitting a
            <strong> Bug-Report</strong> regarding a Pipeline Run, include the output below for better debugging.
          </p>
          <div className='formContainer'>
            <Card
              className='code-inspector-card'
              interactive={false}
              elevation={Elevation.ZERO}
              style={{ padding: '6px 12px', minWidth: '320px', width: '100%', maxWidth: '601px', marginBottom: '20px', overflow: 'auto' }}
            >

              <code>
                <pre style={{ fontSize: '10px' }}>
                  {JSON.stringify(activePipeline, null, '  ')}
                </pre>
              </code>
            </Card>
          </div>
          <p style={{ fontSize: '10px', lineHeight: '120%', opacity: 0.6 }}>
            <Icon icon='info-sign' color={Colors.GRAY5} size={12} style={{ marginRight: '4px' }} />
            <strong>Pipelines</strong> &mdash;
            {' '}For a project with 10k commits and 5k Issues, this can take up to 20 minutes for collecting JIRA, GitLab, and Jenkins data.
            {' '}Collection will take longer for GitHub due to rate limits.
            {' '}You can accelerate the process by configuring multiple tokens.
          </p>
        </div>
      </div>
    </Drawer>

  )
}

export default CodeInspector
