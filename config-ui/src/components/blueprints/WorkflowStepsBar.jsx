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
import { CSSTransition } from 'react-transition-group'
import {
  Button,
  Icon,
  Intent,
  Elevation,
  Card,
  Colors,
} from '@blueprintjs/core'
import { WorkflowSteps } from '@/data/BlueprintWorkflow'

const WorkflowStepsBar = (props) => {
  const { activeStep } = props

  return (
    <div className='workflow-bar'>
      <ul className='workflow-steps'>
        <li className={`workflow-step ${activeStep?.id === 1 ? 'active' : ''}`}>
          <a href='#' className='step-id'>
            1
          </a>
          Add Data Connections
        </li>
        <li className={`workflow-step ${activeStep?.id === 2 ? 'active' : ''}`}>
          <a href='#' className='step-id'>
            2
          </a>
          Set Data Scope
        </li>
        <li className={`workflow-step ${activeStep?.id === 3 ? 'active' : ''}`}>
          <a href='#' className='step-id'>
            3
          </a>
          Add Transformation (Optional)
        </li>
        <li className={`workflow-step ${activeStep?.id === 4 ? 'active' : ''}`}>
          <a href='#' className='step-id'>
            4
          </a>
          Set Sync Frequency
        </li>
      </ul>
    </div>
  )
}

export default WorkflowStepsBar
