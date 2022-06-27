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
  Button,
  Icon,
  Intent,
  TextArea,
  InputGroup,
  ButtonGroup,
  Popover,
  Position,
  Divider,
  Elevation,
  Card,
  Colors,
} from '@blueprintjs/core'
import PipelineConfigsMenu from '@/components/menus/PipelineConfigsMenu'
import BlueprintNameCard from '@/components/blueprints/BlueprintNameCard'

const AdvancedJSON = (props) => {
  const {
    activeStep,
    name,
    runTasksAdvanced = [],
    rawConfiguration,
    blueprintConnections = [],
    connectionsList = [],
    setBlueprintName = () => {},
    fieldHasError = () => {},
    getFieldError = () => {},
    onAdvancedMode = () => {},
    isMultiStagePipeline = () => {},
    setRawConfiguration = () => {},
    isSaving = false,
    isRunning = false,
    isValidConfiguration = false,
    advancedMode = false,
    validationAdvancedError,
    validationErrors = []
  } = props

  return (
    <div className='workflow-step workflow-step-advanced-json' data-step={activeStep?.id}>
      <BlueprintNameCard 
        activeStep={activeStep}
        advancedMode={advancedMode}
        name={name}
        setBlueprintName={setBlueprintName}
      />

      <Card
        className='workflow-card workflow-panel-card'
        elevation={Elevation.TWO}
        style={{ width: '100%' }}
      >
        <h3>
          JSON Configuration
        </h3>
        <Divider className='section-divider' />

        <h4>Task Editor</h4>
        <p>Enter JSON Configuration or preload from a template</p>

        <Card
          className='code-editor-card'
          interactive={false}
          elevation={Elevation.ZERO}
          style={{ padding: '2px', minWidth: '320px', width: '100%', maxWidth: '100%', marginBottom: '20px' }}
        >
          <TextArea
            growVertically={false}
            fill={true}
            className='codeArea'
            style={{ minHeight: '240px' }}
            value={rawConfiguration}
            onChange={(e) => setRawConfiguration(e.target.value)}
          />
          <div
            className='code-editor-card-footer'
            style={{
              display: 'flex',
              justifyContent: 'flex-end',
              padding: '5px',
              borderTop: '1px solid #eeeeee',
              fontSize: '11px'
            }}
          >
            <ButtonGroup
              intent={Intent.PRIMARY}
              minimal
              className='code-editor-controls' style={{
                borderRadius: '3px',
                // boxShadow: '0px 0px 2px rgba(0, 0, 0, 0.30)'
              }}
            >
              <Button
                small text='Reset'
                icon='eraser'
                onClick={() => setRawConfiguration('[[]]')}
              />
              <Popover
                className='popover-options-menu-trigger'
                popoverClassName='popover-options-menu'
                position={Position.TOP}
                usePortal={true}
              >
                <Button
                  disabled={isRunning}
                  rightIcon='caret-down'
                  text='Load Templates'
                />
                <>
                  <PipelineConfigsMenu
                    setRawConfiguration={setRawConfiguration}
                    advancedMode={advancedMode}
                  />
                </>
              </Popover>
              {/* <Button
                disabled={!isValidConfiguration}
                small text='Format' icon='align-left'
                onClick={() => formatRawCode()}
              /> */}
              {/* <Button
                small text='Revert' icon='reset'
                onClick={() => setRawConfiguration(JSON.stringify([runTasks], null, '  '))}
              /> */}
            </ButtonGroup>
          </div>
        </Card>


      </Card>

      <div className='mode-notice normal-mode-notice'>
        <p>To visually define blueprint tasks, please use <a href='#' rel='noreferrer' onClick={() => onAdvancedMode(false)}>Normal Mode</a></p>
      </div>

    </div>
  )
}

export default AdvancedJSON
