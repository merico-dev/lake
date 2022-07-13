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

const AdvancedJSONValidation = (props) => {
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
    onPrev = () => {},
    isSaving = false,
    isRunning = false,
    isValidConfiguration = false,
    advancedMode = false,
    validationAdvancedError = null,
    validationErrors = [],
  } = props

  return (
    <div
      className='workflow-step workflow-step-advanced-json'
      data-step={activeStep?.id}
    >
      <Card
        className='workflow-card workflow-panel-card'
        elevation={Elevation.TWO}
        style={{ width: '100%' }}
      >
        <h3>Validate JSON Tasks</h3>
        <Divider className='section-divider' />

        {isValidConfiguration ? (<p className='alert neutral'>
          <strong>Your Blueprint JSON Configuration is valid</strong>. Please see below for
          detected data providers.
          <br />
          <a
            href='#'
            className='more-link'
            rel='noreferrer'
            style={{
              // color: '#7497F7',
              marginTop: '5px',
              display: 'inline-block',
            }}
          >
            Find out more
          </a>
        </p>) : (
                                   <p className='alert error'>
            <strong>Your Blueprint JSON Configuration is invalid.</strong> {validationAdvancedError}
            <br />
            <a
                                       href='#'
                                       className='more-link'
                                       rel='noreferrer'
                                       style={{
                marginTop: '5px',
                display: 'inline-block',
              }}
                                       onClick={onPrev}
                                     >
              Go Back
                                     </a>
          </p>
        )}

        <code style={{ fontSize: '10px', borderRadius: '8px' }}>
          <pre
            style={{
              borderRadius: '8px',
              backgroundColor: '#f0f0f0',
              padding: '10px',
            }}
          >
            {rawConfiguration}
          </pre>
        </code>
      </Card>
    </div>
  )
}

export default AdvancedJSONValidation
