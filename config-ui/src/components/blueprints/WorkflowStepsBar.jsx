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
  const { activeStep, steps = [] } = props

  const getWorkflowStep = (stepId) => {
    return WorkflowSteps.find((s) => s.id === stepId)
  }

  return (
    <div className='workflow-bar'>
      {steps.length > 0 && (
        <ul className='workflow-steps'>
          {steps.map((step, sIdx) => (
            <li
              key={`workflow-step-key-${sIdx}`}
              className={`workflow-step ${
                activeStep?.id === step?.id ? 'active' : ''
              } ${step.complete ? 'is-completed' : ''}`}
            >
              <a href='#' className='step-id'>
                {step?.complete ? <Icon icon='tick' size={14} /> : step?.id}
              </a>
              {step.title}
            </li>
          ))}
          {/* <li className={`workflow-step ${activeStep?.id === 2 ? 'active' : ''} ${activeStep?.completed ? 'is-completed' : ''}`}>
          <a href='#' className='step-id'>
          {activeStep?.id === 2 && activeStep?.completed ? <Icon icon='tick' size={14} /> : 2}
          </a>
          {getWorkflowStep(2)?.title || 'Set Data Scope'}
        </li>
        <li className={`workflow-step ${activeStep?.id === 3 ? 'active' : ''} ${activeStep?.completed ? 'is-completed' : ''}`}>
          <a href='#' className='step-id'>
            {activeStep?.id === 3 && activeStep?.completed ? <Icon icon='tick' size={14} /> : 3}
          </a>
          {getWorkflowStep(3)?.title || 'Add Transformation (Optional)'}
        </li>
        <li className={`workflow-step ${activeStep?.id === 4 ? 'active' : ''} ${activeStep?.completed ? 'is-completed' : ''}`}>
          <a href='#' className='step-id'>
            {activeStep?.id === 4 && activeStep?.completed ? <Icon icon='tick' size={14} /> : 4}
          </a>
          {getWorkflowStep(4)?.title || 'Set Sync Frequency'}
        </li> */}
        </ul>
      )}
    </div>
  )
}

export default WorkflowStepsBar
